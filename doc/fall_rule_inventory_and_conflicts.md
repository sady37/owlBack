# Fall 规则引擎清点 + 冲突清单

**定位**：穷尽列出 wisefido-sensor 当前 gate-list 规则引擎的**每一条规则**（要点 + file:line +
条件/阈值 + 动作）和**规则间冲突**。用真实数字证明"必须治本"，作 [[room_belief_state_machine.md]]
的论据侧。

**口径**：一个 `if` 分支 = 一条规则。≈100 条，散在 11 个文件。
**数据来源**：2026-05-31 两路代码扫描，**file:line 是某时点快照，使用前复查**（R2）。
**已抽查核验**（2026-05-31）：L1/L3/L5/C8/C13/V6/V7/BD11/BD12 行号全部命中（±1 行内）；
SI1 行号已修正（:460/:550）。冲突 1/2/3 的支柱规则已逐条核到当前代码，冲突属实非推测。

`revision: 1` ・ `created: 2026-05-31`

---

## A. 数字总览

| 判定族 | 规则数 | 主文件 |
|---|---:|---|
| Track verdict 出生/翻转 | 18 | track_manager.go |
| Lost / silent / bedside fall | 12 | track_manager.go |
| Bathroom fall（10a/b/c/d） | 11 | bathroom_fall.go |
| Bedroom fall（11b/c） | 7 | bedroom_fall.go |
| Ghost adjudicator | 5 | bathroom_ghost.go |
| Fall verifier 评分 | 8 | fall_verify.go |
| Bed FSM 贝叶斯 | 12 | zoneengine/bed_bayesian_scorer.go |
| Zone state machine | 6 | zoneengine/state_machine.go |
| Subset invariant | 8 | zoneengine/engine.go |
| Suite census 升格/衰退 | 14 | suite_census.go |
| 兜底/默认值 | 9 | 散落 |
| **合计** | **≈100** | **11 文件** |

---

## B. 逐条规则清单

### B.1 Track verdict 出生/翻转（track_manager.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| V1 | Startup grace 反 ghost | :981 | 新 track 且 service 启动 <5min | Verdict=Real, Penalty=0, StartupGrace=true |
| V2 | 出生打分 BirthScore | :992 | 新 track（非 grace） | 按位置/cell 信任算 BirthScore 0-100 |
| V3 | no_enter_pair 扣分 | birthScore | 出生时无邻近 EnterRoom event 配对 | -60 |
| V4 | enter_pair_bonus | birthScore | 3s 内有 EnterRoom 配对 | +20 |
| V5 | 盲区返回判 Real | :998 | birthScore isBlindsideReturn | Verdict=Real，绕过 ghost 检查 |
| V6 | GhostPenalty≥80 → Ghost | :1165/1185 | Penalty≥80 且 !Anchored 且 !Grace | Verdict=Ghost |
| V7 | LongSurvivalAnchor | :1151 | AgeSec≥300（5min）且 !Anchored | Anchored=true, 翻 Real, **此后豁免 ghost** |
| V8 | Real→Ghost（post-real） | :1165 | Verdict==Real 且 Penalty≥80 且 !Anchored | 翻 Ghost，emit ReasonGhostPostReal |
| V9 | score-based 判定 | :1217 | Pending 到期：Score≥RealTh→Real / <GhostTh→Ghost | 翻 verdict |
| V10 | later_born_with_real_track | birthScore | 已有 real track 时晚生 | 扣分判 ghost |
| V11 | mirror pair penalty | applyLifetimeGhostFactors | motion_symmetry 镜像对 | +60 penalty |
| V12 | 低分 probation 边缘加码 | factor4/5 | penalty∈[70,80) 边缘 | factor4/5 各 +10 跨阈 |
| V13 | Kalman birth-coherence | :1033 | 首帧 Kalman 与历史不一致 | 累积 ghost penalty |
| V14 | frozen 检测 | :2678 | 连续位移<30cm(StillBoxCm) 且 >2min | 记 FrozenRunStart |
| V15 | still box 检测 | :2327 | pose≠Walking 且 位移<30cm | 记 StillSince，>15min 学 AreaSit |
| V16 | pose-mismatch 异常 | detectPoseMismatch | pose=Walk 但速度≈0 / pose=Stand 但 Z<50 | -3 / 累积异常 |
| V17 | Z 噪声检测 | detectZNoise | 单帧 Δz>50cm | ZNoiseCount++ |
| V18 | avgSpeed 低惩罚 | :2403 | age>5s 且 avgSpeed<2 且 !rest | -2 |

### B.2 Lost / silent / bedside fall（track_manager.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| L1 | checkLostFall 几何门 | :2502 | age≥5s 且 离门>30cm(ExitDistMinCm) 且 cell≠AreaEnter | 允许进 pending |
| L2 | lost_fall pending 入池 | :1105 | (Real∨Pending) 且 checkLostFall 且 !bedside_reported | 建 pending |
| L3 | lost_fall fire | :1269 | 等待超时(cell-typed) 且 realTrackCount<2 | **fire Fall**(ReasonLostTrack) |
| L4 | frozen credit 折抵 | :2557 | 算超时时 frozen 时长×50% 折抵 wait | 更快报 lost |
| L5 | ExitRoom 取消全部 pending | :772 | 收到 alarm.ExitRoom | **delete 全部** pendingLostFalls |
| L6 | 新 track 出生取消 | :977 | 新 track 邻近 pending 消失点 | delete 该 pending（盲区返回） |
| L7 | 实时多人取消 | :1258 | realTrackCount≥2 | delete pending |
| L8 | silent_fall gate | scanSilentFallLeftBed | 需 RadarInBedConfirmedMs 赋值 | 允许 silent 判定 |
| L9 | silent_fall fire | :1447 | LeftBed 后 vital/位置矛盾 + 超时(60/120s) | **fire**（person_silent） |
| L10 | bedside_fall R4 | :249 | LeftBed 后 180s 内 track 在床邻域<100cm 静止>900s | AnomalyBedsideFall |
| L11 | bedside dedup | :1099 | 已报 bedside → 跳过 lost pending | 防同事件双报 |
| L12 | RadarInBed 双源一致 | :738 | radar InBed 与 BedSession.LeftBedAtMs ±15s | 设 RadarInBedConfirmedMs |

### B.3 Bathroom fall 四类（bathroom_fall.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| BA1 | 10a StillFall（risk-time） | :276/294 | pose=Stand 静止 ≥10min（risk-time） | fire |
| BA2 | 10a StillFall（normal） | :276 | 静止 ≥12min（non-risk） | fire |
| BA3 | 10b BedsideFall | :312 | 静止 8min + 90s grace | fire |
| BA4 | 10c LostFall 强档 | :395 | bathroom 内 0 track ≥30s 且 BathroomCount≥1 | **fire 最高风险** |
| BA5 | 10d LostFall 弱档 | :436 | SuitePerson 在 bathroom + ghost 滞留 ≥7min | fire |
| BA6-11 | risk-time 判定/grace/阈值参数等 | bathroom_fall.go:46-51 | 10/12min(still)、8min(bedside)、30s/7min(lost) | 阈值常量 |

### B.4 Bedroom fall（bedroom_fall.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| BE1 | 11b BedsideFall | :156 | 夜间静止 15min（LeftBed 后） | fire |
| BE2 | 11c LostFall cell-typed | :295 | cell-typed 阈值：AreaBed 2h / AreaSit 30min / AreaActive 5min / default 10min | fire |
| BE3 | LastActiveMs 完整性 | :313 | LastActiveMs==0 → return（不报） | 抑制 |
| BE4 | silentSec idle 判定 | :353 | idleMs<silentSec → return | 抑制 |
| BE5 | isHumanBedAt 抑制 | :359 | 在人标床区 → 不报 | 抑制 |
| BE6 | hasActiveBedSession 抑制 | :359 | sleepad InBed 中 → 不报 | 抑制 |
| BE7 | sole resident 守卫 | :299 | 非 sole resident in bedroom → return | 抑制 |

### B.5 Ghost adjudicator（bathroom_ghost.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| G1 | rule1 距 entry 近 | :386 | distanceToNearestEntry 小 | 降 ghost 嫌疑 |
| G2 | rule2 motion_symmetry | :137+ | 镜像对称运动 | 判 ghost |
| G3 | rule3 mirror 反射 | bathroom_ghost.go | 镜面区 + 对称 | +penalty |
| G4 | IsPublicBathroom 降级 | :137 | public bathroom | rule2/3 自动降级 |
| G5 | entry 盲区不可信 | :290 | bathroom 入口在 FOV 盲区 | adjudicator 不可信 |

### B.6 Fall verifier（fall_verify.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| F1 | baseline 评分 | :69 | firmware Fall alarm 到 | score=50 baseline |
| F2 | ghost<30 档 | :214 | score<GhostThreshold(30) | verdict=ghost（不报） |
| F3 | real≥70 档 | :214 | score≥RealThreshold(70) | verdict=real（报） |
| F4 | suspect 中间档 | :254 | 30≤score<70 | verdict=suspect |
| F5 | WeakBio force-real 短路 | fall_verify | WeakBio≥80 | **强制 real**（已知唯一例外，[[feedback_no_dynamic_threshold_modulation]]） |
| F6 | ghost_penalty 扣分 | :254 | track ghost_penalty | bd_ghost_penalty 负向 |
| F7-8 | pose2→5 qualification 等 | fall_verify | firmware 30-90s 升级 | Option C 不重判 |

### B.7 Bed FSM 贝叶斯（zoneengine/bed_bayesian_scorer.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| BD1 | sleepad InBed 锚定 | :265 | sleepad "enter" 事件 | sleepadInBedLastTs=now，清 LeftBed |
| BD2 | sleepad LeftBed 锚定 | :277 | sleepad "leave" | sleepadLeftBedLastTs=now；radar 同源 60s 内也 LeftBed→multiSourceLeftBed |
| BD3 | sleepad vital 条件接收 | :292 | vital 到：若 multiSourceLeftBed>0 或自家 LeftBed 活→**拒绝** | 否则记 vital |
| BD4 | radar InBed 锚定 | :308 | radar bed-area "enter" | radarInBedLastTs=now |
| BD5 | radar LeftBed 锚定 | :320 | radar bed-area "leave" | radarLeftBedLastTs=now |
| BD6 | radar vital 条件接收 | :334 | 同 BD3 逻辑 | 拒绝 or 记 |
| BD7 | radar pose_lying 条件接收 | :348 | 同 BD3 逻辑 | 拒绝 or 记 |
| BD8 | sleepad 簇 LR | :449 | LeftBed→-LR_S / InBed→+LR_S / vital→+0.69 | 簇内 max-merge |
| BD9 | radar 簇 LR | :478 | LeftBed→-1.39 / max(InBed 1.39, pose 1.45, vital 0.56) | 簇内 max |
| BD10 | γ tempering 冲突衰减 | :502 | sleepad LeftBed sticky + radar 正向 | γ:1.0(0-60s)→0.5(60-120s)→0.0(≥120s) |
| BD11 | maintain 区 2min 强制 LeftBed | :393 | P∈[0.50,0.70] 维持区 >120s | 强制 LeftBed（跌床安全偏） |
| BD12 | multiSourceLeftBed 粘性 | :112 | 双源 60s 内都 LeftBed | 所有 vital/pose 衍生证据**全拒**，直到 InBed 清零 |

### B.8 Zone state machine（zoneengine/state_machine.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| ZS1 | Vacant→Occupied | :97 | score≥enter_th 且 hysteresis 过 | Occupied |
| ZS2 | Occupied→Leaving | :110 | score≤exit_th 且 hysteresis 过 | Leaving |
| ZS3 | Leaving→Occupied 返回 | :121 | score≥enter_th（不受 hysteresis） | Occupied |
| ZS4 | Leaving→Vacant | :129 | leavingSince 超 LeavingWindowSec(8s) | Vacant |
| ZS5 | hysteresis 防抖 | :86 | now-lastTransition<HysteresisSec | 拒绝翻转 |
| ZS6 | ForceSet/Rollback | :143 | self_contradiction 或 subset_invariant | 无条件强制 |

### B.9 Subset invariant（zoneengine/engine.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| SI1 | bed⊆room lift_parent | :460(mode 判)/:550(lift) | bed Present 且 room Vacant 且有 radar | 强制 room→Occupied |
| SI2 | repair 周期巡检 | :458 | Tick 且超 RepairIntervalSec | 扫 bed candidates |
| SI3 | stale bed 降级 | :477 | bed LastEvidenceTs>24h | 强制 bed→Vacant（信 room） |
| SI4 | drop 无 radar 孤儿房 | :576 | room Present 且 !HasRadar 且无活 bed | 强制 room→Vacant |
| SI5 | self-contradiction 登记 | :283 | 翻→Occupied 方向 | 登记 pendingValidations |
| SI6 | self-contradiction 验证 | :367 | window_sec(5s) 内 score<50 | rollback Vacant + penalty 20 |
| SI7 | maybeReconcileSubset 即时 lift | :726 | Apply 收证据且 bed Present 有 radar | 立即 lift（不等 Tick） |
| SI8 | Leaving window 兜底 | :129 | LeavingWindowSec≤0 | fallback 8s |

### B.10 Suite census（suite_census.go）

| # | 规则 | file:line | 条件/阈值 | 动作 |
|---|---|---|---|---|
| C1 | 双源 sleepad 强升 resident | :214 | sleepadInBed 且 residentID≠"" | 立即升 resident |
| C2 | candidate→resident | :288 | anchor≥5min 且 traverse≥10 cells 且 residentID≠"" | 升 resident |
| C3 | candidate→visitor | :297 | anchor≥2min 且 traverse≥5 cells 且 residentID=="" | 升 visitor |
| C4 | 已升格 person 更新 | :243 | trackID 匹配 anchor | TraverseCount/LastActiveMs 更新 |
| C5 | visitor idle 衰退 | :529 | role=visitor 且 idle>10min | delete person |
| C6 | resident 衰退 | :541 | role=resident 且 idle>6h | delete（warn 设备故障） |
| C7 | candidate 衰退 | :555 | role="" 且 idle>4min | delete |
| C8 | ClearAnchorOnLostTrack | :361 | trackID ≥60s 未收到 | AnchorTrackID=0（**保留 person**） |
| C9 | MarkPersonExitToBathroom | :381 | gate 检测进 bathroom | AnchorRoomType=Bathroom, count++ |
| C10 | MarkPersonReturnToBedroom | :400 | gate 检测离 bathroom | AnchorRoomType=Default, count-- |
| C11 | TryFlipSoleResidentRoomType | :476 | 恰好 1 resident 且 room_type 变 | flip AnchorRoomType |
| C12 | AdjustBathroomCount | :443 | 匿名 track 进/出 | BathroomCount±（public→noop） |
| C13 | MarkActiveCrossZone | :330 | bathroom 内 person 非静止帧 | LastActiveMs=now（防 6h 误衰退） |
| C14 | SaveToRedis | :571 | 周期 | 序列化 census，TTL 24h |

### B.11 兜底/默认值

| # | 规则 | file:line | 要点 |
|---|---|---|---|
| D1 | Bed FSM 初始态 | bed_bayesian_scorer.go:125 | L=0 中性，lastDecision=LeftBed |
| D2 | LeavingWindow 兜底 | state_machine.go:31 | ≤0 → 8s |
| D3 | DefaultSuiteCensusConfig | suite_census.go:95 | 5min/10cells/2min/5cells/10min/6h |
| D4 | BedStatus 默认 | translator.go:34 | IsPresent→0 否则 1 |
| D5 | ~~number_people=0 ExitRoom 兜底~~ | track_manager.go:787 | **PR-9 已删**，改 cardagg 归一化 |
| D6 | AloneSinceTs 锚点 | engine.go:861 | Count==1 起锚 / ≠1 清 0 |
| D7 | count_change 直写 | engine.go:207 | 绕过 FSM 直写 Count |
| D8 | cardagg bed_status 默认 NotInBed | cardagg | 缺 BedState 强制 NotInBed（[[bed_status_default_not_in_bed]]） |
| D9 | ScoreConfirmTh 常量 | track_manager.go | confirmed 阈值 |

---

## C. 冲突清单（真冲突，靠 hack 协调）

筛掉"互斥条件不同时触发"的伪冲突，**4 处真冲突**：

### 冲突 1 — LeftBed 立刻翻 Vacant vs fall 需 bed 仍 Occupied
- **A**（ZS2+ZS4）：sleepad LeftBed → 立即翻 Vacant
- **B**（L2+L10）：fall 检测需要 bed 仍 Occupied 才能定位倒地
- **同触发**：sleepad LeftBed 与 radar track 同帧/前帧消失，cell=bed
- **协调 hack**：维护**两套独立时间轴**（bed FSM `UpdatedAt` + BedSession `LeftBedAtMs`）表示同一物理事实；bedside fall 读 BedSession 时戳不读 bed FSM 状态
- **协调本身是规则**：是（L12 RadarInBed 双源检测，跨两时间轴）→ 冲突没消除，搬成第三条规则

### 冲突 2 — LongSurvivalAnchor 锚定 vs GhostPenalty 翻 Ghost
- **A**（V7）：5min 锚定 Real，豁免 ghost
- **B**（V6）：Penalty≥80 翻 Ghost
- **同触发**：track 第 4-5 分钟 penalty 冲到 80，5min 刚好锚定
- **协调 hack**：**时序竞速** + 护卫 `!LongSurvivalAnchored`（谁先到谁赢）；真 ghost 若 5min 前没冲到 80 就被永久锚成 Real（[[longsurvival_anchor_ghost_gap]] 已知此洞）
- **风险**：中（锚定后 ghost 强信号也翻不动）

### 冲突 3 — track 失锁清 anchor vs person_silent 判倒地（**= 9h bug family**）
- **A**（C8）：track 失锁 60s 清 anchor（暗示人走了）
- **B**（BE2/L9）：person_silent 用 LastActiveMs 判倒地（暗示人还在）
- **同触发**：bathroom 静止淋浴 / 人进雷达盲区，无新 track 更新 LastActiveMs
- **协调 hack**：靠 caller **显式调 C13 MarkActiveCrossZone** 跨房喂活跃时戳
- **协调本身是规则**：是（C13）→ **漏调就误衰退/误报**，这正是 John.Y 9h person_silent 的结构
- **风险**：高（隐式规则易遗漏）

### 冲突 4 — ExitRoom 清全部 pending vs 多人各自 pending
- **A**（L5）：ExitRoom 任一条 **delete 全部** pendingLostFalls
- **B**（L2）：多人时各 track 各自 pending
- **同触发**：2 人房，A 入 pending，外部 ExitRoom 到（B 还在）
- **协调 hack**：靠 **cardagg 保证 ExitRoom 只在全员离开时发** —— 跨服务约定，**本仓无法验证**（[[number_people_zero_exitroom_fallback]] 兜底已删，更依赖此约定）
- **风险**：高（半人离开误发 → pending 误清 → fall 漏报）

---

## D. 更本质的冲突 — 四套独立时钟

4 处真冲突的**共同根**：描述"同一个人此刻状态"的**四个时间戳从不对账**：

```
bed FSM      UpdatedAt        ← bed 何时翻转
BedSession   LeftBedAtMs      ← 床压何时离床
SuitePerson  LastActiveMs     ← 人何时最后活跃
track        LastObservedMs   ← track 何时最后被看到
```

每条规则读其中 1-2 个 → 一致性靠人工在 caller 里手动同步（C13 漏调 = 9h bug）→
**缝活在"两个读不同时钟的 gate 之间"**。

---

## E. 与 belief 的对照（为什么治本）

| gate-list 现状 | belief 模型 |
|---|---|
| ≈100 条规则散在 11 文件 | 1 个 belief 向量 b + 1 个 A + N 行 P(o\|s) |
| **4 套独立时钟无对账** | **单一显式状态 b**：所有源更新同一个 b，时钟统一成 Observation.Ts |
| 4 处真冲突靠 hack 协调，协调本身又是新规则（L12/C13） | 冲突消失：相反证据 = 在同一 b 上加权，贝叶斯**自动**仲裁（Conf 高者影响大），不需第三条规则 |
| 冲突 3（9h bug）靠人工记得调 C13 | 缺证据=Conf 0 不更新，**结构上**不会误衰退 |
| OR 越加越长，封不死 | 状态有限，封闭 |
| 每修一误报 = 加规则 + 加协调规则，规则数与冲突数一起涨 | 加传感器 = 加 P(o\|s) 一行，复杂度 O(N²)→O(N) |

**核心**：今天每修一个误报 = 加一条 gate 或一条协调规则，而**协调规则本身又会和别的规则打架**
（冲突 3 的协调 = C13，它又会漏调）。规则数和冲突数**一起增长**。belief 把"N 条规则 +
O(N²) 个潜在冲突"换成"N 个状态 + 每传感器一组似然"，复杂度从平方降到线性。

---

## F. 用法

本清单 = [[room_belief_state_machine.md]] §6 "gate→矩阵条目对照表"的**输入侧**：
每条规则映射到 belief 模型里 A 或 P(o|s) 的哪个 0/1 角点条目，先复现现状再标定。
4 处冲突 = belief shadow mode 必须验证"修对而非搬过来"的重点。

关联 [[fall_fp_roots_and_todo]]（根因+fixture）/ [[bed_bayesian_review.md]]（B.7 已是 belief 子模块）/
[[belief_input_normalization.md]]（输入规范化）。
