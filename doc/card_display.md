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
│ Section3.down.left          Section3.down.right│
│ Visitor 或 Bed timing       VitalTrend 横条    │
│ fallback                    (WeakBio 风险描述)  │
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
| Section3.down.left (Visitor 或 Bed timing) | display.`visitor_state` + `visitor_anchor_ms` + `bed_anchor_ms` | **VisitorDeriver (cardagg)** 写 /88 room card (visitor 仅父 unit Private) / SensorStateProjector (bed timing leaf) | **是** |
| Section3.down.right (VitalTrend 横条) | display.`vital_trend_level` (0/1/2/3) | cardagg `card_display_builder.pickVitalTrendLevel` 从 `Target.WeakBiometricSignal` 阈值派生（§3.5b） | **是** |

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

### 3.5 Section3.down.left（Visitor + Bed timing fallback）

| 字段 | JSON | 类型 | 取值方式 |
|---|---|---|---|
| `VisitorState` | `visitor_state` | int | `0`=None / `1`=Now / `2`=Today；**仅 /88 room 卡 AND 父 unit `unit_type==1`(Private) 才可能非 0**，算法见 §4.4 |
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

### 3.5b Section3.down.right（VitalTrend 横条 — WeakBio 风险描述符）

| 字段 | JSON | 类型 | 取值方式 |
|---|---|---|---|
| `VitalTrendLevel` | `vital_trend_level` | int | `0`=None (hide) / `1`=Gray (Attention) / `2`=Yellow (Watch) / `3`=Red (Alert)；从 `Target.WeakBiometricSignal` 阈值派生 |

**阈值映射**（详 [[target_state_weak_bio_signal_design]] §"阈值表"）：

| WeakBio score | VitalTrendLevel | 横条配色 |
|---|---|---|
| 0-29  | 0 None    | 不显示 |
| 30-59 | 1 Gray    | 灰（Attention） |
| 60-79 | 2 Yellow  | 黄（Watch） |
| 80-100| 3 Red     | 红（Alert） |

**staleness 守护**（W3）：
- cardagg `target_merger` 在 merge WeakBio 时按 **offline 过滤 + 30min UpdatedAt staleness**（贴 sensor aggregator lazy 滑窗 weakBioWindowMs）
- 空闲老人卡 30min 后 score 自然回 0（避免 sensor 不再 push 时 cardagg 卡老值显示 Red 横条）

**FE 渲染逻辑**：
```
if VitalTrendLevel == 0:  hide 横条
elif == 1:                bar bg=#888  // gray
elif == 2:                bar bg=#FFC107  // yellow
elif == 3:                bar bg=#F44336  // red
```

**不独立触发 alarm**：score ≥80 仅作风险描述符；后续由风险放大消费者（cardagg AlarmRouter / sensor fall verifier / zonealarm Supervisor）读 `Target.WeakBiometricSignal` 在 Warn alarm 上提级 Critical 或缩短阈值（详 weakBio 设计 §"风险放大消费者"）。

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

### 4.3 ActiveState 数据源 — per-device TargetState 模型（v2 拍板 2026-05-18）

**架构**：
- Sensor 按 **per radar device (/128)** 维护 TargetState；不按 room/card 合并
- Cardagg 按 card 组成（card 包含的 device 列表）汇聚多 device → 单 card 显示值
- v3 未来引入 logicID 实现 cross-radar 真融合；v2 暂以 max(deviceTarget) 近似

**Sensor 端（per radar device）**：
- 每条 radar 维护自己的 TargetState 实例
- `TargetState.LastActiveTs` 更新条件（per device）：
  - radar 帧自带 `walk_distance ≥ 2m` **OR** `walk_duration ≥ 6s`
  - **AND** 该 radar 所在 room `total_people == 1`（多人时不更新——无法区分 caregiver / resident）
  - 60s 节流（同 device 已 60s 内更新过则跳过 Redis IO）
- Sensor 发布到 `sensor:derived:stream` per device，category=`target.state`，key=device_addr (/128)

**Cardagg 端（per card 合并）**：
- 维护 per device target snapshot in-memory map
- card.lastActiveTs = **max(deviceTarget.LastActiveTs across all online radars in this card)**
- card:state.target Hash 写**合并后**的单 `TargetState`（FE 看到的还是 1 个）
- 显示规则：`ActiveState = (now - card.lastActiveTs < 60_000) ? Now : Inactive`

**Offline 过滤（2026-05-19 落地）**：
- merge 时跳过 offline device 的 snapshot
- 避免"device 失联 24h，FE 永远显示 Active 24h ago"误导
- 通过 `TargetMerger.SetOnlineChecker(deviceTracker.IsOnline)` 注入 callback
- LastActive 不加时间阈值——历史时间戳是真实事实，配合 SceneState 让 nurse 正确解读（"InRoom + Active 25min ago" = "在房静止 25min"，可能正常可能危险，结合其他线索）

**护士需求语义（2026-05-19 拍板）**：
- 旧含义"resident 步行时刻"过严——多人场景下 sensor `total_people==1` gate 不更新，丢失"房间有人能动 → 互助能力"的安全信号
- v3 改语义为"房间内任一人最后一次步行时刻"（去 single-person gate；visitor 归因走 [[unit_visitor_attribution]]，未来 bed-bound radar 检测床主访客）
- v2 阶段 sensor 仍 single-person gate（FollowUp 暂不动）

**示例（同卡 2 个 radar）**：

| 时刻 | 事件 | R1.lastActive | R2.lastActive | card.lastActive = max | card 显示 |
|---|---|---|---|---|---|
| T=0 | R1 active | T=0 | 0 | T=0 | active now |
| T=3m | R2 active | T=0 (停) | T=3m | T=3m | active now |
| T=5m | R2 停 / R1 续 | T=5m | T=3m | T=5m | active now |
| T=10m | R1 停 | T=10m | T=3m | T=10m | active 0s ago |
| T=12m | （无更新）| T=10m | T=3m | T=10m | active 2min ago |

### 4.3a StandingContinuousMin（per-device，cardagg 合并）

**已不在 RoomState**（schema 2026-05-19 已迁），移到 `TargetState.StandingContinuousMin`（per radar device）：
- Sensor per radar：`stand_duration ≥ 55s` + 该 radar 所在 room `total_people == 1` → +1（封顶 8）
- 坐/走/躺 / 多人 / 空房 → reset 0
- Cardagg：card.standingMin = max(deviceTarget.StandingContinuousMin across online + fresh radars)
- risk_evaluator 不再读 RoomState.StandingContinuousMin —— 改读 card 合并后的（待 cardagg 改造）

**Cardagg 双重过滤（2026-05-19 落地）**：
- **offline**：跳过 offline device snapshot
- **2min UpdatedAt staleness**：snapshot 老于 2min 视为 push 异常，不参与 max
- 必要性：standing 是瞬时态，stale 直接进 risk_evaluator → false Attention/Risk → 误报 alarm

**Sensor 心跳契约（必约束）**：sensor 必须每分钟 push 一次 standing 值（即使无变化），让 cardagg 2min 阈值不误判正常持续站立为 stale。FollowUp-4 sensor `handleMonitorFrame` 实施时必须保证。

**v3 双 radar AND 规则（follow-up）**：
- 用户拍板（[[standing_dual_radar_and_rule]]）：同 card 多 radar 都 standing>0 才信 max；任一报 0 → 整体 0
- 物理前提：同一老人不可能在两个 radar 下一个动一个静；冲突说明信号有问题
- 但 v2 区分不了"sit"和"视野盲区"（snapshot 只有 standing 一个数值），AND 规则会导致盲区配置漏报
- 推后到 v3 sensor logicID 真融合 + device-level visibility 信号时再做；v2 阶段保持简单 max
- 同 room 多 radar 在 v2 是小概率配置，max 偏激进可接受

### 4.3b WeakBiometricSignal — 风险描述符（2026-05-19 设计修订）

**核心修订**：WeakBio score 是**长期状态 / 风险放大系数**，不是事件源——**不独立触发 alarm**。

理由：firmware 已经直发 HR/RR/ApneaH/WeakBio raw alarm 进 alarm pipeline；二次造一个 escalation Critical 是冗余且不可降级（lazy 设计 + 上升沿 → 永远要 ack）。WeakBio 类比温度计：38℃ 不会触发"高温警报"，温度是**上下文**让其他判断更精准。

**计算公式（修订权重）**：

```
score = max(weakBio_raw_in_30min_window)    # 0..60 (state×20)
      + count(HeartRateAlert)   × 5         # 旧 15 → 新 5（HR/RR 单条噪声居多）
      + count(RespRateAlert)    × 5         # 旧 15 → 新 5
      + count(ApneaHypopnea)    × 15        # 旧 25 → 新 15（单条仍有意义但不应单条触红）
final = min(100, score)
```

**临床校准依据**（AASM AHI 标准 5/15/30 = 轻/中/重）：
- 30min 内 6 次 ApneaH = AHI≈12 = 轻度 → score 90 → 红 ✓
- 30min 内 4 次 ApneaH = AHI≈8 = 临界轻度 → score 60 → 黄 ✓
- 12 次 HR alert (持续真异常) → score 60 → 黄 ✓
- 单条 ApneaH = 15 → 不显示（避免单条噪声）

**FE 横条阈值（Section3.down.right，Step2 实施）**：

| score | UI 横条 |
|---|---|
| 0-29 | 不显示 |
| 30-59 | 灰（Attention）|
| 60-79 | 黄（Watch）|
| ≥80 | 红（Alert）|

**风险放大消费者**（待 follow-up）：
- cardagg AlarmRouter 收 Warn alarm 时查 card.Target.WeakBio，超阈值提级 Critical
- sensor roomengine fall verifier：WeakBio≥80 → 跳过 verifier 直接 Confirmed
- sensor zonealarm Supervisor：WeakBio≥80 → 缩短 LeftBed/Stay DurationSec

**Cardagg merge 待做（Step2）**：offline 过滤 + 30min UpdatedAt staleness（贴 sensor 滑窗）。

**注**：v2 max-merge 违反 VitalWeight 加权意图（Sleepad=8 > Radar=4），但因为不再触发 alarm，false positive 代价从"误叫救护车"降到"FE 横条假红"——可接受；v3 vital 融合时再上加权策略（FollowUp-1）。

### 4.4 VisitorState 算法（cardagg VisitorDeriver 写；2026-05-19 加 bed-bound radar 路径）

**架构原则**（2026-05-18 拍板）：
- Sensor 不做 visitor 跨 room 合并 —— sensor 只发 RoomState (含 TotalPeople)，per /88 物理实体
- 跨实体合并是 card 层职责 → **cardagg `VisitorDeriver` 模块负责**
- 不同 room 触发的 visitor 各自独立 tracking

**双路径判定矩阵（2026-05-19 拍板）**：

| unit_type | room level (/88 room card)| bed level (/96 bed card with bed-bound radar) | 净结果 |
|---|---|---|---|
| Private (1) | ✅ 兜底 | ✅ 优先（写 /96 bed card） | 双路径；bed level 优先 |
| **Share (2)** | ❌（多 resident 常态）| ✅ 唯一路径（解锁 share）| 仅 bed level |
| Public (3) | ❌ | ❌ | 跳过 |

bed level 解锁 Share unit visitor 归因——以前 Share 完全跳过（"多人是常态无法判定"），现在 bed-bound radar 视野受 firmware boundary 物理限制到该床区域，看到 2 人=床主+访客。

**bed-bound radar 判定**：
- cardagg 端判定：`card.CardType == "bed" AND card.Devices` 含 `device_type == "Radar"` 的 device
- 物理基础：device IPv6 绑到 /96 bed prefix 即视为 bed-bound radar
- **firmware 自带 boundary 物理裁剪**——share 双床各自 boundary 设到自己床区，cardagg 不查物理安装、只信 IPAM 绑定 + firmware 部署纪律

**优先级冲突处理**：
- bed level 触发了某 /96 bed card visitor → 对应 /88 父 room card 本轮跳过（避免父子双显）

**写入位置**：
- bed level 触发 → `/96 bed card` target
- room level 触发 → `/88 room card` target

**FE 渲染**（统一逻辑，per card 读自己 target）：

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

**Visitor 检测算法（cardagg VisitorDeriver 60s tick）**：

```
for each card in tracked cards:
    case 1: bed level （/96 bed + bed-bound radar）
        peopleCount = bed-bound radar 视野内 people count（per-bed in-memory tracker）
    case 2: room level（/88 + parent unit_type == Private）
        peopleCount = card:state.room_state.total_people
    
    if peopleCount >= 2:
        segment_duration_min += 1
        if segment_duration_min >= 5:
            Target.VisitorStartTs = segment_start_ts
            Target.HasVisitorToday = true
            Target.TodayMaxVisitorMin = max(prev, segment_duration_min)
    else:
        segment_duration_min = 0
    
midnight reset (parent unit timezone): 清三字段 + segment_start_ts
```

**bed level 数据流**（2026-05-19 修订：number_people 走 event 流不是 monitor 流）：

```
radar firmware type=3 NumberPeople event
     ↓
qinglan radar_decoder.go: buildPeopleNumber → category="number_people"
     ↓
iot:event:stream
     ↓
cardagg EventHandler（filter category=number_people）
     ↓ Update(deviceAddr, count, ts)
cardagg BedPeopleTracker per-device snapshot（仅 offline 过滤；无时间窗 staleness）
     ↓
VisitorDeriver 60s tick：ListBedCardsWithBedBoundRadar → CardPeopleCount
     ↓
判定 + 写 /96 bed card target；记录 parent room 跳过本轮
```

**已接受的边界限制**：

| 边界 | 处理 |
|---|---|
| 视野中央盲区漏报（visitor 站两床之间）| 接受；5min 阈值过滤短时——visitor 真长时间停留就会走到某床边触发 |
| segment 不连续（访客离开 bed 几分钟又回来）| 严格 5min 连续；不做间断容忍——短时离开即视作 visit 结束 |
| 路过他床 / 巡房 | 5min 阈值天然过滤 |
| number_people 仅在变化时上报（不是周期心跳）| BedPeopleTracker 不做时间窗 staleness——静态 2 人不变 firmware 就不再报，加 staleness 反误 reset 真实 visitor；仅靠 IsOnline 过滤 device 失联 |

**Sensor 侧**：完全不参与 visitor 累加 —— 只通过现有 RoomState publish（/88 total_people）喂给 cardagg；
radar number_people 由 qinglan 直发 iot:event:stream，cardagg 自己订阅消费。

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

### 7.3 TargetState（per-device, v2 拍板 2026-05-18）

**Sensor 发布粒度**：per radar device（key=/128 device_addr），每条 radar 独立一条 TargetState 消息发到 `sensor:derived:stream` (category=`target.state`)。

```go
type TargetState struct {
    TrackID               int    // radar 当前 active track (per device 维度)
    LogicID               string // 跨 radar 融合 ID 占位（v3 上线）
    LastActiveTs          int64  // per device, sensor 维护（§4.3）
    StandingContinuousMin int    // per device, sensor 维护（§4.3a）— ⭐ 已从 RoomState 移到这
    WeakBiometricSignal   int    // per device, sensor 维护（30min 滑窗，详 [[target_state_weak_bio_signal_design]]）
    VisitorStartTs        int64  // ⭐ 现归 cardagg `VisitorDeriver` 写到 /88 room card target（§4.4）
    TodayMaxVisitorMin    int    // 同上
    HasVisitorToday       bool   // 同上
}
```

**JSON tag**：snake_case（2026-05-18 Task 2.5 落地 commit `8e8732e`）。

**单 vs 多**：
- Sensor 输出：每个 radar device 一条 TargetState（同 unit 多 radar → 多条消息）
- Cardagg 入库：per card 合并多 device target → 单 `CardStatus.Target`（FE 看到的还是 1 个）
- 合并规则：LastActiveTs/StandingContinuousMin/WeakBiometricSignal 全部取 max(across devices in card)

**Visitor 三字段例外**：
- 不由 sensor 写
- cardagg `VisitorDeriver` 在 60s tick 时**直接写**对应 /88 room card 的 target.visitor_*（不走"sensor 发→cardagg 合并"路径）
- 详 §4.4

## 8. 命名约定

- Go 常量：复用 `owl-common/card/card_types.go` 已定义的 `Section2LeftMode*` / `RoomIconKind*` / `ActiveState*` / `SceneState*` / `VisitorState*`、`SleepStage*`、`RiskLevel*`、`RoomKind*`
- JSON tag：snake_case，可选字段加 `omitempty`

## 9. 进度跟踪（2026-05-19 更新）

### 9.1 已完成

**Sensor 侧**：
| 项 | 状态 |
|---|---|
| `RoomState.StaySec` 即时累加（translator） | ✅ commit `8e8732e` |
| `StandingContinuousMin` schema 从 RoomState 挪到 TargetState | ✅ 2026-05-19 |
| `risk_evaluator` 签名加 standingMin 参数（解耦 RoomState 字段）| ✅ 2026-05-19 |
| sensor 清除 card 知识（删 device_meta.go / state_service.go / derive_helpers.go）| ✅ 2026-05-19 |
| sensor `alarm_enablement.go` 改 spatial_config LPM（不依赖 card 概念）| ✅ 2026-05-19 |
| AlarmBackChannel 加 enablement gate（所有 alarm 出口源头查使能）| ✅ 2026-05-19 |
| sensor stream_publisher 去 ReadCardStatus + translator 去 prev 参数 | ✅ 2026-05-19 |
| sensor `TargetStateAggregator.handleAlarmEvent` WeakBio 累加（30min 滑窗 + 修订权重 5/5/15）| ✅ 2026-05-19 |
| WeakBio 砍 escalation alarm 路径（改风险描述符模型）| ✅ 2026-05-19 |

**Cardagg 侧**：
| 项 | 状态 |
|---|---|
| `SensorStateProjector` bed.state 字段级 merge（保留 SleepStage / TrackNumber 等非 sensor owner）| ✅ 2026-05-19 |
| `TargetMerger` per-device → owning card max-merge（LastActive / Standing / WeakBio）| ✅ 2026-05-19 |
| `VisitorDeriver` 新模块（60s tick + Private gate + 5min 阈值，§4.4）| ✅ 2026-05-19 |
| main wiring：TargetMerger + VisitorDeriver + cardChange invalidate | ✅ 2026-05-19 |
| LastActive offline 过滤（`DeviceStatusTracker.IsOnline` 注入）| ✅ 2026-05-19 |
| Standing 双重过滤（offline + 2min UpdatedAt staleness）| ✅ 2026-05-19 |
| Visitor-v2 bed-bound radar 路径（Share unit 解锁；`BedPeopleTracker` + `EventHandler` 订阅 iot:event:stream NumberPeople + `VisitorDeriver` 双路径）| ✅ 2026-05-19 |

### 9.2 待做（Step2）

| 项 | 内容 |
|---|---|
| FE 横条 | `CardDisplay.VitalTrendLevel int` 字段 + `card_display_builder` 派生 30/60/80 阈值 |
| Section3.down 布局拆 left/right | doc §1 ASCII + §3.5 更新 |
| WeakBio merger staleness | offline + 30min UpdatedAt（贴 sensor 滑窗）|

### 9.3 Follow-up backlog

**Sensor 端**：
- `TargetStateAggregator` 订阅 iot:alarm:stream wire 接上（PushAlarmEvent 当前无 producer）
- `handleMonitorFrame` 填实 LastActive / Standing 累加 + **每分钟心跳 push 契约**
- `StreamPublisher.tickPullAndPublish` 真发 target.state（当前 stub）
- vital 融合（Q1=B）：sensor 订阅 SleepStage event + VitalWeight 加权写 BedState.SleepStage；BedState 4 个非 sensor owner 字段当前全栈 0 writer
- WeakBio max-merge → VitalWeight 加权（合到 vital 融合 PR）
- WeakBio sensor 端补主动 expire push（30min 滑窗到期主动 push score=0）或每分钟 tick recompute

**Cardagg 端**：
- pending_arm/cancel sentinel 拆到专用流（违反 CLAUDE.md 规则 #2.1）
- VisitorDeriver 直接写 hash（当前依赖下次 target.state 触发，stale 1 tick）
- `mergeForCard` 锁内调 GetOrLoad（lock-on-IO 重构锁外）
- WeakBio AlarmRouter 提级（风险放大系数消费）
- VisitorDeriver 单测（接口化 metaCache/reader + mock）

**跨模块**：
- alarm_device flow invalidate 信号传到 sensor enablement（当前 cardagg 投到自己 enablement，sensor 不知）
- v3 同 room 双 radar Standing AND 规则（需 sensor 区分 sit vs 盲区）

**遗留破损**：
- roomengine pre-existing test：bathroom_fall_test / bedroom_fall_test / ghost_adjudicator_test / public_bathroom_test 在 v2 cutover 时遗留 import / field 不匹配，独立 PR 修

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
