# AlarmEvents Repository 统一化完成总结

## ✅ 已完成项目

### 1. ✅ 类型统一

**wisefido-alarm**：
- ✅ `models.AlarmEvent` 使用 `json.RawMessage` 处理 JSONB 字段
- ✅ Repository 层直接使用 `json.RawMessage` 扫描和存储

**wisefido-data**：
- ✅ `domain.AlarmEvent` 使用 `json.RawMessage` 处理 JSONB 字段
- ✅ Repository 层直接使用 `json.RawMessage` 扫描和存储

**验证结果**：
```bash
wisefido-alarm models.AlarmEvent:
	TriggerData      json.RawMessage `json:"trigger_data" db:"trigger_data"`

wisefido-data domain.AlarmEvent:
	TriggerData     json.RawMessage `db:"trigger_data"`
```

---

### 2. ✅ Repository 层实现统一

两个 Repository 实现相同的接口方法：

| 接口方法 | wisefido-alarm | wisefido-data | 状态 |
|---------|----------------|---------------|------|
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
- ✅ 都使用 `json.RawMessage` 处理 JSONB 字段（`trigger_data`, `notified_users`, `metadata`）
- ✅ 都使用 `sql.NullInt64` 处理 `iot_timeseries_id`
- ✅ 都使用 `sql.NullTime` 处理 `hand_time`
- ✅ 都使用 `sql.NullString` 处理可空字符串字段（`handler`, `operation`, `notes`）
- ✅ 都实现软删除（通过 metadata 标记 `deleted_at`）
- ✅ 都支持复杂过滤和跨表 JOIN（devices, beds, rooms, units, residents）

---

### 3. ✅ Service 层类型转换

#### wisefido-alarm Service 层（写入场景）

**位置**：`wisefido-alarm/internal/evaluator/alarm_event_builder.go:BuildAlarmEvent`

**转换流程**：
```
业务对象 (TriggerData, map[string]interface{})
    ↓ json.Marshal
json.RawMessage
    ↓ Repository.CreateAlarmEvent
PostgreSQL JSONB
```

**代码示例**：
```go
// 序列化 trigger_data
triggerDataJSON, err := json.Marshal(triggerData)  // []byte → json.RawMessage

// 序列化 metadata
metadataBytes, err := json.Marshal(metadata)
metadataJSON := json.RawMessage(metadataBytes)

// 构建 AlarmEvent（包含 json.RawMessage）
event := &models.AlarmEvent{
    TriggerData:   triggerDataJSON,  // json.RawMessage
    Metadata:      metadataJSON,     // json.RawMessage
    NotifiedUsers: json.RawMessage("[]"),
}
```

**职责**：将业务对象序列化为数据库存储格式。

---

#### wisefido-data Service 层（读取场景）

**位置**：`wisefido-data/internal/service/alarm_event_service.go:convertAlarmEventToDTO`

**转换流程**：
```
PostgreSQL JSONB
    ↓ Repository.GetAlarmEvent
json.RawMessage
    ↓ json.Unmarshal
业务对象 (map[string]interface{}, []interface{})
    ↓ Service.convertAlarmEventToDTO
DTO (用于前端展示)
```

**代码示例**：
```go
// 反序列化 trigger_data
if len(event.TriggerData) > 0 {
    var triggerData map[string]interface{}
    if err := json.Unmarshal(event.TriggerData, &triggerData); err == nil {
        dto.TriggerData = triggerData  // 用于前端展示
    }
}

// 反序列化 notified_users
if len(event.NotifiedUsers) > 0 {
    var notifiedUsers []interface{}
    if err := json.Unmarshal(event.NotifiedUsers, &notifiedUsers); err == nil {
        dto.NotifiedUsers = notifiedUsers  // 用于前端展示
    }
}

// 反序列化 metadata
if len(event.Metadata) > 0 {
    var metadata map[string]interface{}
    if err := json.Unmarshal(event.Metadata, &metadata); err == nil {
        dto.Metadata = metadata  // 用于前端展示
    }
}
```

**职责**：将数据库格式反序列化为前端展示格式。

---

## 架构原则

### Repository 层
- ✅ **统一使用 `json.RawMessage`**（PostgreSQL JSONB 原生类型）
- ✅ **不做业务逻辑转换**
- ✅ **直接与数据库交互**
- ✅ **两个 Repository 各自维护，但实现一致**

### Service 层
- ✅ **负责类型转换**（序列化/反序列化）
- ✅ **将数据库类型转换为业务需要的类型**
- ✅ **处理业务逻辑**

---

## 文件清单

### wisefido-alarm
- ✅ `internal/models/alarm_event.go` - 使用 `json.RawMessage`
- ✅ `internal/repository/alarm_events.go` - Repository 实现（15 个方法，包含辅助方法）
- ✅ `internal/evaluator/alarm_event_builder.go` - Service 层序列化

### wisefido-data
- ✅ `internal/domain/alarm_event.go` - 使用 `json.RawMessage`
- ✅ `internal/repository/alarm_events_repo.go` - Repository 接口定义
- ✅ `internal/repository/postgres_alarm_events.go` - Repository 实现（10 个接口方法 + 辅助方法）
- ✅ `internal/service/alarm_event_service.go` - Service 层反序列化

---

## 编译验证

```bash
✅ wisefido-alarm Repository 编译通过
✅ wisefido-data Repository 编译通过
✅ wisefido-data Service 层编译通过
```

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
4. 两个 Repository 各自维护，但实现一致

---

## 下一步

- [ ] 阶段 4：编写 AlarmEvent Service 测试
- [ ] 阶段 5：实现 AlarmEvent Handler
- [ ] 阶段 6：集成和路由注册
- [ ] 阶段 7：验证和测试

