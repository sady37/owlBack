# AlarmEvent Service 层类型转换说明

## 架构原则

**Repository 层**：统一使用数据库原生类型（`json.RawMessage`），直接与数据库交互，不做业务逻辑转换。

**Service 层**：负责类型转换，将 Repository 层的数据库类型转换为业务需要的类型。

---

## 类型转换场景

### 1. wisefido-alarm Service 层（写入场景）

#### 场景：创建报警事件

**输入**：业务对象（结构体）
```go
// 输入：业务对象
triggerData := &models.TriggerData{
    EventType: "Fall",
    Source: "Radar",
    HeartRate: intPtr(72),
    // ...
}
metadata := map[string]interface{}{
    "trigger_source": "cloud",
    "card_id": "card-789",
}
```

**转换过程**（在 `AlarmEventBuilder.BuildAlarmEvent`）：
```go
// 1. 序列化 triggerData：结构体 → json.RawMessage
triggerDataJSON, err := json.Marshal(triggerData)
// triggerDataJSON 是 []byte，可以直接赋值给 json.RawMessage

// 2. 序列化 metadata：map → json.RawMessage
metadataBytes, err := json.Marshal(metadata)
metadataJSON := json.RawMessage(metadataBytes)

// 3. 构建 AlarmEvent（包含 json.RawMessage 字段）
event := &models.AlarmEvent{
    TriggerData:   triggerDataJSON,  // json.RawMessage
    Metadata:      metadataJSON,     // json.RawMessage
    NotifiedUsers: json.RawMessage("[]"),
    // ...
}
```

**输出**：`*models.AlarmEvent`（包含 `json.RawMessage` 字段，可直接存储到数据库）

**转换位置**：`wisefido-alarm/internal/evaluator/alarm_event_builder.go:BuildAlarmEvent`

---

### 2. wisefido-data Service 层（读取场景）

#### 场景：查询报警事件列表

**输入**：`*domain.AlarmEvent`（从 Repository 获取，包含 `json.RawMessage` 字段）
```go
// Repository 返回
event := &domain.AlarmEvent{
    EventID:     "xxx",
    TriggerData: json.RawMessage(`{"heart_rate": 120, "event_type": "Fall"}`),
    NotifiedUsers: json.RawMessage(`[{"user_id": "123"}]`),
    Metadata:    json.RawMessage(`{"source": "cloud"}`),
    // ...
}
```

**转换过程**（在 `alarmEventService.convertAlarmEventToDTO`）：
```go
// 1. 反序列化 TriggerData：json.RawMessage → map[string]interface{}
if len(event.TriggerData) > 0 {
    var triggerData map[string]interface{}
    if err := json.Unmarshal(event.TriggerData, &triggerData); err == nil {
        dto.TriggerData = triggerData  // 用于前端展示
    }
}

// 2. 反序列化 NotifiedUsers：json.RawMessage → []interface{}
if len(event.NotifiedUsers) > 0 {
    var notifiedUsers []interface{}
    if err := json.Unmarshal(event.NotifiedUsers, &notifiedUsers); err == nil {
        dto.NotifiedUsers = notifiedUsers  // 用于前端展示
    }
}

// 3. 反序列化 Metadata：json.RawMessage → map[string]interface{}
if len(event.Metadata) > 0 {
    var metadata map[string]interface{}
    if err := json.Unmarshal(event.Metadata, &metadata); err == nil {
        dto.Metadata = metadata  // 用于前端展示
    }
}
```

**输出**：`AlarmEventDTO`（包含 `map[string]interface{}` 和 `[]interface{}`，用于 HTTP 响应）

**转换位置**：`wisefido-data/internal/service/alarm_event_service.go:convertAlarmEventToDTO`

---

### 3. wisefido-alarm Service 层（读取场景 - 如果需要）

#### 场景：读取报警事件用于业务逻辑

**输入**：`*models.AlarmEvent`（从 Repository 获取，包含 `json.RawMessage` 字段）

**转换过程**（如果需要解析 TriggerData）：
```go
// 如果需要解析 TriggerData 用于业务逻辑
var triggerData models.TriggerData
if err := json.Unmarshal(event.TriggerData, &triggerData); err == nil {
    // 使用 triggerData 进行业务逻辑处理
    if triggerData.HeartRate != nil && *triggerData.HeartRate > 100 {
        // 处理逻辑
    }
}
```

**注意**：当前 wisefido-alarm 的 Service 层主要关注创建报警事件，读取场景较少。如果需要读取并解析，转换逻辑应该在 Service 层。

---

## 类型转换总结

| 场景 | 输入类型 | 转换操作 | 输出类型 | 转换位置 |
|------|---------|---------|---------|---------|
| **wisefido-alarm 写入** | `*models.TriggerData`<br/>`map[string]interface{}` | `json.Marshal` | `json.RawMessage` | `AlarmEventBuilder.BuildAlarmEvent` |
| **wisefido-data 读取** | `json.RawMessage` | `json.Unmarshal` | `map[string]interface{}`<br/>`[]interface{}` | `alarmEventService.convertAlarmEventToDTO` |
| **wisefido-alarm 读取**（如需要） | `json.RawMessage` | `json.Unmarshal` | `*models.TriggerData` | Service 层（按需） |

---

## 关键原则

1. **Repository 层**：
   - ✅ 使用 `json.RawMessage`（数据库原生类型）
   - ✅ 不做业务逻辑转换
   - ✅ 直接与数据库交互

2. **Service 层**：
   - ✅ 负责类型转换（序列化/反序列化）
   - ✅ 将数据库类型转换为业务需要的类型
   - ✅ 处理业务逻辑

3. **Handler 层**：
   - ✅ 只负责 HTTP 请求/响应
   - ✅ 调用 Service 层
   - ✅ 不做类型转换（由 Service 层完成）

---

## 示例代码

### wisefido-alarm：写入时的转换

```go
// evaluator/alarm_event_builder.go
func (b *AlarmEventBuilder) BuildAlarmEvent(
    eventType string,
    category string,
    alarmLevel string,
    triggerData *models.TriggerData,  // 输入：业务对象
    metadata map[string]interface{},  // 输入：业务对象
) (*models.AlarmEvent, error) {
    // 转换：业务对象 → json.RawMessage
    triggerDataJSON, err := json.Marshal(triggerData)
    if err != nil {
        return nil, err
    }
    
    metadataJSON := json.RawMessage("{}")
    if metadata != nil {
        metadataBytes, err := json.Marshal(metadata)
        if err != nil {
            return nil, err
        }
        metadataJSON = json.RawMessage(metadataBytes)
    }
    
    // 输出：包含 json.RawMessage 的 AlarmEvent
    return &models.AlarmEvent{
        TriggerData:   triggerDataJSON,  // json.RawMessage
        Metadata:      metadataJSON,     // json.RawMessage
        NotifiedUsers: json.RawMessage("[]"),
        // ...
    }, nil
}
```

### wisefido-data：读取时的转换

```go
// service/alarm_event_service.go
func (s *alarmEventService) convertAlarmEventToDTO(
    ctx context.Context,
    tenantID string,
    event *domain.AlarmEvent,  // 输入：包含 json.RawMessage
) (*AlarmEventDTO, error) {
    dto := &AlarmEventDTO{
        EventID: event.EventID,
        // ...
    }
    
    // 转换：json.RawMessage → map[string]interface{}
    if len(event.TriggerData) > 0 {
        var triggerData map[string]interface{}
        if err := json.Unmarshal(event.TriggerData, &triggerData); err == nil {
            dto.TriggerData = triggerData  // 用于前端
        }
    }
    
    // 转换：json.RawMessage → []interface{}
    if len(event.NotifiedUsers) > 0 {
        var notifiedUsers []interface{}
        if err := json.Unmarshal(event.NotifiedUsers, &notifiedUsers); err == nil {
            dto.NotifiedUsers = notifiedUsers  // 用于前端
        }
    }
    
    // 输出：包含 map/[]interface{} 的 DTO
    return dto, nil
}
```

---

## 总结

**Repository 层**：统一使用 `json.RawMessage`，与数据库直接交互。

**Service 层**：
- **写入时**：业务对象 → `json.RawMessage`（序列化）
- **读取时**：`json.RawMessage` → 业务对象（反序列化）

这样确保了：
1. Repository 层统一（都使用数据库原生类型）
2. Service 层负责业务逻辑和类型转换
3. 职责清晰，易于维护

