# 架构分层设计

## 一、现状分析

### 1.1 API 端点统计（按业务领域）

| 领域 | 端点 | 复杂度 | Handler 行数 | 状态 |
|------|------|--------|---------------|------|
| **住户管理** | `/admin/api/v1/residents` | 极高 | 3032 | ❌ 需重构 |
| **用户管理** | `/admin/api/v1/users` | 高 | 1257 | ❌ 需重构 |
| **认证授权** | `/auth/api/v1/*` | 高 | 886 | ❌ 需重构 |
| **Tag 管理** | `/admin/api/v1/tags` | 中 | 576 | ❌ 需重构 |
| **角色权限** | `/admin/api/v1/roles`, `/admin/api/v1/role-permissions` | 中 | 634 | ❌ 需重构 |
| **地址管理** | `/admin/api/v1/buildings/units/rooms/beds` | 中 | 336 | ✅ 已使用 Repository |
| **设备管理** | `/admin/api/v1/devices`, `/admin/api/v1/device-store` | 中 | 924 | ✅ 已使用 Repository |
| **租户管理** | `/admin/api/v1/tenants` | 低 | 266 | ✅ 已使用 Repository |
| **告警管理** | `/admin/api/v1/alarm-*` | 低 | 245 | ⚠️ 待处理 |
| **数据查询** | `/data/api/v1/data/vital-focus/*` | 中 | 305 | ⚠️ 待处理 |

### 1.2 业务领域边界

```
┌─────────────────────────────────────────────────────────┐
│ 平台层（Platform）                                       │
│  - 租户管理（Tenants）                                   │
│  - 系统配置                                             │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ 用户权限层（Auth & Access）                              │
│  - 用户管理（Users）                                     │
│  - 角色管理（Roles）                                     │
│  - 权限管理（RolePermissions）                          │
│  - 认证登录（Auth）                                      │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ 业务层（Business）                                        │
│  ├─ 地址层级（Location）                                 │
│  │   - 楼栋（Buildings）                                │
│  │   - 单元（Units）                                    │
│  │   - 房间（Rooms）                                    │
│  │   - 床位（Beds）                                     │
│  ├─ 住户管理（Resident）                                 │
│  │   - 住户信息                                         │
│  │   - 住户 PHI                                         │
│  │   - 联系人（Contacts）                               │
│  │   - 护工分配（Caregivers）                          │
│  ├─ 设备管理（Device）                                   │
│  │   - 设备（Devices）                                  │
│  │   - 设备库存（Device Store）                        │
│  ├─ 标签管理（Tag）                                      │
│  │   - Tag 目录（TagsCatalog）                         │
│  │   - Tag 查询                                         │
│  └─ 告警管理（Alarm）                                    │
│      - 告警配置                                         │
│      - 告警事件                                         │
└─────────────────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────────────────┐
│ 数据查询层（Data）                                       │
│  - Vital Focus                                          │
│  - 卡片数据                                             │
└─────────────────────────────────────────────────────────┘
```

### 1.3 Repository 现状

**已实现（使用 map[string]any）**：
- ✅ `PostgresUnitsRepo` (1073 行) - 已集成到 Handler
  - 接口：`UnitsRepo` (repository.go)
  - 问题：使用 `map[string]any`，有 `ToJSON()` 方法
- ✅ `PostgresDevicesRepo` (533 行) - 已集成到 Handler
  - 接口：`DevicesRepo` (repository.go)
  - 问题：使用 `map[string]any`
- ✅ `PostgresResidentsRepo` (487 行) - **已实现触发器替代，但未集成到 Handler**
  - 接口：无（直接实现）
  - 问题：使用 `map[string]any`
- ✅ `PostgresDeviceStoreRepo` (391 行) - 已集成到 Handler
  - 接口：`DeviceStoreRepo` (repository.go)
  - 问题：使用 `map[string]any`
- ✅ `PostgresTenantsRepo` (137 行) - 已集成到 Handler
  - 接口：`TenantsRepo` (tenants_types.go)
  - 问题：使用 `map[string]any`

**新定义（使用强类型领域模型）**：
- ✅ `ResidentsRepository` (residents_repo.go) - **接口已定义，实现待完成**
  - 使用：`domain.Resident`, `domain.ResidentPHI`, `domain.ResidentContact`, `domain.ResidentCaregiver`
  - 优势：强类型，不使用 `map[string]any`
- ✅ `UsersRepository` (users_repo.go) - **接口已定义，实现待完成**
  - 使用：`domain.User`
  - 优势：强类型，不使用 `map[string]any`
- ✅ `TagsRepository` (tags_repo.go) - **接口已定义，实现待完成**
  - 使用：`domain.Tag`
  - 优势：强类型，不使用 `map[string]any`

**Repository接口统计**：
- 新定义（强类型）：3个
- 原有（map[string]any）：4个
- **总计：7个**

**问题**：
- 原有实现使用 `map[string]any` 而不是强类型
- 原有实现有 `ToJSON()` 方法（耦合前端）
- `PostgresResidentsRepo` 未集成到 Handler
- 新定义的接口需要实现对应的Postgres实现

### 1.4 Handler 现状

**两种模式**：

1. **简单模式**（Units/Devices/Tenants）：
   ```
   Handler → Repository → Database
   ```
   - Handler 只做 HTTP 处理
   - 业务逻辑在 Repository
   - 代码简洁（~100-200 行/Handler）

2. **复杂模式**（Residents/Users/Tags）：
   ```
   Handler → 直接 SQL → Database
   ```
   - Handler 包含所有逻辑（权限、业务规则、数据转换、SQL）
   - 代码复杂（3000+ 行/Handler）
   - 无法复用、无法测试

---

## 二、分层原则

### 2.1 三层架构

```
┌─────────────┐
│   Handler   │  HTTP 请求/响应处理
│   Layer     │  - 解析请求参数
└──────┬──────┘  - 返回 JSON 响应
       │         - 错误处理
       ↓
┌─────────────┐
│   Service   │  业务逻辑
│   Layer     │  - 权限检查
└──────┬──────┘  - 业务规则验证
       │         - 数据转换
       │         - 业务编排
       ↓
┌─────────────┐
│ Repository  │  数据访问
│   Layer     │  - SQL 操作
└──────┬──────┘  - 数据一致性（替代触发器）
       │         - 事务管理
       ↓
┌─────────────┐
│  Database   │
└─────────────┘
```

### 2.2 职责边界

#### Handler 层
**职责**：
- ✅ 解析 HTTP 请求（URL 参数、请求体、请求头）
- ✅ 生成 HTTP 响应（JSON 格式）
- ✅ 路由分发（根据 HTTP 方法和路径）
- ✅ 错误处理（捕获异常并返回 HTTP 状态码）

**不负责**：
- ❌ 业务规则验证
- ❌ 权限检查
- ❌ 数据转换
- ❌ 数据库操作

#### Service 层
**职责**：
- ✅ 权限检查（调用 PermissionChecker）
- ✅ 业务规则验证（如 nickname 不能为空）
- ✅ 数据转换（JSON ↔ 领域模型）
- ✅ 业务编排（协调多个 Repository）
- ✅ 事务管理（跨 Repository 的事务）

**不负责**：
- ❌ HTTP 请求/响应处理
- ❌ 数据库 SQL 操作
- ❌ 数据一致性维护（属于 Repository）

#### Repository 层
**职责**：
- ✅ 数据访问抽象（封装 SQL）
- ✅ 数据一致性（替代数据库触发器）
- ✅ 事务管理（单 Repository 内的事务）
- ✅ 领域模型映射（数据库记录 ↔ 领域模型）

**不负责**：
- ❌ 业务规则验证
- ❌ 权限检查
- ❌ HTTP 处理

### 2.3 依赖方向

```
Handler → Service → Repository → Database
```

**规则**：
- Handler 只能调用 Service
- Service 只能调用 Repository
- Repository 只能调用 Database
- **不允许反向依赖**

---

## 三、分类原则

### 3.1 业务领域分类

按业务领域组织 Service 和 Repository：

```
internal/
├── service/
│   ├── resident_service.go      # 住户管理
│   ├── user_service.go          # 用户管理
│   ├── auth_service.go          # 认证授权
│   ├── tag_service.go           # Tag 管理
│   ├── role_service.go          # 角色权限
│   ├── location_service.go      # 地址管理（可选，当前直接使用 Repository）
│   ├── device_service.go        # 设备管理（可选，当前直接使用 Repository）
│   └── alarm_service.go         # 告警管理
│
└── repository/
    ├── postgres_residents.go    # 住户 Repository
    ├── postgres_users.go        # 用户 Repository
    ├── postgres_tags_catalog.go # Tag Repository
    ├── postgres_roles.go        # 角色 Repository
    ├── postgres_role_permissions.go # 权限 Repository
    ├── postgres_units.go        # 地址 Repository（已存在）
    ├── postgres_devices.go      # 设备 Repository（已存在）
    └── postgres_tenants.go      # 租户 Repository（已存在）
```

### 3.2 Service 分类原则

**按业务领域分类**，每个领域一个 Service：
- `ResidentService` - 住户相关业务逻辑
- `UserService` - 用户相关业务逻辑
- `AuthService` - 认证授权业务逻辑
- `TagService` - Tag 相关业务逻辑
- `RoleService` - 角色权限业务逻辑

**简单领域可以不设 Service**（直接使用 Repository）：
- `LocationService` - 地址管理（当前直接使用 `UnitsRepo`）
- `DeviceService` - 设备管理（当前直接使用 `DevicesRepo`）

### 3.3 Repository 分类原则

**按数据实体分类**，每个实体一个 Repository：
- `PostgresResidentsRepo` - 住户数据
- `PostgresUsersRepo` - 用户数据
- `PostgresTagsCatalogRepo` - Tag 数据
- `PostgresRolesRepo` - 角色数据
- `PostgresRolePermissionsRepo` - 权限数据

---

## 四、最小集（MVP）

### 4.1 优先级排序

**Phase 1: 最高优先级**（复杂度极高，影响最大）
1. ✅ `ResidentService` + `PostgresResidentsRepo`（3032 行 → 重构）
   - 已实现 Repository，需集成到 Handler
   - 需提取 Service 层

**Phase 2: 高优先级**（复杂度高）
2. ✅ `UserService` + `PostgresUsersRepo`（1257 行 → 重构）
3. ✅ `AuthService`（886 行 → 重构）

**Phase 3: 中优先级**（复杂度中）
4. ✅ `TagService` + `PostgresTagsCatalogRepo`（576 行 → 重构）
5. ✅ `RoleService` + `PostgresRolesRepo` + `PostgresRolePermissionsRepo`（634 行 → 重构）

**Phase 4: 低优先级**（复杂度低或已实现）
6. ⚠️ `AlarmService`（245 行，待处理）
7. ✅ `LocationService`（已直接使用 Repository，可选）
8. ✅ `DeviceService`（已直接使用 Repository，可选）

### 4.2 MVP 最小集

**必须实现**（Phase 1-2）：
- `ResidentService` + `PostgresResidentsRepo`（已实现，需集成）
- `UserService` + `PostgresUsersRepo`（需实现）
- `AuthService`（需实现）

**可选实现**（Phase 3）：
- `TagService` + `PostgresTagsCatalogRepo`（需实现）
- `RoleService` + `PostgresRolesRepo` + `PostgresRolePermissionsRepo`（需实现）

---

## 五、迁移方案

### 5.1 迁移步骤

**Step 1: 分析现有代码**
- ✅ 已完成（见本文档第一部分）

**Step 2: 实现 Repository（如未实现）**
- ✅ `ResidentsRepository` 接口已定义（residents_repo.go）
- ⚠️ `PostgresResidentsRepo` 已实现但使用 `map[string]any`，需重构为强类型实现
- ✅ `UsersRepository` 接口已定义（users_repo.go）
- ❌ `PostgresUsersRepo` 需实现（实现 `UsersRepository` 接口）
- ✅ `TagsRepository` 接口已定义（tags_repo.go）
- ❌ `PostgresTagsRepo` 需实现（实现 `TagsRepository` 接口）
- ❌ `PostgresRolesRepo` 需实现
- ❌ `PostgresRolePermissionsRepo` 需实现

**Step 3: 实现 Service 层**
- ❌ `ResidentService` 需实现
- ❌ `UserService` 需实现
- ❌ `AuthService` 需实现
- ❌ `TagService` 需实现
- ❌ `RoleService` 需实现

**Step 4: 重构 Handler**
- ❌ `admin_residents_handlers.go` 需重构（3032 行 → ~200 行）
- ❌ `admin_users_handlers.go` 需重构（1257 行 → ~200 行）
- ❌ `auth_handlers.go` 需重构（886 行 → ~200 行）
- ❌ `admin_tags_handlers.go` 需重构（576 行 → ~200 行）
- ❌ `admin_roles_handlers.go` + `admin_role_permissions_handlers.go` 需重构（634 行 → ~200 行）

**Step 5: 集成测试**
- 确保功能不变
- 确保性能不降

### 5.2 迁移示例

**当前代码**（`admin_residents_handlers.go`）：
```go
func (s *StubHandler) AdminResidents(w http.ResponseWriter, r *http.Request) {
    // 3000+ 行代码
    // 权限检查（200+ 行）
    // 业务规则验证
    // 数据转换
    // 直接 SQL（100+ 行）
}
```

**目标代码**：
```go
// Handler 层（~50 行）
type ResidentHandler struct {
    service *service.ResidentService
}

func (h *ResidentHandler) CreateResident(w http.ResponseWriter, r *http.Request) {
    var payload map[string]any
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        writeJSON(w, http.StatusBadRequest, Fail("invalid body"))
        return
    }
    
    tenantID, _ := getTenantIDFromRequest(r)
    userID := r.Header.Get("X-User-Id")
    userRole := r.Header.Get("X-User-Role")
    
    residentID, err := h.service.CreateResident(r.Context(), tenantID, payload, userID, userRole)
    if err != nil {
        handleError(w, err)
        return
    }
    
    writeJSON(w, http.StatusOK, Ok(map[string]any{"resident_id": residentID}))
}

// Service 层（~100 行）
type ResidentService struct {
    repo            *repository.PostgresResidentsRepo
    permissionChecker *PermissionChecker
}

func (s *ResidentService) CreateResident(ctx, tenantID, payload, userID, userRole) {
    // 权限检查
    if !s.permissionChecker.CanCreateResident(ctx, tenantID, userID, userRole) {
        return "", ErrPermissionDenied
    }
    
    // 业务规则验证
    if err := s.validateResidentPayload(payload); err != nil {
        return "", err
    }
    
    // 数据转换
    resident := s.convertPayloadToResident(payload)
    
    // 调用 Repository
    return s.repo.CreateResident(ctx, tenantID, resident)
}

// Repository 层（已实现）
func (r *PostgresResidentsRepo) CreateResident(ctx, tenantID, resident) {
    // 数据访问 + 数据一致性（替代触发器）
}
```

---

## 六、设计原则总结

### 6.1 分层原则
- **Handler**: HTTP 处理
- **Service**: 业务逻辑
- **Repository**: 数据访问

### 6.2 分类原则
- **按业务领域分类** Service
- **按数据实体分类** Repository

### 6.3 最小集原则
- **Phase 1**: ResidentService（最高优先级）
- **Phase 2**: UserService, AuthService
- **Phase 3**: TagService, RoleService

### 6.4 迁移原则
- **逐步迁移**，确保功能不变
- **参考现有模式**（Units/Devices Handler → Repository）
- **保持接口稳定**，避免破坏性变更

