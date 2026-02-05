# Alarm 处理流程 - 双重比对策略

## 整体架构

```
Redis Stream (alarm/event) 
    ↓
EventAlarmHandler (JSON 解析)
    ↓
EventAlarmService.HandleMessage()
    ├─ 【统一入口】5s 时间窗口检查：msg.Timestamp + 5s > realtimeData.Timestamp
    │   └─ 如果失败 → 直接拒绝（消息太旧）
    │
    ├─ handleEvent() 或 handleAlarm() 分流
    │   └─ 遍历 data_value 数组
    │       └─ updateActiveAlarm()
    │           ├─ 【比对 1】检查 alarm.timestamp > existing.timestamp
    │           │   └─ 如果失败 → 不更新（时间戳没有更新）
    │           │
    │           ├─ 【比对 2】检查 alarm_level 优先级（防止低级覆盖高级）
    │           │   └─ 如果新级别不如现有高 → 不更新
    │           │
    │           └─ ✓ 双重比对通过 → 更新计数和 NowAlarm
    │
    └─ 保存 RealtimeData 到 Redis（TTL 5秒/12小时）
```

## 关键步骤详解

### 步骤 1：入口处统一时间窗口检查

**位置**：`EventAlarmService.HandleMessage()` 开始

```go
// 如果 msg.Timestamp + 5s <= realtimeData.Timestamp，说明消息太旧，直接拒绝
if msg.Timestamp+5 < realtimeData.Timestamp {
    // 记录并返回，不进行进一步处理
    return nil
}
```

**目的**：防止接收网络延迟或队列延迟导致的古老消息

**时间窗口**：5 秒（可覆盖所有现实网络场景）

---

### 步骤 2：比对 1 - 时间戳比较

**位置**：`EventAlarmService.updateActiveAlarm()`

```go
// 【比对 1】检查 timestamp：只有 alarm.timestamp > existing.timestamp 才更新
if alarmTimestamp <= realtimeData.ActiveAlarms.Timestamp {
    // 时间戳没有更新，不处理
    return
}
```

**目的**：防止同一时刻的多个报警消息相互覆盖，保证最新的报警信息被保留

**工作方式**：
- 新报警的 `alarm.timestamp` 必须大于当前 `ActiveAlarms.Timestamp`
- 如果相同或更旧，直接返回

---

### 步骤 3：比对 2 - 报警级别比较

**位置**：`EventAlarmService.updateActiveAlarm()`

```go
// 【比对 2】检查 alarm_level：防止低级别报警覆盖高级别报警
highestLevelStr := ""
if realtimeData.ActiveAlarms.NowAlarm != "" {
    parts := splitAlarmCategory(realtimeData.ActiveAlarms.NowAlarm) // 格式："EMERG.Fall"
    highestLevelStr = parts[0]
}

newPriority := alarmLevelPriority[alarmLevel]
existingPriority := alarmLevelPriority[highestLevelStr]

// 如果新报警级别不如现有的高（优先级数字更大），不更新
if highestLevelStr != "" && newPriority > existingPriority {
    return
}
```

**报警级别优先级**（数字越小优先级越高）：
- `EMERG` = 0（最高）
- `ALERT` = 1
- `CRIT` = 2
- `ERR` = 3
- `WARNING` = 4
- `NOTICE` = 5（最低）

**工作方式**：
- 新报警的优先级必须 ≥ 当前最高级别（数字需要 ≤）
- 例如：`WARNING` 不能覆盖 `EMERG`，但 `EMERG` 可以覆盖 `WARNING`

---

### 步骤 4：更新 ActiveAlarms

**双重比对通过后**：

```go
// 更新 timestamp
realtimeData.ActiveAlarms.Timestamp = alarmTimestamp

// 根据级别增加计数
switch alarmLevel {
case "EMERG":
    realtimeData.ActiveAlarms.EMERG++
case "ALERT":
    realtimeData.ActiveAlarms.ALERT++
// ... 其他级别
}

// 更新 NowAlarm（格式："AlarmLevel.Alarm"，如 "EMERG.Alarm"）
realtimeData.ActiveAlarms.NowAlarm = alarmLevel + ".Alarm"
```

**更新的字段**：
- `Timestamp`：最后一次更新的时间（用于比对 1）
- `EMERG`/`ALERT`/`CRIT`/`ERR`/`WARNING`/`NOTICE`：各级别未处理数量（+1）
- `NowAlarm`：当前最高级别的报警（格式："AlarmLevel.Alarm"）

---

## 数据流示例

### 场景：同一卡片收到多个报警

**初始状态**：`RealtimeData.ActiveAlarms = nil`

**消息 1**：`ALERT.Fall` 在 `ts=1000`
- 时间窗口：✓ (首次)
- 比对 1：✓ (首次)
- 比对 2：✓ (首次)
- **结果**：`ALERT++`, `NowAlarm="ALERT.Alarm"`, `Timestamp=1000`

**消息 2**：`WARNING.SignalPoor` 在 `ts=1005`
- 时间窗口：✓ (1005+5=1010 > 1000)
- 比对 1：✓ (1005 > 1000)
- 比对 2：✗ (WARNING 优先级 4 > ALERT 优先级 1)
- **结果**：不更新

**消息 3**：`EMERG.Fall` 在 `ts=1010`
- 时间窗口：✓ (1010+5=1015 > 1000)
- 比对 1：✓ (1010 > 1000)
- 比对 2：✓ (EMERG 优先级 0 < ALERT 优先级 1)
- **结果**：`EMERG++`, `NowAlarm="EMERG.Alarm"`, `Timestamp=1010`

**消息 4**：`EMERG.Fall` 重复 在 `ts=1010`（网络重复）
- 时间窗口：✓ (1010+5=1015 > 1000)
- 比对 1：✗ (1010 = 1010，不是 >)
- **结果**：不更新（防止重复计数）

**最终状态**：
```json
{
  "ALERT": 1,
  "EMERG": 1,
  "NowAlarm": "EMERG.Alarm",
  "Timestamp": 1010
}
```

---

## 关键设计决策

| 决策 | 原因 |
|------|------|
| 5s 时间窗口（Handler 层） | 覆盖所有现实网络/队列延迟场景 |
| 时间戳双重检查 | 防止时间戳相同的消息相互覆盖 |
| 级别优先级检查 | 防止低级别报警刷掉高级别报警（UX 问题） |
| `NowAlarm` 格式统一 | "AlarmLevel.Alarm" 便于前端识别和显示 |
| 计数独立维护 | 每个级别的计数可独立增长，支持多级报警同时存在 |

---

## 常见问题

**Q：为什么需要两个时间戳检查（Handler 层 + Service 层）？**

A：不需要两个。现在已经统一移到 Service 层的 HandleMessage 方法开始处理，在分流到 event/alarm 前进行。

**Q：为什么不直接用级别排序而是用优先级数字？**

A：数字比较更快，且便于排序。0 最高，5 最低，直观清晰。

**Q：NowAlarm 为什么总是后缀 ".Alarm"？**

A：因为在 Service 层不需要知道具体的报警类型（Fall/SignalPoor 等），只关心级别。前端可根据业务逻辑解释 ".Alarm"。

**Q：如果同时收到 EMERG 和 WARNING，计数都会增加吗？**

A：是的。
- 第一条 EMERG → `EMERG=1, NowAlarm=EMERG`
- 第二条 WARNING → `EMERG=1, WARNING=1, NowAlarm=EMERG`（不覆盖，因为优先级低）

前端可根据 `NowAlarm` 显示最高级别，同时显示 `EMERG` 和 `WARNING` 的计数。
