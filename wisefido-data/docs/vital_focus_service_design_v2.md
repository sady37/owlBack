# VitalFocusService 方法设计（基于实际使用）

## 当前 Overview.vue 的实际使用

### 1. API 调用
- **Overview.vue** 使用 `getVitalFocusCardsApi()` 
- 每 2 秒调用一次 `GET /data/api/v1/data/vital-focus/cards`
- 返回的数据**已经包含实时数据和报警数据**（heart, breath, sleep_stage, bed_status, alarms 等）

### 2. 数据来源
- 从 Redis `vital-focus:card:{card_id}:full` 读取
- 数据已经由 `wisefido-card-aggregator` 聚合完成，包含：
  - 基础信息（card_name, card_address, residents, devices）
  - 实时数据（heart, breath, sleep_stage, bed_status, postures）
  - 报警数据（alarms 数组，unhandled_alarm_* 统计）

### 3. Detail.vue 的使用
- Detail.vue 使用 `cardStore.loadCardDetail()` -> `getCardOverviewApi()`
- 这是不同的 API（`/admin/api/v1/card-overview`），不是 vital-focus API
- 用于详情页，返回更详细的卡片信息

## Service 层设计

基于用户说明和实际使用，VitalFocusService 应该提供两个核心方法：

### 1. ListCards - 获取卡片列表（根据角色权限过滤）

**方法签名**：
```go
func (s *VitalFocusService) ListCards(
    ctx context.Context,
    req ListCardsRequest,
) (*ListCardsResponse, error)
```

**请求参数**：
```go
type ListCardsRequest struct {
    TenantID string
    
    // 用户信息（从 HTTP Header 获取）
    UserID   string // X-User-Id
    UserRole string // X-User-Role（可选，如果提供可减少 DB 查询）
    UserType string // "staff" | "resident"（从登录类型推断）
    
    // 分页参数（可选，当前 Handler 支持但 Overview.vue 不使用）
    Page     int // 默认 1
    PageSize int // 默认 10（但 Overview.vue 不使用分页）
}
```

**响应**：
```go
type ListCardsResponse struct {
    Items      []models.VitalFocusCard // 已包含实时数据和报警数据
    Pagination models.BackendPagination // 分页信息（可选）
}
```

**实现逻辑**：
1. 从 Redis 扫描所有 `vital-focus:card:*:full` 键
2. 批量读取并解析所有卡片数据（使用 `MGET` 或循环 `GET`）
3. 按 `tenant_id` 过滤
4. **权限过滤**（核心功能）：
   - 如果 `UserID` 为空：返回所有卡片（向后兼容，但实际场景应该都有 UserID）
   - 如果 `UserType == "resident"`：调用 `filterCardsForResident()`
   - 如果 `UserType == "staff"` 或为空：调用 `filterCardsForStaff()`
5. 排序（按 `card_id`，后续可扩展）
6. 分页（如果请求了分页）
7. 返回结果

**关键点**：
- 返回的数据**已经包含实时数据和报警数据**，无需额外查询
- 主要职责是**权限过滤**

### 2. GetCard - 获取单个卡片的实时数据和报警数据

**方法签名**：
```go
func (s *VitalFocusService) GetCard(
    ctx context.Context,
    req GetCardRequest,
) (*models.VitalFocusCard, error)
```

**请求参数**：
```go
type GetCardRequest struct {
    TenantID string
    CardID   string
    
    // 权限参数
    UserID   string
    UserRole string // 可选
    UserType string // "staff" | "resident"
}
```

**实现逻辑**：
1. 从 Redis 读取 `vital-focus:card:{card_id}:full`
2. 解析卡片数据
3. 验证 `tenant_id` 匹配
4. **权限验证**：
   - 调用 `hasCardPermission()` 检查用户是否有权限查看该卡片
   - 如果没有权限：返回错误（`ErrCardNotFound` 或 `ErrPermissionDenied`）
5. 返回卡片（已包含实时数据和报警数据）

**关键点**：
- 数据从 Redis 读取，已经包含实时数据和报警数据
- 主要职责是**权限验证**

**注意**：
- 当前 Detail.vue 使用的是 `getCardOverviewApi()`（不同的 API），不是这个方法
- 这个方法可能用于：
  - 根据 `resident_id` 查询卡片（`GetCardByIDOrResident`）
  - 或者作为备用接口

## 权限过滤辅助方法（内部方法）

### 1. filterCardsForStaff

**功能**：过滤 Staff 用户可见的卡片

```go
func (s *VitalFocusService) filterCardsForStaff(
    ctx context.Context,
    userID, tenantID string,
    cards []models.VitalFocusCard,
) ([]models.VitalFocusCard, error)
```

**权限规则**：
1. **Admin 角色**：返回所有 tenant 下的卡片
2. **ALL scope**：返回所有 tenant 下的卡片
3. **LOCATION scope**：
   - 查询用户的 `tags`（JSONB 数组）
   - 查询卡片的 `location_id` 对应的 `units.location_tag`
   - 过滤 `location_tag` 在 `users.tags` 中的卡片
4. **ASSIGNED_ONLY scope**：
   - 查询 `resident_caregivers` 表，获取用户负责的 `resident_id` 列表
   - 过滤 `primary_resident_id` 或 `residents[].resident_id` 在列表中的卡片

### 2. filterCardsForResident

**功能**：过滤 Resident 用户可见的卡片

```go
func (s *VitalFocusService) filterCardsForResident(
    ctx context.Context,
    residentID, tenantID string,
    cards []models.VitalFocusCard,
) ([]models.VitalFocusCard, error)
```

**权限规则**：
1. **ActiveBed 卡片**：`card_type == "ActiveBed"` 且 `bed_id == resident.bed_id` 且 `primary_resident_id == resident_id`
2. **Location 卡片**：`card_type == "Location"` 且 `location_id == resident.location_id` 且住户在 `card.residents` 中

### 3. hasCardPermission

**功能**：检查用户是否有权限查看指定卡片

```go
func (s *VitalFocusService) hasCardPermission(
    ctx context.Context,
    userID, userRole, userType, tenantID string,
    card models.VitalFocusCard,
) (bool, error)
```

**实现逻辑**：
1. 根据 `userType` 调用相应的过滤方法
2. 检查卡片是否在过滤后的列表中
3. 返回 `true` 如果有权限，`false` 如果没有权限

## Service 依赖

```go
type VitalFocusService struct {
    kv            store.KV                       // Redis 操作
    usersRepo     repository.UsersRepository     // 用户信息查询（role, alarm_scope, tags）
    residentsRepo repository.ResidentsRepository // 住户信息查询（bed_id, location_id）
    unitsRepo     repository.UnitsRepository     // 单元信息查询（location_tag）
    db            *sql.DB                        // 复杂 SQL 查询（resident_caregivers）
    logger        *zap.Logger
}
```

## 与当前 Handler 的对比

| 当前 Handler 方法 | Service 方法 | 主要差异 |
|------------------|-------------|---------|
| `GetCards` | `ListCards` | **新增权限过滤** |
| `GetCardByIDOrResident` | `GetCard` | **新增权限验证** |
| `SaveSelection` | （保留在 Handler） | 主要是数据持久化，暂不迁移 |

## 实现优先级

1. **高优先级**：
   - `ListCards`（带权限过滤）- Overview.vue 主要使用
   - `filterCardsForStaff`（Admin、ALL、LOCATION、ASSIGNED_ONLY）
   - `filterCardsForResident`

2. **中优先级**：
   - `GetCard`（带权限验证）
   - `hasCardPermission`

3. **低优先级**：
   - `filterCardsForFamily`（如果系统支持 Family 登录）

## 关键理解

1. **数据已经聚合**：Redis `vital-focus:card:{card_id}:full` 已经包含实时数据和报警数据，Service 层不需要额外查询
2. **主要职责是权限过滤**：Service 层的核心功能是根据用户权限过滤卡片列表
3. **ListCards 是主要方法**：Overview.vue 每 2 秒调用一次，是高频接口
4. **GetCard 是辅助方法**：用于单个卡片查询，可能用于根据 resident_id 查询等场景
