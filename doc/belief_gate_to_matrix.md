# Gate → 矩阵条目对照表（room_belief_state_machine §9 第 1 步）

**定位**：把 [[fall_rule_inventory_and_conflicts.md]] 清点的 **≈100 条 gate-list 规则**，逐条映射到
belief 模型里的**一个矩阵条目**——是 `P(o|s)` 的某一行、`A` 的某一条转移、决策层阈值，还是
**被单一 b 直接消灭**。这是 §6"确定性角点先复现现状再标定"的输入侧账本。

**配套代码**：`wisefido-sensor/internal/roomengine/belief/`（本 PR 新包，shadow-only）。
表里 "→ likelihood.go" / "→ model.go" 指向骨架里已落的条目。

**口径**：规则编号沿用 inventory（V/L/BA/BE/G/F/BD/ZS/SI/C/D）。一条 gate 可映射多条目。
`revision: 1` ・ `created: 2026-05-31`

---

## 0 三种去向（每条 gate 必落其一）

| 去向 | 含义 | 代码落点 |
|---|---|---|
| **P(o\|s)** | gate 本质是"某观测 → 偏向某状态"，变成似然行 | `likelihood.go` rawLikelihood |
| **A** | gate 本质是"状态怎么随时间演化/不演化" | `model.go` transitionPropensity |
| **决策层** | gate 本质是"信念到某程度才动作"（阈值/代价） | `belief.go` thFire/thUncertain/Decide |
| **消灭** | gate 是"为协调两套时钟/两条规则而存在的 hack"，单一 b 后不需要 | — |

**消灭** 是最重要的一类：inventory §D 的"四套独立时钟"派生出的协调规则（L12/C13/multiSourceLeftBed
sticky 等）在单一 b 下整类蒸发——相反证据在同一个 b 上加权，贝叶斯自动仲裁，不需第三条规则。

---

## 1 ObsKind → P(o|s) 标定行（§4 观测模型，已落 likelihood.go）

每格 = 似然权重（中性 1.0，>1 偏好 <1 压制）。Conf 退火 w^Conf，Conf=0→全 1.0=不更新（命门）。
Geom 作条件（pose=Lying@InBed vs @OpenFloor 完全相反）。

| ObsKind \ S | Empty | BedLying | BedRestless | Sit | StandWalk | **Fallen** | Trans | Left | Artifact |
|---|--:|--:|--:|--:|--:|--:|--:|--:|--:|
| Pose=Walking | | 0.3 | | | **6** | 0.3 | | 3@Enter | |
| Pose=Standing | | | | 1.5 | **6** | 0.4 | | | |
| Pose=Sitting | | | | **6** | 0.8 | | | | |
| Pose=Fallen@OpenFloor | | | | 0.3 | 0.3 | **10** | | | |
| Pose=Fallen@InBed | | **4** | | | | 1.5↓ | | | |
| Pose=Lying@InBed | | **6** | 3 | | | 0.3 | | | |
| Pose=Lying@OpenFloor | | 0.5 | | | 0.4 | **4** | | | |
| Kinematics(fall sig f) | 1-.5f | 1-.5f | | | 1+1.5(1-f) | **1+7f** | | | |
| VitalPresent | **0.2** | | | | | | | | **0.3** |
| BedOccupied(p) | 1-.5p | **1+5p** | 1+3p | 1-.5p | 1-.6p | **1-.7p** | | | |
| FirmwareFall | 0.3 | 0.3 | | | 0.3 | **10** | | | |
| EnterExit=+1 | **0.2** | | | | 2 | | | 0.3 | |
| EnterExit=−1 | 2 | | | | 0.5 | **0.2** | | **8** | |
| NumberPeople=0 | **6** | 0.3 | 0.3 | 0.3 | 0.3 | **0.3** | | 2 | |
| StandDuration@Toilet | | | | | | **1+.4·min** | | | |
| TrackPresent(ghost g) | | | | | 1-.5g | 1-.5g | | | **1+6g** |
| Neighbor(occ n) | **1+3n** | 1-.4n | | | 1-.4n | **1-.7n** | | 1+2n | |

读法：**进 Fallen 的灵敏度全在这张表**（FirmwareFall×10 / Kinematics×8 / Fallen@OpenFloor×10），
A 的 →Fallen 条目刻意极小（见 §2），不靠时间转移造跌倒。

---

## 2 A 转移条目（§5，已落 model.go）

| A 条目 | 值（行归一前倾向） | 编码的物理常识 | 对应 gate |
|---|--:|---|---|
| Fallen→Fallen | 92 | **不自愈**：倒地不会自己变回站立 | L3/BA4/BE2 lost-fall"持续报"前提 |
| Fallen→StandWalk | 0.5 | 极小：起身必经 Transition | — |
| StandWalk→**Fallen** ★ | **0.2** | 跌倒主入口，但**极小**——造跌倒交给观测 | V18/still-box 的"久站可疑"软化版 |
| BedLying→Fallen ★ | 0.05 | 跌床极罕见 | L10/BE1 bedside 的转移先验 |
| Left→Empty | 85 | 离门后收敛空房 | L5 ExitRoom 清 pending 的状态版 |
| StandWalk→Left | 3 | 行走可能正离场（配 Pose@Enter 似然） | L1/checkLostFall 几何门 |
| Artifact→Empty | 30 | ghost 逐步消散 | V6/G2/G3 ghost 判定后的归宿 |
| Empty→Empty | 90 | 空房维持，进入靠观测 | V1 startup grace 的状态版 |
| *@all→Transition | 4-33 | 过渡枢纽吸收瞬时不确定 | V12/birth Pending 中间态 |

**★ 极小是治本核心**：inventory 冲突 3（9h bug）与 CABB 误报的共同机制 = "缺证据时系统继续
推断倒地"。A 的 →Fallen 给小值后，纯 Predict（缺证据期）不会把 b 渗进 Fallen，从沉默里
**构造上无法凭空造跌倒**。

---

## 3 逐族 gate → 条目（≈100 条）

### 3.1 Track verdict V1–V18（track_manager.go）

| gate | 去向 | 条目 |
|---|---|---|
| V1 startup grace 反 ghost | A | Empty→Empty 自持 + prior（启动期信念中性偏空，不急判 ghost） |
| V2 BirthScore / V9 score 判定 | P(o\|s) + 决策层 | ObsTrackPresent ghost-ness → Artifact；阈值并入 thUncertain |
| V3 no_enter_pair −60 / V4 enter_pair +20 | P(o\|s) | ObsEnterExit=+1（有配对→压 Empty/Artifact，抬 StandWalk） |
| V5 盲区返回判 Real | **消灭** | 不需特例：新观测自然抬真人态，无需"绕过 ghost 检查"的 hack |
| V6/V8 Penalty≥80→Ghost | P(o\|s) → Artifact | ObsTrackPresent ghost g 高 → Artifact×(1+6g) |
| V7 LongSurvivalAnchor 5min 豁免 | **消灭** | 冲突 2 的一根支柱；anchor 是"信念锁死"的 hack，b 持续可被新证据翻 |
| V10 later_born_with_real_track | P(o\|s) | ObsTrackPresent（多 track 时晚生 ghost-ness 高） |
| V11 mirror pair +60 | P(o\|s) → Artifact | 镜像对称 → ObsTrackPresent ghost 分量（v1 由 adapter 合成进 ghost-ness） |
| V13 Kalman birth-coherence | P(o\|s) | 并入 ObsTrackPresent（运动学不自洽→ghost） |
| V14 frozen 检测 / V15 still box | **命门** | 冻结 → adapter 置 Fresh=false → Conf=0 → 不更新（不再当活观测） |
| V16 pose-mismatch / V17 Z 噪声 / V18 avgSpeed 低 | P(o\|s) | ObsKinematics（运动学矛盾压低真人态，抬 Artifact） |

### 3.2 Lost / silent / bedside L1–L12（track_manager.go）

| gate | 去向 | 条目 |
|---|---|---|
| L1 checkLostFall 几何门（离门>30cm 且 cell≠Enter） | A + P(o\|s) | StandWalk→Left + Pose@GeomInEnter 抬 Left（门口丢=离场） |
| L2 lost pending 入池 / L3 fire | 决策层 | 不再有"pending 池"；P(Fallen)>thFire 即报，靠观测非计时 |
| L4 frozen credit 折抵 | **消灭** | 计时折抵是 gate 时钟 hack；b 无计时器 |
| L5 ExitRoom 清全部 pending | P(o\|s) → Left | ObsEnterExit=−1 抬 Left×8、压 Fallen×0.2（**冲突 4 在此消失**，见 §4） |
| L6 新 track 出生取消 | **消灭** | 新观测自然更新 b，无需显式 cancel |
| L7 实时多人取消 | 决策层（v2 中耦合） | ObsNumberPeople≥2 + 守恒（v1 弱耦合近似） |
| L8/L12 silent gate + RadarInBed 双源一致 | **消灭** | L12 是冲突 1 的协调规则（跨两时间轴）；单一 b 不需要 |
| L9 silent_fall fire（LeftBed 后矛盾+超时） | 决策层 + 命门 | 缺证据 Conf=0，P(Fallen) 不起来 → **9h bug 治本** |
| L10 bedside_fall R4 | A + P(o\|s) | BedLying→Fallen★ + Pose=Lying@OpenFloor |
| L11 bedside dedup | **消灭** | 单一 b 无双报路径 |

### 3.3 Bathroom fall BA1–BA11（bathroom_fall.go）

| gate | 去向 | 条目 |
|---|---|---|
| BA1/BA2 10a StillFall（risk 10/normal 12min） | P(o\|s) + A | ObsStandDuration@Toilet 抬 Fallen；时长门 v1 硬阈值，**v2 HSMM 状态时长替代** |
| BA3 10b BedsideFall（8min+grace） | 决策层 | 时长 → HSMM duration（v2） |
| BA4 10c LostFall 强档（0 track≥30s） | P(o\|s) + 命门 | Geom=InToilet 缺证据；不再凭"0 track"硬报 |
| BA5 10d LostFall 弱档（ghost 滞留≥7min） | P(o\|s) → Artifact | ObsTrackPresent ghost |
| BA6-11 阈值常量 | 决策层/A | 并入 thFire 与 duration gate |

### 3.4 Bedroom fall BE1–BE7（bedroom_fall.go）

| gate | 去向 | 条目 |
|---|---|---|
| BE1 11b BedsideFall（夜 15min） | A + 决策层 | BedLying→Fallen★ + risk-time prior（ObsTimeContext） |
| BE2 11c LostFall cell-typed（Bed 2h/Sit 30m/...） | A duration（v2 HSMM） | cell type → Geom；时长 → 状态停留分布 |
| BE3 LastActiveMs==0 抑制 | **命门** | 缺证据=Conf 0 不更新，天然不报 |
| BE4 silentSec idle 抑制 | **消灭** | 计时器抑制 hack 不需要 |
| BE5/BE6 isHumanBedAt / hasActiveBedSession 抑制 | P(o\|s) → BedOccupied | sleepad/bed 贝叶斯 P(InBed) 抬床态、压 Fallen（**嵌套 belief**） |
| BE7 sole resident 守卫 | 决策层（v2 多人/守恒） | v1 单实体默认成立 |

### 3.5 Ghost adjudicator G1–G5（bathroom_ghost.go）

| gate | 去向 | 条目 |
|---|---|---|
| G1 距 entry 近降嫌疑 | P(o\|s) | Pose@GeomInEnter（门口=真人入口先验） |
| G2 motion_symmetry / G3 mirror | P(o\|s) → Artifact | ObsTrackPresent ghost-ness（adapter 合成镜像分） |
| G4 IsPublicBathroom 降级 | 决策层 | 房型 prior（ObsTimeContext / 配置位） |
| G5 entry 盲区不可信 | P(o\|s) Conf | 盲区 → Geom Unknown + Conf 低（似然偏平） |

### 3.6 Fall verifier F1–F8（fall_verify.go）

| gate | 去向 | 条目 |
|---|---|---|
| F1 baseline 50 / F2 ghost<30 / F3 real≥70 / F4 suspect | 决策层 | score 档 → thFire/thUncertain；连续概率替代三档 |
| F5 WeakBio force-real 短路 | **禁入**（铁律） | 派生信号不进 belief（feedback_no_dynamic_threshold_modulation），此唯一例外不迁移 |
| F6 ghost_penalty 扣分 | P(o\|s) → Artifact | ObsTrackPresent |
| F7/F8 pose2→5 qualification | P(o\|s) | ObsFirmwareFall×10（Option C 不重判） |

### 3.7 Bed FSM 贝叶斯 BD1–BD12（zoneengine/bed_bayesian_scorer.go）

**整族 = 一个嵌套 belief 子模块**，输出 P(InBed) 直接作 `ObsBedOccupied`。不逐条迁移——
床 scorer 已生产验证，本 PR 复用其输出，不重写。

| gate | 去向 |
|---|---|
| BD1-7 各源锚定/条件接收 | 仍在 bed scorer 内部；输出汇成 ObsBedOccupied(P+conf) |
| BD8/BD9 簇 LR | = ObsBedOccupied 的标定来源（§1 BedOccupied 行直取 LR 表） |
| BD10 γ tempering | 同构本包 temper(w,Conf)（Conf 退火 = γ 衰减） |
| BD11 maintain 区 2min 强制 LeftBed | bed scorer 内部决策；room belief 只读其 P |
| BD12 multiSourceLeftBed sticky | **消灭**（在 room 层）：跨源矛盾由 b 加权仲裁，不需 sticky hack |

### 3.8 Zone state machine ZS1–ZS6 + Subset invariant SI1–SI8

| gate | 去向 | 条目 |
|---|---|---|
| ZS1-4 Vacant↔Occupied↔Leaving 翻转 | A | Empty↔StandWalk↔Left 转移（hysteresis=A 自持概率） |
| ZS5 hysteresis 防抖 | A | 高自持对角（防单帧翻转 = 概率惯性） |
| ZS6 ForceSet/Rollback | **消灭** | 无条件强制 = 子集不变量 hack；联合 b 一致性天然成立 |
| SI1/SI7 bed⊆room lift_parent | **消灭** → 中耦合 | "床有人→房有人"= §5.5.3 守恒（v1 弱耦合 ObsNeighbor 近似） |
| SI2/SI3 repair 巡检 / stale bed 降级 | **命门** | stale → Conf 0；不需周期巡检对账 |
| SI4 drop 无 radar 孤儿房 | P(o\|s) | ObsNumberPeople=0 / 无观测自然衰减 |
| SI5/SI6 self-contradiction 登记+验证 | **消灭** | 矛盾证据在 b 上加权，不需 pendingValidations 二次确认 |
| SI8 Leaving window 兜底 | A | Left→Empty 收敛率 |

### 3.9 Suite census C1–C14（suite_census.go）

census 在 v1 仍是 suiteID-key 人数管理；多数 gate 是 **§5.5.3 中耦合守恒的宿主**（v2）。

| gate | 去向 |
|---|---|
| C1-C4 升格 resident/visitor/candidate | v2 中耦合：census 从计数器升归一化器 |
| C5-C7 idle 衰退（visitor 10m/resident 6h/candidate 4m） | **命门 + HSMM**：缺证据 Conf 0 + 状态时长，替代硬 idle 阈值 |
| C8 ClearAnchorOnLostTrack（60s 清 anchor 保留 person） | **消灭** | 冲突 3 支柱；单一 b 无 anchor/person 双账 |
| C9-C13 跨房 mark（exit bathroom / active cross-zone） | **消灭/弱耦合** | C13（冲突 3 的协调规则=9h 漏调根源）整个蒸发 |
| C14 SaveToRedis | — | 持久化，belief 同样需要 snapshot（实现细节，非规则） |

### 3.10 兜底/默认 D1–D9

| gate | 去向 |
|---|---|
| D1 Bed FSM 初始态 / D2 LeavingWindow / D3 census config | A prior / 常量 |
| D4/D8 BedStatus 默认 NotInBed | P(o\|s) | 无 BedState → ObsBedOccupied Conf 0（不强写零值） |
| D6 AloneSinceTs / D7 count_change 直写 | **消灭/禁入** | 派生时长禁入；count 走 ObsNumberPeople |
| D9 ScoreConfirmTh | 决策层 |

---

## 4 四冲突在 belief 里如何消失（inventory §C 的兑现）

| 冲突 | 现状 hack | belief 解 |
|---|---|---|
| **1** LeftBed 立翻 Vacant vs fall 需 Occupied | 维护 bed FSM `UpdatedAt` + BedSession `LeftBedAtMs` 两套时钟（L12 协调） | 同一 b：sleepad LeftBed 压床态、radar lying 抬床态，**贝叶斯加权仲裁**，无第二时钟 |
| **2** LongSurvivalAnchor vs GhostPenalty | 时序竞速 + `!Anchored` 护卫（V7 锁死） | b 永远可被新证据翻；ObsTrackPresent ghost 持续作用，无"锚定豁免" |
| **3** 失锁清 anchor vs person_silent（**9h bug**） | 靠 caller 显式调 C13 跨房喂活跃时戳（**漏调=误报**） | 缺证据=Conf 0 不更新（命门）；P(Fallen) 不起来，**构造上消灭整类** |
| **4** ExitRoom 清全部 vs 多人各自 pending | 靠 cardagg 保证 ExitRoom 只在全员离开发（跨服务约定，本仓不可验证） | v1：ObsEnterExit 在单实体 b 上加权；v2：§5.5.3 人数守恒，半人离开不误清 |

**共同根**（inventory §D 四套独立时钟）→ 单一 b + 统一 Observation.Ts，**结构上对账**。

---

## 5 v1 scope 边界（明确不覆盖）

| 项 | v1 | 何时 |
|---|---|---|
| 单实体 forward（A + P(o\|s) + 命门） | ✓ | 本 PR |
| §5.5.2 弱耦合 ObsNeighbor | ✓ | 本 PR（9h 用 sleepad InBed 作 Neighbor 近似治本） |
| shadow 旁路只 log | ✓ | 本 PR（§9 第 3 步） |
| 3-case 闭环（CABB/John.Y/真跌倒） | ✓ | 本 PR（belief_test.go 已绿） |
| §5.5.3 中耦合人数守恒 | ✗ | v2（治本 9h 的完全体；C1-4/SI1 迁此） |
| HSMM 状态时长（替 still-fall 硬阈值） | ✗ | v2（BA1-3/BE1-2 时长门迁此） |
| 多人 per-track belief | ✗ | v2 |
| cutover（删 gate-list） | ✗ | shadow 分歧全 triage 为"修真 bug"后 |

---

## 6 与代码骨架对应

| 文档条目 | 代码 |
|---|---|
| S 九态（§3 设计） | `state.go` State/Vector |
| ObsKind 12 种 + 七元组 | `observation.go` |
| P(o\|s) 标定行（本表 §1） | `likelihood.go` rawLikelihood/poseLikelihood |
| A 转移条目（本表 §2） | `model.go` transitionPropensity |
| forward 两步 + 命门 + 决策 | `belief.go` Predict/Observe/Decide |
| 3-case 回归 oracle | `belief_test.go` |

关联 [[room_belief_state_machine.md]]（总设计）/ [[belief_input_normalization.md]]（输入侧）/
[[fall_rule_inventory_and_conflicts.md]]（本表输入：100 条 gate）。
