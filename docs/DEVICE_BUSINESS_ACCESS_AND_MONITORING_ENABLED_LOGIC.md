# 设备 business_access 和 monitoring_enabled 字段处理逻辑说明

## 一、字段定义与业务含义

### 1.1 business_access（业务访问权限）

**数据库定义**（`owlRD/db/17_devices.sql`）：
```sql
business_access VARCHAR(20) NOT NULL DEFAULT 'pending' 
  CHECK (business_access IN ('pending', 'approved', 'rejected'))
```

**业务含义**：
- `pending`：等待审批（新设备默认状态）
- `approved`：已批准（租户管理员批准后，设备可以访问业务系统）
- `rejected`：已拒绝（租户管理员拒绝设备访问）

**权限控制逻辑**：
设备访问业务系统需要同时满足：
1. `device_store.allow_access = TRUE`（系统级，由系统管理员控制）
2. `devices.business_access = 'approved'`（租户级，由租户管理员批准）

### 1.2 monitoring_enabled（监控启用状态）

**数据库定义**（`owlRD/db/17_devices.sql`）：
```sql
monitoring_enabled BOOLEAN NOT NULL DEFAULT FALSE
```

**业务含义**：
- `TRUE`：监控已激活（设备正常监控状态）
- `FALSE`：休眠模式/等待分配（设备已配置但监控未启用）

## 二、数据流转

### 2.1 设备创建时的默认值

#### 方式一：系统管理员从 device_store 创建设备（`postgres_devices.go:CreateDevice`）

```go
// 默认值设置
status := device.Status
if status == "" {
    status = "offline"
}
businessAccess := device.BusinessAccess
if businessAccess == "" {
    businessAccess = "pending"  // 默认值：等待审批
}
// monitoring_enabled 使用传入的值，默认为 false
```

**创建时的默认状态**：
- `status`: `'offline'`
- `business_access`: `'pending'`
- `monitoring_enabled`: `FALSE`

#### 方式二：设备首次连接时自动创建（`wisefido-radar/wisefido-sleepace`）

```sql
INSERT INTO devices (
    tenant_id, device_id, device_name,
    serial_number, uid,
    status, business_access, monitoring_enabled
) VALUES ($1, $2, $3, $4, $5, 'online', 'pending', FALSE)
```

**创建时的默认状态**：
- `status`: `'online'`（设备已连接）
- `business_access`: `'pending'`（等待租户管理员审批）
- `monitoring_enabled`: `FALSE`（监控未启用）

### 2.2 设备更新流程

#### 前端 → Handler → Service → Repository

**完整流程**：

```
前端 (DeviceList.vue)
  ↓ PUT /admin/api/v1/devices/:id
  ↓ { business_access: 'approved' }
Handler (device_handler.go)
  ↓ payloadToDevice() 转换
  ↓ 检查 payload 中是否包含字段
  ↓ UpdateDeviceRequest{ UpdateBusinessAccess: true }
Service (device_service.go)
  ↓ 参数验证
  ↓ 获取旧设备信息（用于监控状态变化检测）
  ↓ UpdateDeviceWithFlags()
Repository (postgres_devices.go)
  ↓ 检查 updateBusinessAccess 标志
  ↓ 如果为 true 且值不为空，则更新
  ↓ UPDATE devices SET business_access = 'approved' WHERE ...
```

## 三、详细处理逻辑

### 3.1 business_access 处理逻辑

#### 3.1.1 前端处理（`DeviceList.vue`）

```typescript
// 用户在下拉框中选择 business_access 值
const handleBusinessAccessChange = async (record: Device, value: 'pending' | 'approved' | 'rejected') => {
  // 如果值没有变化，不执行更新
  if (record.business_access === value) {
    return
  }
  
  // 调用 API 更新
  await updateDeviceApi(record.device_id, {
    business_access: value,  // 只传递 business_access 字段
  })
}
```

**特点**：
- 只传递需要更新的字段（`business_access`）
- 不传递其他字段（如 `monitoring_enabled`）

#### 3.1.2 Handler 层处理（`device_handler.go` + `admin_units_devices_impl.go`）

**步骤 1：payload 转换**（`payloadToDevice`）

```go
// Handle business_access: 检查字段是否存在
if val, exists := payload["business_access"]; exists {
    if v, ok := val.(string); ok {
        device.BusinessAccess = v  // 允许空字符串（用于验证）
    }
}
```

**关键点**：
- 使用 `exists` 检查字段是否在 payload 中
- 区分"字段不存在"和"字段值为空字符串"
- 如果字段存在，则设置到 `device.BusinessAccess`

**步骤 2：检查字段标志**

```go
// 检查 payload 中是否包含 business_access
_, hasBusinessAccess := payload["business_access"]

req := service.UpdateDeviceRequest{
    // ...
    UpdateBusinessAccess: hasBusinessAccess,  // 传递标志
}
```

**关键点**：
- 使用 `hasBusinessAccess` 标志标记字段是否在 payload 中
- 这个标志会传递到 Repository 层，决定是否更新该字段

#### 3.1.3 Service 层处理（`device_service.go`）

```go
// 调用 Repository（传递更新标志）
err := s.devicesRepo.UpdateDeviceWithFlags(
    ctx, req.TenantID, req.DeviceID, req.Device,
    req.UpdateBoundRoomID, req.UpdateBoundBedID,
    req.UpdateBusinessAccess,      // 传递 business_access 更新标志
    req.UpdateMonitoringEnabled,
)
```

**关键点**：
- Service 层只是传递标志，不做业务逻辑判断
- 实际的更新逻辑在 Repository 层

#### 3.1.4 Repository 层处理（`postgres_devices.go:UpdateDeviceWithFlags`）

```go
// Handle business_access: only update if flag is set
if updateBusinessAccess && device.BusinessAccess != "" {
    add("business_access", device.BusinessAccess)
}
```

**关键点**：
- **双重检查**：
  1. `updateBusinessAccess`：字段是否在 payload 中
  2. `device.BusinessAccess != ""`：值是否不为空
- 只有两个条件都满足，才会添加到 UPDATE 语句中
- 如果字段不在 payload 中，即使 `device.BusinessAccess` 有值也不会更新

**为什么需要双重检查？**

1. **防止误更新**：如果 payload 中没有 `business_access` 字段，说明前端不想更新它
2. **防止空值更新**：如果值为空字符串，不应该更新（数据库约束要求 NOT NULL）

### 3.2 monitoring_enabled 处理逻辑

#### 3.2.1 前端处理（`DeviceList.vue`）

```typescript
// 用户切换监控开关
const handleMonitoringChange = async (record: Device, checked: boolean) => {
  await updateDeviceApi(record.device_id, {
    monitoring_enabled: checked,  // 只传递 monitoring_enabled 字段
  })
}
```

**特点**：
- 只传递需要更新的字段（`monitoring_enabled`）
- 不传递其他字段（如 `business_access`）

#### 3.2.2 Handler 层处理

**步骤 1：payload 转换**

```go
// Handle monitoring_enabled: 检查字段是否存在
if val, exists := payload["monitoring_enabled"]; exists {
    if v, ok := val.(bool); ok {
        device.MonitoringEnabled = v
    }
}
```

**关键点**：
- 使用 `exists` 检查字段是否在 payload 中
- 如果字段不存在，`device.MonitoringEnabled` 保持默认值 `false`
- 如果字段存在，则设置到 `device.MonitoringEnabled`

**步骤 2：检查字段标志**

```go
// 检查 payload 中是否包含 monitoring_enabled
_, hasMonitoringEnabled := payload["monitoring_enabled"]

req := service.UpdateDeviceRequest{
    // ...
    UpdateMonitoringEnabled: hasMonitoringEnabled,  // 传递标志
}
```

#### 3.2.3 Service 层处理

```go
// 获取旧设备信息（用于比较 monitoring_enabled 是否变化）
oldDevice, _ := s.devicesRepo.GetDevice(ctx, req.TenantID, req.DeviceID)
if oldDevice != nil && req.Device.MonitoringEnabled != oldDevice.MonitoringEnabled {
    monitoringEnabledChanged = true
}

// 如果 monitoring_enabled 变化，触发相关服务更新
if monitoringEnabledChanged {
    // 清除 wisefido-iot-timeseries 的位置信息缓存
    // 通知 wisefido-card-aggregator 更新卡片数据
}
```

**关键点**：
- Service 层会检测 `monitoring_enabled` 是否变化
- 如果变化，会触发相关服务的更新（如清除缓存、更新卡片数据）

#### 3.2.4 Repository 层处理

```go
// Handle monitoring_enabled: only update if flag is set
if updateMonitoringEnabled {
    add("monitoring_enabled", device.MonitoringEnabled)
}
```

**关键点**：
- **只检查标志**：只要 `updateMonitoringEnabled` 为 true，就更新
- **不需要检查值**：因为 `monitoring_enabled` 是布尔类型，不存在"空值"的概念
- 如果字段不在 payload 中，即使 `device.MonitoringEnabled` 有值也不会更新

**为什么只需要检查标志？**

1. **布尔类型特性**：`monitoring_enabled` 是 `bool` 类型，只有 `true` 或 `false`，不存在"空值"
2. **防止误更新**：如果 payload 中没有 `monitoring_enabled` 字段，说明前端不想更新它

## 四、关键设计决策

### 4.1 为什么需要 `UpdateBusinessAccess` 和 `UpdateMonitoringEnabled` 标志？

**问题**：如果不使用标志，会出现什么问题？

**场景 1：只更新 business_access**

```json
// 前端请求
PUT /admin/api/v1/devices/xxx
{ "business_access": "approved" }
```

**没有标志的情况**：
```go
// Repository 层
if device.BusinessAccess != "" {
    add("business_access", device.BusinessAccess)  // ✓ 会更新
}
add("monitoring_enabled", device.MonitoringEnabled)  // ✗ 问题：也会更新！
```

**问题**：
- `device.MonitoringEnabled` 的默认值是 `false`
- 如果 payload 中没有 `monitoring_enabled` 字段，`device.MonitoringEnabled` 保持为 `false`
- 会导致 `monitoring_enabled` 被错误地更新为 `false`，覆盖现有的 `true` 值

**有标志的情况**：
```go
// Repository 层
if updateBusinessAccess && device.BusinessAccess != "" {
    add("business_access", device.BusinessAccess)  // ✓ 会更新
}
if updateMonitoringEnabled {
    add("monitoring_enabled", device.MonitoringEnabled)  // ✓ 不会更新（标志为 false）
}
```

**结果**：只有 `business_access` 被更新，`monitoring_enabled` 保持不变

### 4.2 business_access 为什么需要双重检查？

**检查 1：`updateBusinessAccess`**
- 确保字段在 payload 中
- 防止误更新

**检查 2：`device.BusinessAccess != ""`**
- 确保值不为空字符串
- 数据库约束要求 `business_access NOT NULL`

**为什么不能只检查一个？**

如果只检查 `updateBusinessAccess`：
```go
if updateBusinessAccess {
    add("business_access", device.BusinessAccess)  // 如果值为空字符串，会违反 NOT NULL 约束
}
```

如果只检查 `device.BusinessAccess != ""`：
```go
if device.BusinessAccess != "" {
    add("business_access", device.BusinessAccess)  // 如果字段不在 payload 中，也会更新（误更新）
}
```

### 4.3 monitoring_enabled 为什么只需要检查标志？

**原因**：
1. **布尔类型**：`monitoring_enabled` 是 `bool` 类型，只有 `true` 或 `false`，不存在"空值"
2. **默认值安全**：如果字段不在 payload 中，`device.MonitoringEnabled` 保持默认值 `false`，但不会更新（因为标志为 `false`）

## 五、状态组合与业务场景

### 5.1 设备生命周期状态

| 阶段 | status | business_access | monitoring_enabled | 说明 |
|------|--------|-----------------|-------------------|------|
| 1. 设备出库 | offline | pending | FALSE | 新设备已添加，等待租户管理员审批 |
| 2. 设备连接 | online | pending | FALSE | 设备已连接，但未审批 |
| 3. 审批通过 | online | approved | FALSE | 设备已批准，等待分配（休眠模式） |
| 4. 分配并激活 | online | approved | TRUE | 设备已分配且监控已激活（正常监控状态） |
| 5. 审批拒绝 | offline | rejected | FALSE | 设备已拒绝或已移除 |

### 5.2 典型操作流程

**场景 1：租户管理员审批设备**

```
1. 设备连接 → status='online', business_access='pending', monitoring_enabled=FALSE
2. 租户管理员在前端选择 business_access='approved'
3. 前端发送：PUT /admin/api/v1/devices/xxx { "business_access": "approved" }
4. 后端更新：business_access='approved'
5. 结果：status='online', business_access='approved', monitoring_enabled=FALSE
```

**场景 2：激活设备监控**

```
1. 设备已审批 → status='online', business_access='approved', monitoring_enabled=FALSE
2. 管理员在前端开启监控开关
3. 前端发送：PUT /admin/api/v1/devices/xxx { "monitoring_enabled": true }
4. 后端更新：monitoring_enabled=TRUE
5. 后端触发：清除缓存、更新卡片数据
6. 结果：status='online', business_access='approved', monitoring_enabled=TRUE
```

**场景 3：同时更新两个字段**

```
1. 前端发送：PUT /admin/api/v1/devices/xxx 
   { 
     "business_access": "approved",
     "monitoring_enabled": true 
   }
2. 后端更新：两个字段都会被更新
3. 结果：status='online', business_access='approved', monitoring_enabled=TRUE
```

## 六、错误处理

### 6.1 常见错误场景

**错误 1：business_access 值为空字符串**

```go
// payload: { "business_access": "" }
if updateBusinessAccess && device.BusinessAccess != "" {  // false，不会更新
    add("business_access", device.BusinessAccess)
}
```

**结果**：不会更新，避免违反 NOT NULL 约束

**错误 2：business_access 值不在允许范围内**

```sql
-- 数据库约束
CHECK (business_access IN ('pending', 'approved', 'rejected'))
```

**结果**：数据库会拒绝更新，返回错误

**错误 3：字段不在 payload 中但被误更新**

```go
// payload: { "device_name": "New Name" }  // 没有 business_access
// 没有标志的情况
if device.BusinessAccess != "" {  // 如果 device.BusinessAccess 有旧值，会误更新
    add("business_access", device.BusinessAccess)
}
```

**结果**：使用标志后，不会误更新

## 七、总结

### 7.1 business_access 处理要点

1. **双重检查**：检查标志 + 检查值不为空
2. **字段存在性检查**：使用 `exists` 区分"字段不存在"和"字段值为空"
3. **默认值**：新设备默认为 `'pending'`

### 7.2 monitoring_enabled 处理要点

1. **标志检查**：只检查标志，不需要检查值（布尔类型）
2. **变化检测**：Service 层会检测变化并触发相关服务更新
3. **默认值**：新设备默认为 `FALSE`

### 7.3 设计原则

1. **部分更新**：只更新 payload 中存在的字段
2. **防止误更新**：使用标志确保不会更新未传递的字段
3. **数据安全**：检查约束，避免违反数据库约束
