# 报警事件写入功能说明

## ✅ 已实现功能

### 1. 报警事件构建器 (`alarm_event_builder.go`)

- ✅ `AlarmEventBuilder` - 报警事件构建器
- ✅ `BuildAlarmEvent` - 构建报警事件（自动生成 event_id，序列化 trigger_data 和 metadata）
- ✅ `BuildTriggerData` - 构建触发数据（包含 HR/RR、姿态、SNOMED 编码等）
- ✅ `CheckDuplicate` - 检查是否重复报警（在 Evaluator 中使用）

### 2. 报警事件写入 (`evaluator.go`)

- ✅ 在 `Evaluator.Evaluate` 中，评估完成后自动写入报警事件
- ✅ 遍历所有生成的报警事件，调用 `alarmEventsRepo.CreateAlarmEvent` 写入 PostgreSQL
- ✅ 记录日志（成功/失败）

### 3. 报警缓存更新 (`cache_consumer.go`)

- ✅ 只更新活跃的报警（`alarm_status = 'active'`）
- ✅ 过滤非活跃报警，避免缓存污染

## 📊 数据流

```
Evaluator.Evaluate()
    ↓
事件1-4评估（返回 []AlarmEvent）
    ↓
遍历报警事件
    ↓
alarmEventsRepo.CreateAlarmEvent() → PostgreSQL (alarm_events 表)
    ↓
cache.UpdateAlarmCache() → Redis (vital-focus:card:{card_id}:alarms)
```

## 🔍 使用示例

### 在事件评估器中创建报警事件

```go
// 1. 创建构建器
builder := NewAlarmEventBuilder(tenantID, deviceID)

// 2. 构建触发数据
triggerData := BuildTriggerData(
    "Fall",                    // event_type
    "Radar",                   // source
    nil,                       // heart_rate
    nil,                       // respiratory_rate
    &posture,                  // posture
    &postureDisplay,           // posture_display
    &snomedCode,               // snomed_code
    &snomedDisplay,            // snomed_display
    &confidence,               // confidence
    &durationSec,              // duration_sec
)

// 3. 构建元数据
metadata := map[string]interface{}{
    "trigger_source": "cloud",
    "card_id": cardID,
}

// 4. 构建报警事件
alarmEvent, err := builder.BuildAlarmEvent(
    "Fall",                    // event_type
    "safety",                  // category
    "ALERT",                   // alarm_level
    triggerData,               // trigger_data
    metadata,                  // metadata
)

// 5. 检查是否重复（可选）
isDuplicate, err := e.CheckDuplicate(tenantID, deviceID, "Fall", 5) // 5分钟内
if isDuplicate {
    return nil, nil // 跳过重复报警
}

// 6. 返回报警事件（Evaluator.Evaluate 会自动写入）
return []models.AlarmEvent{*alarmEvent}, nil
```

## 📝 注意事项

1. **报警去重**：
   - 使用 `CheckDuplicate` 检查最近 N 分钟内是否已有相同类型的报警
   - 建议在事件评估器中调用，避免重复报警

2. **设备ID**：
   - 报警事件需要 `device_id`，可以从卡片绑定的设备中获取
   - 如果卡片有多个设备，需要选择合适的设备ID（通常是触发报警的设备）

3. **序列化**：
   - `trigger_data` 和 `metadata` 会自动序列化为 JSON 字符串
   - `notified_users` 默认为空数组 `[]`

4. **时间戳**：
   - `triggered_at`、`created_at`、`updated_at` 自动设置为当前时间

## 🔗 相关文档

- `IMPLEMENTATION_SUMMARY.md` - 实现总结
- `REPOSITORY_LAYER_SUMMARY.md` - Repository 层总结
- `owlRD/db/15_alarm_events.sql` - 数据库表定义

