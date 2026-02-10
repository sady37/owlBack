# VitalFocusService 方法设计（基于实际需求重新设计）

## 问题分析

### 1. 为什么从 Redis full cache 读取，而不是直接从 cards 表？

**原因**：Redis full cache (`vital-focus:card:{card_id}:full`) 包含完整数据：
- ✅ 基础信息（来自 cards 表）
- ✅ 实时数据（heart, breath, sleep_stage, bed_status, postures）- 来自 Redis realtime cache
- ✅ 报警数据（alarms 数组）- 来自 Redis alarms cache

**cards 表**只包含基础信息和报警统计，**不包含实时数据**，所以必须从 Redis 读取。

### 2. 为什么不复用 card_service.go 的权限过滤逻辑？

**原因**：card_service.go 已经有完整的权限过滤逻辑：
- `getResourcePermission` - 查询权限配置
- `PermissionFilter` - 权限过滤参数（Resident, BranchOnly, AssignedOnly）
- `ListCards` - Repository 层已经实现了权限过滤的 SQL 查询

**vital-focus 应该复用这些逻辑**，而不是重复实现。

## 设计方案

### 方案：扩展 card_service.go，添加 ListVitalFocusCards 方法

**设计思路**：
1. **先从 cards 表获取 card_id 列表（应用权限过滤）** - 复用现有的权限过滤逻辑
2. **然后从 Redis full cache 读取这些 card_id 的完整数据** - 获取包含实时数据的 VitalFocusCard

**优点**：
- ✅ 权限过滤逻辑复用（不用重复实现）
- ✅ 先从 PostgreSQL 获取 card_id 列表（应用权限过滤），然后再从 Redis 读取（避免扫描所有 Redis 键）
- ✅ 代码更清晰，职责分离
- ✅ 性能更好（不用扫描所有 Redis 键，只读取需要的卡片）

---

## 接口设计

### 1. 接口定义

```go
// CardService 卡片服务接口（扩展）
type CardService interface {
	// GetCardOverview 获取卡片概览列表（返回所有可见的卡片）
	GetCardOverview(ctx context.Context, req GetCardOverviewRequest) (*GetCardOverviewResponse, error)
	
	// ListVitalFocusCards 获取 Vital Focus 卡片列表（包含实时数据和报警数据）
	// 复用权限过滤逻辑，先从 cards 表获取 card_id 列表，然后从 Redis full cache 读取完整数据
	ListVitalFocusCards(ctx context.Context, req ListVitalFocusCardsRequest) (*ListVitalFocusCardsResponse, error)
}
```

---

## 输入/输出结构

### 1. 输入：ListVitalFocusCardsRequest

```go
// ListVitalFocusCardsRequest 获取 Vital Focus 卡片列表请求
type ListVitalFocusCardsRequest struct {
	// 基础过滤参数
	TenantID string // 租户 ID（必填）
	Page     int    // 页码（默认 1）
	PageSize int    // 每页大小（默认 10）
	
	// 权限相关参数（复用 GetCardOverviewRequest 的权限逻辑）
	CurrentUserID   string // 当前用户 ID（必填）
	CurrentUserType string // 当前用户类型："resident" | "staff"（必填）
	CurrentUserRole string // 当前用户角色："Nurse" | "Caregiver" | "Manager" | "SystemAdmin"（staff 必填）
}
```

**字段说明**：
- `TenantID`: 租户 ID，用于过滤卡片
- `Page`: 页码，用于分页（默认 1）
- `PageSize`: 每页大小，用于分页（默认 10）
- `CurrentUserID`: 当前用户 ID（resident_id 或 user_id）
- `CurrentUserType`: 当前用户类型
  - `"resident"`: 住户用户
  - `"staff"`: 员工用户
- `CurrentUserRole`: 当前用户角色（仅 staff 需要）
  - `"Nurse"`: 护士
  - `"Caregiver"`: 护理员
  - `"Manager"`: 管理员
  - `"SystemAdmin"`: 系统管理员

### 2. 输出：ListVitalFocusCardsResponse

```go
// ListVitalFocusCardsResponse 获取 Vital Focus 卡片列表响应
type ListVitalFocusCardsResponse struct {
	Items      []models.VitalFocusCard  // 卡片列表（包含实时数据和报警数据）
	Pagination models.BackendPagination // 分页信息
}
```

**字段说明**：
- `Items`: VitalFocusCard 列表，每个卡片包含：
  - 基础信息（card_id, card_name, card_address, residents, devices）
  - 实时数据（heart, breath, sleep_stage, bed_status, postures）
  - 报警数据（alarms 数组，unhandled_alarm_* 统计）
- `Pagination`: 分页信息（size, page, count, sort, direction）

### 3. 数据结构：models.VitalFocusCard

```go
// VitalFocusCard 完整的卡片对象（聚合后的数据）
type VitalFocusCard struct {
	// 基础信息（来自 cards 表）
	CardID          string   `json:"card_id"`
	TenantID        string   `json:"tenant_id"`
	CardType        string   `json:"card_type"` // "ActiveBed" | "Location"
	BedID           *string  `json:"bed_id,omitempty"`
	LocationID      *string  `json:"location_id,omitempty"`
	CardName        string   `json:"card_name"`
	CardAddress     string   `json:"card_address"`

	
	// 住户和设备（来自 cards.residents 和 cards.devices JSONB）
	Residents       []CardResident `json:"residents"`
	Devices         []CardDevice   `json:"devices"`
	DeviceCount     int            `json:"device_count"`
	ResidentCount   int            `json:"resident_count"`
	
	// 报警统计（来自 cards 表）
	UnhandledAlarm0 *int `json:"unhandled_alarm_0,omitempty"`
	UnhandledAlarm1 *int `json:"unhandled_alarm_1,omitempty"`
	UnhandledAlarm2 *int `json:"unhandled_alarm_2,omitempty"`
	UnhandledAlarm3 *int `json:"unhandled_alarm_3,omitempty"`
	UnhandledAlarm4 *int `json:"unhandled_alarm_4,omitempty"`
	IconAlarmLevel  *int `json:"icon_alarm_level,omitempty"`
	PopAlarmEmerge  *int `json:"pop_alarm_emerge,omitempty"`
	
	// 设备连接状态
	RConnection     *int `json:"r_connection,omitempty"`
	SConnection     *int `json:"s_connection,omitempty"`
	
	// 实时数据（来自 Redis: vital-focus:card:{card_id}:realtime）
	Heart           *int    `json:"heart,omitempty"`
	Breath          *int    `json:"breath,omitempty"`
	HeartSource     *string `json:"heart_source,omitempty"`
	BreathSource    *string `json:"breath_source,omitempty"`
	SleepStage      *int    `json:"sleep_stage,omitempty"`
	BedStatus       *int    `json:"bed_status,omitempty"`
	PersonCount     *int    `json:"person_count,omitempty"`
	Postures        []int   `json:"postures,omitempty"`
	
	// 时间信息
	BedStatusTimestamp *string `json:"bed_status_timestamp,omitempty"`
	StatusDuration     *string `json:"status_duration,omitempty"`
	
	// 报警列表（来自 Redis: vital-focus:card:{card_id}:alarms）
	Alarms          []AlarmItem `json:"alarms,omitempty"`
}
```

---

## 业务逻辑

### 流程概览

```
1. 安全验证：验证 TenantID（不信任前端传入的 TenantID）
   ↓
2. 构建权限过滤请求（复用 GetCardOverview 的逻辑）
   ↓
3. 从 cards 表查询 card_id 列表（应用权限过滤）
   ↓
4. 从 Redis full cache 读取完整数据（包含实时数据）
   ↓
5. 数据规范化（decodeAndNormalizeFullCard）
   ↓
6. 排序和分页
   ↓
7. 返回结果
```

### 详细步骤

#### 步骤 1：安全验证 - 验证用户信息（TenantID, CurrentUserID, CurrentUserRole）

**目标**：不信任前端传入的所有用户信息，从后端数据库验证所有信息是否一致

**逻辑**：
- 根据 `CurrentUserType` 判断用户类型，从数据库查询用户信息：
  - `"resident"`: 从 `residents` 表查询（使用 `CurrentUserID` 作为 `resident_id`）
    - 查询字段：`tenant_id`, `role`, `status`
  - `"staff"`: 从 `users` 表查询（使用 `CurrentUserID` 作为 `user_id`）
    - 查询字段：`tenant_id`, `role`, `status`
- **验证 1：用户是否存在**
  - 如果用户不存在（查询失败），返回错误：`user not found: user_id={CurrentUserID}`
- **验证 2：TenantID 是否一致**
  - 如果前端传入的 `TenantID` 与后端查询的 `tenant_id` 不一致：
    - **安全漏洞**：记录严重错误日志（包含用户 ID、前端传入的 tenant_id、后端查询的 tenant_id）
    - 返回错误：`invalid tenant_id: tenant_id mismatch (security violation)`
- **验证 3：CurrentUserRole 是否一致**（仅 staff 用户）
  - 如果前端传入的 `CurrentUserRole` 与后端查询的 `role` 不一致：
    - **安全漏洞**：记录严重错误日志（包含用户 ID、前端传入的 role、后端查询的 role）
    - 返回错误：`invalid user_role: role mismatch (security violation)`
- **验证 4：用户状态是否有效**
  - 如果用户状态不是 `'active'`，返回错误：`user is not active: status={status}`
- 如果所有验证通过，继续后续流程（使用后端查询的 tenant_id、user_id、role）

**输入**：
- `req.TenantID`（前端传入，不信任）
- `req.CurrentUserID`（前端传入，不信任）
- `req.CurrentUserType`（前端传入，不信任）
- `req.CurrentUserRole`（前端传入，不信任）

**输出**：
- `validatedTenantID`（后端查询的 tenant_id，信任）
- `validatedUserID`（后端查询的 user_id/resident_id，信任）
- `validatedUserRole`（后端查询的 role，信任）
- 如果验证失败，返回错误

**安全考虑**：
- **不信任前端传入的所有用户信息**：前端可能被篡改，必须从后端验证
- **全面验证**：验证 tenant_id、role、用户存在性、用户状态
- **记录安全日志**：如果发现不一致，记录严重错误日志，便于安全审计
- **拒绝请求**：如果验证失败，直接返回错误，不继续执行，避免数据泄露
- **可选：kill session**：如果发现安全漏洞，可以 kill session（需要会话管理支持）

**代码示例**：
```go
// 1. 从后端查询用户信息（根据 CurrentUserType 查询不同的表）
var actualTenantID, actualUserRole, userStatus string
if req.CurrentUserType == "resident" {
	// 从 residents 表查询
	var resident domain.Resident
	err := s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text, role, status 
		 FROM residents 
		 WHERE resident_id = $1`,
		req.CurrentUserID,
	).Scan(&actualTenantID, &actualUserRole, &userStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("resident not found: resident_id=%s", req.CurrentUserID)
		}
		return nil, fmt.Errorf("failed to get resident: %w", err)
	}
} else if req.CurrentUserType == "staff" {
	// 从 users 表查询（不指定 tenant_id，因为我们要验证）
	err := s.db.QueryRowContext(ctx,
		`SELECT tenant_id::text, role, COALESCE(status, 'active') 
		 FROM users 
		 WHERE user_id = $1`,
		req.CurrentUserID,
	).Scan(&actualTenantID, &actualUserRole, &userStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: user_id=%s", req.CurrentUserID)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
} else {
	return nil, fmt.Errorf("invalid user type: %s", req.CurrentUserType)
}

// 2. 验证用户状态是否有效
if userStatus != "active" {
	s.logger.Warn("User is not active",
		zap.String("user_id", req.CurrentUserID),
		zap.String("user_type", req.CurrentUserType),
		zap.String("status", userStatus),
	)
	return nil, fmt.Errorf("user is not active: status=%s", userStatus)
}

// 3. 验证前端传入的 TenantID 是否与后端查询的一致
if req.TenantID != "" && req.TenantID != actualTenantID {
	// 安全漏洞：记录严重错误日志
	s.logger.Error("Security violation: tenant_id mismatch",
		zap.String("user_id", req.CurrentUserID),
		zap.String("user_type", req.CurrentUserType),
		zap.String("frontend_tenant_id", req.TenantID),
		zap.String("backend_tenant_id", actualTenantID),
	)
	// 返回错误，拒绝请求
	return nil, fmt.Errorf("invalid tenant_id: tenant_id mismatch (security violation)")
}

// 4. 验证前端传入的 CurrentUserRole 是否与后端查询的一致（仅 staff 用户）
if req.CurrentUserType == "staff" && req.CurrentUserRole != "" && req.CurrentUserRole != actualUserRole {
	// 安全漏洞：记录严重错误日志
	s.logger.Error("Security violation: user_role mismatch",
		zap.String("user_id", req.CurrentUserID),
		zap.String("user_type", req.CurrentUserType),
		zap.String("frontend_role", req.CurrentUserRole),
		zap.String("backend_role", actualUserRole),
	)
	// 返回错误，拒绝请求
	return nil, fmt.Errorf("invalid user_role: role mismatch (security violation)")
}

// 5. 使用后端查询的信息（信任）
validatedTenantID := actualTenantID
validatedUserID := req.CurrentUserID  // CurrentUserID 已验证存在
validatedUserRole := actualUserRole   // 使用后端查询的 role（信任）
```

**复用代码**：
- `usersRepo.GetUser`（staff 用户，如果需要完整用户信息）
- `residentsRepo.GetResident`（resident 用户，如果存在该方法）
- 或者直接使用 SQL 查询（更轻量，只需要 tenant_id、role、status）

#### 步骤 2：构建权限过滤请求

**目标**：复用 `GetCardOverview` 的权限过滤逻辑，构建 `repository.ListCardsRequest`

**逻辑**：
- 根据 `CurrentUserType` 判断用户类型：
  - `"resident"`: 设置 `PermissionFilter.UserID = CurrentUserID`, `PermissionFilter.UserType = "resident"`
  - `"staff"`: 
    1. 调用 `getResourcePermission(ctx, CurrentUserRole, "cards", "R")` 查询权限配置
    2. 如果 `perm.BranchOnly = true`: 查询用户的 `branch_name`，设置 `PermissionFilter.UserBranchTag`
    3. 如果 `perm.AssignedOnly = true`: 设置 `PermissionFilter.AssignedOnly = true`, `PermissionFilter.UserIDForAssignment = CurrentUserID`

**输入**：
- `validatedTenantID`（步骤 1 验证后的 tenant_id，信任）
- `validatedUserID`（步骤 1 验证后的 user_id/resident_id，信任）
- `validatedUserRole`（步骤 1 验证后的 role，信任）
- `req.CurrentUserType`（已验证用户存在，可以信任）

**输出**：
- `repository.ListCardsRequest`（包含 `PermissionFilter`）

**复用代码**：
- `cardService.getResourcePermission`（查询权限配置）
- `cardService.GetCardOverview` 的权限过滤逻辑（第 91-193 行）

**注意**：
- 使用步骤 1 验证后的 `validatedTenantID`、`validatedUserID`、`validatedUserRole`（信任）
- 不再使用前端传入的 `req.TenantID`、`req.CurrentUserID`、`req.CurrentUserRole`（不信任）

#### 步骤 3：从 cards 表查询 card_id 列表

**目标**：应用权限过滤，获取可见的卡片 ID 列表

**逻辑**：
- 调用 `cardsRepo.ListCards(ctx, repoReq)` 查询卡片列表
- Repository 层会根据 `PermissionFilter` 应用 SQL 过滤：
  - Resident 权限：`WHERE card.resident_id = $1` 或 `WHERE card.unit_id IN (SELECT unit_id FROM residents WHERE resident_id = $1)`
  - BranchOnly 权限：`WHERE unit.branch_name = $1`
  - AssignedOnly 权限：使用 CTE JOIN `resident_caregivers` 表过滤

**输入**：
- `repository.ListCardsRequest`（包含权限过滤参数）
- **注意**：`repoReq.TenantID` 使用步骤 1 验证后的 `validatedTenantID`（信任）

**输出**：
- `[]*domain.CardWithUnitInfo`（包含 `Card.CardID`）

**复用代码**：
- `repository.PostgresCardsRepository.ListCards`（已实现权限过滤的 SQL 查询）

#### 步骤 4：从 Redis full cache 读取完整数据

**目标**：根据 card_id 列表，从 Redis 读取完整的 VitalFocusCard 数据

**逻辑**：
- 遍历 card_id 列表
- 对每个 card_id，构建 Redis key：`vital-focus:card:{card_id}:full`
- 调用 `kv.Get(ctx, key)` 读取数据
- 如果 Redis 中没有数据（aggregator 还没聚合），跳过该卡片
- 使用 `decodeAndNormalizeFullCard(raw)` 解析和规范化数据

**输入**：
- `[]string`（card_id 列表）

**输出**：
- `[]models.VitalFocusCard`（完整的卡片数据）

**复用代码**：
- `service.decodeAndNormalizeFullCard`（从 `vital_focus_util.go`）

#### 步骤 5：数据规范化

**目标**：解析和规范化 Redis 数据，确保与前端模型一致

**逻辑**：
- 使用 `decodeAndNormalizeFullCard(raw)` 解析 JSON
- 规范化字段类型（device_type 从 string 转 number，heart_source/breath_source 转 's'/'r'/'-'）
- 确保必填字段有默认值（如 residents[].last_name）

**输入**：
- `string`（Redis 原始 JSON 数据）

**输出**：
- `models.VitalFocusCard`（规范化后的数据）

**复用代码**：
- `service.decodeAndNormalizeFullCard`（从 `vital_focus_util.go`）

#### 步骤 6：排序和分页

**目标**：对结果进行排序和分页处理

**逻辑**：
- 排序：按 `card_id` 排序（使用 `sortCardsByID` 函数）
- 分页：根据 `Page` 和 `PageSize` 计算 `start` 和 `end` 索引
- 计算总数：`count = len(items)`

**输入**：
- `[]models.VitalFocusCard`（所有卡片）
- `req.Page`（页码）
- `req.PageSize`（每页大小）

**输出**：
- `[]models.VitalFocusCard`（分页后的卡片列表）
- `models.BackendPagination`（分页信息）

#### 步骤 6：返回结果

**目标**：构建响应结构并返回

**逻辑**：
- 构建 `ListVitalFocusCardsResponse`
- 设置 `Items`（分页后的卡片列表）
- 设置 `Pagination`（分页信息）

**输入**：
- `[]models.VitalFocusCard`（分页后的卡片列表）
- `models.BackendPagination`（分页信息）

**输出**：
- `*ListVitalFocusCardsResponse`

---

## 依赖关系

### 1. 服务依赖

```go
type cardService struct {
	cardsRepo     repository.CardsRepository   // 查询 cards 表（已有）
	residentsRepo repository.ResidentsRepository // 查询 residents 表（已有）
	devicesRepo   repository.DevicesRepository   // 查询 devices 表（已有）
	usersRepo     repository.UsersRepository     // 查询 users 表（已有）
	kv            store.KV                       // 新增：读取 Redis full cache
	db            *sql.DB                        // 查询权限配置（已有）
	logger        *zap.Logger                    // 日志（已有）
}
```

### 2. 新增依赖

- `store.KV`: 用于读取 Redis full cache
  - 接口方法：`Get(ctx, key) -> (string, error)`

### 3. 复用代码

- `cardService.getResourcePermission`: 查询权限配置
- `cardService.GetCardOverview` 的权限过滤逻辑（第 91-193 行）
- `repository.PostgresCardsRepository.ListCards`: 查询卡片列表（已实现权限过滤）
- `service.decodeAndNormalizeFullCard`: 解析和规范化 Redis 数据

---

## 错误处理

### 1. 用户不存在（安全漏洞）

**情况**：根据 `CurrentUserID` 和 `CurrentUserType` 查询用户/resident 失败

**处理**：
- 记录错误日志（包含用户 ID、用户类型）
- 返回错误，拒绝请求，不继续执行

```go
if err == sql.ErrNoRows {
	return nil, fmt.Errorf("user not found: user_id=%s, user_type=%s", req.CurrentUserID, req.CurrentUserType)
}
```

### 2. TenantID 验证失败（安全漏洞）

**情况**：前端传入的 `TenantID` 与后端查询的 `tenant_id` 不一致

**处理**：
- 记录严重错误日志（包含用户 ID、前端传入的 tenant_id、后端查询的 tenant_id）
- 返回错误，拒绝请求，不继续执行
- **可选**：kill session（需要会话管理支持）

```go
if req.TenantID != "" && req.TenantID != actualTenantID {
	s.logger.Error("Security violation: tenant_id mismatch",
		zap.String("user_id", req.CurrentUserID),
		zap.String("user_type", req.CurrentUserType),
		zap.String("frontend_tenant_id", req.TenantID),
		zap.String("backend_tenant_id", actualTenantID),
	)
	return nil, fmt.Errorf("invalid tenant_id: tenant_id mismatch (security violation)")
}
```

### 3. CurrentUserRole 验证失败（安全漏洞）

**情况**：前端传入的 `CurrentUserRole` 与后端查询的 `role` 不一致（仅 staff 用户）

**处理**：
- 记录严重错误日志（包含用户 ID、前端传入的 role、后端查询的 role）
- 返回错误，拒绝请求，不继续执行
- **可选**：kill session（需要会话管理支持）

```go
if req.CurrentUserType == "staff" && req.CurrentUserRole != "" && req.CurrentUserRole != actualUserRole {
	s.logger.Error("Security violation: user_role mismatch",
		zap.String("user_id", req.CurrentUserID),
		zap.String("user_type", req.CurrentUserType),
		zap.String("frontend_role", req.CurrentUserRole),
		zap.String("backend_role", actualUserRole),
	)
	return nil, fmt.Errorf("invalid user_role: role mismatch (security violation)")
}
```

### 4. 用户状态无效

**情况**：用户状态不是 `'active'`

**处理**：
- 记录警告日志（包含用户 ID、状态）
- 返回错误，拒绝请求，不继续执行

```go
if userStatus != "active" {
	s.logger.Warn("User is not active",
		zap.String("user_id", req.CurrentUserID),
		zap.String("status", userStatus),
	)
	return nil, fmt.Errorf("user is not active: status=%s", userStatus)
}
```

### 5. 权限配置查询失败

**情况**：`getResourcePermission` 返回错误

**处理**：返回错误，不继续执行

```go
perm, err := s.getResourcePermission(ctx, req.CurrentUserRole, "cards", "R")
if err != nil {
	return nil, fmt.Errorf("failed to get resource permission: %w", err)
}
```

### 6. cards 表查询失败

**情况**：`cardsRepo.ListCards` 返回错误

**处理**：返回错误，不继续执行

```go
cards, err := s.cardsRepo.ListCards(ctx, repoReq)
if err != nil {
	return nil, fmt.Errorf("failed to list cards: %w", err)
}
```

### 7. Redis 读取失败（单个卡片）

**情况**：某个 card_id 的 Redis key 不存在或读取失败

**处理**：跳过该卡片，记录日志，继续处理其他卡片

```go
raw, err := s.kv.Get(ctx, key)
if err != nil {
	s.logger.Debug("Failed to get card from Redis", 
		zap.String("card_id", cardID), 
		zap.Error(err))
	continue // 跳过该卡片
}
```

### 8. 数据解析失败（单个卡片）

**情况**：`decodeAndNormalizeFullCard` 返回 false

**处理**：跳过该卡片，记录日志，继续处理其他卡片

```go
card, ok := decodeAndNormalizeFullCard(raw)
if !ok {
	s.logger.Debug("Failed to decode and normalize card", 
		zap.String("card_id", cardID))
	continue // 跳过该卡片
}
```

### 9. Redis 完全不可用

**情况**：所有 Redis 读取都失败

**处理**：返回空列表（联调友好，避免前端报错）

**注意**：这种情况在步骤 3 中会自然地返回空列表（所有卡片都被跳过）

---

## 性能考虑

### 1. 避免扫描所有 Redis 键

**优化**：先从 cards 表获取 card_id 列表（应用权限过滤），然后再从 Redis 读取

**好处**：
- 只读取需要的卡片（权限过滤后的卡片）
- 避免扫描所有 Redis 键（`ScanKeys` 操作较慢）
- 性能更好，特别是卡片数量很多时

### 2. 批量读取（可选优化）

**当前实现**：逐个读取 Redis key（串行）

**可选优化**：使用 `pipeline` 或 `MGet` 批量读取（需要 store.KV 接口支持）

**注意**：当前 store.KV 接口可能不支持批量操作，先使用串行实现

### 3. 缓存策略

**当前实现**：直接从 Redis full cache 读取（由 wisefido-card-aggregator 维护）

**TTL**：Redis full cache 的 TTL 为 10 秒

**说明**：不需要额外的缓存层，直接使用 wisefido-card-aggregator 的缓存

---

## 总结

### 数据流

```
用户请求（包含前端传入的 TenantID，不信任）
  ↓
步骤 1：安全验证 - 验证 TenantID
  - 从后端查询用户的实际 tenant_id
  - 验证前端传入的 TenantID 是否与后端查询的一致
  - 如果不一致：记录安全日志，返回错误（拒绝请求）
  - 如果一致：使用后端查询的 tenant_id（信任）
  ↓
步骤 2：构建权限过滤请求（复用 GetCardOverview 的逻辑）
  - 使用验证后的 tenant_id（信任）
  ↓
步骤 3：从 cards 表查询 card_id 列表（应用权限过滤）
  ↓
步骤 4：从 Redis full cache 读取完整数据（包含实时数据）
  ↓
步骤 5：数据规范化（decodeAndNormalizeFullCard）
  ↓
步骤 6：排序和分页
  ↓
步骤 7：返回结果
  ↓
VitalFocusCard 列表（包含实时数据和报警数据）
```

### 关键优势

- ✅ **权限过滤逻辑复用**：不用重复实现，复用 `GetCardOverview` 的逻辑
- ✅ **性能更好**：不用扫描所有 Redis 键，只读取需要的卡片
- ✅ **代码更清晰**：职责分离，与现有的 card_service 架构一致
- ✅ **错误处理完善**：单个卡片失败不影响整体结果

### 实现要点

1. **扩展 CardService 接口**：添加 `ListVitalFocusCards` 方法
2. **添加依赖**：在 `cardService` 中添加 `kv store.KV`
3. **复用权限过滤逻辑**：复用 `GetCardOverview` 的权限过滤代码
4. **复用工具函数**：使用 `decodeAndNormalizeFullCard` 解析数据
5. **错误处理**：单个卡片失败不影响整体结果
