# 事件3：Bathroom 跌倒检测 - 重新设计

## 📋 需求分析

### 用户需求

1. **进出门事件触发**：不是轮询，而是事件驱动（检测到进入 bathroom 时开始监测）
2. **房间名确认**：检查 `room_name` 是否包含：
   - `Bathroom` / `restRoom`（完整词）
   - `bath` / `rest`（子串）
3. **持续监测站姿**：站姿持续时间 >= 10分钟 → 报警 `SuspectedFall` (WARNING)
4. **位置保持不动 >30分钟**：报警 `Fall` (ALERT)（更严重）
   - **原因**：没有人能站立不动超过10分钟，很可能是房间干扰导致跌倒，但雷达无法检测

---

## 🎯 触发机制

### 方式1：事件驱动（推荐）

**触发条件**：检测到进入 bathroom 时开始监测

**如何检测进入**：
1. **通过 area_id 变化**：
   - 从实时数据中检测到 `area_id` 匹配 bathroom 的 area_id
   - 之前 `area_id` 不匹配，现在匹配 → 进入事件

2. **通过 room_id 匹配**：
   - 从实时数据中的设备信息获取 `room_id`
   - 检查 `room_id` 对应的房间是否是 bathroom

3. **订阅进入事件**（如果系统支持）：
   - 订阅 `ENTER_ROOM` 事件
   - 过滤 `room_id` 是 bathroom 的事件

### 方式2：轮询检测（备选）

**触发条件**：每次轮询时检查是否在 bathroom 内

**逻辑**：
- 如果检测到在 bathroom 内，且之前不在 → 进入事件
- 如果检测到不在 bathroom 内，且之前在 → 离开事件

---

## ✅ 核心判断逻辑

### 阶段1：进入检测

```
检测到进入 bathroom
    ↓
检查前置条件
    ├─ 房间名包含 "Bathroom"/"restRoom" 或 "bath"/"rest"？ → 否 → 退出
    ├─ 仅1人？ → 否 → 退出
    └─ 仅1个 track_id？ → 否 → 退出
    ↓
初始化监测状态（T0 = now）
```

### 阶段2：持续监测（每10秒轮询一次）

```
检查是否仍在 bathroom
    ├─ 不在？ → 清除状态，退出
    └─ 仍在？ → 继续
    ↓
检查 track_id 数量
    ├─ > 1个 track_id？ → 清除状态，退出（中途有新的track_id进入）
    └─ == 1个 track_id？ → 继续
    ↓
检查人员数量
    ├─ != 1人？ → 清除状态，退出
    └─ == 1人？ → 继续
    ↓
检查姿态
    ├─ 不是站立？ → 清除状态，退出（可能是坐下，正常行为）
    └─ 是站立？ → 继续
    ↓
检查位置变化
    ├─ 位置变化 >= 10cm？ → 重置计时（人在移动，继续监测）
    └─ 位置变化 < 10cm？ → 继续计时
    ↓
检查时间阈值
    ├─ 位置不动持续时间 >= 30分钟？ → 触发 Fall (ALERT) 🚨
    └─ 站姿持续时间 >= 10分钟？ → 触发 SuspectedFall (WARNING) ⚠️
```

---

## 🔄 状态管理

### Event3State 结构（更新）

```go
type Event3State struct {
    TrackID          string   // 跟踪的 track_id
    EnterTime        *int64   // T0：进入 bathroom 的时间
    StandingTime     *int64   // 开始站立的时间（如果进入时就是站立）
    LastPosition     *struct {
        X float64 // 最后位置 X（厘米）
        Y float64 // 最后位置 Y（厘米）
    }
    PositionChange   float64  // 位置变化（cm）
    LastPositionTime *int64   // 最后位置的时间
    StandingDuration int64    // 站姿持续时间（秒）
    StillDuration    int64    // 位置不动持续时间（秒）
    AlarmTriggered   bool     // 是否已触发报警（避免重复报警）
}
```

### 状态键

```
alarm:state:{card_id}:track_{track_id}:bathroom_monitor
```

### 状态生命周期

1. **初始化**（进入 bathroom）：
   - `EnterTime = now`
   - `LastPosition = (current_x, current_y)`
   - `LastPositionTime = now`
   - 如果进入时就是站立：`StandingTime = now`

2. **持续监控**（每10秒）：
   - 检查是否仍在 bathroom
   - 检查人员数量
   - 检查姿态
   - 更新位置变化
   - 更新持续时间

3. **退出条件**：
   - 离开 bathroom（`area_id` 不匹配）
   - 人员数量 != 1
   - 姿态不再是站立（坐下、躺下等）
   - track_id 消失（人离开）

4. **清除状态**：
   - 满足退出条件时
   - 报警触发后（避免重复报警）

---

## 📊 完整流程

### 流程图

```
事件触发：检测到进入 bathroom
    ↓
前置条件检查
    ├─ 房间名匹配？ → 否 → 退出
    ├─ 仅1人？ → 否 → 退出
    └─ 仅1个 track_id？ → 否 → 退出
    ↓
初始化状态（T0 = now）
    ↓
持续监测（每10秒轮询）
    ↓
┌─────────────────────────────────────┐
│ 检查是否仍在 bathroom                │
└─────────────────────────────────────┘
    ├─ 不在？ → 清除状态，退出
    └─ 仍在？ → 继续
    ↓
┌─────────────────────────────────────┐
│ 检查人员数量                         │
└─────────────────────────────────────┘
    ├─ != 1人？ → 清除状态，退出
    └─ == 1人？ → 继续
    ↓
┌─────────────────────────────────────┐
│ 检查姿态                             │
└─────────────────────────────────────┘
    ├─ 不是站立？ → 清除状态，退出（正常行为）
    └─ 是站立？ → 继续
    ↓
┌─────────────────────────────────────┐
│ 检查位置变化                         │
└─────────────────────────────────────┘
    ├─ 位置变化 >= 10cm？ → 重置计时（人在移动）
    └─ 位置变化 < 10cm？ → 继续计时
    ↓
┌─────────────────────────────────────┐
│ 检查时间阈值                         │
└─────────────────────────────────────┘
    ├─ 站姿持续时间 >= 10分钟？
    │   └─ 是 → 触发 SuspectedFall (WARNING)
    └─ 位置不动持续时间 >= 30分钟？
        └─ 是 → 触发 Fall (ALERT)
    ↓
保存状态（更新 LastPosition, 持续时间）
```

---

## 🚨 报警触发

### 报警1：站姿持续时间 >= 10分钟

**触发条件**：
- 站姿持续时间：`>= 10分钟`（600秒）
- 位置变化：`< 10cm`（持续10分钟）

**报警信息**：
```go
{
    EventType: "SuspectedFall",
    Category: "safety",
    AlarmLevel: "WARNING",
    TriggerData: {
        EventType: "SuspectedFall",
        Source: "Radar",
        Posture: "102538003", // STANDING
        PostureDisplay: "Standing position",
        SNOMEDCode: "248220002", // Suspected Fall
        SNOMEDDisplay: "Suspected Fall",
        PositionX: current_x,
        PositionY: current_y,
        DurationSec: standing_duration,
    },
    Metadata: {
        card_id: card.CardID,
        track_id: trackID,
        trigger_point: "bathroom_stand_still_10min",
        enter_time: state.EnterTime,
        standing_duration: state.StandingDuration,
        position_change: state.PositionChange,
    }
}
```

### 报警2：位置不动持续时间 >= 30分钟（更严重）

**触发条件**：
- 位置不动持续时间：`>= 30分钟`（1800秒）
- 位置变化：`< 10cm`（持续30分钟）

**报警信息**：
```go
{
    EventType: "Fall",
    Category: "safety",
    AlarmLevel: "ALERT", // 更严重
    TriggerData: {
        EventType: "Fall",
        Source: "Radar",
        Posture: "102538003", // STANDING
        PostureDisplay: "Standing position",
        SNOMEDCode: "248220002", // Fall
        SNOMEDDisplay: "Fall",
        PositionX: current_x,
        PositionY: current_y,
        DurationSec: still_duration,
    },
    Metadata: {
        card_id: card.CardID,
        track_id: trackID,
        trigger_point: "bathroom_still_30min",
        enter_time: state.EnterTime,
        still_duration: state.StillDuration,
        position_change: state.PositionChange,
        reason: "No one can stand still for more than 10 minutes, likely fall due to room interference",
    }
}
```

---

## 🔍 关键逻辑点

### 1. 房间名匹配 ⚠️

**匹配规则**（不区分大小写）：
- 完整词：`Bathroom`, `restRoom`
- 子串：`bath`, `rest`, `toilet`

**实现**：
```go
func isBathroomRoom(roomName string) bool {
    roomNameLower := strings.ToLower(roomName)
    return strings.Contains(roomNameLower, "bathroom") ||
           strings.Contains(roomNameLower, "restroom") ||
           strings.Contains(roomNameLower, "bath") ||
           strings.Contains(roomNameLower, "rest") ||
           strings.Contains(roomNameLower, "toilet")
}
```

### 2. 进入检测

**方式1：通过 area_id 变化**（推荐）
- 从 `config_version` 表获取 bathroom 的 `area_id`
- 检测 `Posture.AreaID` 是否匹配
- 如果之前不匹配，现在匹配 → 进入事件

**方式2：通过 room_id 匹配**
- 从设备信息获取 `room_id`
- 检查 `room_id` 对应的房间是否是 bathroom
- 如果之前不在，现在在 → 进入事件

### 3. 站姿持续时间计算

**逻辑**：
- 如果进入时就是站立：`StandingTime = EnterTime`
- 如果进入时不是站立，后来变成站立：`StandingTime = now`
- 如果从站立变成非站立：重置 `StandingTime = nil`
- 持续时间：`elapsed = now - StandingTime`

### 4. 位置不动持续时间计算

**逻辑**：
- 每次位置变化 < 10cm：继续计时
- 每次位置变化 >= 10cm：重置计时
- 持续时间：`elapsed = now - LastPositionTime`（如果位置未变化）

### 5. 双重报警机制

**报警优先级**：
1. **位置不动 >= 30分钟** → `Fall` (ALERT)（更严重）
2. **站姿 >= 10分钟** → `SuspectedFall` (WARNING)

**逻辑**：
- 如果位置不动 >= 30分钟，直接触发 `Fall`（不触发 `SuspectedFall`）
- 如果站姿 >= 10分钟但位置不动 < 30分钟，触发 `SuspectedFall`

---

## 📝 数据源

### 输入数据（Redis）

- **RealtimeData**：从 `vital-focus:card:{card_id}:realtime` 读取
  - `PersonCount`：人数
  - `Postures[]`：姿态列表
    - `TrackingID`：跟踪ID
    - `PostureCode`：姿态SNOMED编码
    - `PositionX`：位置X（厘米）
    - `PositionY`：位置Y（厘米）
    - `AreaID`：区域ID

### 房间信息（PostgreSQL）

- **cards.devices JSONB**：包含 `room_name`
- **rooms 表**：包含 `room_name`
- **config_versions 表**：包含 bathroom 的 `area_id`

### 状态存储（Redis）

- **状态键**：`alarm:state:{card_id}:track_{track_id}:bathroom_monitor`
- **TTL**：60分钟（状态最长持续时间）

### 报警存储（PostgreSQL）

- **表**：`alarm_events`
- **缓存**：`vital-focus:card:{card_id}:alarms`（Redis）

---

## ⚙️ 配置项

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `BATHROOM_STANDING_DURATION` | 600秒（10分钟） | 站姿持续时间阈值 |
| `BATHROOM_STILL_DURATION` | 1800秒（30分钟） | 位置不动持续时间阈值 |
| `BATHROOM_POSITION_CHANGE_THRESHOLD` | 10cm | 位置变化阈值 |
| `EVENT3_STATE_TTL` | 3600秒（60分钟） | 状态TTL |
| `EVENT3_POLL_INTERVAL` | 10秒 | 轮询间隔 |

---

## 🔄 与 Event 1 的对比

| 特性 | Event 1 (床上跌落) | Event 3 (Bathroom跌倒) |
|------|-------------------|------------------------|
| **触发方式** | 事件驱动（BED_LEFT） | 事件驱动（进入bathroom） |
| **检测区域** | 床区域（area_id） | Bathroom房间（room_id/area_id） |
| **关键条件** | sleepad离床 + radar仍在床上 | 站立 + 长时间不动 |
| **时间窗口** | T0+10秒, T0+120秒 | 持续10分钟/30分钟 |
| **报警级别** | Fall (ALERT) / SuspectedFall (WARNING) | SuspectedFall (WARNING) / Fall (ALERT) |
| **状态管理** | 复杂（多阶段检测） | 中等（单阶段持续监测） |

---

## ✅ 实现要点

### 1. 模块化函数

- `checkBathroomRoomName()`：检查房间名是否匹配
- `detectEnterBathroom()`：检测进入 bathroom 事件
- `checkStandingPosture()`：检查是否是站立状态
- `checkPositionChange()`：检查位置变化
- `checkStandingDuration()`：检查站姿持续时间
- `checkStillDuration()`：检查位置不动持续时间

### 2. 状态管理

- 使用 `Event3State` 存储状态
- 使用 Redis 持久化状态
- 设置合理的 TTL（60分钟）

### 3. 进入检测

- 方式1：通过 `area_id` 变化（推荐）
- 方式2：通过 `room_id` 匹配（备选）
- 方式3：订阅进入事件（如果系统支持）

### 4. 退出条件

- 离开 bathroom（`area_id` 不匹配）
- 人员数量 != 1
- 姿态不再是站立（坐下、躺下等）
- track_id 消失（人离开）

### 5. 报警触发

- 仅在满足所有条件时触发
- 避免重复报警（使用 `AlarmTriggered` 标志）
- 双重报警机制（10分钟/30分钟）

---

## 📌 待确认问题

1. **进入检测方式**：
   - 是否通过 `area_id` 变化检测？
   - 还是通过 `room_id` 匹配？
   - 是否有进入事件可以订阅？

2. **SNOMED 编码**：
   - 站立：`102538003` (STANDING) 还是 `10904000` (ORTHOSTATIC)？
   - 坐位：`102491009` (SITTING) 还是 `33586001`？

3. **位置单位**：
   - 确认 `PositionX` 和 `PositionY` 的单位是厘米（cm）

4. **时间阈值**：
   - 10分钟和30分钟是否合适？是否需要可配置？

5. **位置变化阈值**：
   - 10cm 是否合适？是否需要可配置？

---

## 🎯 下一步

1. ✅ 需求分析完成
2. ⏳ 确认进入检测方式
3. ⏳ 确认 SNOMED 编码
4. ⏳ 实现模块化函数
5. ⏳ 实现完整逻辑
6. ⏳ 添加单元测试

