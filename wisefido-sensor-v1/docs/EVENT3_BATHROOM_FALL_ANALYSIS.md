# 事件3：Bathroom 可疑跌倒检测 - 逻辑分析

## 📋 概述

**目的**：检测卫生间内长时间站立不动，可能是跌倒后无法移动。

**报警级别**：`SuspectedFall` (WARNING 级别)

**报警类型**：`AI-suspected fall`

---

## 🎯 触发条件

### 前置条件（必须全部满足）

1. **房间类型**：在 bathroom（卫生间）房间内
   - 识别方式：通过 `room_name` 或 `unit_name` 中是否包含以下词（不区分大小写）：
     - `bathroom`
     - `restroom`
     - `toilet`

2. **人员数量**：雷达检测范围内仅 **1人**
   - `realtimeData.PersonCount == 1`
   - 排除多人干扰

3. **Track ID 数量**：房间内仅有 **1个 track_id**
   - `len(realtimeData.Postures) == 1`
   - 确保只跟踪一个人

---

## ✅ 核心判断逻辑

### 条件检查（必须全部满足）

1. **在 bathroom 房间内** ✅
   - 已通过前置条件验证

2. **雷达检测范围内仅1人** ✅
   - 已通过前置条件验证

3. **1个人处于站立状态（不是坐着）** ⚠️
   - **姿态码判断**：
     - **站立**：`posture.PostureCode == "102538003"` (STANDING) 或 `"10904000"` (ORTHOSTATIC)
     - **排除坐着**：`posture.PostureCode != "102491009"` (SITTING) 且 `!= "33586001"`
   - **关键逻辑**：
     - ✅ **站立状态** + 长时间不动 → 可能是跌倒
     - ❌ **坐着状态**（坐在马桶上）+ 长时间不动 → **正常行为，不报警**

4. **位置未有变化（位置变化小于10cm，超过10分钟）** ⏱️
   - **位置变化阈值**：`< 10cm`（可配置）
   - **时间阈值**：`>= 10分钟`（可配置，默认 600秒）
   - **计算方法**：
     ```
     位置变化 = sqrt((current_x - last_x)² + (current_y - last_y)²)
     ```
   - **状态管理**：
     - 记录 `StandingTime`：开始站立的时间（T0）
     - 记录 `LastPosition`：最后记录的位置 (x, y)
     - 每次收到数据时：
       - 如果位置变化 < 10cm：继续计时
       - 如果位置变化 >= 10cm：重置计时（退出检测）

5. **房间内仅有1个 track_id** ✅
   - 已通过前置条件验证

---

## 🔄 状态管理

### Event3State 结构

```go
type Event3State struct {
    TrackID      string   // 跟踪的 track_id
    StandingTime *int64   // 开始站立的时间（T0）
    LastPosition *struct {
        X float64 // 最后位置 X（厘米）
        Y float64 // 最后位置 Y（厘米）
    }
    PositionChange float64 // 位置变化（cm）
}
```

### 状态键

```
alarm:state:{card_id}:track_{track_id}:bathroom_stand
```

### 状态生命周期

1. **初始化**：
   - 当检测到：bathroom + 1人 + 站立状态
   - 记录 `StandingTime = now`
   - 记录 `LastPosition = (current_x, current_y)`

2. **持续监控**：
   - 每次收到数据时（轮询间隔：10秒）
   - 检查位置变化：
     - 如果 `位置变化 < 10cm`：继续计时
     - 如果 `位置变化 >= 10cm`：清除状态（退出检测）
   - 检查时间：
     - 如果 `elapsed >= 10分钟`：触发报警

3. **退出条件**：
   - 位置变化 >= 10cm（人在移动）
   - 姿态不再是站立（坐下、躺下等）
   - 人员数量 != 1（多人进入）
   - track_id 消失（人离开）

4. **清除状态**：
   - 满足退出条件时
   - 报警触发后（避免重复报警）

---

## 📊 完整流程

```
轮询开始（每10秒）
    ↓
检查前置条件
    ├─ 是 bathroom？ → 否 → 退出
    ├─ 仅1人？ → 否 → 退出
    └─ 仅1个 track_id？ → 否 → 退出
    ↓
检查姿态
    ├─ 是站立状态？ → 否 → 退出
    └─ 不是坐着状态？ → 否 → 退出（正常行为）
    ↓
获取/初始化状态
    ├─ 状态不存在？ → 初始化（T0 = now, LastPosition = current）
    └─ 状态存在？ → 继续监控
    ↓
检查位置变化
    ├─ 位置变化 >= 10cm？ → 是 → 清除状态，退出（人在移动）
    └─ 位置变化 < 10cm？ → 继续
    ↓
检查时间阈值
    ├─ elapsed >= 10分钟？ → 是 → 触发报警
    └─ elapsed < 10分钟？ → 继续监控
    ↓
保存状态（更新 LastPosition）
```

---

## 🚨 报警触发

### 触发条件

- **时间阈值**：`elapsed >= 10分钟`（600秒）
- **位置变化**：`< 10cm`（持续10分钟）

### 报警信息

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
    },
    Metadata: {
        card_id: card.CardID,
        track_id: trackID,
        trigger_point: "bathroom_stand_still",
        standing_time: state.StandingTime,
        duration_seconds: elapsed,
        position_change: state.PositionChange,
    }
}
```

---

## 🔍 关键逻辑点

### 1. 姿态区分 ⚠️

**必须**：站立状态（`STANDING`）
- 站立 + 长时间不动 → 可能是跌倒

**排除**：坐着状态（`SITTING`）
- 坐着（坐在马桶上）+ 长时间不动 → **正常行为，不报警**
- 这是关键的业务逻辑，避免误报

### 2. 位置未变化判断

- **阈值**：`< 10cm`（可配置）
- **计算方法**：欧几里得距离
- **单位**：厘米（cm）

### 3. 时间阈值

- **默认值**：`10分钟`（600秒，可配置）
- **可配置项**：`BATHROOM_STAND_STILL_DURATION`

### 4. 单人检测

- **必须**：仅1人（`PersonCount == 1`）
- **必须**：仅1个 track_id（`len(Postures) == 1`）
- **目的**：排除多人干扰

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

### 状态存储（Redis）

- **状态键**：`alarm:state:{card_id}:track_{track_id}:bathroom_stand`
- **TTL**：15分钟（状态最长持续时间）

### 报警存储（PostgreSQL）

- **表**：`alarm_events`
- **缓存**：`vital-focus:card:{card_id}:alarms`（Redis）

---

## ⚙️ 配置项

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `BATHROOM_STAND_STILL_DURATION` | 600秒（10分钟） | 站立不动的时间阈值 |
| `BATHROOM_POSITION_CHANGE_THRESHOLD` | 10cm | 位置变化阈值 |
| `EVENT3_STATE_TTL` | 900秒（15分钟） | 状态TTL |

---

## 🔄 与 Event 1 的对比

| 特性 | Event 1 (床上跌落) | Event 3 (Bathroom跌倒) |
|------|-------------------|------------------------|
| **触发方式** | 事件驱动（BED_LEFT） | 轮询（每10秒） |
| **检测区域** | 床区域 | Bathroom房间 |
| **关键条件** | sleepad离床 + radar仍在床上 | 站立 + 长时间不动 |
| **时间窗口** | T0+10秒, T0+120秒 | 持续10分钟 |
| **报警级别** | Fall (ALERT) / SuspectedFall (WARNING) | SuspectedFall (WARNING) |
| **状态管理** | 复杂（多阶段检测） | 简单（单阶段计时） |

---

## ✅ 实现要点

1. **模块化函数**：
   - `checkBathroom()`：检查是否是bathroom
   - `checkStandingPosture()`：检查是否是站立状态
   - `checkPositionChange()`：检查位置变化
   - `checkStandingDuration()`：检查站立持续时间

2. **状态管理**：
   - 使用 `Event3State` 存储状态
   - 使用 Redis 持久化状态
   - 设置合理的 TTL

3. **退出条件**：
   - 位置变化 >= 10cm
   - 姿态不再是站立
   - 人员数量 != 1
   - track_id 消失

4. **报警触发**：
   - 仅在满足所有条件时触发
   - 避免重复报警（使用 `AlarmTriggered` 标志）

---

## 📌 待确认问题

1. **SNOMED 编码**：
   - 站立：`102538003` (STANDING) 还是 `10904000` (ORTHOSTATIC)？
   - 坐位：`102491009` (SITTING) 还是 `33586001`？
   - 需要查看实际的 `PostureCode` 值

2. **位置单位**：
   - 确认 `PositionX` 和 `PositionY` 的单位是厘米（cm）

3. **时间阈值**：
   - 10分钟是否合适？是否需要可配置？

4. **位置变化阈值**：
   - 10cm 是否合适？是否需要可配置？

---

## 🎯 下一步

1. ✅ 分析完成
2. ⏳ 确认 SNOMED 编码
3. ⏳ 实现模块化函数
4. ⏳ 实现完整逻辑
5. ⏳ 添加单元测试

