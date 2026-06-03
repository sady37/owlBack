# P2：absence 发射 `P(no-detect | T, S)` 设计

> DBN 迁移分期的第 2 步。前置 = P1 Track 层（`belief_dbn_completeness.md` §6.1 已部署）。
> 本文只设计、shadow-only 落地；不碰 fire 路径，不删 gate-list（删 gate 是 P4）。
> 关联：[belief_dbn_completeness.md](belief_dbn_completeness.md)（§3②缺发射 / §6 P2）、
> [belief_input_normalization.md](belief_input_normalization.md)、[belief_gate_to_matrix.md](belief_gate_to_matrix.md)（#2 门区 / #14 np=0）、
> [fall_rule_inventory_and_conflicts.md](fall_rule_inventory_and_conflicts.md)。

---

## 1. 目的

P2 = 把「这帧没测到 track」当成**带似然的观测**来建模，而不是当成空白。
`belief_dbn_completeness.md` §3② 把它列为三段缺口之二：

> "这帧没测到 track" 本身带似然：`P(no-detect|Fallen)` 高（贴地遮挡）、`P(no-detect|Empty)` 高、
> `P(no-detect|StandWalk)` 低。现在"消失"不是发射，而是临时合成 `lostWhileMovingToObs` ramp 硬塞。

P2 用一个显式的 `P(no-detect | T, S)` 发射，替掉那段合成 ramp，并把「为什么没测到」的判别
（走出去了 vs 倒地遮挡了）交给**可解释的乘性似然因子**，而不是硬阈值闸。

---

## 2. 现状的三个缺陷（带证据）

### 2.1 合成 ramp 不是发射
`belief_adapter.go:124 lostWhileMovingToObs`：track 丢失后按「丢失后时长归一」线性抬 `P(Fallen)`
（`likelihood.go:92 ObsLostWhileMoving` → `SFallen: 1 + lostMovingFallGain*a`）。这是手调的时间斜坡，
不带「这次没测到，在不同状态下概率不同」的贝叶斯语义；且 gain 在 2 个 case 上标定，易过拟合。

### 2.2 离门距离是硬阈值悬崖（实证 = CABB-2247）
判「人是不是走到门口出去了」当前是 **0/1 硬 cutoff**：
- gate-list：`track_manager.go:2548 NearestEntryDist(px,py) <= FallRulesParam.Lost.ExitDistMinCm` → 判离场；
  `fall_rules_param.go:153 ExitDistMinCm = 30`（2026-04-30 从 100cm 收紧）。
- belief：`belief_adapter.go:20 beliefEnterMarginCm = 30`，同一刀切。

**CABB-2247 实测**：人最后停在 `(50,-70)`，离 Enter 多边形 `x[70,100] y[0,80]` = **73cm**，pose4 站立，
随后 track 消失 + `number_people=0`（0.9s 后）。`73 > 30` → 几何门不触发 → 当「开阔地板丢失」→ 误报。
旧阈值 100cm 时 73cm 会被抑制；收紧到 30cm（为减少门口附近真跌倒漏报）就把这类「走到门附近停一下再出去」
漏成 FP。**任何单一 cutoff 都两头不讨好**：松了漏报门口真跌倒，紧了把退场当跌倒。
同型：两条 09E7 线上 FP（人走到 FOV 边缘 `x300`、离门更远但 `np=0`）。

### 2.3 np=0 似然 over-trust（违铁律）
`likelihood.go:60 ObsNumberPeople`，np=0（Value<0.5）当前：
```
{SEmpty: 6, SLeft: 2, SBedLying/Restless/Sit/StandWalk/Fallen: 0.3}
```
`SEmpty=6` + 把 `SFallen` 压到 0.3 = **过度信任 np=0**。违反铁律
（`lost_track_ghost_npzero_door_exit`：**np=0 是 corroboration 不是 substitution**——金属垃圾桶/镜子→ghost、
淋浴水气→信号衰减都会假报 np=0）。真跌倒若固件误报 np=0，会被这条强压住 → 漏报。

---

## 3. 目标：`P(no-detect | T, S)` = 三因子相乘

track 丢失后每 tick 发一条 absence 观测，其「倒地 vs 离场」的方向由**三个软因子的乘积**决定，
而非任何硬 if：

```
P(no-detect | justleft/Left)  ∝  f_dist(d) · f_reach(v, d, Δt) · f_np0
P(no-detect | lost/Fallen)    ∝  (1 - 上述) 方向的补，叠加丢失后时长（dwell→P3）
```

- **离场方向证据越强（近门 + 可达 + np=0）** → absence 似然偏 `Real-justleft → Room:Left`，压 Fallen。
- **离场方向证据弱（远离门 + 不可达 + 无 np=0）** → absence 似然偏 `Real-lost → Fallen`，候选倒地。
- 同一个「73cm + np=0」由三因子共同涌现结论，不再是「过 30cm 线 = 一律地板丢失」。

### 3.1 因子①：软距离到 Enter `f_dist(d)`
`d = NearestEntryDist(lastX, lastY)`（已有，`grid.NearestEntryDist`）。把 0/1 cutoff 换成平滑衰减：

```
f_dist(d) = exp(-d / d0)      // d0 ≈ 80cm（CABB 73cm 落在显著非零区）
```
- `d=0`（门内）→ 1.0；`d=73`（CABB）→ ≈0.40；`d=190`（D5F7 镜面 ghost 区）→ ≈0.09。
- d0 是唯一新标定量，初值 80cm（≈1 步），由 oracle 标定（§9）。**保留** `ExitDistMinCm` 仅作 gate-list 现役，
  belief 侧不再用硬 margin。

### 3.2 因子②：reachability `f_reach(v, d, Δt)`
「以该人的步速，能否在一个上报间隔 Δt 内跨过离门距离 d」——单帧可达 = 退场极可能：

```
reach = (v * Δt) / max(d, ε)          // v=步速 cm/s, Δt=上报间隔(~1s), d=离门距离
f_reach = clamp01(reach)              // ≥1 → 满（一帧即可走出）
```
- **步速 v 的两个来源（主 + 兜底）**：
  - **观测速度（主）** = 固件分钟级聚合 `walk_distance / walk_duration`。**已落地**：
    `target_state_aggregator.go:61-62 WalkDistanceMeters / WalkDurationSec`（源 `iot:event:stream category=activity`，
    字段 `observation.FieldWalkDistance / FieldWalkDuration`，`fields.go:46-47`）。
    **必须用聚合口径，不用逐帧速度**（逐帧位置量化重、走动也常 median=0/帧；D5F7 真跌倒逐帧 max 仅 82cm/s 会漏）。
  - **PHI 步速先验（兜底/上界）** = 见 §4。近期无 walk stat 或样本噪声时用；并给 reach 一个生理上界，
    避免把噪声当「超人速度走出去」。
- **CABB 验算**：d=73cm，Δt≈1s。观测速度缺则用先验：行动不便 0.6 m/s → `60*1/73 = 0.82` → f_reach≈0.82；
  健康老人 1.0 m/s → `100/73 = 1.37` → f_reach=1.0。**单帧可达成立**。

### 3.3 因子③：np=0 弱乘性 `f_np0`（修 #14，遵铁律）
`likelihood.go:60 ObsNumberPeople` np=0 重标定：
```
旧: {SEmpty: 6,   SLeft: 2, ..., SFallen: 0.3}     // over-trust
新: {SEmpty: 1.5, SLeft: 2,      SFallen: 1.0}     // 弱倾向空/离，不反驳已倒地
```
- np=0 降为**弱乘性因子**：只轻微倾向 Empty/Left，`SFallen` 保持中性 1.0（不清零）。
- 强 Empty 拉力交给 `ObsEnterExit<0`（`likelihood.go:57` 已有 `SLeft:8`）和 absence 的 justleft 方向。
- 铁律落地：真跌倒 + 固件误报 np=0 时，`SFallen=1.0` 不被反驳，倒地的 absence/姿态证据仍能竞争。
- np=0 仅作 `f_dist·f_reach` 的**确认乘数**（corroboration），不单独决定离场（substitution）。

---

## 4. PHI 步速先验：mobility-tier（HIPAA 边界派生）

| 人群 | 室内步速 v | 判据（resident PHI） |
|---|---|---|
| 正常成年人 | **1.2 m/s** | 默认 |
| 健康老人 | **1.0 m/s** | age ≤ 79 |
| 行动不便 / 高龄 | **0.6 m/s** | age ≥ 80 或 mobility 受限 |

（室内/bathroom 已较临床基准下调；bathroom 场景步速进一步降，由 oracle 标定是否再乘 0.8。）

**铁律：原始 age / DOB / mobility 不进 sensor/roomengine 的 fall 路径**（min-necessary，对齐
[`phi_encryption`](kms.md) / K-service 双因子）。改在 **PHI 边界**（解密侧，wisefido-data 或 K-service）
把 age+mobility 派生成一个**粗 3 档 `mobility_tier`（非 PHI）**，随 resident/card 配置下发；sensor 只读 tier→映射 v。

**schema（走 dbv2 CREATE 评审，铁律 `feedback_schema_review_via_dbv2`）**：
- 候选落点：`resident`/`card` 配置增一列 `mobility_tier SMALLINT`（0=normal/1=elder/2=impaired），
  或进现有 spatial/card config 投影。**先改 `owlRD/dbv2/<NN>_*.sql` 的 CREATE 提审，再 ALTER + 改 Go。**
- 0 样本（无 PHI / 未派生）→ 默认 tier=0（1.2 m/s）= 最保守（最易判可达 = 最易抑制）；
  注意这与「保守 = 不漏报」张力：默认偏快会更易把丢失判成退场。**默认应取最慢 0.6 m/s 更安全**（少抑制、少漏真跌倒）——
  待 §9 决策。

---

## 5. 数据来源就绪盘点

| 因子 | 数据 | 现状 |
|---|---|---|
| 软距离 d | `grid.NearestEntryDist(x,y)` | ✅ 已有 |
| 观测走速 | `WalkDistanceMeters/WalkDurationSec`（activity stat） | ✅ 已解析（`target_state_aggregator.go:61-62`），但**当前未喂 belief**，需桥接 |
| PHI 步速先验 | `mobility_tier` | ❌ 待建（§4，dbv2 评审） |
| np=0 | `ObsNumberPeople` | ✅ 已有，需重标定 likelihood |
| Δt 上报间隔 | 固定 ~1s | ✅ 常量 |

唯一新增管道 = 把固件聚合走速接到 belief（observation 七元组多一个 `ObsReachability` 或并进 absence 的 Geom/Value）。

---

## 6. 代码落点（shadow-only）

1. **新观测/重定义**（`belief/observation.go`）：
   - 把 `ObsLostWhileMoving` 升级语义为 absence 发射输入，或新增 `ObsReachability`（Value=f_reach，Geom 带 d）。
2. **likelihood**（`belief/likelihood.go`）：
   - `ObsNumberPeople` 重标定（§3.3，#14）。
   - 新 absence 发射：`P(no-detect|·)` = 三因子合成（§3），替 `ObsLostWhileMoving` 的纯时长 ramp（§2.1）。
   - `lostMovingFallGain` 的时间斜坡退化为 dwell 分量（**时长本身归 P3 HSMM**，P2 只管方向）。
3. **adapter**（`belief_adapter.go`）：
   - `lostWhileMovingToObs` → 计算 `f_dist·f_reach·f_np0`，发 absence 观测；
   - 新增「固件走速注入」：从 target_state_aggregator 取 `WalkDistanceMeters/WalkDurationSec` → v_obs；
   - mobility_tier → v_prior 兜底。
4. **shadow wire**（`belief_shadow.go:183`）：现 `lostWhileMovingToObs(age, geom, nowMs)` 调用点换成新 absence；
   **仍 log-only，不 fire**。

---

## 7. gate→DBN 对照（P2 内化哪几条）

来自 `belief_gate_to_matrix.md` / `belief_dbn_completeness.md` §5：

| # | gate | P2 内化为 |
|---|---|---|
| #2 | 门区 exit 推断（门区+np=0） `bathroom_fall.go inferredDoorExit` | `f_dist·f_np0` 进 absence 发射的 justleft 方向 |
| #8 | checkLostFall 离门/门区不入池 | `f_dist` 软化（不再 30cm 硬线） |
| #10 | lostWhileMoving age-ramp | 时长分量剥离 → P3；方向分量 → P2 三因子 |
| #14 | np=0「强证据空房」压 Fallen 0.3 | 降为弱乘性 `f_np0`（SFallen 中性） |

#1/#3/#12（ghost→benign）已在 P1；#9/#11（dwell/确认窗）留 P3；#7/#13（多人）留 P5。

---

## 8. Scope 边界：P2 vs P3

- **P2（本文）**：方向判别 = 「这次没测到，是走出去还是倒地」。靠 `f_dist·f_reach·f_np0` 的发射似然。
- **P3（不在本文）**：时长判别 = 「倒地态要维持多久才确认」（`confirmMs=90s` / `MovingPreconditionMs=60s` / `StillSec`）→ HSMM 状态时长分布。
- 走速/距离是**方向证据非时长**，全在 P2。`lostWhileMovingToObs` 的时间 ramp 拆分：方向→P2，时长→P3。

---

## 9. 标定与 oracle（防过拟合）

- **新标定量**：`d0`（软距离尺度，初值 80cm）、bathroom 步速折扣、mobility_tier 默认档、`f_np0` 权重。
- **oracle**（`belief_replay_test.go TestReplayOracle`，shadow-only）：
  - **CABB-2247**（73cm + np=0 走出）：现仅诊断（confirm=false 靠未定稿 np=0 likelihood）→ P2 后应可 **assert 不确认**（三因子结构性抑制）。
  - **09E7 ×2**（FOV 边缘 + np=0）：导出 fixture 后加入，期望不确认。
  - **D5F7-1031 真跌倒**（远离门、无 np=0、长时倒地）：必须仍 **confirm=true**（f_dist/f_reach 低 → absence 偏 Fallen）。
  - cd2b（返回）/ MoM（有 ExitRoom）/ D523（无床 + 09E7 suite）：保持不确认。
- **决策待定**（§4）：mobility_tier 默认档取 1.2（保守抑制）还是 0.6（保守不漏报）——用 oracle 在「漏真跌倒 vs 误报退场」上定。
- **诚实边界**：P2 不创造可观测性。雷达分不开「门口倒地 vs 门口走出」的极端帧时，三因子也只能输出适当不确定度
  → 决策层不确定时不抑制（保跌倒优先）。

---

## 10. 铁律遵守清单

- ✅ np=0 corroboration 非 substitution（§3.3 弱乘性、SFallen 中性）。
- ✅ 派生信号不进 alarm 决策路径——走速/距离是**物理观测**（固件 stat + 几何），非 WeakBio/RiskLevel 类派生（`feedback_no_dynamic_threshold_modulation`）。
- ✅ PHI min-necessary（§4 边界派生 tier，原始 age 不进 sensor）。
- ✅ schema 改动先过 dbv2 CREATE 评审（§4）。
- ✅ shadow-only，不碰 fire；gate-list 现役保留到 P4 cutover。
- ✅ 与现状一致可证：三因子初始化到「现行硬阈值的软化角点」先复现现状，再标定（对齐 `belief_state_rule_engine_reframe` 确定性角点原则）。
