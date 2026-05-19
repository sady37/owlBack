# card_display.md — `card:state.display` 字段权威定义

> 适用范围：`owlBack/owl-common/card/card_types.go` 的 `CardDisplay` struct + Redis hash `card:state:<cardID>` 的 `display` 字段。
>
> 设计原则：BE 收口所有 picker / scene / risk / time-window 派生；FE 退化为 dumb renderer。FE Overview 只读 `display` + `realtime stream` + `device_status` 三者就能渲染完整卡片。

## 1. UI 4-Section 布局

```
┌──────────────────────────────────────────────┐
│ Section1.up.left          Section1.up.right  │
│ unit + card_name          alarmBell          │
├──────────────────────────────────────────────┤
│ Section1.down.left        Section1.down.right│
│ active_room/"WholeUnit"   alarmEvent 简称    │
├──────────────────────────────────────────────┤
│ Section2.left   Section2.middle  Section2.right│
│ Status 区        Vital 区          Posture 区  │
│ sleepstage 或    HR/BR             前 3 + badge│
│ RoomStatus       (仅 bed_id!=null) PoseDisplay │
│                                    Priority    │
├──────────────────────────────────────────────┤
│ Section3.up.left      Section3.up.right       │
│ ActiveState +Time     SceneState + Time       │
├──────────────────────────────────────────────┤
│ Section3.down: Visitor 或 Bed timing fallback│
└──────────────────────────────────────────────┘
```

## 2. UI 区 ↔ 数据来源总览

| UI 区 | 数据来源 | producer | 在 display 里？ |
|---|---|---|---|
| Section1.up.left (unit + card_name) | `cards` 表 + `residents` JOIN（静态） | wisefido-data card_static API | 否 |
| Section1.up.right (alarmBell 颜色) | `alarm_state.active_emerg/alert/crit/err/warning` 计数器 | AlarmRouter | 否 |
| Section1.down.left (active room 名) | FE 业务侧自决（card.bed_id == null 时） | — | 否 |
| Section1.down.right (alarmEvent 简称) | `alarm_state.pop_alarm`（FE 查 `alarm.GetAlarmDisplayName`） | AlarmRouter | 否 |
| Section2.left (icon + badge) | display 内 `section2_left_mode` + `bed_status` + `sleep_stage` + `room_*` 字段族 | UnitPicker (/80) / SensorStateProjector (leaf) | **是** |
| Section2.middle (HR / BR) | realtime stream | wisefido-data realtime SSE | 否 |
| Section2.right (postures) | realtime stream | wisefido-data realtime SSE | 否 |
| Section3.up.left (ActiveState + Time) | display.`active_state` + `active_anchor_ms` | UnitPicker / SensorStateProjector | **是** |
| Section3.up.right (SceneState + Time) | display.`scene_state` + `scene_anchor_ms` | UnitPicker / SensorStateProjector | **是** |
| Section3.down (Visitor 或 Bed timing) | display.`visitor_state` + `visitor_anchor_ms` + `bed_anchor_ms` | UnitPicker (visitor 仅 /80 private) / SensorStateProjector (bed timing leaf) | **是** |

**约定**：
- `display` 是 FE 渲染的唯一契约
- `alarm_state` 由 AlarmRouter 独立写，不进 display
- `room_state` / `bed_state` / `target` 是 BE 内部 state，**FE Overview 不直接消费**（Detail / RadarCanvas 等诊断视图另走专属 API）

## 3. CardDisplay struct 字段表

> 位置：`owlBack/owl-common/card/card_types.go` 内 `CardDisplay` struct。
>
> 通配：**leaf 卡** (/88 /96) display 由 `SensorStateProjector` 调 `BuildCardDisplay(self)` 写；**/80 unit 卡** 由 `UnitPicker.HandleChildEvent` 写。

### 3.1 Meta

| 字段 | JSON | 类型 | 取值方式 |
|---|---|---|---|
| `UpdatedAt` | `updated_at` | int64 (unix ms) | 写入时取 `time.Now().UnixMilli()` |

### 3.2 Section2.left（status 区）

| 字段 | JSON | 类型 | 取值方式 |
|---|---|---|---|
| `Section2LeftMode` | `section2_left_mode` | int | `0`=None / `1`=SleepStage / `2`=RoomStatus，picker 算法见 §4.1 |
| `BedStatus` | `bed_status` | int | `0`=InBed / `1`=NotInBed / `8`=Unknown；直接 copy `bs.bed_status`；FE 仅在 `CardStatic.bed_id != null` 时消费 |
| `SleepStage` | `sleep_stage` | int | `0/1/2/4/8`（复用 owl-common `SleepStage*`）；mode=SleepStage AND BedStatus=0 时 = `bs.sleep_stage` |
| `RoomPersonCount` | `room_person_count` | int | mode=RoomStatus 时 = `rs.total_people`（右上角 badge）|
| `RoomIconKind` | `room_icon_kind` | int | `0`=Room / `1`=Bathroom；按 `rs.kind == "bathroom"` 决定 |
| `RoomRiskLevel` | `room_risk_level` | int | mode=RoomStatus 时 = `rs.risk_level`（0/1/2/3，由 sensor `risk_evaluator.go` 算）|

### 3.3 Section3.up.left（ActiveState）

| 字段 | JSON | 类型 | 取值方式 |
|---|---|---|---|
| `ActiveState` | `active_state` | int | `0`=Inactive / `1`=Now。算法：`now - ActiveAnchorMs < 60_000` → Now |
| `ActiveAnchorMs` | `active_anchor_ms` | int64 | = `Target.LastActiveTs`（resident 最后一次身体活动时刻；详 §4.3）|

### 3.4 Section3.up.right（SceneState）

| 字段 | JSON | 类型 | 取值方式 |
|---|---|---|---|
| `SceneState` | `scene_state` | int | `0`=OOR / `1`=InRoom / `2`=InBath / `3`=InBed / `4`=OOB；算法见 §4.2 |
| `SceneAnchorMs` | `scene_anchor_ms` | int64 | 按 SceneState 取墙钟：InRoom/InBath→`rs.last_enter_time`；InBed/OOB→`bs.start_time`；OOR→`rs.last_exit_time` |

### 3.5 Section3.down（Visitor + Bed timing fallback）

| 字段 | JSON | 类型 | 取值方式 |
|---|---|---|---|
| `VisitorState` | `visitor_state` | int | `0`=None / `1`=Now / `2`=Today；**仅 /80 卡 AND `unit_type==1`(private) 才可能非 0**，算法见 §4.4 |
| `VisitorAnchorMs` | `visitor_anchor_ms` | int64 | VisitorState ≠ None 时 = `Target.VisitorStartTs` |
| `BedAnchorMs` | `bed_anchor_ms` | int64 | = `bs.start_time`（FE 在 VisitorState=None 时渲 "InBed/LeftBed Xm ago"，配合 Section2.BedStatus） |

**FE 渲染逻辑**：
```
if VisitorState == Now:    "Visitor now · " + (serverNow - VisitorAnchorMs)
elif VisitorState == Today: "Visitor today · " + (serverNow - VisitorAnchorMs)
else:                      // fallback bed timing
  if CardStatic.bed_id == null: "-"
  elif BedStatus == 0:           "InBed " + (serverNow - BedAnchorMs)
  elif BedStatus == 1:           "LeftBed " + (serverNow - BedAnchorMs)
  else:                          "-"
```

## 4. 决策算法

### 4.1 Section2LeftMode picker

输入：rs (winner 的 RoomState, 可能 nil)、bs (winner 的 BedState, 可能 nil)。

```
bedHas  = bs != nil && bs.updated_at > 0
roomHas = rs != nil && rs.updated_at > 0

if !bedHas && !roomHas:
    → None

elif roomHas && rs.risk_level > 0 && rs.total_people > 0:
    → RoomStatus           // Risk 强制走 RoomStatus

elif bedHas && bs.bed_status == 0:
    → SleepStage           // 在床

elif roomHas && rs.total_people > 0:
    → RoomStatus           // 房间有人

else:
    // 都无人 → recency
    if bedHas && (!roomHas || bs.updated_at >= rs.updated_at):
        → SleepStage
    else:
        → RoomStatus
```

### 4.2 SceneState picker

输入：rs, bs。

```
if roomHas && rs.total_people > 0:
    if rs.kind == "bathroom":
        → InBath,  anchor = rs.last_enter_time
    elif bedHas && bs.bed_status == 0:
        → InBed,   anchor = bs.start_time
    else:
        → InRoom,  anchor = rs.last_enter_time

elif bedHas && bs.bed_status == 1:
    → OOB,         anchor = bs.start_time

elif roomHas:
    → OOR,         anchor = rs.last_exit_time

else:
    → OOR,         anchor = 0
```

**注**：OOU 状态已删除——unit 是不是空可以由"所有子卡都 OOR"间接表达，picker 不需特别区分。`rs.last_exit_to_outside` 字段保留作 risk/alarm 的原始信号，**不参与 SceneState 派生**。

### 4.3 ActiveState 数据源（统一公式）

三类卡（/80 unit、/88 room、/96 bed）公式相同：
- `ActiveAnchorMs = Target.LastActiveTs`
- `ActiveState = (now - ActiveAnchorMs < 60_000) ? Now : Inactive`

`Target.LastActiveTs` 由 sensor 在以下条件更新：
- radar 帧自带 `walk_distance ≥ 2m` **OR** `walk_duration ≥ 6s`
- **AND** 当前 room `total_people == 1`（多人时不更新——无法区分 caregiver / resident）
- 写到 radar 所在 cardID（/88 room 或 /96 bed）的 Target.LastActiveTs

leaf 卡读自己的 Target.LastActiveTs。/80 unit 卡：picker 把 winner 的 Target.LastActiveTs 投到 /80 display。

### 4.4 VisitorState picker（仅 /80 unit private）

**前置条件**：当前卡是 /80 unit 卡 AND `unit_type == 1` (private/single)。否则永远 `VisitorState = None`。

```
if Target.VisitorStartTs > 0 AND (now - VisitorStartTs < 2h):
    → Now,   anchor = VisitorStartTs
elif Target.VisitorStartTs > 0:
    → Today, anchor = VisitorStartTs
elif Target.HasVisitorToday:
    → Today, anchor = 0
else:
    → None
```

**Visitor 检测**（sensor 侧）：
- 任一 radar 上报 `multi_person_duration ≥ 30s`（`MultiPersonDurationSec`，sensor weights.go）
- 写到该 unit /80 的 Target.VisitorStartTs（不写子卡）
- 持续 ≥ 10min 时（`VisitorMinThreshold`）累加到 Target.TodayMaxVisitorMin
- 当日午夜 reset `HasVisitorToday`

## 5. /80 UnitPicker 优先级算法

### 5.1 优先级 tuple

```
priorityFromState(rs, bs):
    if rs != nil && rs.risk_level > 0 && rs.total_people > 0:
        return P1, risk=rs.risk_level, ts=max(rs.ts, bs.ts)
    if rs != nil && rs.total_people > 0:
        return P2, risk=0, ts=max(rs.ts, bs.ts)
    if bs != nil && bs.bed_status == 0:
        return P3, risk=0, ts=max(rs.ts, bs.ts)
    return P4, risk=0, ts=max(rs.ts, bs.ts)

priorityFromDisplay(d):
    if d == nil:                          return PNone
    if d.room_risk_level > 0:             return P1, risk=d.room_risk_level, ts=d.active_anchor_ms
    if d.scene_state in (InRoom, InBath): return P2, ts=d.active_anchor_ms
    if d.scene_state == InBed:            return P3, ts=d.active_anchor_ms
    else:                                 return P4, ts=d.active_anchor_ms
```

### 5.2 决策

```
HandleChildEvent(childID, rs, bs):
    unit = derive_/80(childID)
    if unit invalid or unit == childID: return

    prev    = ReadCardStatus(unit)
    curPrio = priorityFromDisplay(prev.Display)
    newPrio = priorityFromState(rs, bs)

    if newPrio >= curPrio:
        // 高 risk 或同 risk：覆盖（同档内 latest wins）
        write_display(rs, bs, prev.Target, prev.AlarmState)
    elif newPrio.risk == 0 && unit_has_bed():
        // 跌档到 0 + unit 有床：fallback 显 sleepstage 默认
        write_sleepstage_default(prev)
    else:
        // 新 risk 低且不为 0，或 unit 无床 → 保留 cur（接受滞后）
        return
```

**注**：无 WinnerCardID 字段——picker 不区分 winner 自更 vs 挑战者，简化为纯 (risk, time) 比较。代价：winner 静默清空时 /80 display 滞后到下次活动来。可接受。

### 5.3 race 处理

`UnitPicker` 写 display Hash 子字段，`AlarmRouter` 写 alarm_state Hash 子字段——**完全不重叠**，无 race。

## 6. Producer 职责切割

| Producer | 写 Hash 字段 | 触发 |
|---|---|---|
| **AlarmRouter** | `alarm_state` | `iot:alarm:stream` |
| **SensorStateProjector** | `room_state` / `bed_state` / `target` / `display`（**leaf 卡**） | `sensor:derived:stream` |
| **UnitPicker** | `display`（**/80 unit 卡**） | SensorStateProjector 收 room.state / bed.state 后调 `HandleChildEvent` |

**单 owner 原则**：每张卡的 display 唯一写入方——leaf 是 SensorStateProjector，/80 是 UnitPicker。AlarmRouter **不触** display。

## 7. RoomState / BedState / TargetState 字段（producer 视角）

### 7.1 RoomState（在 card_types.go）

```go
type RoomState struct {
    Kind                  string // bedroom/bathroom/living/...
    UpdatedAt             int64
    TotalPeople           int    // 当前证据推断的人数（radar number_people + sleepad in_bed），不是绝对真实总人数
    LastEnterTime         int64
    LastExitTime          int64
    LastExitToOutside     bool   // raw 信号，留作 risk/alarm 用，不进 display SceneState
    StaySec               int    // 当前占用秒；sensor 累加
    StandingContinuousMin int    // 连续站立分钟（封顶 8）；sensor 累加
    RiskLevel             int    // 0/1/2/3，sensor risk_evaluator 算
}
```

已删：`HasMulti`（= TotalPeople>1 冗余）、`AreaPeople`（全栈无 reader）。

### 7.2 BedState（无字段变更）

### 7.3 TargetState（关键字段）

```go
LastActiveTs        int64  // resident 最后身体活动时刻（详 §4.3）
VisitorStartTs      int64  // 当前 visitor 进入时刻
HasVisitorToday     bool   // 当日是否曾有 visitor
TodayMaxVisitorMin  int    // 当日最长 visitor 时长
```

**注**：JSON tag 当前是 `VisitorStartTs / HasVisitorToday / TodayMaxVisitorMin`（PascalCase）——违反 owlBack CLAUDE.md 规则 #1.1（snake_case）。需独立 PR 改名为 `visitor_start_ts / has_visitor_today / today_max_visitor_min`。

## 8. 命名约定

- Go 常量：复用 `owl-common/card/card_types.go` 已定义的 `Section2LeftMode*` / `RoomIconKind*` / `ActiveState*` / `SceneState*` / `VisitorState*`、`SleepStage*`、`RiskLevel*`、`RoomKind*`
- JSON tag：snake_case，可选字段加 `omitempty`

## 9. 已知 wiring gap（producer 待修）

| 字段 | 现状 | 待修 |
|---|---|---|
| `Target.LastActiveTs` | `UpdateTargetLastActive` 函数有，**零 caller** | sensor 收 radar 帧时按 §4.3 条件调用 |
| `Target.VisitorStartTs` / `HasVisitorToday` / `TodayMaxVisitorMin` | 字段定义在，**全栈无 writer** | sensor 按 §4.4 条件累加 |
| `RoomState.StaySec` | risk_evaluator 读，**无 writer** | sensor 进房=0 起累，空房复位 |
| `RoomState.StandingContinuousMin` | risk_evaluator 读，**无 writer** | sensor 按 stand_duration ≥ 55s 累 +1，封顶 8 |

修复顺序（按当前讨论的约定）：
1. 修 producer（sensor + RoomState / Target）
2. 再修 consumer（cardagg picker / display_builder）

## 10. unit_type schema 校准（独立 task）

三处定义不一致：

| 位置 | 字段 | 值域 |
|---|---|---|
| `owlRD/dbv2/13_units.sql` | `unit_type SMALLINT` | `1`=single / `2`=share / `3`=public |
| `owlRD/dbv2/13_units.sql` | `unit_property SMALLINT` | `0`=Home / `1`=Facility |
| `owl-common/ipam/types.go` | `UnitType string` | `"residential" / "public" / "shared"` |
| `owl-common/card/card_types.go:89` | `UnitType string` | `"facility" \| "home"` |

需要：
1. 统一 schema 值（推荐 DB SMALLINT 是正源，对应 `1=single→private` 命名展示给 FE）
2. Go 侧加 `UnitTypePrivate=1` 等常量
3. cardagg DeviceMetaCache 加载 CardMeta 时一并查 `unit_type` 缓存

private 判定：`unit_type == 1`。
