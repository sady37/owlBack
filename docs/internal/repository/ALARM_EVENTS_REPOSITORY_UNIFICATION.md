# AlarmEvents Repository 统一化总结

## ✅ 已完成

### 1. 类型统一

**wisefido-alarm**：
- ✅ `models.AlarmEvent` 使用 `json.RawMessage`（PostgreSQL JSONB 原生类型）
- ✅ Repository 层直接使用 `json.RawMessage` 扫描和存储

**wisefido-data**：
- ✅ `domain.AlarmEvent` 使用 `json.RawMessage`（PostgreSQL JSONB 原生类型）
- ✅ Repository 层直接使用 `json.RawMessage` 扫描和存储

**统一结果**：两个 Repository 都使用 `json.RawMessage`，与数据库原生类型一致。

---

### 2. Repository 层实现统一

两个 Repository 实现相同的接口和逻辑：

| 方法 | wisefido-alarm | wisefido-data | 状态 |
|------|----------------|---------------|------|
| `ListAlarmEvents` | ✅ | ✅ | 实现一致 |
| `GetAlarmEvent` | ✅ | ✅ | 实现一致 |
| `CreateAlarmEvent` | ✅ | ✅ | 实现一致 |
| `UpdateAlarmEvent` | ✅ | ✅ | 实现一致 |
| `DeleteAlarmEvent` | ✅ | ✅ | 实现一致 |
| `AcknowledgeAlarmEvent` | ✅ | ✅ | 实现一致 |
| `UpdateAlarmEventOperation` | ✅ | ✅ | 实现一致 |
| `GetRecentAlarmEvent` | ✅ | ✅ | 实现一致 |
| `CountAlarmEvents` | ✅ | ✅ | 实现一致 |

**关键实现点**：
- ✅ 都使用 `json.RawMessage` 处理 JSONB 字段
- ✅ 都使用 `sql.NullInt64` 处理 `iot_timeseries_id`
- ✅ 都使用 `sql.NullTime` 处理 `hand_time`
- ✅ 都使用 `sql.NullString` 处理可空字符串字段
- ✅ 都实现软删除（通过 metadata 标记）
- ✅ 都支持复杂过滤和跨表 JOIN

---

### 3. Service 层类型转换

#### wisefido-alarm Service 层（写入场景）

**位置**：`wisefido-alarm/internal/evaluator/alarm_event_builder.go:BuildAlarmEvent`

**转换**：业务对象 → `json.RawMessage`
```go
// 输入：业务对象
triggerData := &models.TriggerData{...}
metadata := map[string]interface{}{...}

// 转换：序列化
triggerDataJSON, err := json.Marshal(triggerData)  // []byte
metadataBytes, err := json.Marshal(metadata)       // []byte

// 输出：json.RawMessage（可直接存储到数据库）
event := &models.AlarmEvent{
    TriggerData:   triggerDataJSON,  // json.RawMessage
    Metadata:      json.RawMessage(metadataBytes),
    NotifiedUsers: json.RawMessage("[]"),
}
```

**职责**：将业务对象序列化为数据库存储格式。

---

#### wisefido-data Service 层（读取场景）

**位置**：`wisefido-data/internal/service/alarm_event_service.go:convertAlarmEventToDTO`

**转换**：`json.RawMessage` → 业务对象
```go
// 输入：json.RawMessage（从 Repository 获取）
event := &domain.AlarmEvent{
    TriggerData:   json.RawMessage(`{"heart_rate": 120}`),
    NotifiedUsers: json.RawMessage(`[{"user_id": "123"}]`),
    Metadata:      json.RawMessage(`{"source": "cloud"}`),
}

// 转换：反序列化
var triggerData map[string]interface{}
json.Unmarshal(event.TriggerData, &triggerData)

var notifiedUsers []interface{}
json.Unmarshal(event.NotifiedUsers, &notifiedUsers)

var metadata map[string]interface{}
json.Unmarshal(event.Metadata, &metadata)

// 输出：DTO（用于前端展示）
dto := &AlarmEventDTO{
    TriggerData:   triggerData,    // map[string]interface{}
    NotifiedUsers: notifiedUsers,   // []interface{}
    Metadata:      metadata,        // map[string]interface{}
}
```

**职责**：将数据库格式反序列化为前端展示格式。

---

## 架构原则总结

### Repository 层
- ✅ **统一使用 `json.RawMessage`**（数据库原生类型）
- ✅ **不做业务逻辑转换**
- ✅ **直接与数据库交互**

### Service 层
- ✅ **负责类型转换**（序列化/反序列化）
- ✅ **将数据库类型转换为业务需要的类型**
- ✅ **处理业务逻辑**

### 类型转换流程

```
┌─────────────────────────────────────────────────────────┐
│ wisefido-alarm (写入)                                    │
├─────────────────────────────────────────────────────────┤
│ 业务对象 (TriggerData)                                   │
│         ↓ json.Marshal                                  │
│ json.RawMessage                                         │
│         ↓ Repository.CreateAlarmEvent                   │
│ PostgreSQL JSONB                                        │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ wisefido-data (读取)                                     │
├─────────────────────────────────────────────────────────┤
│ PostgreSQL JSONB                                        │
│         ↓ Repository.GetAlarmEvent                     │
│ json.RawMessage                                         │
│         ↓ json.Unmarshal                                │
│ 业务对象 (map[string]interface{})                      │
│         ↓ Service.convertAlarmEventToDTO                │
│ DTO (用于前端)                                           │
└─────────────────────────────────────────────────────────┘
```

---

## 验证清单

- [x] wisefido-alarm 的 `models.AlarmEvent` 使用 `json.RawMessage`
- [x] wisefido-data 的 `domain.AlarmEvent` 使用 `json.RawMessage`
- [x] wisefido-alarm 的 Repository 使用 `json.RawMessage` 扫描和存储
- [x] wisefido-data 的 Repository 使用 `json.RawMessage` 扫描和存储
- [x] wisefido-alarm 的 Service 层负责序列化（业务对象 → json.RawMessage）
- [x] wisefido-data 的 Service 层负责反序列化（json.RawMessage → 业务对象）
- [x] 两个 Repository 实现相同的接口和逻辑
- [x] 所有方法编译通过

---

## 文件清单

### wisefido-alarm
- `internal/models/alarm_event.go` - 使用 `json.RawMessage`
- `internal/repository/alarm_events.go` - Repository 实现
- `internal/evaluator/alarm_event_builder.go` - Service 层序列化

### wisefido-data
- `internal/domain/alarm_event.go` - 使用 `json.RawMessage`
- `internal/repository/alarm_events_repo.go` - Repository 接口定义
- `internal/repository/postgres_alarm_events.go` - Repository 实现
- `internal/service/alarm_event_service.go` - Service 层反序列化

---

## 总结

✅ **类型统一**：两个 Repository 都使用 `json.RawMessage`（数据库原生类型）

✅ **实现统一**：两个 Repository 实现相同的接口和逻辑，但各自维护（受 Go 的 `internal` 包限制）

✅ **职责清晰**：
- Repository 层：统一使用 `json.RawMessage`，直接与数据库交互
- Service 层：负责类型转换（序列化/反序列化），处理业务逻辑

这样的设计确保了：
1. Repository 层统一（都使用数据库原生类型）
2. Service 层负责业务逻辑和类型转换
3. 职责清晰，易于维护

