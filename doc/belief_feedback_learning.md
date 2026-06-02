# Belief 反馈自学习闭环 — 人工标注间接再估计概率表（白盒）

状态：设计文档（落地施工图）。最后更新 2026-06-01。
关联：[[belief_state_rule_engine_reframe]]、`belief_dbn_completeness.md`（结构完备性，本 doc 的**前置依赖**）、
`room_belief_state_machine.md`（S/A/L 总设计）、`belief_input_normalization.md`（Observation 七元组）、
`belief_gate_to_matrix.md`（gate→L/A 对照）、`feedback_no_dynamic_threshold_modulation`（时序边界铁律）。

代码现状锚点：`wisefido-sensor/internal/roomengine/feedback.go`（婴儿版摄取）、
`cell.go`（cell 计数器 + `toleranceFactor`）、`belief/`（纯模型）、`belief_shadow.go`（production shadow）。

---

## 0. 北极星与本 doc 的位置

最初设计意图：

> **持续输入设备实时值 → 白盒信念滤波算连续"危险系数" → 人工 feedback 离线再估计概率表 → 模型间接自学习。**

belief 状态机（`belief/`）已实现前两段（滤波 + 决策）。本 doc 设计**第三、四段**：把人工 feedback 变成"下一版概率表"的训练信号，且全程白盒、可解释、每个数字可追溯到支持它的人工标注。

关键定性：这**不是**机器学习黑箱，是**有标签的监督式参数再估计**（监督版 Baum-Welch = 带先验的计数归一）。人类规则进先验伪计数，feedback 进观测计数，后验 = 先验 + 计数。模型"学到"的每一格都能回答"几条人工标注支持它"。

---

## 1. 前置认知（命门）：DBN 结构补全是自学习的前提

**在残缺单层 HMM 上学参数会学到扭曲值。** 原因见 `belief_dbn_completeness.md`：本会话所有 lost_track 修复（method-2 ghost-last / 门区+np=0 exit / ghost-skip guard / lostWhileMoving ramp）**全在 adapter 构造 observation 那一层，无一在 L 或 A 内**。也就是说，喂进 `L` 的"观测"本身已是带历史记忆的推断产物。

后果：若现在直接对 `P(o|s)` 做计数再估计，学到的数字会把 adapter 的合成逻辑（ramp 形状、ghost-skip 时机、np=0 推断）一起"吸收"进概率表 →

- 参数不再可解释（"这格 0.7 是物理似然还是 adapter ramp 的副作用?"无法回答）；
- 参数过拟合到 gate 的缝（adapter 改一行，学到的表全失效）；
- 违背白盒承诺。

**所以学习范围必须与结构完备度对齐**（本会话决策：**先定义安全子集，可在当前模型学**）：

| 学习目标 | 是否纠缠 adapter | 当前可学? | 解锁前置 |
|---|---|---|---|
| **geom-conditioned 发射 `P(o\|s, Geom)`**（cell 空间先验） | 否（cell 几何是真实测量，非合成） | ✅ **安全子集，现在可学** | — |
| **基础姿态发射 `P(ObsPose\|s)` / `P(ObsKinematics\|s)`** | 否（直接物理测量） | ✅ **安全子集，现在可学** | — |
| **`P(no-detect\|s)` 缺测发射** | 是（现合成 ramp 替代） | ❌ 推迟 | `belief_dbn_completeness` ② absence 发射（P2） |
| **转移 A 条目** | 是（含 lost↔justleft↔present 簇 + frozen-credit） | ❌ 推迟 | DBN Track 层 A_T（P1→P3） |
| **HSMM 驻留分布**（confirmMs / MovingPreconditionMs / StillSec） | 是（现外挂 timer） | ❌ 推迟 | HSMM 状态（P3） |

**本 doc 的 v1 学习 scope = 仅上表前两行（geom-conditioned + 基础姿态发射），其余标注解锁阶段、写设计不实施。** 这样既启动闭环，又不在残缺结构上学扭曲值。

---

## 2. 标签 schema：什么是一条训练样本

**训练样本 = 一段 replay episode + 一个人工 ground-truth 标签 + 触发位置 geom。**

### 2.1 标签来源与信任策略（硬约束）

来源同婴儿版：`alarm_events.operation ∈ {false_alarm, verified}` + `handler_notes` 的 ☑ checkbox（复用 `parseConditions`）。

**信任策略（本会话决策：需人工 ground-truth 闸）**——这是与婴儿版的根本区别：

- gate-list 自己的报警输出**只作候选**，**不直接当标签**。理由：用户已实证 gate-list 的 FP 标签会毒化（cabb-fall-A 的 "real" 标签几乎肯定错、pose 不可靠）。用 gate-list 的 FP 当训练标签 = belief 学回同一个 bug。
- 一条样本进训练集，**必须有人工显式确认**：护理员/admin 标了 `verified`（=真摔）或 `false_alarm` **且勾了具体 condition**（=确认没摔/确认是 ghost/坐具）。
- **废掉婴儿版的无勾选兜底**：`feedback.go:373` 现在 `pc.FAOther || !pc.AnyFASelected() → FakeAlarmCount++`，admin 啥都没勾也兜底计数。新策略下**无 condition 勾选的 false_alarm 不进训练集**（视为标注缺失噪声跳过）。

### 2.2 标签置信度分层（不直接当二值）

人工标签本身也有可信度差异，进 Dirichlet 计数时**带权重**：

| 标签类型 | 证据 | 权重 |
|---|---|---|
| `verified` + ☑Fall + 护理员现场确认 | 强 ground truth | 1.0 |
| `false_alarm` + ☑具体 condition（Ghost/Chair/Sofa/Wheelchair） | 强 ground truth | 1.0 |
| `false_alarm` 仅 still_box=0（干净消失，机器高置信 FP）但无人工 note | **不进 v1**（待人工补标） | — |
| still_box>0 短静止（疑 FP，无人工确认） | 需人工 ground-truth 才进 | — |

> 注：still_box 类机器置信度在 [[belief_state_rule_engine_reframe]] 已讨论可作*诊断*分层；但本 doc 的训练集**只收人工确认过的**，机器置信度仅用于**给人工标注排优先级**（先让人确认高价值的歧义 case），不替代人工。

### 2.3 episode 重建

一条样本不只是"标签 + 位置"，要能重放出 belief 的输入序列：

- 触发时刻 ±窗（90s lookback，复用 `findRadarPositionAt` 思路，但读 `monitor_stream` 全序列非单帧）→ 重建 track 帧序列；
- 配套 `event_log` 的 EnterRoom/ExitRoom/number_people；
- 配套 layout geom（`room_visual_layout.canvas` /128）→ cell 语义；
- = 与 `doc/cases/` fixture 同格式（用 `scripts/export_case_v2.sh` 导出），直接进 `TestReplayOracle` 当回归 oracle。

**人工确认过的每条样本 → 导成 fixture → 进 oracle 集**。学习闭环与回归 oracle 是**同一批数据**：你既用它训练，也用它防回归。

---

## 3. 可学 vs 锁死

映射到 [[feedback_no_dynamic_threshold_modulation]] 的三类阈值表：

| 参数类别 | 例子 | 可离线学? | 说明 |
|---|---|---|---|
| **belief 发射 L（geom-conditioned + 基础姿态）** | `P(ObsPose\|s,Geom)` / `P(ObsKinematics\|s)` | ✅ v1 学 | §1 安全子集；走 §4 promote 闸 |
| **belief 转移 A** | StandWalk→Fallen 等条目 | ⏸ 推迟 | DBN 补全后（P1+） |
| **Fallen 近吸收行** | `Floor-Fallen` 的 99/0.7 | ❌ **锁死** | 物理不变量（倒地不自愈），不是统计量 |
| **决策阈值 θ_fire / confirmMs** | belief.go thFire=0.55 / 90s | ⚠️ 谨慎学 | 改报警灵敏度；可学但需更高 promote 门槛 + 人工签 |
| **Sensor 内部启发式常量** | GhostPenaltyThreshold / MovingPreconditionMs | ⚠️ git PR 改 OK | **禁** runtime 派生信号自动调（铁律） |
| **Firmware / 客户面阈值** | `alarm.cloud_config` HR/RR/fall 灵敏度 | ❌ **绝对锁死** | 只能 layout/admin/API 人工配置（铁律） |

**白盒可解释要求**：每个可学参数在版本化产物里附 `support_count`（支持它的人工标注条数）+ `prior`（人类规则伪计数）+ `posterior`（学后值）。任何人能读出"这格为什么是这个数"。

---

## 4. EM / 计数更新规则（监督版 Baum-Welch，白盒）

### 4.1 为什么是"计数归一"不是梯度黑箱

有 ground-truth 标签时，episode 的状态路径**几乎已知**（fall episode → 末段是 Fallen；false_alarm episode → 全程非 Fallen）。所以无需完整 EM 的 E 步软分配，退化成**监督式计数**：

```
对每条人工确认 episode e（权重 w_e）:
  沿已知/Viterbi 状态路径 s_t，对每帧观测 o_t：
    count[s_t][bucket(o_t, geom_t)] += w_e        # 发射计数
```

### 4.2 Dirichlet 伪计数 = 人类规则注入

人类规则不是和数据"二选一"，而是当**先验伪计数**坐底：

```
prior[s][o]  =  现行 gate 的确定性角点 × κ        # κ = 先验强度（人类规则可信几条样本之力）
posterior[s][o]  =  normalize( prior[s][o] + count[s][o] )
```

- 新装机 / 该格 0 样本 → posterior = prior = 现行行为（**精确复现现状**，对齐 belief_state_rule_engine_reframe 的"确定性角点先 0/1 复现"约束）；
- 样本累积 → 平滑偏离先验，但 κ 保证少量噪声标注翻不动；
- 每格可读：`posterior = (prior κ 票 + 人工 N 票) 归一` → 完全可追溯。

### 4.3 geom-conditioned 发射 = cell 计数器的升维（见 §6）

v1 唯一实学的就是这块：cell 层的人工计数（拆出的 (a) 部分）→ 离线聚合成 `P(o|s, Geom)` 的 posterior，替代婴儿版 cell 直接调 `toleranceFactor`。

---

## 5. 时序边界：与 no-live-modulation 铁律（命门）

**这是本 doc 与婴儿版的根本分野，也是不可破的纪律。**

```
运行时（hot path）          ┃   离线批（offline）
─────────────────────────  ╋  ─────────────────────────
belief 用冻结的概率表 v_k    ┃   摄取人工确认标注
决策纯净、可复现、可审计      ┃   ↓ §4 计数再估计 → 候选表 v_{k+1}
不读任何当批 feedback         ┃   ↓ §4 跑 TestReplayOracle（零回归）
                            ┃   ↓ shadow 对账分歧（§4 蓝图 P1 风格）
                            ┃   ↓ 人工签 promote → v_{k+1} 上线（下次重启/热加载）
```

- **反馈是"下一版模型"的训练信号，不是"这次报警"的调制信号。** 一条 false_alarm 进来，**不影响任何在途/当下报警**，只进离线批，影响**下一版表**。
- 运行时参数对一次会话内**完全冻结** → 同样输入永远同样输出（决策可复现 = 可审计 = HIPAA 友好）。
- 这与铁律一致：铁律禁的是**派生信号实时改决策**；本设计连**人工 feedback** 都不许实时改决策（更严），只走离线 promote。
- **婴儿版的 (b) inline modulation 恰恰破这条**（`FakeAlarmCount` 一进当帧就改 `EffectiveStillTimeoutSec`）→ §6 处理。

---

## 6. cell-counter 的归宿：拆 (a) 保留 / (b) 退役

婴儿版 cell-counter（`cell.go` / `feedback.go`）是**两个东西焊在一起**，本会话决策**拆开**：

### (a) 空间先验账本 — **保留并升级为 belief 的 geom 发射标定源**

- 5 计数器（`GhostCount` / `RestZoneConfirmed` / `RealFallCount` / `FakeAlarmCount` / `ToleratedStillCount`）+ `Decay()` 半衰期，编码"房间**哪个位置**ghost 多发 / 真人坐卧区 / 真摔高发"。
- 这是 belief 全局 per-state 模型**缺的空间分辨率**（`P(o|s)` 是全局的；"地板中央 vs 化妆台前"的 fall 先验天然 per-cell）→ 有真价值。
- **改造**：① 只接 §2.1 人工 ground-truth 闸过的标注（**去掉无勾选兜底** `feedback.go:373`）；② 计数**不再当场调任何决策参数**，而是作为 §4 离线再估计 `P(o|s, Geom)` 的输入。

### (b) 运行时 inline modulation — **belief cutover 后退役**

- `FakeAlarmCount + ToleratedStillCount → toleranceFactor∈[1,2] → ×EffectiveStillTimeoutSec`（`cell.go:374`）= 标一次当帧就改活决策，无 shadow/promote 闸。
- **违反 §5 时序边界** → 在 belief 体系内**不复刻**。
- **gate-list 现役不动**（本 PR 不碰 gate-list 决策路径）；doc 明确标 (b) 为 "belief 接管后退役"项，cutover 时随 gate-list 一起删。

> 注：(b) 不算违反 no-modulation **铁律**（那条管派生信号；人工 feedback 是合法学习通道），但违反**本 doc 的时序纪律**。区别要讲清，避免被误当铁律违规回退 gate-list。

---

## 7. 版本化产物与 promote 流程

### 7.1 模型表是版本化文件，不是 DB 行

- 概率表（L 的 posterior + support_count + prior + κ）落**版本化产物**（YAML/JSON，类比 `ai_fall_parameter.yaml`），git 管理、PR review。
- 运行时加载冻结版 `v_k`；离线批产出 `v_{k+1}` 候选；promote = 改指针 + 重启/热加载。
- **不入 DB 实时表**（决策不依赖 DB 派生值，对齐 [[feedback_store_raw_not_derived]]）。

### 7.2 promote 门槛（逐级严格）

| 学习目标 | 门槛 |
|---|---|
| geom-conditioned / 基础姿态发射 | TestReplayOracle 零回归 + shadow 分歧全 triage 为"修真 bug" |
| θ_fire / confirmMs（谨慎类） | 上述 + **人工显式签**（改报警灵敏度需人定） |
| 转移 A / HSMM | 不在 v1（DBN 补全后另立 promote 流程） |

### 7.3 shadow 对账（复用现有 belief_shadow）

- `belief_shadow.go` 已 log-only 旁路在跑。promote 候选表 `v_{k+1}` 先在 shadow 跑 → 对账 `belief_shadow_fall` log vs gate-list alarm_events vs 人工标签三方分歧 → 逐条 triage → 全是修真 bug 才 promote。
- = belief_dbn_completeness 蓝图 P1 "Track 层只读 shadow 当零风险 oracle" 的同款纪律。

---

## 8. v1 落地 scope（本 doc 对应 PR）

**做**：
1. §2 人工 ground-truth 闸的摄取改造（`feedback.go`：去无勾选兜底 + 权重字段 + 导 fixture 进 oracle）。
2. §6 (a)：cell 计数器接闸后只作账本，断开 → `toleranceFactor` 的运行时调用（belief 侧不复刻；gate-list 侧保留）。
3. §4 geom-conditioned + 基础姿态发射的离线再估计器（计数 → Dirichlet posterior → 版本化产物），先**只产候选表 + 报告，不 wire 进 fire**。
4. §7 promote 流程文档化 + shadow 对账脚本。

**不做（标注解锁阶段）**：
- 转移 A / HSMM 学习（DBN 补全 P1-P3 后）；
- absence 发射 `P(no-detect|s)` 学习（P2 后）；
- gate-list (b) inline modulation 退役（belief cutover 时）；
- θ_fire/confirmMs 自动学（先人工，攒够样本再评估）。

---

## 9. 与现行的一致性证明（可证）

沿用 [[belief_state_rule_engine_reframe]] 的"一致 = 与已验证正确结果（fixture 标签）一致"：

1. **0 样本 → posterior=prior=现状**（§4.2 κ 坐底）→ 新装机精确复现现行 belief 行为；
2. **5-case oracle 仍全绿**（D5F7 真摔 confirm / cabb / cd2b / MoM / D523 不报）→ 学习不破已验证 case；
3. **每个学到的格可追溯**（support_count + prior + κ）→ 白盒，可逐格人审。

失效模式从"补 gate"变成"补样本 / 调 κ / 标更多人工 ground-truth"—— 局部可观测、可回归。
