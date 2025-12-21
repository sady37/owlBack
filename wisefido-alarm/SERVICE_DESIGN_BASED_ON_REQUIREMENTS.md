# Service 层设计（基于实际需求）

## 📋 前端页面和功能分析

### 前端页面

1. **AlarmCloud.vue** - 报警策略配置页面
   - 功能：查看和编辑报警策略配置
   - 权限：需要 `canEdit` 权限检查
   - API：
     - `GET /admin/api/v1/alarm-cloud` - 获取配置
     - `PUT /admin/api/v1/alarm-cloud` - 更新配置

2. **AlarmRecord.vue** - 报警记录页面
   - 功能：查看报警记录（Pending/Resolved两个tab）
   - API：
     - `GET /admin/api/v1/alarm-events` - 获取报警事件列表

3. **AlarmRecordList.vue** - 报警记录列表组件
   - 功能：显示报警列表，处理报警
   - 权限：处理报警需要权限检查（Facility vs Home）
   - API：
     - `GET /admin/api/v1/alarm-events` - 获取列表
     - `PUT /admin/api/v1/alarm-events/:id/handle` - 处理报警

---

## 🔍 后端 API 需求分析

### 1. Alarm Cloud API

#### GET /admin/api/v1/alarm-cloud
**功能**：获取报警策略配置

**需求**：
- 查询 `alarm_cloud` 表
- 支持租户配置和系统默认配置（tenant_id = NULL）
- 需要权限检查（查看权限）

**复杂度**：中等
- 需要权限检查
- 需要数据转换（JSONB 字段）

**是否需要 Service**：✅ **需要**

---

#### PUT /admin/api/v1/alarm-cloud
**功能**：更新报警策略配置

**需求**：
- 更新 `alarm_cloud` 表
- 需要权限检查（编辑权限）
- 需要业务规则验证（数据格式验证）
- 需要数据转换（JSONB 字段）

**复杂度**：高
- 需要权限检查
- 需要业务规则验证
- 需要数据转换

**是否需要 Service**：✅ **需要**

---

### 2. Alarm Events API

#### GET /admin/api/v1/alarm-events
**功能**：获取报警事件列表

**需求**：
- 复杂查询（多条件过滤）
  - 状态过滤（active/resolved）
  - 时间范围过滤
  - 住户搜索（通过 device_id → beds → residents）
  - 位置搜索（branch_tag, unit_name）
  - 设备搜索（device_name）
  - 事件类型过滤（多选）
  - 分类过滤（多选）
  - 报警级别过滤（多选）
- 分页支持
- **权限过滤**（重要）：
  - Resident：只能看到自己相关的报警
  - Family：只能看到家庭成员相关的报警
  - Staff（Nurse, Caregiver）：根据卡片权限过滤
  - Admin/Manager/IT：看到租户内所有报警
- JOIN 多个表：
  - `alarm_events` → `devices` → `beds` → `residents`
  - `alarm_events` → `devices` → `rooms` → `units`
  - 需要返回完整信息（住户信息、地址信息等）

**复杂度**：极高
- 需要权限检查
- 需要复杂的权限过滤逻辑
- 需要复杂的查询（多表 JOIN）
- 需要数据转换（返回前端需要的格式）

**是否需要 Service**：✅ **需要**

---

#### PUT /admin/api/v1/alarm-events/:id/handle
**功能**：处理报警事件

**需求**：
- 更新报警状态（active → acknowledged/resolved）
- **权限检查**（重要）：
  - Facility 类型卡片：只有 Nurse 或 Caregiver 可以处理
  - Home 类型卡片：所有角色都可以处理
  - 需要通过 `event_id` → `device_id` → `card` → `unit_type` 查询
- 业务规则验证：
  - 只能处理 active 状态的报警
  - 验证 handle_type 值（verified/false_alarm/test）
- 更新处理信息（handler_id, hand_time, operation, notes）

**复杂度**：高
- 需要权限检查（复杂的权限规则）
- 需要业务规则验证
- 需要状态管理
- 需要跨表查询（event → device → card）

**是否需要 Service**：✅ **需要**

---

## 📊 修正后的决策矩阵

| Repository | API 端点 | 功能 | 权限检查 | 业务规则 | 复杂查询 | **是否需要 Service** | **原因** |
|-----------|---------|------|---------|---------|---------|---------------------|---------|
| **AlarmCloudRepository** | GET /admin/api/v1/alarm-cloud | 获取配置 | ✅ 需要 | ❌ 不需要 | ❌ 简单 | ✅ **需要** | 需要权限检查、数据转换 |
| **AlarmCloudRepository** | PUT /admin/api/v1/alarm-cloud | 更新配置 | ✅ 需要 | ✅ 需要 | ❌ 简单 | ✅ **需要** | 需要权限检查、业务规则验证、数据转换 |
| **AlarmEventsRepository** | GET /admin/api/v1/alarm-events | 获取列表 | ✅ 需要 | ❌ 不需要 | ✅ **复杂** | ✅ **需要** | 需要权限过滤、复杂查询（多表JOIN）、数据转换 |
| **AlarmEventsRepository** | PUT /admin/api/v1/alarm-events/:id/handle | 处理报警 | ✅ 需要 | ✅ 需要 | ✅ **复杂** | ✅ **需要** | 需要权限检查（Facility vs Home）、业务规则验证、状态管理、跨表查询 |
| AlarmDeviceRepository | - | 内部使用 | ❌ 不需要 | ❌ 不需要 | ❌ 简单 | ❌ **不需要** | 后台服务使用，不需要 Service |
| CardRepository | - | 内部使用 | ❌ 不需要 | ❌ 不需要 | ❌ 简单 | ❌ **不需要** | 后台服务使用，不需要 Service |
| DeviceRepository | - | 内部使用 | ❌ 不需要 | ❌ 不需要 | ❌ 简单 | ❌ **不需要** | 后台服务使用，不需要 Service |
| RoomRepository | - | 内部使用 | ❌ 不需要 | ❌ 不需要 | ❌ 简单 | ❌ **不需要** | 后台服务使用，不需要 Service |

---

## 🎯 修正后的结论

### HTTP API 场景（需要 Service）

1. **AlarmCloudService** ✅ **需要**
   - 原因：需要权限检查、业务规则验证、数据转换
   - 方法：
     - `GetAlarmCloudConfig(ctx, tenantID, userID, userRole)`
     - `UpdateAlarmCloudConfig(ctx, tenantID, userID, userRole, config)`

2. **AlarmEventService** ✅ **需要**
   - 原因：需要权限检查、权限过滤、复杂查询、状态管理、业务规则验证
   - 方法：
     - `ListAlarmEvents(ctx, tenantID, userID, userRole, filters, page, size)` - 需要权限过滤
     - `HandleAlarmEvent(ctx, tenantID, userID, userRole, eventID, params)` - 需要权限检查（Facility vs Home）

### 后台服务场景（不需要 Service）

- AlarmDeviceRepository - 直接使用
- CardRepository - 直接使用
- DeviceRepository - 直接使用
- RoomRepository - 直接使用

---

## 🏗️ Service 层设计

### 1. AlarmCloudService

**职责**：
1. 权限检查（查看/编辑权限）
2. 业务规则验证（数据格式验证）
3. 数据转换（JSONB 字段 ↔ 领域模型）

**接口**：
```go
type AlarmCloudService struct {
    alarmCloudRepo *repository.AlarmCloudRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// GetAlarmCloudConfig 获取报警策略配置
func (s *AlarmCloudService) GetAlarmCloudConfig(
    ctx context.Context,
    tenantID, userID, userRole string,
) (*models.AlarmCloudConfig, error) {
    // 1. 权限检查
    if !s.permissionChecker.CanViewAlarmConfig(ctx, tenantID, userID, userRole) {
        return nil, ErrPermissionDenied
    }
    
    // 2. 调用 Repository
    return s.alarmCloudRepo.GetAlarmCloudConfig(ctx, tenantID)
}

// UpdateAlarmCloudConfig 更新报警策略配置
func (s *AlarmCloudService) UpdateAlarmCloudConfig(
    ctx context.Context,
    tenantID, userID, userRole string,
    config *models.AlarmCloudConfig,
) error {
    // 1. 权限检查
    if !s.permissionChecker.CanEditAlarmConfig(ctx, tenantID, userID, userRole) {
        return ErrPermissionDenied
    }
    
    // 2. 业务规则验证
    if err := s.validateAlarmCloudConfig(config); err != nil {
        return err
    }
    
    // 3. 调用 Repository
    return s.alarmCloudRepo.UpdateAlarmCloudConfig(ctx, tenantID, config)
}
```

---

### 2. AlarmEventService

**职责**：
1. 权限检查（查看/处理权限）
2. 权限过滤（根据用户角色过滤可查看的报警）
3. 业务规则验证（处理报警的规则）
4. 数据转换（返回前端需要的格式，包含 JOIN 的数据）

**接口**：
```go
type AlarmEventService struct {
    alarmEventsRepo *repository.AlarmEventsRepository
    cardRepo        *repository.CardRepository
    deviceRepo      *repository.DeviceRepository
    permissionChecker *PermissionChecker
    logger *zap.Logger
}

// ListAlarmEvents 获取报警事件列表（需要权限过滤）
func (s *AlarmEventService) ListAlarmEvents(
    ctx context.Context,
    tenantID, userID, userRole string,
    filters repository.AlarmEventFilters,
    page, size int,
) ([]*models.AlarmEvent, int, error) {
    // 1. 权限检查
    if !s.permissionChecker.CanViewAlarmEvents(ctx, tenantID, userID, userRole) {
        return nil, 0, ErrPermissionDenied
    }
    
    // 2. 权限过滤（根据用户角色添加过滤条件）
    filters = s.applyPermissionFilters(ctx, tenantID, userID, userRole, filters)
    
    // 3. 调用 Repository
    events, total, err := s.alarmEventsRepo.ListAlarmEvents(ctx, tenantID, filters, page, size)
    if err != nil {
        return nil, 0, err
    }
    
    // 4. 数据转换（添加 JOIN 的数据：住户信息、地址信息等）
    return s.enrichAlarmEvents(ctx, events), total, nil
}

// HandleAlarmEvent 处理报警事件（需要权限检查）
func (s *AlarmEventService) HandleAlarmEvent(
    ctx context.Context,
    tenantID, userID, userRole, eventID string,
    params HandleAlarmEventParams,
) error {
    // 1. 获取报警事件
    event, err := s.alarmEventsRepo.GetAlarmEvent(ctx, tenantID, eventID)
    if err != nil {
        return err
    }
    
    // 2. 权限检查（Facility vs Home）
    if !s.canHandleAlarm(ctx, tenantID, userID, userRole, event) {
        return ErrPermissionDenied
    }
    
    // 3. 业务规则验证
    if err := s.validateHandleParams(event, params); err != nil {
        return err
    }
    
    // 4. 调用 Repository 更新状态
    return s.alarmEventsRepo.UpdateAlarmEventOperation(
        ctx, tenantID, eventID, params.HandleType, userID, params.Remarks,
    )
}

// canHandleAlarm 检查用户是否可以处理报警
func (s *AlarmEventService) canHandleAlarm(
    ctx context.Context,
    tenantID, userID, userRole string,
    event *models.AlarmEvent,
) bool {
    // 1. 通过 device_id 获取卡片信息
    device, err := s.deviceRepo.GetDeviceBindingInfo(ctx, tenantID, event.DeviceID)
    if err != nil {
        return false
    }
    
    // 2. 通过 device 获取卡片（需要查询 cards 表，找到包含该 device 的卡片）
    card, err := s.cardRepo.GetCardByDeviceID(ctx, tenantID, event.DeviceID)
    if err != nil {
        return false
    }
    
    // 3. 权限规则
    if card.UnitType == "Facility" {
        // Facility：只有 Nurse 或 Caregiver 可以处理
        return userRole == "Nurse" || userRole == "Caregiver"
    } else if card.UnitType == "Home" {
        // Home：所有角色都可以处理
        return true
    }
    
    return true
}
```

---

## 📋 总结

### 修正后的决策

| Repository | HTTP API | 是否需要 Service | 原因 |
|-----------|---------|----------------|------|
| **AlarmCloudRepository** | ✅ 有 | ✅ **需要** | 需要权限检查、业务规则验证 |
| **AlarmEventsRepository** | ✅ 有 | ✅ **需要** | 需要权限检查、权限过滤、复杂查询、状态管理 |
| AlarmDeviceRepository | ❌ 无 | ❌ **不需要** | 后台服务使用 |
| CardRepository | ❌ 无 | ❌ **不需要** | 后台服务使用 |
| DeviceRepository | ❌ 无 | ❌ **不需要** | 后台服务使用 |
| RoomRepository | ❌ 无 | ❌ **不需要** | 后台服务使用 |

### 关键发现

1. **AlarmCloudRepository 也需要 Service**
   - 之前错误地认为它只是只读操作，不需要 Service
   - 实际上需要权限检查（canEdit）和业务规则验证

2. **AlarmEventService 需要更复杂的功能**
   - 权限过滤（根据用户角色过滤可查看的报警）
   - 跨表查询（event → device → card → unit_type）
   - 数据转换（添加 JOIN 的数据）

---

## 🚀 下一步

1. ✅ 重新设计 AlarmCloudService
2. ✅ 完善 AlarmEventService（添加权限过滤、跨表查询）
3. ⏳ 实现 PermissionChecker
4. ⏳ 实现数据转换逻辑

