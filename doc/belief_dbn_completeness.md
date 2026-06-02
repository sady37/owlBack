# Belief 完备性蓝图 — 从单层 HMM 到 (Track × Room) DBN + HSMM

状态：设计蓝图（下一步 DBN PR 的施工图）。
关联：[[belief_state_rule_engine_reframe]]、belief_gate_to_matrix.md（gate→似然对照）、
本会话 lost_track 误报修复（method-2 / 门区 exit / belief_shadow ghost-skip）。

---

## 1. 当前实现与"完备"的定义

`belief/belief.go` 跑标准 HMM forward：

```
bel_t  ∝  L(o_t)  ⊙  (Aᵀ · bel_{t-1})
        └─输入─┘     └─转移─┘
Predict: b ← A·b      (belief.go:Predict)
Observe: b ← norm(diag(w^conf)·b)   w = rawLikelihood(o)   (belief.go:Observe)
```

状态空间 S（`state.go`，单实体 v1，一房一人）：
`{Empty, BedLying, BedRestless, Sit, StandWalk, Fallen, Transition, Left, Artifact}`。

**`L × A = state` 完备（= 精确后验）需三前提：**

| | 条件 | 含义 |
|---|---|---|
| C1 | 状态充分(Markov) | `s_t ⊥ 过去 \| s_{t-1}` |
| C2 | 观测条件独立 + 无记忆发射 | `o_t ⊥ 其它 \| s_t`，`L=P(o\|s)` 只依赖当前测量 |
| C3 | A、L 已知且正确 | 转移/发射参数标定对 |

## 2. 判定：不成立（且从未完全成立）

本会话所有修复**无一在 A 或 L 内**，全在"构造 observation"的 adapter/engine 层：
喂进 `L` 的输入本身已是推断产物 → 同时破 C1（状态不充分）与 C2（合成观测带历史记忆）。

**每个打地鼠 guard = 一条缺失的 Track-层转移被手补成了 if。**

## 3. 缺的三段

### ① 缺一整层：per-track「存在/真伪」隐链（最大）
S 全是房间级姿态；"blip 是真人还是反射 / track 在不在"是**每条 track 自己的隐过程**，不在 S 里。
ghost adjudicator + track 生命周期就是这条缺链，现被拆进观测层（ghost-ness 当 obs + "ghost 消失不发 lost"当 gate）。

### ② 缺发射：`P(无观测 | s)` 未建模
"这帧没测到 track"本身带似然：`P(no-detect|Fallen)` 高（贴地遮挡）、`P(no-detect|Empty)` 高、
`P(no-detect|StandWalk)` 低。现在"消失"不是发射，而是临时合成 `lostWhileMovingToObs` ramp 硬塞。

### ③ 缺时长：dwell-time 不在状态（应 HSMM）
lost ramp 按"消失多久"、Decider `confirmMs=90s`、still-fall 的 `StillSec≥阈值`、`MovingPreconditionMs=60s`
全是驻留时钟。HMM 驻留是无记忆几何分布，装不下"维持数分钟才算真"。完整写法 = 半马尔可夫 HSMM。

## 4. 目标结构：(Track × Room) 两层耦合 DBN + HSMM

```
Track 层  T_t ∈ {None, Real-present, Ghost-present, Real-justleft, Real-lost}
            └─ 自带转移阵 A_T（缺失的那张表）+ 入边由 verdict/几何/事件发射驱动
Room  层  S_t | T_t      （现有 9 态，条件于 Track 层）
发射       含显式 P(no-detect | T,S)；frozen 帧 Conf=0→I（已对）
时长       Fallen/Left/Still 用 HSMM 状态时长分布替代 ramp+timer
```

核心：`Real-present→Real-lost`（走动中丢=fall-suspect）、`Ghost-present→None`（反射闪灭=benign）、
`Real-present(门区)→Real-justleft→Room:Left` 全部变成 **A_T 的转移概率**，不再是外部 if。

## 5. 对照表：现存 gate ↔ 它实为哪条 DBN 转移/发射/驻留

| # | 现存 gate / 似然 | 位置 | 现在的形态 | DBN 内化为 |
|---|---|---|---|---|
| 1 | method-2 失锁前全 ghost 不 fire | bathroom_fall.go:524 `lastBasesAllGhost` | obs-gate | A_T: `Ghost-present→None` ⇒ Room→Empty（非 Fallen） |
| 2 | 门区 exit 推断（门区+np=0） | bathroom_fall.go:546 `inferredDoorExit` | obs-gate（双证据合取） | A_T: `Real-present(门)→Real-justleft`；np=0 = `P(no-detect\|justleft)` 发射分量 |
| 3 | belief_shadow ghost 移出追踪 | belief_shadow.go:92 `delete(sh.tracks)` | obs-gate | A_T: `*→Ghost-present` 后不可达 `Real-lost` |
| 4 | 池入场仅 Real/Pending | track_manager.go:1124 | admit-gate | A_T: 仅 `Real/Pending-present→Real-lost` 有非零概率 |
| 5 | 池 ExitRoom 取消 | track_manager.go:785 | event-cancel | A_T: `Real-lost→Real-justleft`（门事件发射） |
| 6 | 池 birth-recovery 取消 | track_manager.go:996 `cancelPendingLostFallByBirth` | event-cancel | A_T: `Real-lost→Real-present`（重捕获发射） |
| 7 | 池 多人(NumberPeople≥2)取消 | track_manager.go pending | count-gate | 需 **多-track/计数层**（超 v1 单实体 scope，§7） |
| 8 | checkLostFall：离门/门区不入池 | track_manager.go `checkLostFall` | geom-gate | A_T: 门区消失走 `→justleft` 而非 `→lost` |
| 9 | checkLostFall：still-box≥60s 不入池（归 Still-fall） | 同上 + replay/shadow `MovingPreconditionMs` | dwell-gate | HSMM: `StandWalk` 驻留时长分流 lost vs still |
| 10 | lostWhileMoving age-ramp | belief_adapter.go `lostWhileMovingToObs` | 合成 obs | 发射 `P(no-detect\|Real-lost)` + HSMM 驻留 |
| 11 | Decider 90s 确认窗 | belief.go `confirmMs` | 外部 timer | HSMM: `Fallen` 状态时长分布 |
| 12 | still-fall 排除 ghost track | bathroom_fall.go:354 | obs-gate | 同 #1（Ghost 不作姿态主体） |
| 13 | 多人抑制（非"1真人+1ghost"） | bathroom_fall.go `suppressedByMultiResident` ~386 | obs-gate | 多-track/计数层（§7） |
| 14 | np=0 似然"强证据空房"+压 Fallen 0.3 | likelihood.go:61 | L 误标定 | 改为弱 + SFallen 中性；强 Empty 拉力归门区发射（见 #2） |
| 15 | frozen 帧 Conf=0→不更新 | belief_adapter.go `motionFresh` | L 门控（已对） | 保留：等价 `P(o\|s)=I`，是正确的缺证据处理 |

观察：#1/#3/#12 同一条 `Ghost→benign`；#2/#5/#6/#8 同一簇 `Real-lost↔justleft↔present` 转移；
#9/#10/#11 同一类 HSMM 驻留。**15 个手补点坍缩成 3 类 DBN 参数。**

## 6. 迁移分期（建议）

- **P1 Track 层只读 shadow** ✅ **【2026-06-01 已实施 + 部署】**：在 belief_shadow 旁加 A_T 推断，log T_t，对照现有 gate 是否一致（零风险 oracle）。见 §6.1。
- **P2 absence 发射**：把 `lostWhileMovingToObs` 换成 `P(no-detect|T,S)` 正式发射，删合成 ramp。
- **P3 HSMM 驻留**：Fallen/Left/Still 三态上时长分布，吃掉 confirmMs / MovingPreconditionMs / StillSec 阈值。
- **P4 收敛**：gate（method-2/门区/池 admit）逐条删除，行为由 A_T 接管；belief_gate_to_matrix.md 标"已内化"。
- **P5 多-track 层**：解 #7/#13 多人场景，退 v1 单实体限制。

### 6.1 P1 实施记录（2026-06-01）

**代码**：`belief/track.go`（纯模型：T 五态 + A_T + 发射 + TrackBelief 滤波）+ `belief/track_test.go`（合成结构验证）+ `belief_replay_test.go` 的 `TestTrackLayerOracle`（真机 oracle）+ `belief_shadow.go` wire（per-track T_t，**只 log `belief_shadow_track_lost` 不 fire**）。build/vet/test（-race）全绿，sensor 已 restart 部署（2026-06-01 15:34 PDT，0 panic）。

**T 状态**：`{None, Real, Ghost, JustLeft, Lost}`，各带 A_T 行。

**三条结构核心（替代 gate，非外部 if）**：
1. **method-2「失锁前全 ghost 不 fire」= `A_T: Ghost→Lost = floor`**。ghost 闪灭经 A_T 走 →None。
2. **门区 exit 推断 = absent 发射 geom 条件 + `TObsExit` 事件**：门区/ExitRoom → JustLeft，不 Lost。
3. **关键纪律：absent 发射不区分 Lost vs None**（二者都"无检出"，P(absent|·) 等价）；谁是 Lost 谁是 None **只由消失前 T 先验经 A_T 决定**（Ghost 先验→只通 None；Real 先验→通 Lost/JustLeft）。这是缺失段①的本质——存在性记忆在 T 层，不在发射。

**真机 oracle 结果**（`maxTLost`=峰值"判定丢失"，喂 Room 层候选的触发信号）：

| case | maxTLost | T 判定 | 说明 |
|---|---|---|---|
| D5F7 真跌倒(101测试) | 0.740 | **Lost** ✓ | 真人开阔地板走动消失 |
| D5F7 +ghost verdict | 0.050 | None ✓ | `Ghost→Lost≈0` 生效 |
| MoM 走出 exit | 0.097 | JustLeft ✓ | ExitRoom 事件路由 |
| cd2b 人返回 | 0.093 | None | recapture（额外正确） |
| **D523 9h同型** | 0.096 | None | **床区消失→None，9h bug 结构性修对** |
| cabb / Hunzi 静止站立 | 0.74 / 0.72 | Lost | **P1 不分**（FP 性=dwell=P3），诚实诊断 |
| D5F7 裸replay(无verdict) | 0.783 | Lost | 证明 verdict 是 ghost→None 的关键依赖 |

**P1 能结构化分开**：ghost 类（→None）+ 门区/exit 类（→JustLeft），即 §5 表 4 大冲突里的 2 类。**P1 还分不开**：静止-vs-走动丢失（cabb/D523-static/Hunzi 在存在性 T 层都是真 Lost；FP 性来自 moving-precondition = dwell 层）→ P3 HSMM。`endTLost` 因真跌倒被重新检出为 lying / 人返回 recapture 而衰减，故用 `maxTLost`（峰值）当判据——peak 是触发，衰减归 dwell 轴。

**关键依赖确认**：ghost→None 完全依赖 production ghost adjudicator 的 verdict（裸 replay 无 verdict→误判 Lost 0.78）。production shadow 沿用 Verdict，故结构成立。

## 7. Scope / 待决

- v1 单实体（一房一人）；#7/#13 多人/计数需 track 层扩成多实例或显式 count 维 → P5。
- A_T 标定数据源：ai:track:verdict:stream（ghost verdict）+ doc/cases 真机回放当 oracle。
- 仍守 [[feedback_no_dynamic_threshold_modulation]]：派生信号不进 alarm 决策——DBN 是白盒确定性滤波，
  verdict 作发射证据而非阈值调制，符合此铁律。
