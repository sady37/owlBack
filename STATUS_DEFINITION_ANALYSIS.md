# 设备在线/离线状态定义统一性分析报告

## 问题概述
代码中关于设备状态的定义存在**三个不同层面的状态表示**，且存在**不一致**的地方：
1. 内存中的设备状态（DeviceSubscription.Status）
2. Redis Stream 中发送的设备状态（PublishDeviceStatus）
3. 文档中定义的设备状态（Reside_stream_stand.md）

---

## 一、三个层面的状态定义

### 1️⃣ 内存层（Memory）- DeviceSubscription.Status
**位置**：`wisefido-qinglan/internal/subscriber/device_subscription_manager.go`

**状态值**（字符串）：
- `"online"` - 设备在线
- `"offline"` - 设备离线（90秒无消息后）
- `"unsubscribed"` - 已取消订阅（180秒无消息后）

**使用位置**：
```go
sub.Status = "online"      // 初始状态
sub.Status = "offline"     // markDeviceOffline() 中标记
sub.Status = "unsubscribed" // unsubscribeDevice() 中标记
```

---

### 2️⃣ Redis Stream 层（IoT Stream）- PublishDeviceStatus
**位置**：
- `wisefido-qinglan/internal/consumer/stream_publisher.go` - 发送方
- `owl-common/redis/message_types.go` - 消息定义

**Stream 名称**：`iot:deviceStatus:stream`

**消息格式**（Map[string]int）：
```json
{
  "device_id": "uuid",
  "device_type": "Radar",
  "card_id": "uuid",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "status",
  "category": "deviceStatus",
  "data_value": [
    {
      "category": "deviceStatus",
      "statuses": {
        "online": 1,          // ✅ 1=在线, 0=离线
        "signal_poor": 0,
        "angle_abnormal": 0
      }
    }
  ]
}
```

**状态值**（const.StatusField*）：
```go
const (
    StatusFieldOnline        = "online"         // 0=离线, 1=在线 ✓ 定义正确
    StatusFieldAngleAbnormal = "angle_abnormal" // 1=异常, 0=正常
    StatusFieldSignalPoor    = "signal_poor"    // 1=信号差, 0=正常
    StatusFieldDetached      = "detached"       // 1=脱落, 0=正常
    StatusFieldDeviceFailure = "device_failure" // 1=故障, 0=正常
)
```

**发送示例**：
```go
// 设备上线时
map[string]int{constDevice.StatusFieldOnline: 1}

// 设备离线时
map[string]int{constDevice.StatusFieldOnline: 0}
```

---

### 3️⃣ 文档层（Documentation）- Reside_stream_stand.md
**位置**：`docs/Reside_stream_stand.md` - 第 395-420 行

**定义的消息格式**（注意：这是 Event type=5，不是当前使用的 Status 类型）：
```json
{
  "topic_type": "event",
  "category": "isOnline",
  "data_value": [
    {
      "category": "isOnline",
      "device_status": "offline",     // ❌ 字符串格式，不是数字
      "device_uid": "E598A2ACD523"
    }
  ]
}
```

**字段说明**：
- `device_status`: "online" 或 "offline"（**字符串**）
- `device_uid`: 设备 UID

---

## 二、发现的不一致问题

### ❌ 问题 1：message_types.go 注释错误
**文件**：`owl-common/redis/message_types.go` 第 52-59 行

**当前注释**：
```go
// statuses: 设备状态 JSON（map[string]int，其中 0=正常 1=异常/故障）
// 支持字段（使用 const.StatusField* 常量）：
//   - offline: 0=在线, 1=离线        // ❌ 错误：key应为"online"不是"offline"
//   - angle_abnormal: 0=正常, 1=倾角异常
//   - signal_poor: 0=正常, 1=信号差
```

**正确注释应为**：
```go
// statuses: 设备状态 JSON（map[string]int，其中 0=异常/离线, 1=正常/在线）
// 支持字段（使用 const.StatusField* 常量）：
//   - online: 0=离线, 1=在线        // ✓ 正确
//   - angle_abnormal: 0=正常, 1=倾角异常
//   - signal_poor: 0=正常, 1=信号差
```

---

### ❌ 问题 2：文档格式与实现不符（Event vs Status）
**问题**：Reside_stream_stand.md 中定义的设备在线状态事件（type=5）与实际代码使用的格式不同

| 方面 | Reside_stream_stand.md | 实际代码 |
|------|----------------------|-----------|
| Stream | 不明确 | `iot:deviceStatus:stream` |
| category | `"isOnline"` | `"deviceStatus"` |
| 状态值格式 | 字符串 `"online"/"offline"` | Map[string]int `{"online":1/0}` |
| data_value.category | `"isOnline"` | `"deviceStatus"` |
| 设备标识 | `device_uid` | 无（在顶层字段） |

**影响**：消费端（如 wisefido-cardagg）可能看到两种不同的消息格式

---

### ⚠️ 问题 3：PublishOnlineForConnectedDevices 实现
**文件**：`wisefido-qinglan/internal/subscriber/device_subscription_manager.go` 第 1067-1092 行

```go
func (m *DeviceSubscriptionManager) PublishOnlineForConnectedDevices(ctx context.Context, deviceUIDs []string) {
    published := 0
    for _, deviceUID := range deviceUIDs {
        m.mu.RLock()
        sub, ok := m.subscriptionsByUID[deviceUID]
        m.mu.RUnlock()
        if !ok {
            continue
        }
        sub.mu.RLock()
        status, deviceID, deviceType, tenantID := sub.Status, sub.DeviceID, sub.DeviceType, sub.TenantID
        sub.mu.RUnlock()
        if status != "online" {
            continue
        }
        // ✅ 正确：发送 online:1
        go m.streamPublisher.PublishDeviceStatus(context.Background(), deviceID, deviceType, tenantID, deviceUID, map[string]int{
            constDevice.StatusFieldOnline: 1,
        })
        published++
    }
    m.logger.Info("Published online notification for connected devices after startup",
        zap.Int("published", published),
        zap.Int("requested", len(deviceUIDs)),
    )
}
```

**状态**：✅ 这个实现是**正确的**

---

## 三、消费端处理方式

### event_alarm_service.go（wisefido-cardagg）
**位置**：`wisefido-cardagg/internal/service/event_alarm_service.go` 第 731-780 行

**处理方式**：
```go
func (s *EventAlarmService) convertToDeviceStatus(msg *redis.IoTStreamMessage) *card.DeviceStatus {
    if len(msg.DataValue) == 0 {
        return nil
    }
    
    // 从data_value[0]提取statuses
    dataItem, ok := msg.DataValue[0].(map[string]interface{})
    if !ok {
        return nil
    }
    
    category, _ := dataItem["category"].(string)
    if category != "deviceStatus" {    // ✅ 检查的是"deviceStatus"
        return nil
    }
    
    // 提取statuses map[string]interface{}，转换为map[string]int
    statusesRaw, ok := dataItem["statuses"].(map[string]interface{})
    if !ok {
        return nil
    }
    
    // 转换为map[string]int
    statuses := make(map[string]int)
    for k, v := range statusesRaw {
        switch val := v.(type) {
        case float64:
            statuses[k] = int(val)
        case int:
            statuses[k] = val
        case int64:
            statuses[k] = int(val)
        }
    }
    
    return &card.DeviceStatus{
        DeviceID:   msg.DeviceID,
        DeviceType: msg.DeviceType,
        Timestamp:  msg.Timestamp,
        Statuses:   statuses,
    }
}
```

**card.DeviceStatus 结构体**：
```go
type DeviceStatus struct {
    DeviceID   string         `json:"device_id"`
    DeviceType string         `json:"device_type"`
    Timestamp  int64          `json:"timestamp"`
    Statuses   map[string]int `json:"statuses,omitempty"`  // 正确接收map[string]int
}
```

---

## 四、状态转换流程总结

```
内存状态          →  Redis发送          →  消费端接收
("online")        →  online:1           →  convertToDeviceStatus()
("offline")       →  online:0           →  card.DeviceStatus{
("unsubscribed")  →  不发送             →    Statuses:map[string]int
                                          }
```

---

## 五、建议修复方案

### 修复 1️⃣：修正 message_types.go 注释（低风险）
**文件**：`owl-common/redis/message_types.go`

**改动**：修正注释中的错误描述

```go
// BuildDeviceStatusMessage 构建设备状态消息（发送到 iot:deviceStatus:stream）
// statuses: 设备状态 JSON（map[string]int，其中 0=异常/离线, 1=正常/在线）
// 支持字段（使用 const.StatusField* 常量）：
//   - online: 0=离线, 1=在线          // ✓ 修正key从"offline"改为"online"
//   - angle_abnormal: 0=正常, 1=倾角异常
//   - signal_poor: 0=正常, 1=信号差
//   - detached: 0=正常, 1=传感器脱落
//   - device_failure: 0=正常, 1=设备故障
```

### 修复 2️⃣：更新 Reside_stream_stand.md 文档（中风险）
**现状**：文档定义的格式（type=5，category=isOnline）与实际代码不符

**选项A**：更新文档以匹配当前实现
```markdown
### 5.5 type=5（设备状态）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "status",
  "category": "deviceStatus",
  "data_value": [
    {
      "category": "deviceStatus",
      "statuses": {
        "online": 1
      }
    }
  ]
}
```
```

**选项B**：保持向后兼容，支持两种格式
- 发送时：统一使用 BuildDeviceStatusMessage（当前方式）
- 消费时：同时支持 event type=5 (isOnline) 和 status type (deviceStatus)

---

## 六、现状总体评估

| 层面 | 统一性 | 说明 |
|------|--------|------|
| **内存↔Redis** | ✅ 统一 | 内存"online"→Redis online:1，映射正确 |
| **Redis发送** | ✅ 统一 | 所有调用都使用 constDevice.StatusFieldOnline |
| **Redis消费** | ✅ 统一 | event_alarm_service 正确处理 deviceStatus |
| **代码↔文档** | ❌ **不统一** | message_types 注释错误，Reside_stream_stand 格式不匹配 |

---

## 七、结论

### ✅ 代码实现是一致的
- 内存状态 ("online"/"offline"/"unsubscribed") 正确映射到 Redis (online: 1/0)
- PublishDeviceStatus 和 PublishOnlineForConnectedDevices 使用方式一致
- 消费端（event_alarm_service）正确解析消息

### ❌ 文档需要修正
1. **message_types.go** 的注释有错误（"offline" 应为 "online"）
2. **Reside_stream_stand.md** 的设备在线状态事件定义与实际实现不符
   - 文档说是 event type=5，category=isOnline
   - 实际是 status type，category=deviceStatus

### 🔧 建议
1. 修正 message_types.go 的注释
2. 更新 Reside_stream_stand.md 以反映实际实现
3. 考虑在代码中添加转换函数，支持两种格式的向后兼容性
