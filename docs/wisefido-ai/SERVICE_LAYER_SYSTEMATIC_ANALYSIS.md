# Service 层系统性分析

## 📋 分析维度

基于以下维度判断是否需要 Service：
1. **权限检查**：是否需要复杂的权限验证（角色层级、资源权限等）
2. **业务规则验证**：是否有复杂的业务规则（状态转换、依赖检查、唯一性约束等）
3. **数据转换**：是否需要复杂的数据转换（前端格式 ↔ 领域模型、JSONB 处理等）
4. **业务编排**：是否需要协调多个 Repository 或发布事件
5. **Handler 复杂度**：Handler 代码行数和复杂度

---

## 🔍 前端 API 需求分析（API_FRONTEND_BACKEND_MATRIX.md）

### 1. 住户管理（Residents）

**API 端点**：
- GET `/admin/api/v1/residents` - 列表（需要权限过滤）
- GET `/admin/api/v1/residents/:id` - 详情（需要权限检查）
- POST `/admin/api/v1/residents` - 创建（需要权限检查、业务规则验证）
- PUT `/admin/api/v1/residents/:id` - 更新（需要权限检查、业务规则验证）
- DELETE `/admin/api/v1/residents/:id` - 删除（需要权限检查、依赖检查）
- PUT `/admin/api/v1/residents/:id/phi` - 更新 PHI（需要权限检查、HIPAA 合规）
- PUT `/admin/api/v1/residents/:id/contacts` - 更新联系人（需要权限检查）

**复杂度分析**：
- ✅ **权限检查**：复杂（角色层级、资源权限、Resident/Family 自检）
- ✅ **业务规则验证**：复杂（resident_account 必填、email/phone hash 处理、HIPAA 合规）
- ✅ **数据转换**：复杂（前端格式 ↔ 领域模型、PHI 数据转换）
- ✅ **业务编排**：复杂（创建住户时同时创建 PHI、联系人、账户）
- ✅ **Handler 复杂度**：3032 行

**结论**：✅ **需要 Service** - ResidentService

---

### 2. 用户管理（Users）

**API 端点**：
- GET `/admin/api/v1/users` - 列表（需要权限过滤）
- POST `/admin/api/v1/users` - 创建（需要权限检查、角色层级验证）
- PUT `/admin/api/v1/users/:id` - 更新（需要权限检查、角色层级验证）
- DELETE `/admin/api/v1/users/:id` - 删除（需要权限检查、依赖检查）
- POST `/admin/api/v1/users/:id/reset-password` - 重置密码（需要权限检查、业务规则）
- POST `/admin/api/v1/users/:id/reset-pin` - 重置 PIN（需要权限检查、业务规则）

**复杂度分析**：
- ✅ **权限检查**：复杂（角色层级验证、同级或下级角色创建规则）
- ✅ **业务规则验证**：复杂（密码规则、PIN 规则、角色层级规则）
- ✅ **数据转换**：中等（前端格式 ↔ 领域模型）
- ✅ **业务编排**：中等（创建用户时同时创建认证信息）
- ✅ **Handler 复杂度**：1257 行

**结论**：✅ **需要 Service** - UserService

---

### 3. 标签管理（Tags）

**API 端点**：
- GET `/admin/api/v1/tags` - 列表
- POST `/admin/api/v1/tags` - 创建（需要权限检查、业务规则验证）
- PUT `/admin/api/v1/tags/:id` - 更新（需要权限检查、业务规则验证）
- DELETE `/admin/api/v1/tags` - 删除（需要权限检查、依赖检查）
- POST `/admin/api/v1/tags/:id/objects` - 添加标签对象（需要权限检查、业务规则验证）
- DELETE `/admin/api/v1/tags/:id/objects` - 删除标签对象（需要权限检查）
- DELETE `/admin/api/v1/tags/types` - 删除标签类型（需要权限检查、依赖检查）
- GET `/admin/api/v1/tags/for-object` - 获取对象标签（需要权限过滤）

**复杂度分析**：
- ✅ **权限检查**：中等（标签管理权限）
- ✅ **业务规则验证**：中等（标签类型验证、对象关联验证）
- ✅ **数据转换**：中等（前端格式 ↔ 领域模型）
- ✅ **业务编排**：中等（删除标签时同时删除关联对象）
- ✅ **Handler 复杂度**：576 行

**结论**：✅ **需要 Service** - TagService

---

### 4. 角色管理（Roles）

**API 端点**：
- GET `/admin/api/v1/roles` - 列表
- POST `/admin/api/v1/roles` - 创建（需要权限检查、业务规则验证）
- PUT `/admin/api/v1/roles/:id` - 更新（需要权限检查、业务规则验证）
- DELETE `/admin/api/v1/roles/:id` - 删除（需要权限检查、依赖检查）
- PUT `/admin/api/v1/roles/:id/status` - 更新状态（需要权限检查、业务规则验证）

**复杂度分析**：
- ✅ **权限检查**：中等（角色管理权限）
- ✅ **业务规则验证**：中等（角色层级验证、状态验证）
- ✅ **数据转换**：简单（前端格式 ↔ 领域模型）
- ✅ **业务编排**：简单（无跨 Repository 操作）
- ✅ **Handler 复杂度**：~250 行

**结论**：✅ **需要 Service** - RoleService

---

### 5. 权限管理（RolePermissions）

**API 端点**：
- GET `/admin/api/v1/role-permissions` - 列表
- POST `/admin/api/v1/role-permissions` - 创建（需要权限检查、业务规则验证）
- POST `/admin/api/v1/role-permissions/batch` - 批量创建（需要权限检查、事务管理）
- PUT `/admin/api/v1/role-permissions/:id` - 更新（需要权限检查、业务规则验证）
- DELETE `/admin/api/v1/role-permissions/:id` - 删除（需要权限检查）
- PUT `/admin/api/v1/role-permissions/:id/status` - 更新状态（需要权限检查、业务规则验证）
- GET `/admin/api/v1/role-permissions/resource-types` - 获取资源类型

**复杂度分析**：
- ✅ **权限检查**：复杂（只有 SystemAdmin 可以修改全局权限）
- ✅ **业务规则验证**：中等（权限冲突验证、资源类型验证）
- ✅ **数据转换**：中等（前端格式 ↔ 领域模型）
- ✅ **业务编排**：复杂（批量操作需要事务管理）
- ✅ **Handler 复杂度**：~230 行

**结论**：✅ **需要 Service** - RolePermissionService

---

### 6. 地址层级管理（Units - Branch/Building/Floor/Unit/Room/Bed）

**API 端点**：
- **Building**：
  - GET `/admin/api/v1/buildings` - 列表
  - POST `/admin/api/v1/buildings` - 创建（需要权限检查、业务规则验证）
  - PUT `/admin/api/v1/buildings/:id` - 更新（需要权限检查、业务规则验证）
  - DELETE `/admin/api/v1/buildings/:id` - 删除（需要权限检查、依赖检查：是否有 Units）
- **Unit**：
  - GET `/admin/api/v1/units` - 列表（需要权限过滤）
  - POST `/admin/api/v1/units` - 创建（需要权限检查、业务规则验证、标签同步）
  - GET `/admin/api/v1/units/:id` - 详情
  - PUT `/admin/api/v1/units/:id` - 更新（需要权限检查、业务规则验证、标签同步）
  - DELETE `/admin/api/v1/units/:id` - 删除（需要权限检查、依赖检查：Rooms/Beds/Devices/Residents/Caregivers）
- **Room**：
  - GET `/admin/api/v1/rooms` - 列表（需要权限过滤）
  - POST `/admin/api/v1/rooms` - 创建（需要权限检查、业务规则验证）
  - PUT `/admin/api/v1/rooms/:id` - 更新（需要权限检查、业务规则验证）
  - DELETE `/admin/api/v1/rooms/:id` - 删除（需要权限检查、依赖检查：Beds/Devices）
- **Bed**：
  - GET `/admin/api/v1/beds` - 列表（需要权限过滤）
  - POST `/admin/api/v1/beds` - 创建（需要权限检查、业务规则验证）
  - PUT `/admin/api/v1/beds/:id` - 更新（需要权限检查、业务规则验证）
  - DELETE `/admin/api/v1/beds/:id` - 删除（需要权限检查、依赖检查：Devices/Residents）

**复杂度分析**：
- ✅ **权限检查**：中等（地址管理权限）
- ✅ **业务规则验证**：复杂（层级依赖检查、唯一性约束、标签同步）
- ✅ **数据转换**：中等（前端格式 ↔ 领域模型、层级结构转换）
- ✅ **业务编排**：复杂（标签同步到 tags_catalog、依赖检查、层级结构管理）
- ✅ **Handler 复杂度**：~200 行（但业务逻辑复杂）

**结论**：✅ **需要 Service** - UnitService

---

### 7. 设备管理（Devices）

**API 端点**：
- GET `/admin/api/v1/devices` - 列表（需要权限过滤、状态过滤）
- GET `/admin/api/v1/devices/:id` - 详情
- PUT `/admin/api/v1/devices/:id` - 更新（需要权限检查、业务规则验证、绑定验证、card 更新事件）
- DELETE `/admin/api/v1/devices/:id` - 删除（需要权限检查、依赖检查）
- GET `/device/api/v1/device/:id/relations` - 设备关系（需要权限过滤）

**复杂度分析**：
- ✅ **权限检查**：中等（设备管理权限）
- ✅ **业务规则验证**：复杂（设备状态转换规则、绑定规则验证）
- ✅ **数据转换**：中等（前端格式 ↔ 领域模型）
- ✅ **业务编排**：复杂（设备绑定变更后发布 card 更新事件）
- ✅ **Handler 复杂度**：~200 行（但业务逻辑复杂）

**结论**：✅ **需要 Service** - DeviceService

---

### 8. 告警配置（AlarmCloud）

**API 端点**：
- GET `/admin/api/v1/alarm-cloud` - 获取配置（需要权限检查、数据转换）
- PUT `/admin/api/v1/alarm-cloud` - 更新配置（需要权限检查、业务规则验证、数据转换）

**复杂度分析**：
- ✅ **权限检查**：中等（告警配置权限）
- ✅ **业务规则验证**：中等（配置数据格式验证）
- ✅ **数据转换**：复杂（JSONB 字段 ↔ 领域模型）
- ✅ **业务编排**：简单（无跨 Repository 操作）
- ✅ **Handler 复杂度**：~240 行

**结论**：✅ **需要 Service** - AlarmCloudService

---

### 9. 告警事件（AlarmEvents）

**API 端点**：
- GET `/admin/api/v1/alarm-events` - 列表（需要权限过滤、复杂查询、多表 JOIN）
- PUT `/admin/api/v1/alarm-events/:id/handle` - 处理告警（需要权限检查、业务规则验证、状态管理、跨表查询）

**复杂度分析**：
- ✅ **权限检查**：复杂（Facility vs Home 权限规则、权限过滤）
- ✅ **业务规则验证**：复杂（处理告警的规则、状态转换）
- ✅ **数据转换**：复杂（返回前端需要的格式，包含 JOIN 的数据）
- ✅ **业务编排**：复杂（跨表查询：event → device → card → unit_type）
- ✅ **Handler 复杂度**：~240 行（但业务逻辑复杂）

**结论**：✅ **需要 Service** - AlarmEventService

---

### 10. 设备库存（DeviceStore）

**API 端点**：
- GET `/admin/api/v1/device-store` - 列表
- PUT `/admin/api/v1/device-store/batch` - 批量更新
- POST `/admin/api/v1/device-store/import` - 导入
- GET `/admin/api/v1/device-store/import-template` - 导入模板
- GET `/admin/api/v1/device-store/export` - 导出

**复杂度分析**：
- ❌ **权限检查**：简单（设备库存管理权限）
- ❌ **业务规则验证**：简单（数据格式验证）
- ❌ **数据转换**：简单（Excel 导入导出）
- ❌ **业务编排**：简单（无跨 Repository 操作）
- ❌ **Handler 复杂度**：~200 行（主要是 Excel 处理）

**结论**：❌ **不需要 Service** - 直接使用 Repository

---

### 11. 租户管理（Tenants）

**API 端点**：
- GET `/admin/api/v1/tenants` - 列表
- POST `/admin/api/v1/tenants` - 创建（需要权限检查：只有 SystemAdmin）
- PUT `/admin/api/v1/tenants/:id` - 更新（需要权限检查：只有 SystemAdmin）
- DELETE `/admin/api/v1/tenants/:id` - 删除（需要权限检查：只有 SystemAdmin）

**复杂度分析**：
- ✅ **权限检查**：简单（只有 SystemAdmin）
- ❌ **业务规则验证**：简单（数据格式验证）
- ❌ **数据转换**：简单（前端格式 ↔ 领域模型）
- ❌ **业务编排**：简单（无跨 Repository 操作）
- ❌ **Handler 复杂度**：~100 行

**结论**：❌ **不需要 Service** - 直接使用 Repository（权限检查在 Handler 层即可）

---

### 12. 认证授权（Auth）

**API 端点**：
- POST `/auth/api/v1/login` - 登录（需要业务规则验证、密码验证）
- GET `/auth/api/v1/institutions/search` - 搜索机构
- POST `/auth/api/v1/forgot-password/send-code` - 发送验证码（需要业务规则验证）
- POST `/auth/api/v1/forgot-password/verify-code` - 验证验证码（需要业务规则验证）
- POST `/auth/api/v1/forgot-password/reset` - 重置密码（需要业务规则验证）

**复杂度分析**：
- ✅ **权限检查**：简单（登录不需要权限检查）
- ✅ **业务规则验证**：复杂（密码验证、验证码验证、密码重置规则）
- ✅ **数据转换**：中等（前端格式 ↔ 领域模型）
- ✅ **业务编排**：中等（登录时创建 session、发送验证码等）
- ✅ **Handler 复杂度**：886 行

**结论**：✅ **需要 Service** - AuthService

---

### 13. VitalFocus 数据查询（Redis 缓存）

**API 端点**：
- GET `/data/api/v1/data/vital-focus/cards` - 获取卡片列表（从 Redis 读取缓存）
- GET `/data/api/v1/data/vital-focus/card/:id` - 获取卡片详情（从 Redis 读取缓存，支持 card_id 和 resident_id）
- POST `/data/api/v1/data/vital-focus/selection` - 保存用户选择（写入 Redis）

**复杂度分析**：
- ✅ **权限检查**：中等（tenant_id 过滤）
- ❌ **业务规则验证**：简单（无复杂业务规则）
- ✅ **数据转换**：复杂（decodeAndNormalizeFullCard - 字段类型规范化、数据源规范化）
- ❌ **业务编排**：简单（无跨 Repository 操作）
- ✅ **Handler 复杂度**：~305 行（数据转换逻辑复杂）

**结论**：✅ **需要 Service** - VitalFocusService（数据转换逻辑复杂，需要规范化处理）

---

### 14. 巡检管理（Rounds/RoundDetails）

**Repository**：
- RoundsRepository - 巡房记录
- RoundDetailsRepository - 巡房详细记录

**前端 API 需求**：
- 当前未在前端 API 矩阵中找到，但 Repository 已实现
- 如果有前端 API，可能需要：
  - GET `/admin/api/v1/rounds` - 列表（需要权限过滤）
  - POST `/admin/api/v1/rounds` - 创建（需要权限检查、业务规则验证）
  - PUT `/admin/api/v1/rounds/:id` - 更新（需要权限检查、业务规则验证）
  - DELETE `/admin/api/v1/rounds/:id` - 删除（需要权限检查、依赖检查）
  - PUT `/admin/api/v1/rounds/:id/status` - 更新状态（需要权限检查、状态转换规则）
  - GET `/admin/api/v1/round-details` - 列表（需要权限过滤）
  - POST `/admin/api/v1/round-details` - 创建（需要权限检查、业务规则验证）

**复杂度分析**（假设有前端 API）：
- ✅ **权限检查**：中等（巡检管理权限）
- ✅ **业务规则验证**：中等（巡房状态转换规则、依赖检查）
- ✅ **数据转换**：中等（前端格式 ↔ 领域模型）
- ❌ **业务编排**：简单（无跨 Repository 操作）
- ❌ **Handler 复杂度**：未知（未实现）

**结论**：⚠️ **待定** - 如果有前端 API，需要 RoundService；如果没有，则不需要

---

### 15. 其他 API（ServiceLevel, CardOverview, Settings, Reports, Address）

**复杂度分析**：
- **ServiceLevel**：简单（只读，无业务规则）
- **CardOverview**：简单（只读，无业务规则）
- **Settings**：中等（设备监控配置，可能需要 Service）
- **Reports**：简单（只读，无业务规则）
- **Address**：简单（地址管理，可能需要 Service，但当前未实现）

**结论**：❌ **暂不需要 Service** - 等实现后再判断

---

## 📊 最终决策矩阵

| 领域 | API 端点 | 权限检查 | 业务规则 | 数据转换 | 业务编排 | Handler 行数 | **是否需要 Service** |
|------|---------|---------|---------|---------|---------|-------------|-------------------|
| **Residents** | 7 个端点 | ✅ 复杂 | ✅ 复杂 | ✅ 复杂 | ✅ 复杂 | 3032 | ✅ **需要** |
| **Users** | 6 个端点 | ✅ 复杂 | ✅ 复杂 | ✅ 中等 | ✅ 中等 | 1257 | ✅ **需要** |
| **Tags** | 8 个端点 | ✅ 中等 | ✅ 中等 | ✅ 中等 | ✅ 中等 | 576 | ✅ **需要** |
| **Roles** | 5 个端点 | ✅ 中等 | ✅ 中等 | ✅ 简单 | ❌ 简单 | ~250 | ✅ **需要** |
| **RolePermissions** | 7 个端点 | ✅ 复杂 | ✅ 中等 | ✅ 中等 | ✅ 复杂 | ~230 | ✅ **需要** |
| **Units** | 16 个端点 | ✅ 中等 | ✅ 复杂 | ✅ 中等 | ✅ 复杂 | ~200 | ✅ **需要** |
| **Devices** | 5 个端点 | ✅ 中等 | ✅ 复杂 | ✅ 中等 | ✅ 复杂 | ~200 | ✅ **需要** |
| **AlarmCloud** | 2 个端点 | ✅ 中等 | ✅ 中等 | ✅ 复杂 | ❌ 简单 | ~240 | ✅ **需要** |
| **AlarmEvents** | 2 个端点 | ✅ 复杂 | ✅ 复杂 | ✅ 复杂 | ✅ 复杂 | ~240 | ✅ **需要** |
| **Auth** | 5 个端点 | ❌ 简单 | ✅ 复杂 | ✅ 中等 | ✅ 中等 | 886 | ✅ **需要** |
| **VitalFocus** | 3 个端点 | ✅ 中等 | ❌ 简单 | ✅ 复杂 | ❌ 简单 | ~305 | ✅ **需要** |
| **Rounds** | 待定 | ⚠️ 待定 | ⚠️ 待定 | ⚠️ 待定 | ⚠️ 待定 | 未知 | ⚠️ **待定** |
| **DeviceStore** | 5 个端点 | ❌ 简单 | ❌ 简单 | ❌ 简单 | ❌ 简单 | ~200 | ❌ **不需要** |
| **Tenants** | 4 个端点 | ✅ 简单 | ❌ 简单 | ❌ 简单 | ❌ 简单 | ~100 | ❌ **不需要** |

---

## 🎯 最终 Service 清单（13个 + 2个待定）

### 已确认的 Service（13个）

1. ✅ **ResidentService** - 住户管理（3032 行 Handler）
2. ✅ **UserService** - 用户管理（1257 行 Handler）
3. ✅ **TagService** - 标签管理（576 行 Handler）
4. ✅ **RoleService** - 角色管理（~250 行 Handler）
5. ✅ **RolePermissionService** - 权限管理（~230 行 Handler）
6. ✅ **UnitService** - 地址层级管理（Branch → Building → Floor → Unit → Room → Bed）
7. ✅ **DeviceService** - 设备管理（设备状态、设备绑定、card 更新事件）
8. ✅ **AlarmCloudService** - 告警配置（JSONB 数据转换）
9. ✅ **AlarmEventService** - 告警事件（权限过滤、复杂查询、跨表查询）
10. ✅ **AuthService** - 认证授权（密码验证、验证码验证、密码重置）
11. ✅ **VitalFocusService** - VitalFocus 数据查询（Redis 缓存、数据规范化转换）
12. ✅ **SleepaceReportService** - 睡眠报告（从时间序列数据聚合生成报告）
13. ✅ **DeviceMonitorSettingsService** - 设备监控配置（配置参数验证、数据转换）

---

## ⚠️ 待定的 Service（2个）

1. **RoundService** - 巡检管理（Rounds/RoundDetails）
   - 如果前端有 API 需求，则需要 Service
   - 如果只是后台服务使用，则不需要 Service

2. **RadarRealtimeService** - 雷达实时轨迹
   - 如果只是简单的数据库查询，可以不需要 Service
   - 如果需要复杂的数据聚合（如轨迹点聚合、时间窗口计算），则需要 Service

---

## ❌ 不需要 Service 的 Repository（2个）

1. **DeviceStoreRepository** - 设备库存管理（简单领域，Excel 导入导出）
2. **TenantsRepository** - 租户管理（简单领域，权限检查在 Handler 层即可）

---

## 📋 后台服务说明

### wisefido-card-aggregator（卡片聚合服务）

**功能**：
- 数据聚合：从 PostgreSQL + Redis 读取，组装成完整的 VitalFocusCard
- 缓存管理：写入 Redis 缓存（`vital-focus:card:{card_id}:full`）

**架构**：
- 后台服务（不是 HTTP API）
- 已有 Service 层（`internal/service/aggregator.go`）
- 不需要额外的 Service 层

**结论**：❌ **不需要额外的 Service** - 它本身就是后台服务，已有 Service 层

---

## 📋 分析总结

### 需要 Service 的原因（共同特征）

1. **复杂的权限检查**：角色层级验证、资源权限、权限过滤
2. **复杂的业务规则验证**：状态转换、依赖检查、唯一性约束、数据格式验证
3. **复杂的数据转换**：JSONB 处理、前端格式 ↔ 领域模型、多表 JOIN 数据转换
4. **复杂的业务编排**：跨 Repository 操作、事件发布、事务管理
5. **Handler 复杂度高**：代码行数多，业务逻辑复杂

### 不需要 Service 的原因（共同特征）

1. **简单的权限检查**：单一权限检查，可在 Handler 层完成
2. **简单的业务规则验证**：数据格式验证，可在 Repository 层完成
3. **简单的数据转换**：直接映射，无需复杂转换
4. **无业务编排**：单一 Repository 操作，无跨 Repository 操作
5. **Handler 复杂度低**：代码行数少，业务逻辑简单

---

## 🚀 实现优先级

### Phase 1: 最高优先级（复杂度极高）
1. ✅ **ResidentService** - 3032 行 Handler 需要重构

### Phase 2: 高优先级（复杂度高）
2. ✅ **UserService** - 1257 行 Handler 需要重构
3. ✅ **AuthService** - 886 行 Handler 需要重构

### Phase 3: 中优先级（复杂度中）
4. ✅ **TagService** - 576 行 Handler 需要重构
5. ✅ **RoleService** - ~250 行 Handler 需要重构
6. ✅ **RolePermissionService** - ~230 行 Handler 需要重构
7. ✅ **AlarmCloudService** - ~240 行 Handler 需要重构
8. ✅ **AlarmEventService** - 复杂查询、权限过滤、跨表查询
9. ✅ **UnitService** - 地址层级管理、依赖检查、标签同步
10. ✅ **DeviceService** - 设备状态管理、设备绑定管理、card 更新事件
11. ✅ **VitalFocusService** - VitalFocus 数据查询（Redis 缓存、数据规范化转换）

### Phase 4: 待定（根据前端需求）
12. ⚠️ **RoundService** - 巡检管理（如果有前端 API 需求）
13. ✅ **SleepaceReportService** - 睡眠报告（前端已使用，需要从时间序列数据聚合生成报告）
14. ✅ **DeviceMonitorSettingsService** - 设备监控配置（前端已使用，配置参数验证、数据转换）
15. ⚠️ **RadarRealtimeService** - 雷达实时轨迹（如果前端需要，取决于数据聚合复杂度）

