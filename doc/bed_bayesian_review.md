# Bed Bayesian Scorer — 审查文件

**目的**：下次会话独立审查 bed FSM 的贝叶斯模型（数学公式、决策、参数、实现一致性）。本文件自包含，无需读外部上下文。

---

## 1 背景与目标

### 1.1 待解问题

| # | 问题 | 现状（legacy Scorer+StateMachine） | 期望 |
|---|---|---|---|
| P1 | **夜间跌床检测**（核心 safety 场景）| sleepad LeftBed event 来时立刻翻 Vacant，但 fall.suspect 需要 bed 仍 Occupied 才能定位 → 检测窗口闭合 | LeftBed 单源不足以立刻翻 Vacant；保留 maintain 区让 fall 探测器有时间 confirm |
| P2 | **大床翻身假阳 LeftBed**（B2C home）| sleepad 翻到床沿 → LeftBed 真实触发，但人还在 → cardagg bed_status flip → 误报 | LeftBed 期间若 vital 持续 → 视为冲突，γ 衰减 LeftBed 权重 |
| P3 | **Yang.R 5-7s 周期翻转**（已 root-cause）| Self-contradiction rollback 在 sustain 死后 fire → bed FSM 抖动 | Bayesian 模型用 maintain 区 + γ tempering 取代 self_contradiction 机制 |
| P4 | **B2B/B2C 床型差异** | 一刀切（small/large bucket 仅微调 hysteresis）| 引入 unit_property（0=Home / 1=Facility），LR 强度档不同 |

### 1.2 核心 trade-off

跌床 FN cost ≫ FP cost：宁可慢一点确认 Vacant，也不能在老人跌床时让 bed FSM 翻 Vacant 导致 fall 探测器找不到 anchor。**决策阈值非对称**：P>0.70 → InBed / P<0.50 → LeftBed / 中间 → 维持上一态。

---

## 2 数学模型

### 2.1 递归 log-odds 累加

定义：
- $L_t = \log \frac{P(\text{InBed at } t)}{P(\text{¬InBed at } t)}$（log-odds）
- $P_t = \sigma(L_t) = \frac{1}{1 + e^{-L_t}}$（probability）

递归更新：
$$L_t = \text{clamp}_{[-5, +5]}(L_{t-1} + \Delta L_t)$$

$\Delta L_t$ 由每个证据簇贡献，**簇内 max-merge，簇间累加**。

### 2.2 证据簇 LR 表

**sleepad 簇**（max-merge：簇内多证据不重复加，取最强）：

| 证据 | mode=Facility (B2B) | mode=Home (B2C) | 备注 |
|---|---|---|---|
| InBed event fresh | +2.9444 (ln 19) | +2.1972 (ln 9) | 5% / 10% miss rate 反推 |
| LeftBed event fresh **且无 vital** | -2.9444 | -2.1972 | 同上 |
| Vital alive（HR/RR/body_move/turn_over/bed_status=0 任一） | +2.9444 | +2.1972 | 与 InBed event 同档 |

**簇合并规则**（每分钟最多贡献一次该簇的 max LR）：
```
sleepadLR =
  if leftBedFresh && !vitalAlive  →  -LR_S   (负向冲突)
  else if inBedFresh || vitalAlive →  +LR_S
  else                             →   0
```

**radar 簇**（同 max-merge）：

| 证据 | LR | 备注 |
|---|---|---|
| InBed event fresh | +1.3863 (ln 4) | firmware bed-area enter |
| pose=lying fresh | +1.4469 (ln 4.25) | track pose（仅 install_mod=full+bed area 才返） |
| Vital alive（HR>0 / RR>0） | +0.5596 (ln 1.75) | radar 接触式不足，弱 LR |
| LeftBed event fresh | -1.3863 | 任何正向 fresh 时不触发 |

```
positive = max(InBed:+1.39, pose_lying:+1.45, vital:+0.56) over fresh ones
radarLR =
  positive > 0    →  positive
  else leftBed    →  -1.3863
  else            →   0
```

**fresh** = `nowMs - lastEvidenceTs <= 60_000`（60s window，匹配 4G sleepad 30s 上报 + 2x margin）

### 2.3 γ tempering（sleepad 冲突衰减）

冲突 = sleepad LeftBed fresh **且** sleepad vital alive。冲突意味着传感器自我矛盾，应折扣 radar 正向证据（不影响 sleepad 自身）。

| 冲突持续时长 dur | γ | 含义 |
|---|---|---|
| 0 - 60s | 1.0 | 早期信任 radar 全力（vital 真的还在 → 翻身假阳）|
| 60 - 120s | 0.5 | 中期半折扣（仍倾向于人在床，但开始怀疑）|
| ≥ 120s | 0.0 | 晚期完全不信 radar 正向（人确实离床很久了）|

γ **仅应用到 radar 正向 LR**；sleepad LR 不衰减；radar 负向（LeftBed）不衰减。

### 2.4 阈值与 hysteresis

```
P > 0.70  → InBed   （清空 maintain timer）
P < 0.50  → LeftBed （清空 maintain timer）
0.50 ≤ P ≤ 0.70 → maintain previous decision
                  若 maintain 已持续 ≥ 120s → 强制 LeftBed
```

**非对称设计**：跌床场景中，老人跌下床后 sleepad LeftBed 真实，但 radar 正向证据可能仍在（pose=lying 在地上）。我们宁可保守翻 Vacant 也不能漏（120s 强制 leave 是 safety net）。

### 2.5 Prior 初始化

按 unit 本地小时 (要查 unit timezone)：
- 21:00 - 06:59 night → $L_0 = +1.39$（P=0.8，假定老人夜间在床）
- 07:00 - 20:59 day → $L_0 = -0.85$（P=0.3，假定老人白天不在床）

仅在 scorer 创建或重置时调一次。

### 2.6 关键 invariant

| Invariant | 实现位置 | 验证 |
|---|---|---|
| `L_t ∈ [-5, +5]` | `clampLogOdds` | clamp 函数 + 单测 case 3 验证 t=7min L→-5 |
| 每分钟每簇最多贡献一次 max LR | `contributeClusterLocked` (lastMin 守护) | 单测 case 1 InBed 反复触发只一次 |
| 同分钟簇值翻转触发 delta 补偿 | `contributeClusterLocked` (lastValue 差分) | 单测 case 4 同瞬间双源 leave |
| Vital 无须 InBed event 即可正向贡献 | `sleepadClusterLRLocked` (`inBedFresh || vitalAlive`) | 单测 case 2 P3 |
| 冲突期 radar 不污染 L | `gammaLocked` × `evaluateAndContributeLocked` 仅乘到 radarContrib | 单测 case 2 t=5-7min |
| Decision 维持 2min 后强制 LeftBed | `Decision` (maintainStartedTs) | `TestBayesian_MaintainTimeoutForceLeftBed` |

---

## 3 单元测试（6 例，全过）

| # | 名字 | 场景 | 期望 |
|---|---|---|---|
| 1 | NormalInBed_Facility | sleepad InBed + radar InBed + radar pose lying + vital 全活 | t=0 立刻 InBed，t=60s L→+5 cap |
| 2 | BigBedRoll_HomeMode_MaintainInBed | home 模式，5min vital → LeftBed 但 vital 仍活（翻身假阳）| t=5min maintain，t=6min L≈-0.87 maintain，t=7min γ→0 后 vital 撑住 P≈0.04（应 Vacant）— 注：这个 case 的 expected outcome 是 Vacant 因为 6min 后冲突 γ=0，radar 正向被砍只剩 sleepad LeftBed 持续负向。**审查关注**：是否应让 vital alive 在 home 模式独自扛住？|
| 3 | FallFromBed_Facility | 5min 在床 → LeftBed 来，无 vital，γ 累积 | t=5min P=0.29 already Vacant；3min 内打到 L=-5 |
| 4 | NormalLeave_Facility | t=0 sleepad + radar 双 LeftBed | 立即 Vacant，L=-4.33 |
| 5 | PriorByHour | InitPriorByHour(23) vs (10) | L=+1.39 / -0.85 |
| 6 | MaintainTimeoutForceLeftBed | 进入 maintain 区 → 120s 后 | 自动 Vacant |

**Engine 集成测试（2 例，全过）**：
- `TestEngine_Bayesian_SleepaceInBedFlipsOccupied` — sleepace InBed event → ZoneEvent (bed Occupied, Confidence ≥ 70)
- `TestEngine_Bayesian_DualLeaveFlipsVacant` — sleepad+radar 双 LeftBed → ZoneEvent (bed Vacant)

---

## 4 实现位置

| 文件 | 角色 |
|---|---|
| `owlBack/wisefido-sensor/internal/zoneengine/bed_bayesian_scorer.go` | 核心 scorer（独立可测试）|
| `owlBack/wisefido-sensor/internal/zoneengine/bed_bayesian_scorer_test.go` | 6 单测 |
| `owlBack/wisefido-sensor/internal/zoneengine/engine.go` | Engine wiring：SetUseBedBayesian / Apply dispatch / Tick dispatch / repairSubsetInvariant 适配 |
| `owlBack/wisefido-sensor/internal/zoneengine/engine_test.go` | 2 集成单测（line ~637+）|
| `owlBack/wisefido-sensor/internal/zoneengine/wiring/unit_property_lookup.go` | unit_property cache（已写，未 wire 到 BedModeLookup）|

---

## 5 SignalEvidence → BedBayesianScorer 路由

`IngestEvidence` 派发表（engine.Apply 自动调用）：

| Source | Kind | → 方法 |
|---|---|---|
| sleepace | enter | OnSleepadInBed |
| sleepace | leave | OnSleepadLeftBed |
| sleepace | sustain | OnSleepadVital |
| vital_sleepad | sustain | OnSleepadVital |
| radar | enter | OnRadarInBed |
| radar | leave | OnRadarLeftBed |
| radar | pose_lying | OnRadarPoseLying |
| vital_radar | sustain | OnRadarVital |

不在表内的 source/kind → 该 evidence 被丢弃（`IngestEvidence` 返回 false）。**审查关注**：是否有遗漏 source？是否应允许 fallthrough？

---

## 6 待审查关键点（recommended audit checklist）

### 6.1 数学层

- [ ] LR 系数推导：facility 5% miss → ln(19=0.95/0.05)、home 10% miss → ln(9=0.9/0.1) 是否正确？
- [ ] radar pose=lying +1.45 是否过高？依据：firmware 已 gate (install_mod=full + bed area)，相当于"地面验证后的 lying"，比裸 +1.39 InBed event 高 5% 合理？
- [ ] radar vital +0.56 是否过低？依据：mmWave HR ±5-10 BPM 不精确（[[radar_hr_no_critical]]）；接触式 sleepad vital 同档 +2.94 显示差距合理？
- [ ] γ 阶梯 1.0/0.5/0.0 是否平滑度足够？是否需要 sigmoid 平滑而非阶跃？
- [ ] L cap ±5 是否对应 P∈[0.0067, 0.9933]，足以承载 facility/home 两挡的累加？
- [ ] 阈值 0.70/0.50 是否需要按 mode 不同（home 模式更保守）？

### 6.2 工程层

- [ ] `IngestEvidence` 路由表是否完整覆盖现有 adapter source/kind？
- [ ] `LastEvidenceTs()` 取 max(7 个 *LastTs) 是否合 `repairSubsetInvariant` 的语义（"最后一次任何证据")？
- [ ] `neverTs = -(evidenceWindowMs+1) = -60001` 是否在所有可能的 nowMs 下都正确（包括 nowMs=0 测试时）？
- [ ] Tick 频率建议 1s — Decision 内部 maintainStartedTs 是否能在 Tick 间正确演进？
- [ ] `BedModeLookup` 接口设计：按 /96 bed prefix 查 mode，由 Stage 4 wire 到 UnitPropertyLookup → /80 unit 查 unit_property。**审查关注**：bed → unit 推导是 prefix 截断（前 80 bits），需 Stage 4 实现。

### 6.3 与现有代码契合度

- [ ] `useBedBayesian=false` 时，legacy Scorer/StateMachine 路径完全未受影响（60+ 历史单测全过）—— 已验证
- [ ] `repairSubsetInvariant` stale_bed 判定改用 `bedBayesian.LastEvidenceTs()` 后语义一致 —— 已改但未跑专项测试，**审查关注**
- [ ] Bayesian 不再需要 `pendingValidations` self_contradiction —— 已在 Apply 路径绕过（直接 return），**确认无遗漏**
- [ ] `applyBedBayesianLocked` 中 Confidence (0-100) 直接进 `state.Score`，下游消费 Score 的代码是否会被破坏（原 Score 是 raw int 累加，现在是 0-100 概率）？

---

## 7 开放问题（Stage 3-7 影响审查决策）

1. **vital_source 改造**：现行 `vital_source.go` 用 HR>0||RR>0||body_move>0||turn_over>0||bed_status==0 判 sustain → 发 `Source="vital_sleepad"` SignalEvidence。**问题**：是否应把 sleepad vital 与 InBed event 视为同一簇（max-merge）？当前模型这么做了，但 vital_source 的发射节奏（每 30s）与 InBed event（状态翻转才发）频率差很大，evidenceWindow=60s 是否合适？

2. **radar pose 提取**：现行 adapter_radar 未单独发 pose_lying SignalEvidence。需要 Stage 6 增加：当 firmware 返回 track.pose=lying 且 install_mod=full 且 bed area 内 → 发 SignalEvidence{Source="radar", Kind="pose_lying"}。**问题**：track 1Hz，触发频率太高，是否应限速？

3. **Home 模式 BedMode 切换**：BedModeLookup 是按 /96 bed prefix 查 unit_property。多个 bed 在同一 unit 下时，BedMode 一致。**问题**：什么时机重新查（unit 改 property 时如何 invalidate）？当前设计在 getOrCreate 时查一次，之后 zoneInstance 缓存。需要 hot-reload 机制吗？

4. **Tick 频率**：Engine.Tick() 默认 1s。Bayesian 模型 contribute 是 per-minute 节奏。**问题**：每秒 Tick 在同分钟内只做 evaluate 不 contribute（lastMin 守护），开销可接受；但 maintainStartedTs 是 ms 级，Tick 1s 足以收敛 120s 强制 LeftBed。是否需要保留 1s Tick？

5. **观测与告警**：Bayesian 是黑盒，调参困难。**问题**：是否应输出 `bed_status.bayesian_log` 给 zone state，包含 L / P / 各簇 LR / γ 当前值？方便 production debugging。

---

## 8 commit message draft（Stage 1+2 合并后）

```
zoneengine: Bayesian bed FSM (Stage 1+2)

替换 bed zone 的 Scorer+StateMachine 为 log-odds 累加模型：
- 簇级 max-merge（sleepad / radar 内部不双倍计 LR）
- γ tempering（sleepad LeftBed 与 vital 冲突时 3 阶段折扣 radar 正向）
- L cap ±5；非对称阈值 P>0.70 InBed / P<0.50 LeftBed / 维持区 120s 强制 leave
- Facility / Home 双 LR 档（unit_property 驱动，pending Stage 4 wire）
- 21-07 夜间 prior L₀=+1.39（pending unit timezone wire）

Engine 加 SetUseBedBayesian opt-in 切换；legacy 路径保留供 room/bathroom 与
opt-out 测试。8 新单测全过，60+ 历史单测无回归。

解决 P1 跌床检测 / P2 大床翻身假阳 / P3 Yang.R 5-7s 抖动（self_contradiction
rollback 机制由 Bayesian maintain 区取代）。Stage 3-7 wiring 接续。
```

---

## 9 审查反馈闭环（2026-05-23 收口）

### 9.1 Must-fix
- ✅ **ZoneState.Score 下游排查**：搜全 `wisefido-sensor` + `wisefido-cardagg` + `owl-common/card` 三处 codebase。结论：
  - `ZoneState.Score` **仅 engine 内部用**（4 处全在 engine.go 内部赋值）。
  - 外部消费者（zonealarm Supervisor / translator / cardagg unit_picker）**只读 NewState.Status / Count / LastEnterTs/LastExitTs / LastSource**，**不读 Score**。
  - FE `BedConfidence` 是 source ladder（sleepace=90 / radar=60），由 translator 从 `NewState.LastSource` 派生，与 Score 无关。
  - **结论**：Bayesian Score 语义改为 0-100 安全，不影响下游。
- ⚠️ **附带发现 P0 bug**：Bayesian 路径下 `applyBedBayesianLocked` 未设置 `z.state.LastSource` → translator 拿不到 source → BedConfidence 永远=0 → FE 显示 "—"。**已修**（[engine.go:227](owlBack/wisefido-sensor/internal/zoneengine/engine.go#L227)）。

### 9.2 Recommended (已实施)
- ✅ **Home 模式更高阈值**：`pInBedThresholdHome=0.75 / pInBedThresholdFacility=0.70`，Decision 按 `s.mode` 自动选阈值（[bed_bayesian_scorer.go](owlBack/wisefido-sensor/internal/zoneengine/bed_bayesian_scorer.go) `Decision`）。
- ✅ **Bayesian debug 字段**：ZoneState 新增 `BayesianLogOdds / BayesianProb / BayesianGamma`（`json:omitempty`，仅 bed+bayesian 时填），engine.applyBedBayesianLocked 每次写。Gamma 通过 `BedBayesianScorer.Gamma(nowMs)` 公开。
- ✅ **stale_bed 专项单测**：`TestEngine_Bayesian_RepairDropsStaleBed` 验证 `bedBayesian.LastEvidenceTs()` 走 repair 路径。
- ✅ **Home 阈值单测**：`TestEngine_Bayesian_HomeModeUsesHigherInBedThreshold` 验证 home 模式单 InBed event (P=0.90) 仍能过 0.75 阈值。

### 9.3 测试结果（全 10 例 PASS）
```
TestBayesian_Case1_NormalInBed_Facility          PASS
TestBayesian_Case2_BigBedRoll_HomeMode_Maintain  PASS
TestBayesian_Case3_FallFromBed_Facility          PASS
TestBayesian_Case4_NormalLeave_Facility          PASS
TestBayesian_PriorByHour                         PASS
TestBayesian_MaintainTimeoutForceLeftBed         PASS
TestEngine_Bayesian_SleepaceInBedFlipsOccupied   PASS
TestEngine_Bayesian_RepairDropsStaleBed          PASS (new)
TestEngine_Bayesian_HomeModeUsesHigherInBedTh    PASS (new)
TestEngine_Bayesian_DualLeaveFlipsVacant         PASS
```
60+ 历史 zoneengine 测试全过，零回归。

### 9.4 仍在 Stage 4 闭环的 audit 项
- **BedModeLookup wire**：UnitPropertyLookup → BedModeLookup 适配器（Stage 4 实现，bed /96 → /80 unit prefix 截断 → unit_property → BedMode）。
- **radar pose_lying 限速**：1Hz 上报，bayesian per-minute dedup 已自动限速，但 IngestEvidence 仍每秒 mutex lock。Stage 6 增加 1s adapter-side dedup（如已有 BedEventDedup 复用）。
- **BedMode 热重载**：当前 getOrCreate 缓存。Stage 4 加 `ReloadRules → 同步刷新 bedMode` 或单独 `OnUnitPropertyInvalidate(unitPrefix)` 通道。

---

## 10 下次会话审查入口指令

> 审查 `owlBack/doc/bed_bayesian_review.md`，重点核对 §2 数学公式、§6 audit checklist。
> 验证位置：`owlBack/wisefido-sensor/internal/zoneengine/bed_bayesian_scorer{,_test}.go` + `engine.go` Apply/Tick/repairSubsetInvariant 三处 dispatch。
> 跑测试：`cd owlBack/wisefido-sensor && go test ./internal/zoneengine/ -run "TestBayesian|TestEngine_Bayesian" -v`
> 不允许跳过 §6 任何一条 checkbox，全部回答 ok/not-ok + 依据。
