# Belief 输入规范化 — 统一观测矩阵

**定位**：把 radar / sleepad / room / bed / suite 各自参差不齐的输出，规范成**统一格式的观测向量**，
作为 [[room_belief_state_machine.md]] §4 `P(o|s)` 层的唯一输入。规范化是 belief 运算的前提：
不统一格式，矩阵乘法无从谈起。

**原则**：每条观测 = 一个 `Observation` 记录，**统一七元组 schema**，不管来自哪个源。
belief 引擎只认这个 schema，不认设备类型 —— 加新传感器 = 加几行 P(o|s) 标定，不改引擎。

`revision: 1` ・ `created: 2026-05-31` ・ 字段来源已逐项实查代码(file:line 见末附录)

---

## 1 统一 Observation schema（七元组）

每个源把自己的原始输出翻译成 0..N 条这样的记录：

```go
type Observation struct {
    Source    SourceID    // 谁产出的：device /128 addr 或 room /88 或 bed /96
    Kind      ObsKind     // 观测种类（枚举，见 §2）— 决定查 P(o|s) 哪一行
    Value     float64     // 归一化数值（连续量）或枚举序号（离散量）
    Conf      float64     // 信号源可信度 [0,1]（0=缺失/stale → P(o|s)=I 不更新）
    Ts        int64       // 观测时刻 unix ms
    Fresh     bool        // 是否在 TTL 内（stale → 当作缺失）
    Geom      GeomTag     // 位置语义：InBed / InEnter / OpenFloor / InToilet / Unknown（由 grid 交集算）
}
```

**关键设计**：
- **`Conf` 统一表达"缺失/stale/弱信号"** —— 这是 belief 命门(§2.3 of belief doc)。Conf=0 →
  该观测 `P(o|s)=I`，不更新 belief。所有源的"缺失行为"(nil/0/未观测)统一映射成 Conf=0，
  引擎不再 per-source 特判 nil。
- **`Geom` 把位置从坐标降成语义标签** —— belief 的隐状态 S 不含坐标(§3 of belief doc)，
  位置只作 P(o|s) 的条件。Geom 由 device 坐标 ∩ layout polygon 现算（依赖 ⑤ 已加载几何）。
- **`Value` 连续量归一到 [0,1] 或物理单位**，离散量用枚举序号 —— 统一成 float 让矩阵能算。

---

## 2 ObsKind 枚举（belief 引擎认的全部观测种类）

把各源的原始字段**收敛**成有限的、与隐状态相关的观测种类。一个原始字段可能映射成一条
ObsKind（如 radar pose → ObsPose），多个原始字段也可能合成一条（如 Δz+位移 → ObsKinematics）。

| ObsKind | Value 语义 | 主要区分的隐状态 S | 原始字段来源 |
|---|---|---|---|
| `ObsPose` | pose 枚举序号(0-12)归一 | S1 Lying/S3 Sit/S4 Stand/S5 Fallen | radar Track.Pose |
| `ObsKinematics` | Δz↓ + 位移 + 隐含速度 合成分 [0,1] | S5 Fallen(z骤降+静止) vs S4 Walk | roomengine 派生(Kalman 速度/位移/LastZ-Z) |
| `ObsVitalPresent` | HR>0 ∨ RR>0 → 1，否则 0 | 有生命体征 → ¬Empty,¬Artifact | radar/sleepad HR/RR |
| `ObsBedOccupied` | bed 贝叶斯 P(InBed) [0,1] | S1/S2 Bed vs 其余 | **BedState.BedStatus + BedConfidence**(已是贝叶斯输出) |
| `ObsSleepStage` | stage 枚举(0/1/2/4/8) | S1 Lying 细分(睡/醒) | BedState.SleepStage(sleepad 自带);radar 有条件(install_mod=full+bed area 返 pose lying/bed_sit_up,粗粒度 conf 低) |
| `ObsFirmwareFall` | firmware pose2→5 升级 → 1 | S5 Fallen 高似然 | radar event Fall/SuspectedFall |
| `ObsEnterExit` | EnterRoom=+1/ExitRoom=-1 | S0 Empty/S7 Left 翻转 | radar event EnterRoom/ExitRoom |
| `ObsNumberPeople` | 人数(0-8) | S0 Empty(=0) vs 有人 | radar event number_people |
| `ObsStandDuration` | 连续站立分钟(0-8 cap) | S4 久站(bathroom 风险) | RoomState.StandingContinuousMin |
| `ObsTrackPresent` | track 存活 + Verdict + GhostPenalty 合成 | S8 Artifact(ghost) vs 真人 | TrackState.Verdict/GhostPenalty/Score |
| `ObsTimeContext` | 夜/昼 + room kind → prior 调整 | 调 A 的 prior 非硬观测 | time + RoomType |

**铁律**：派生信号(WeakBiometricSignal/RiskLevel/AloneContinuousMin 等**聚合/趋势量**)**不进**
ObsKind —— 它们是别的 belief/规则的**输出**，进了会双重计数(对齐
[[feedback_no_dynamic_threshold_modulation]])。belief 只吃**原始物理观测**。AloneContinuousMin
这类"已经算好的时长"由 belief 内部的状态停留时长(HSMM duration)替代，不作输入。

---

## 3 输入矩阵：每个实体能提供什么（规范化后）

横轴 = 实体类型，纵轴 = ObsKind。✓=该实体能产出此观测，空=不产出。
这张表就是"每个 Device/Room/Bed 能提供什么输入"的规范化答案。

| ObsKind \ 实体 | radar device(/128) | sleepad device(/128) | room(/88) | bed(/96) |
|---|:---:|:---:|:---:|:---:|
| ObsPose | ✓(0-12) | | | |
| ObsKinematics | ✓(派生) | | | |
| ObsVitalPresent | ✓(mmWave HR/RR) | ✓(接触式 HR/RR,更准) | | |
| ObsBedOccupied | △(radar bed_status,conf=60) | ✓(sleepad 原生,conf=90) | | ✓(bed 贝叶斯融合输出) |
| ObsSleepStage | △(有条件) | ✓(自带分类) | | ✓(SleepStage) |
| ObsFirmwareFall | ✓(pose2→5) | | | |
| ObsEnterExit | ✓(firmware event) | | △(room transition) | |
| ObsNumberPeople | ✓(0-8) | | ✓(TotalPeople 融合) | |
| ObsStandDuration | △(stat 分钟) | | ✓(StandingContinuousMin) | |
| ObsTrackPresent | ✓(Verdict/Ghost) | | | |
| ObsTimeContext | | | ✓(RoomType+tz) | |

图例：✓=主产出 / △=次级或派生 / 空=不产出。

**读法（对 belief 的意义）**:
- **radar = 运动/姿态/几何主力**，但生命体征不达临床(mmWave HR ±5-10,[[radar_hr_no_critical]])。
- **sleepad = 床上接触式真值**(HR/RR/睡眠分期准),但只覆盖床、无位置/姿态。
- **room = 聚合层**(人数/久站/时段),给 belief 提供"房间整体"约束,不是 raw。
- **bed = 已经是一个 belief 子模块的输出**(bed 贝叶斯 scorer,[[bed_bayesian_review.md]])——
  它产出的 `ObsBedOccupied` 已带概率+置信度,是 room belief 的**高质量子证据**,不是 raw 信号。
  这正是"床二态 belief → 房间多态 belief"的嵌套点:bed belief 的输出 = room belief 的一条观测。

---

## 4 缺失/可信度的统一映射（Conf 字段怎么填）

各源"缺失行为"五花八门,统一成 Conf:

| 源原始表现 | Conf | belief 处理 |
|---|---|---|
| 字段 nil / 值=0 表"无信号"(HR=0,position=nil) | 0 | P(o|s)=I 不更新 |
| 有值但超 TTL(Fresh=false,radar 1Hz 默认 4s) | 0 | 同上,当缺失 |
| radar 派生 bed_status | 0.6 | 弱观测,似然偏平 |
| sleepad 原生 / firmware 确认事件 | 0.9 | 强观测,似然尖锐 |
| track confidence / pose confidence 字段 | 字段值/100 | 直接用设备自报置信度 |
| bed 贝叶斯输出 BedConfidence(0/60/90) | /100 | 嵌套 belief 的置信度透传 |

**这一列是规范化的核心价值**:今天 census 把"track 冻住(无观测)"误当"人静止(强观测)",
就是因为没区分这两者的 Conf。统一后,冻住→Fresh=false→Conf=0→不更新→9h bug 消失。

---

## 5 实体 → Observation 翻译器(adapter 层)

每个源写一个 adapter,把自己的 struct 翻成 []Observation。引擎只见 Observation,不见原始 struct。

```
radarAdapter(Track, TrackState, grid)  → []Obs{ObsPose, ObsKinematics, ObsVitalPresent,
                                                ObsTrackPresent, ObsFirmwareFall?, ObsEnterExit?}
sleepadAdapter(SleepadObservation)     → []Obs{ObsVitalPresent, ObsBedOccupied(conf .9), ObsSleepStage}
bedAdapter(BedState)                    → []Obs{ObsBedOccupied(P+conf), ObsSleepStage}  ← 嵌套 belief 输出
roomAdapter(RoomState)                 → []Obs{ObsNumberPeople, ObsStandDuration, ObsEnterExit}
timeAdapter(now, RoomType)             → []Obs{ObsTimeContext}
```

adapter 负责:① 填 Geom(device 坐标 ∩ layout polygon);② 填 Conf(按 §4);③ 判 Fresh(TTL)。
**引擎下游完全统一** —— 拿一批 Observation,逐条按 ObsKind 查 P(o|s) 行,更新 belief。

---

## 6 下一步

1. 定 `ObsKind` 最终枚举 + 每个的 `P(o|s)` 标定行(对每个隐状态 S 的似然) —— 这是
   [[room_belief_state_machine.md]] §9 第 1 步"gate→矩阵条目对照表"的输入侧。
2. 写 5 个 adapter(§5),先只产 Observation 打日志,不接 belief(shadow 前的前置)。
3. 用 CABB / John.Y case 的真实流,验证 adapter 产出的 Observation 序列符合预期
   (尤其 John.Y 9h:D523 track 冻住后 ObsPose/ObsKinematics 的 Conf 应转 0)。

---

## 附录 A：字段来源(实查 file:line)

**radar Track**: `owl-common/observation/track.go:6-34`(Track struct) +
`owl-common/observation/fields.go:232-246`(Pose 枚举) +
`wisefido-sensor/internal/roomengine/track_parse.go:19-93`(ParseRadarTracks)。
**radar event**: `wisefido-qinglan/internal/decode/radar_decoder.go:486-623`(进出/姿态/人数/设备状态) +
`:323-416`(activity stat: walk/stand/lie_duration)。
**sleepad**: `wisefido-sensor/internal/roomengine/sensor_fusion.go:27-103`(SleepadObservation:
InBed/HR/RR/BodyMove/TurnOver) + `:105-155`(SleepadBedEvent)。
**TrackState 派生**: `wisefido-sensor/internal/roomengine/track.go:53-186`(Score/Verdict/GhostPenalty/
LastZ/FrozenRunStart/MaxImpliedSpeedFromBirth/LongSurvivalAnchored)。
**BedState**: `owl-common/card/card_types.go:258-302` ←
`wisefido-sensor/internal/zoneengine/bed_bayesian_scorer.go:89-140`(贝叶斯 scorer) +
`translator.go:22-62`(TranslateBedState);debug `types.go:112-116`(LogOdds/Prob/Gamma)。
**RoomState**: `owl-common/card/card_types.go:379-435` ← `zoneengine/translator.go:88-98` +
`stream_publisher.go`(applyAloneAndRisk: AloneContinuousMin/StandingContinuousMin/RiskLevel)。
**SuitePerson/Census**: `wisefido-sensor/internal/roomengine/suite_census.go:40-74`
(PersonID/Role/AnchorTrackID/AnchorRoomType/LastActiveMs/SleepadAnchored/BathroomCount)。
**静态 layout**: `wisefido-sensor/internal/roomengine/engine.go:29-114`(RoomConfig:
RoomType/Beds/Enters/EnterTargets/Toilets/Radars/RadarAddrs/Sleepads)。

## 附录 B：归一化前的"参差不齐"实例(为什么必须规范化)

| 问题 | 实例 | 规范化解法 |
|---|---|---|
| 缺失表达不一 | radar HR=0 表无信号 / position=nil 表缺失 / bed_status=nil 表未知 | 统一 → Conf=0 |
| 可信度不可比 | radar bed_status vs sleepad InBed 强度天差地别,但都叫"InBed" | 统一 → Conf 0.6 vs 0.9 |
| 同语义多源 | radar/sleepad/bed 三个都报 InBed,格式各异 | 统一 → 都映射 ObsBedOccupied,Conf 区分 |
| 离散/连续混杂 | pose 是枚举,位移是 cm,人数是计数 | 统一 → Value float + ObsKind 标类型 |
| 派生混进原始 | RiskLevel/AloneContinuousMin 是算出来的却和 raw 平级 | 派生**禁入** ObsKind(§2 铁律) |
| 位置形态不一 | 有的给坐标,有的给 area_type 字符串,有的给 cell | 统一 → Geom 语义标签(grid 交集算) |
