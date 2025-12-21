# Service 层完整设计（基于 Repository 和前端 API 需求）

## 📋 分析依据

### 1. Repository 清单（wisefido-data）

| Repository | 接口文件 | 实现文件 | 状态 |
|-----------|---------|---------|------|
| ResidentsRepository | residents_repo.go | postgres_residents.go | ✅ 已实现 |
| UsersRepository | users_repo.go | postgres_users.go | ✅ 已实现 |
| TagsRepository | tags_repo.go | postgres_tags.go | ✅ 已实现 |
| RolesRepository | roles_repo.go | postgres_roles.go | ✅ 已实现 |
| RolePermissionsRepository | role_permissions_repo.go | postgres_role_permissions.go | ✅ 已实现 |
| UnitsRepository | units_repo.go | postgres_units.go | ✅ 已实现 |
| DevicesRepository | devices_repo.go | postgres_devices.go | ✅ 已实现 |
| DeviceStoreRepository | device_store_repo.go | postgres_device_store.go | ✅ 已实现 |
| TenantsRepository | tenants_repo.go | postgres_tenants.go | ✅ 已实现 |
| AlarmCloudRepository | alarm_cloud_repo.go | postgres_alarm_cloud.go | ✅ 已实现 |
| AlarmDeviceRepository | alarm_device_repo.go | postgres_alarm_device.go | ✅ 已实现 |

### 2. 前端 API 需求（API_FRONTEND_BACKEND_MATRIX.md）

| 前端模块 | API 端点 | 方法 | 功能 |
|---------|---------|------|------|
| `src/api/resident/resident.ts` | `/admin/api/v1/residents` | GET, POST, PUT, DELETE | 住户管理 |
| | `/admin/api/v1/residents/:id/phi` | PUT | 更新 PHI |
| | `/admin/api/v1/residents/:id/contacts` | PUT | 更新联系人 |
| `src/api/admin/user/user.ts` | `/admin/api/v1/users` | GET, POST, PUT, DELETE | 用户管理 |
| | `/admin/api/v1/users/:id/reset-password` | POST | 重置密码 |
| | `/admin/api/v1/users/:id/reset-pin` | POST | 重置 PIN |
| `src/api/admin/tags/tags.ts` | `/admin/api/v1/tags` | GET, POST, PUT, DELETE | 标签管理 |
| | `/admin/api/v1/tags/:id/objects` | POST, DELETE | 标签对象管理 |
| | `/admin/api/v1/tags/types` | DELETE | 删除标签类型 |
| | `/admin/api/v1/tags/for-object` | GET | 获取对象标签 |
| `src/api/admin/role/role.ts` | `/admin/api/v1/roles` | GET, POST, PUT, DELETE | 角色管理 |
| | `/admin/api/v1/roles/:id/status` | PUT | 更新角色状态 |
| `src/api/admin/role-permission/rolePermission.ts` | `/admin/api/v1/role-permissions` | GET, POST, PUT, DELETE | 权限管理 |
| | `/admin/api/v1/role-permissions/batch` | POST | 批量创建权限 |
| | `/admin/api/v1/role-permissions/:id/status` | PUT | 更新权限状态 |
| | `/admin/api/v1/role-permissions/resource-types` | GET | 获取资源类型 |
| `src/api/units/unit.ts` | `/admin/api/v1/buildings` | GET, POST, PUT, DELETE | 楼栋管理 |
| | `/admin/api/v1/units` | GET, POST, PUT, DELETE | 单元管理 |
| | `/admin/api/v1/rooms` | GET, POST, PUT, DELETE | 房间管理 |
| | `/admin/api/v1/beds` | GET, POST, PUT, DELETE | 床位管理 |
| `src/api/devices/device.ts` | `/admin/api/v1/devices` | GET, PUT, DELETE | 设备管理 |
| | `/device/api/v1/device/:id/relations` | GET | 设备关系 |
| `src/api/alarm/alarm.ts` | `/admin/api/v1/alarm-cloud` | GET, PUT | 告警配置 |
| | `/admin/api/v1/alarm-events` | GET | 告警事件列表 |
| | `/admin/api/v1/alarm-events/:id/handle` | PUT | 处理告警 |

### 3. Handler 现状（wisefido-data）

| Handler | 文件 | 行数 | 复杂度 | 是否需要 Service |
|---------|------|------|--------|----------------|
| AdminResidents | admin_residents_handlers.go | 3032 | 极高 | ✅ **需要** |
| AdminUsers | admin_users_handlers.go | 1257 | 高 | ✅ **需要** |
| AdminTags | admin_tags_handlers.go | 576 | 中 | ✅ **需要** |
| AdminRoles | admin_roles_handlers.go | ~250 | 中 | ✅ **需要** |
| AdminRolePermissions | admin_role_permissions_handlers.go | ~230 | 中 | ✅ **需要** |
| AdminUnits | admin_units_devices_handlers.go | ~200 | 低 | ❌ **不需要**（已用 Repository） |
| AdminDevices | admin_units_devices_handlers.go | ~200 | 低 | ❌ **不需要**（已用 Repository） |
| AdminAlarm | admin_alarm_handlers.go | ~240 | 中 | ✅ **需要** |

---

## 🎯 Service 层设计决策

### 需要 Service 的领域（基于 Handler 复杂度和业务需求）

| Service | 对应 Repository | 对应 Handler | 需要原因 |
|---------|----------------|-------------|---------|
| **ResidentService** | ResidentsRepository | AdminResidents | ✅ 权限检查、业务规则验证、数据转换（3032行） |
| **UserService** | UsersRepository | AdminUsers | ✅ 权限检查、业务规则验证、密码重置逻辑（1257行） |
| **TagService** | TagsRepository | AdminTags | ✅ 权限检查、业务规则验证、标签对象管理（576行） |
| **RoleService** | RolesRepository | AdminRoles | ✅ 权限检查、业务规则验证、角色状态管理（~250行） |
| **RolePermissionService** | RolePermissionsRepository | AdminRolePermissions | ✅ 权限检查、业务规则验证、批量操作（~230行） |
| **AlarmCloudService** | AlarmCloudRepository | AdminAlarm | ✅ 权限检查、业务规则验证、数据转换（~240行） |
| **AlarmEventService** | AlarmEventsRepository | AdminAlarm | ✅ 权限检查、权限过滤、复杂查询、状态管理 |

### 需要 Service 的领域（设备管理）

| Service | 对应 Repository | 对应 Handler | 需要原因 |
|---------|----------------|-------------|---------|
| **DeviceService** | DevicesRepository | AdminDevices | ✅ 权限检查、设备状态管理、设备绑定管理、业务编排（card 更新事件） |

### 需要 Service 的领域（Unit 管理）

| Service | 对应 Repository | 对应 Handler | 需要原因 |
|---------|----------------|-------------|---------|
| **UnitService** | UnitsRepository | AdminUnits | ✅ 权限检查、业务规则验证（依赖检查、标签同步）、数据转换、业务编排（层级结构管理） |

### 不需要 Service 的领域（简单领域或已用 Repository）

| Repository | 对应 Handler | 不需要原因 |
|-----------|-------------|-----------|
| DeviceStoreRepository | AdminDevices | ✅ 已直接使用 Repository，业务逻辑简单 |
| TenantsRepository | AdminTenants | ✅ 已直接使用 Repository，业务逻辑简单 |

---

## 📊 完整 Service 清单

### 1. ResidentService（住户管理）

**职责**：
- 权限检查（创建/更新/删除住户）
- 业务规则验证（PHI 数据验证、联系人验证）
- 数据转换（前端格式 ↔ 领域模型）
- 业务编排（创建住户时同时创建 PHI、联系人）

**方法**：
```go
type ResidentService struct {
    repo *repository.ResidentsRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// CRUD
func (s *ResidentService) ListResidents(ctx, tenantID, userID, userRole, filters, page, size)
func (s *ResidentService) GetResident(ctx, tenantID, userID, userRole, residentID)
func (s *ResidentService) CreateResident(ctx, tenantID, userID, userRole, payload)
func (s *ResidentService) UpdateResident(ctx, tenantID, userID, userRole, residentID, payload)
func (s *ResidentService) DeleteResident(ctx, tenantID, userID, userRole, residentID)

// PHI 管理
func (s *ResidentService) UpdateResidentPHI(ctx, tenantID, userID, userRole, residentID, phiData)

// 联系人管理
func (s *ResidentService) UpdateResidentContacts(ctx, tenantID, userID, userRole, residentID, contacts)
```

---

### 2. UserService（用户管理）

**职责**：
- 权限检查（创建/更新/删除用户、角色层级检查）
- 业务规则验证（密码规则、PIN 规则、角色层级）
- 数据转换（前端格式 ↔ 领域模型）
- 业务编排（创建用户时同时创建认证信息）

**方法**：
```go
type UserService struct {
    repo *repository.UsersRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// CRUD
func (s *UserService) ListUsers(ctx, tenantID, userID, userRole, filters, page, size)
func (s *UserService) GetUser(ctx, tenantID, userID, userRole, targetUserID)
func (s *UserService) CreateUser(ctx, tenantID, userID, userRole, payload)
func (s *UserService) UpdateUser(ctx, tenantID, userID, userRole, targetUserID, payload)
func (s *UserService) DeleteUser(ctx, tenantID, userID, userRole, targetUserID)

// 密码管理
func (s *UserService) ResetPassword(ctx, tenantID, userID, userRole, targetUserID)
func (s *UserService) ResetPin(ctx, tenantID, userID, userRole, targetUserID)
```

---

### 3. TagService（标签管理）

**职责**：
- 权限检查（创建/更新/删除标签）
- 业务规则验证（标签类型验证、对象关联验证）
- 数据转换（前端格式 ↔ 领域模型）
- 业务编排（删除标签时同时删除关联对象）

**方法**：
```go
type TagService struct {
    repo *repository.TagsRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// CRUD
func (s *TagService) ListTags(ctx, tenantID, userID, userRole, filters, page, size)
func (s *TagService) GetTag(ctx, tenantID, userID, userRole, tagID)
func (s *TagService) CreateTag(ctx, tenantID, userID, userRole, payload)
func (s *TagService) UpdateTag(ctx, tenantID, userID, userRole, tagID, payload)
func (s *TagService) DeleteTag(ctx, tenantID, userID, userRole, tagID)

// 标签对象管理
func (s *TagService) AddTagObjects(ctx, tenantID, userID, userRole, tagID, objects)
func (s *TagService) RemoveTagObjects(ctx, tenantID, userID, userRole, tagID, objects)
func (s *TagService) GetTagsForObject(ctx, tenantID, userID, userRole, objectType, objectID)

// 标签类型管理
func (s *TagService) DeleteTagType(ctx, tenantID, userID, userRole, tagType)
```

---

### 4. RoleService（角色管理）

**职责**：
- 权限检查（创建/更新/删除角色）
- 业务规则验证（角色层级验证、状态验证）
- 数据转换（前端格式 ↔ 领域模型）

**方法**：
```go
type RoleService struct {
    repo *repository.RolesRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// CRUD
func (s *RoleService) ListRoles(ctx, tenantID, userID, userRole, filters, page, size)
func (s *RoleService) GetRole(ctx, tenantID, userID, userRole, roleID)
func (s *RoleService) CreateRole(ctx, tenantID, userID, userRole, payload)
func (s *RoleService) UpdateRole(ctx, tenantID, userID, userRole, roleID, payload)
func (s *RoleService) DeleteRole(ctx, tenantID, userID, userRole, roleID)

// 状态管理
func (s *RoleService) UpdateRoleStatus(ctx, tenantID, userID, userRole, roleID, status)
```

---

### 5. RolePermissionService（权限管理）

**职责**：
- 权限检查（创建/更新/删除权限）
- 业务规则验证（权限冲突验证、资源类型验证）
- 数据转换（前端格式 ↔ 领域模型）
- 业务编排（批量创建权限）

**方法**：
```go
type RolePermissionService struct {
    repo *repository.RolePermissionsRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// CRUD
func (s *RolePermissionService) ListRolePermissions(ctx, tenantID, userID, userRole, filters, page, size)
func (s *RolePermissionService) GetRolePermission(ctx, tenantID, userID, userRole, permissionID)
func (s *RolePermissionService) CreateRolePermission(ctx, tenantID, userID, userRole, payload)
func (s *RolePermissionService) UpdateRolePermission(ctx, tenantID, userID, userRole, permissionID, payload)
func (s *RolePermissionService) DeleteRolePermission(ctx, tenantID, userID, userRole, permissionID)

// 批量操作
func (s *RolePermissionService) BatchCreateRolePermissions(ctx, tenantID, userID, userRole, permissions)

// 状态管理
func (s *RolePermissionService) UpdateRolePermissionStatus(ctx, tenantID, userID, userRole, permissionID, status)

// 资源类型
func (s *RolePermissionService) GetResourceTypes(ctx, tenantID, userID, userRole)
```

---

### 6. AlarmCloudService（告警配置）

**职责**：
- 权限检查（查看/编辑配置）
- 业务规则验证（配置数据格式验证）
- 数据转换（JSONB 字段 ↔ 领域模型）

**方法**：
```go
type AlarmCloudService struct {
    repo *repository.AlarmCloudRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

func (s *AlarmCloudService) GetAlarmCloudConfig(ctx, tenantID, userID, userRole)
func (s *AlarmCloudService) UpdateAlarmCloudConfig(ctx, tenantID, userID, userRole, config)
```

---

### 7. AlarmEventService（告警事件）

**职责**：
- 权限检查（查看/处理告警）
- 权限过滤（根据用户角色过滤可查看的告警）
- 业务规则验证（处理告警的规则）
- 数据转换（返回前端需要的格式，包含 JOIN 的数据）

**方法**：
```go
type AlarmEventService struct {
    repo *repository.AlarmEventsRepository
    cardRepo *repository.CardRepository
    deviceRepo *repository.DeviceRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

func (s *AlarmEventService) ListAlarmEvents(ctx, tenantID, userID, userRole, filters, page, size)
func (s *AlarmEventService) HandleAlarmEvent(ctx, tenantID, userID, userRole, eventID, params)
```

---

### 8. DeviceService（设备管理）

**职责**：
- 权限检查（创建/更新/删除设备、设备绑定）
- 业务规则验证（设备状态转换规则、绑定规则）
- 数据转换（前端格式 ↔ 领域模型）
- 业务编排（设备绑定变更后发布 card 更新事件）

**方法**：
```go
type DeviceService struct {
    repo *repository.DevicesRepository
    unitsRepo *repository.UnitsRepository
    permissionChecker *PermissionChecker
    eventPublisher *EventPublisher // 用于发布 card 更新事件
    logger *zap.Logger
}

// CRUD
func (s *DeviceService) ListDevices(ctx, tenantID, userID, userRole, filters, page, size)
func (s *DeviceService) GetDevice(ctx, tenantID, userID, userRole, deviceID)
func (s *DeviceService) UpdateDevice(ctx, tenantID, userID, userRole, deviceID, payload)

// 设备状态管理
func (s *DeviceService) UpdateDeviceStatus(ctx, tenantID, userID, userRole, deviceID, status)
func (s *DeviceService) DisableDevice(ctx, tenantID, userID, userRole, deviceID)

// 设备绑定管理（绑定到 Room 或 Bed）
func (s *DeviceService) BindDeviceToRoom(ctx, tenantID, userID, userRole, deviceID, roomID)
func (s *DeviceService) BindDeviceToBed(ctx, tenantID, userID, userRole, deviceID, bedID)
func (s *DeviceService) UnbindDevice(ctx, tenantID, userID, userRole, deviceID)

// 业务规则验证
func (s *DeviceService) validateStatusTransition(oldStatus, newStatus string) error
func (s *DeviceService) validateBinding(deviceID, roomID, bedID string) error
```

**设备状态管理规则**：
- 状态值：`online`, `offline`, `error`, `disabled`
- 状态转换规则：
  - `disabled` → `online`：需要业务访问权限为 `approved`
  - `online` → `disabled`：允许（禁用设备）
  - `offline` → `online`：允许（设备上线）
  - `error` → `online`：允许（错误恢复）

**设备绑定管理规则**：
- 设备可以绑定到 `bound_room_id` 或 `bound_bed_id`（互斥）
- 绑定验证：
  - 验证 room/bed 是否属于该租户
  - 验证 room/bed 是否存在
- 绑定变更后：
  - 发布 card 更新事件（通知 card-aggregator 重新聚合）
  - 更新设备状态（如果需要）

---

### 9. UnitService（地址层级管理）

**职责**：
- 权限检查（创建/更新/删除 Building/Unit/Room/Bed）
- 业务规则验证（依赖检查、唯一性约束、标签同步）
- 数据转换（前端格式 ↔ 领域模型）
- 业务编排（层级结构管理、标签同步到 tags_catalog）

**方法**：
```go
type UnitService struct {
    repo *repository.UnitsRepository
    tagsRepo *repository.TagsRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// Building 管理
func (s *UnitService) ListBuildings(ctx, tenantID, userID, userRole, branchTag)
func (s *UnitService) CreateBuilding(ctx, tenantID, userID, userRole, payload)
func (s *UnitService) UpdateBuilding(ctx, tenantID, userID, userRole, buildingID, payload)
func (s *UnitService) DeleteBuilding(ctx, tenantID, userID, userRole, buildingID) // 检查是否有 Units

// Unit 管理
func (s *UnitService) ListUnits(ctx, tenantID, userID, userRole, filters, page, size)
func (s *UnitService) CreateUnit(ctx, tenantID, userID, userRole, payload) // 同步 branch_tag, area_tag
func (s *UnitService) UpdateUnit(ctx, tenantID, userID, userRole, unitID, payload) // 同步 branch_tag, area_tag
func (s *UnitService) DeleteUnit(ctx, tenantID, userID, userRole, unitID) // 检查依赖：rooms, beds, devices, residents, caregivers

// Room 管理
func (s *UnitService) ListRoomsWithBeds(ctx, tenantID, userID, userRole, unitID)
func (s *UnitService) CreateRoom(ctx, tenantID, userID, userRole, unitID, payload)
func (s *UnitService) UpdateRoom(ctx, tenantID, userID, userRole, roomID, payload)
func (s *UnitService) DeleteRoom(ctx, tenantID, userID, userRole, roomID) // 检查依赖：beds, devices

// Bed 管理
func (s *UnitService) ListBeds(ctx, tenantID, userID, userRole, roomID)
func (s *UnitService) CreateBed(ctx, tenantID, userID, userRole, roomID, payload)
func (s *UnitService) UpdateBed(ctx, tenantID, userID, userRole, bedID, payload)
func (s *UnitService) DeleteBed(ctx, tenantID, userID, userRole, bedID) // 检查依赖：devices, residents
```

**业务规则**：
- 删除 Building：检查是否有 Units
- 删除 Unit：检查是否有 Rooms, Beds, Devices, Residents, Caregivers
- 删除 Room：检查是否有 Beds, Devices
- 删除 Bed：检查是否有 Devices, Residents
- 标签同步：创建/更新 Unit 时，同步 `branch_tag` 和 `area_tag` 到 `tags_catalog`

---

## 📋 总结### Service 清单（13个 + 2个待定）

#### 已确认的 Service（13个）

1. ✅ **ResidentService** - 住户管理（3032 行 Handler，复杂权限检查、业务规则验证、数据转换、业务编排）
2. ✅ **UserService** - 用户管理（1257 行 Handler，角色层级验证、密码重置逻辑）
3. ✅ **TagService** - 标签管理（576 行 Handler，标签对象管理、依赖检查）
4. ✅ **RoleService** - 角色管理（~250 行 Handler，角色状态管理）
5. ✅ **RolePermissionService** - 权限管理（~230 行 Handler，批量操作、权限冲突验证）
6. ✅ **UnitService** - 地址层级管理（Branch → Building → Floor → Unit → Room → Bed，依赖检查、标签同步）
7. ✅ **DeviceService** - 设备管理（设备状态管理、设备绑定管理、card 更新事件）
8. ✅ **AlarmCloudService** - 告警配置（JSONB 数据转换、权限检查）
9. ✅ **AlarmEventService** - 告警事件（权限过滤、复杂查询、跨表查询、状态管理）
10. ✅ **AuthService** - 认证授权（密码验证、验证码验证、密码重置）
11. ✅ **VitalFocusService** - VitalFocus 数据查询（Redis 缓存、数据规范化转换）
12. ✅ **SleepaceReportService** - 睡眠报告（从时间序列数据聚合生成报告，或调用 Sleepace 厂家服务）
13. ✅ **DeviceMonitorSettingsService** - 设备监控配置（配置参数验证、数据转换、可能需要同步更新到设备）

### 待定的 Service（2个）

1. ⚠️ **RoundService** - 巡检管理（Rounds/RoundDetails）
   - 如果前端有 API 需求，则需要 Service
   - 如果只是后台服务使用，则不需要 Service

2. ⚠️ **RadarRealtimeService** - 雷达实时轨迹
   - 如果只是简单的数据库查询，可以不需要 Service
   - 如果需要复杂的数据聚合（如轨迹点聚合、时间窗口计算），则需要 Service

### 不需要 Service 的 Repository（2个）

1. **DeviceStoreRepository** - 设备库存管理（简单领域，Excel 导入导出，无复杂业务规则）
2. **TenantsRepository** - 租户管理（简单领域，权限检查在 Handler 层即可）

### 后台服务说明

**wisefido-card-aggregator**（卡片聚合服务）：
- 后台服务（不是 HTTP API）
- 已有 Service 层（`internal/service/aggregator.go`）
- 不需要额外的 Service 层

---

## 📊 系统性分析依据

详细分析见：`SERVICE_LAYER_SYSTEMATIC_ANALYSIS.md`

**分析维度**：
1. 权限检查复杂度
2. 业务规则验证复杂度
3. 数据转换复杂度
4. 业务编排复杂度
5. Handler 代码行数和复杂度

**决策原则**：
- 满足 3 个及以上维度为"复杂" → 需要 Service
- 满足 2 个及以下维度为"简单" → 不需要 Service

---

## 🎯 设计原则

### Service 层职责（居中调度、权限控制）

1. **权限检查**：调用 PermissionChecker 验证用户权限
2. **业务规则验证**：验证业务规则（数据格式、状态转换等）
3. **数据转换**：前端格式 ↔ 领域模型
4. **业务编排**：协调多个 Repository 完成复杂业务
5. **事务管理**：跨 Repository 的事务管理（如需要）

### Repository 层职责（数据访问）

1. **数据访问**：SQL 操作
2. **数据一致性**：替代触发器，保证数据一致性
3. **事务管理**：单 Repository 的事务管理

### Handler 层职责（HTTP 处理）

1. **HTTP 处理**：解析请求、生成响应
2. **路由分发**：根据 HTTP 方法和路径分发
3. **错误处理**：捕获异常并返回 HTTP 状态码

---

## 🚀 实现优先级

### Phase 1: 最高优先级（复杂度极高）
1. ✅ **ResidentService** - 3032 行 Handler 需要重构

### Phase 2: 高优先级（复杂度高）
2. ✅ **UserService** - 1257 行 Handler 需要重构
3. ✅ **AlarmEventService** - 复杂查询、权限过滤

### Phase 3: 中优先级（复杂度中）
4. ✅ **TagService** - 576 行 Handler 已完成重构 ✅
   - Service: `internal/service/tag_service.go`
   - Handler: `internal/http/admin_tags_handler.go`
   - 已注册路由，功能完整
5. ✅ **RoleService** - ~250 行 Handler 需要重构
6. ✅ **RolePermissionService** - ~230 行 Handler 需要重构
7. ✅ **AlarmCloudService** - ~240 行 Handler 需要重构
8. ✅ **DeviceService** - 设备状态管理、设备绑定管理、业务编排（card 更新事件）
9. ✅ **UnitService** - 地址层级管理（Branch → Building → Floor → Unit → Room → Bed）、依赖检查、标签同步

