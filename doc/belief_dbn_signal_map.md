# DBN 信号/规则总映射 — 底稿

状态:**底稿(single source of truth)**。用途:(1) fall 后验定量论证的标定清单;(2) 后续 DBN 设计施工图。
关联:[[belief_dbn_completeness]](Track×Room DBN 蓝图)、fall_unified_statemachine.md(silent/lost/moving)、
bed_bayesian_review.md、sensor_v2_known_limitations.md(L1–L5)。

本文把**两个来源**合并归位:
- **A 类(对话标定)**:与用户五轮讨论得到的信号物理特性 + 人类先验。
- **B 类(代码现存)**:扫 sensor 全量文件抽出的规则/阈值/机制(标 `file:line`)。

归位原则:每条信号/规则归到 DBN 的一个角色 —— **节点(hidden state) / 发射 L(o|s) / 转移 A / 先验 π / 驻留 HSMM S_vol(t) / 决策 τ\* / 标定 C3 / 上游门**。归不进的进 §10 待决。

---

## 0. 网络结构总览

```
[Cell Engine] 独立子系统(慢,自学习/自衰减) ── 不是 DBN 的层,DBN 只读其结果
   输入: track still-box(传入,不重算)/traverse/z档/feedback
   输出: per-cell AreaType + Confidence + Source
        │ exogenous, read-only(唯一耦合边)
        ▼
[DBN] 快层(per-tick,单时间尺度)
   观测 o_t ──L(o|s)──> 隐节点 ──A──> 隐节点' ──读出──> 决策 τ*
    pose/z/XY/sleepad    R_i realness         fall / count / risk
    vital/door-dist      M_i motion
    /event               T   existence(5态, belief/track.go P1)
                         O_b bed(bed_bayesian) / S_room 9态(belief/) / N_r room 0/1
                         ↑ 读 Z_cell(AreaType) 当 π(zone) + 选 S_vol(t|zone)
上游门  device_fitness / dedup / 防loop → 决定哪些 o_t 进网
```

核心:**cell engine 是独立子系统,DBN 只读它的 AreaType 结果(§4 / §11.5 接口契约)。** DBN 自身要收敛的是**两个 Bayesian 层(bed_bayesian / room 9态 HMM)+ 一个加性 scorer(fall_verify)**,补 realness/dwell 两段,不是 greenfield(详 §9)。**DBN 单时间尺度(快层);慢层(cell 语义学习)归 cell engine,不进 DBN 联合推断。**

---

## 1. 节点层(hidden states)

| 节点 | 状态空间 | 时间尺度 | 现有实现 | 现态(硬/软/学) | 来源 |
|---|---|---|---|---|---|
| **R_i** realness | real / ghost | track 生命(秒–分) | ghost_adjudicator 硬 verdict + GhostPenalty 软累积 | 半软 | bathroom_ghost.go, track.go |
| **M_i** motion | moving / static | 秒 | still-box(BoxRangeWithinMs) | 硬 | track.go:350 |
| **posture** | (lying / sit / stand) | 秒 | CorePose 一次性映射 | 硬,**部分不可辨识** | cell.go:57 |
| **O_b** bed 占用 | empty / occupied | 秒–分 | **bed_bayesian_scorer(log-odds HMM)** ✓ | 软 | bed_bayesian_scorer.go |
| **N_r** room 人数 | 0 / 1 (/2+) | 秒 | max(radar_np, bed_np) 硬 | 硬 | stream_publisher.go:426 |
| **S_room** 9态 | S0–S8 | 秒 | **belief/ HMM forward** ✓ | 软 | belief/state.go, likelihood.go |
| **T** existence | None/Real/Ghost/JustLeft/Lost | 秒 | **belief/track.go(P1 已部署)** ✓ | 软(仅 shadow log) | belief/track.go |
| **P_id** suite 身份 | resident / visitor / none(+ AnchorRoomType) | 分–时 | **suite_census**(radar age≥5min∧traverse≥10 提 resident;visitor 2min∧5;**sleepad 双锚即时**;idle 衰减 visitor 10min/resident 6h;失锁 60s clear) | 硬+衰减 | suite_census.go:97; engine.go:945 |

> **Z_cell 不是 DBN 节点** —— cell 语义由独立 cell engine 学,DBN 只读其 AreaType 当 π(zone),见 §4 / §11.5。

**可辨识性边界(A 类结论,经 z 三档修正)**:
- **moving vs static**:XY 可信,辨识。
- **stand vs sit**:**z>30 时由 z 档直接分**(z>80→stand / 30–80→sit,round-z3);仅 **z<30(假低)时**退回 cell 语义 + dwell。
- **lying-on-ground**:靠 **pose-lying(正向 ln18)+ immobility + 非休息区 + realness**,**不靠 z**(z<30 段假低,且 pose/z 不进 fall 压制,见 §2)。
- **in-bed**:sleepad 接触(独立模态)。

---

## 2. 发射层 L(o|s)

每条观测按**真实方向性**标 LR。硬闸做不到方向性,这是软 > 硬的根。

| 观测 | 可信性(A类) | LR 方向 | 喂节点 | 来源 |
|---|---|---|---|---|
| XY 位置 / still-box(50×50 box) | 可信 | moving↔static | M_i, R_i, dwell | track.go:350; StillBoxCm=50 fall_rules_param.go:162 |
| **pose**(对 fall) | 混淆矩阵 | **只正向**:lying→+fall;**stand/sit→中性(LR=1),永不压 fall** | fall | A类round2/round-posonly; cell.go:57 |
| **pose**(对 posture) | 混淆矩阵 | lying 强 / sit 特异 / stand 模糊 | posture, S_room | 同上 |
| **z>80** | 可信(高值多真) | **+stand** | posture, Z_cell(传 cell engine) | A类round-z3 |
| **z 30–80** | 较可信 | **+sit** | posture, Z_cell | A类round-z3 |
| **z<30/0** | 噪声(静止/sit 假低) | **无信息(LR≈1)**;**z 全程不进 fall**(不确认不否决) | — | A类round-z3/round-posonly |
| sleepad 接触(InBed/LeftBed) | 可信,独立模态 | facility ±ln19 / home ±ln9 | O_b | bed_bayesian:31 |
| sleepad vital sustain | 弱 | +ln2 | O_b | bed_bayesian:39 |
| radar vital / pose_lying | 弱(pose 不可信) | 现 ln4.25,**应降到 ≈0** | O_b | bed_bayesian:35 ←§10 待修 |
| **WeakBio(HR/RR/Apnea)** | **可信,独立模态** | **≥80→force Real**;score=max(raw)+5·HR+5·RR+15·Apnea | **R_i + 健康事件** | fall_verify.go:189; aggregator:467 |
| firmware-fall pose=5 | 低(pose 不可信) | belief 现 ×10,**应降** | S_room/fall | likelihood.go:52 ←§10 待修 |
| **door-distance 趋势↓** | **可信 XY** | **补缺失的 enter/left event**(信号丢失时):朝门↓→JustLeft | T, lost-fall | A类round5; ExitDistMinCm=30 |
| ObsNoDetect(消失) | — | Fallen ×1.6(贴地遮挡)/ StandWalk ×0.3 | S_room, T | likelihood.go:100 |
| ~~kinematics Δz~~ | **z 不可信** | **删**:z↓→0 是环境噪声,不能当 fall 正向证据 | — | likelihood.go:19 ←§10#3 待删 |
| Kalman 残差 / 隐含速度 | 可信 XY | 跳变→ghost(**逐步极值,不平均**) | R_i | track.go:167; MaxKalmanResidual |
| **avg-speed(室内-老人带)** | 可信 XY(**仅 moving**) | 带内→real / 0→无信息 / **超室内天花板→ghost** | R_i | A类round4/round-室内; per-device EWMA belief_adapter:37 |
| **空间跳跃(逐步)** | 可信 XY | **P(瞬移\|real elderly)≈0 → 跳变=近确定性 ghost** | R_i | A类round-室内; isGhostJump |
| **冻结伪迹(方差)** | 可信 XY | **真目标有 ±10–20cm 抖;ghost/冻结近零方差 + pose/z 锁死 → ghost**(≠ box 内真静止) | R_i | A类round-cd2b; cf. still-box 只看 box 范围 |
| mirror 3 不变量 | 可信 XY | 方向<15°+中点共线⊥v+同步<0.4 → ghost | R_i | mirror_detect.go:288 |
| EnterRoom event | **可信,正向**(event 在=真进入) | +real / S_room +stand ×2;**absence≠没进** | R_i, S_room | likelihood.go:58 |
| ExitRoom event | **可信,正向**(event 在=真离开) | →Left ×8;**absence≠还在**(信号丢失时不发 event)→ 缺失由 door-distance 补 | S_room, T | likelihood.go:58 |
| ObsBedOccupied(O_b 投影) | 软 | S_BedLying(1+5p)/S_Fallen(1−0.7p) | S_room | likelihood.go:34 |
| SleepStage | 弱 | stage{1,2,3}→Lying×2 / {0,8}→Restless×2 | S_room, O_b | likelihood.go:45; sleepstage_consumer |

---

## 3. 转移层 A(含 A_T existence chain)

| 转移 | 触发判据 | 现态 | DBN 内化 | 来源 |
|---|---|---|---|---|
| static→lost(silent 后消失) | prev=still 且 dwell≥阈 | track_manager 段4b | A_T silent→lost | fall_unified §2 |
| moving→lost | prev=moving | (缺,未分) | A_T moving-fall(新) | §2 |
| ghost→None(不→lost) | Ghost→Lost=floor | belief/track.go | A_T 结构核心 | belief_dbn §6.1 |
| real→lost vs justleft | **door-distance↓→justleft** | inferredDoorExit | A_T 发射门控 | A类round5; bathroom_fall:546 |
| recapture→real | birth 重生 | track_manager:1830 | A_T lost→real(**软恢复,非硬 cancel**:人返回可能是跌后自救,不能用 recapture 抹掉真摔;真摔自救留低 severity) | A类round-cd2b; belief_dbn §5#6 |
| 占用 enter/exit | score ±50 | state_machine.go:97 | (或并入 O_b) | state_machine |
| hysteresis 防抖 | 3s(bed)/2s(room) | state_machine | 转移平滑 | state_machine.go:89 |
| enter/leave latch | 10s/2s 窗拒反向 | scorer.go:132 | 转移不应性 | scorer |
| self-contradiction 回滚 | 翻 Occupied 后 15s 内 score<30 → rollback | zoneengine engine.go:368 | 低置信转移撤销 | engine |
| subset-invariant 修复 | bed present→room occupied;stale-bed 5min 降级;orphan-room drop | zoneengine engine.go:512 | 结构约束(节点间一致性) | engine |
| LeavingTimeout | leaving 8s/5s → vacant | state_machine.go:129 | dwell 转移 | state_machine |
| **bathroom room-type 锚翻转** | BathroomCount 0↔1(恰好 1 resident) | AdjustBathroomCount + 60s member timeout → 翻 sole-resident AnchorRoomType(default↔bathroom) | A 结构约束(dispatch→选 dwell 档) | bathroom_gate.go:37; suite_census AdjustBathroomCount; engine.go:974 |

---

## 4. 先验层 π

| 先验 | 值 | 喂 | 现态 | 来源 |
|---|---|---|---|---|
| **Z_cell 空间语义**(**外部只读**,来自独立 cell engine) | AreaType+conf+Source,语义半衰期(床9d/沙发7d/卫浴60d/走道14d/即时15min) | π(zone)→posture/fall;选 S_vol(t\|zone) | 学习+遗忘(**cell engine 内部,DBN 不卷入**) | cell.go:278;接口契约 §11.5 |
| risk-time 夜间 | **22:00–06:30**(config 默认 23:30–07:30) | dwell 阈切换 | 配置 | math_util.go:16; config.go:371 |
| birth 出生地 | 距门 dist(>70cm bathroom→ghost) | R_i 先验 | 硬 | bathroom_ghost.go:50; birth score |
| bed 夜/日 prior | L0 +1.39(夜)/−0.85(日) | O_b 初值 | 软 | bed_bayesian:65 |
| birth score 5 因子 | 进口速度/出生cell/后生ID/ghost多发区/enter配对 | R_i 先验 | 加性 | track_manager birthScore |
| Source 信任层级 | Human 1.0 > Feedback 0.6 > Learned 0.4 | π 折扣 | 软 | belief_adapter.go:79 |
| 多源亲和权重 | 同床5>同房3>同unit1;pose Sleepad8>Radar6;vital 8>4 | 多源融合 LR | 权重 | weights.go:36,69 |

---

## 5. 驻留层 HSMM S_vol(t)(生存函数 = 不同场景不同时长)

A 类核心:**老人站立静止 >8min 几乎不可能**;代码用 zone 条件化(避开 pose 不可信)。$\ln\mathrm{LR}_{fall}(d)=-\ln S_{vol}(d)$,平滑 ramp 取代硬阈值。

| 上下文 | 生存尾 | 现态 | 来源 |
|---|---|---|---|
| stand 开阔地 | **8min** | 硬阈 EffectiveStillTimeoutSec | stillTimeoutBase default; StandingMin **cap 8** aggregator:302 |
| toilet/shower | 10min(风险)/12min | 硬 | bathroom_fall.go:46 |
| bed/sofa(RestZone) | ∞(不限) | 硬 | IsRestZone cell.go:326 |
| deny zone | 5min | 硬 | DenyZoneSec |
| leftBed 床边 | 15min(夜) / 窗 3min | 硬 | bedroom_fall.go:41; BedsideFall.WindowSec=180 |
| ground lying | ~即时 | (经 lying+非bed) | — |
| **lost-fall wait by area** | bed 60min / walk 5min / deny 5min | 硬 | fall_rules_param.go:157 |
| moving precondition | 60s(still≥60s→不进lost) | 硬 | MovingPreconditionMs:165 |
| per-cell 自适应放宽 | ×[1, 2.0] by FakeAlarm/ToleratedStill | 学习 | toleranceFactor cell.go:400 |
| **Stay(浴室独处)** | 45min(日)/30min(夜) | 硬 | zonealarm rules.go:111 |
| **LeftBed 缺席** | 30min | 硬 | zonealarm rules.go:122 |
| **NightAbsence** | 30min(21:00–07:00) | 硬 | zonealarm rules.go:133 |

> 注:上半是 **fall 静止驻留**;下半三条(Stay/LeftBed/NightAbsence)是 **缺席驻留** —— 同族 S_vol,不同 anchor(独处/离床/离室)。DBN scope 应一并含。

---

## 6. 决策层 τ\*

| 决策 | 阈值 | 现态 | 来源 |
|---|---|---|---|
| firmware-fall 验证 | baseline 50 → 30(ghost)/70(real)/中间 suspect | **已是加性软 scorer** | fall_verify.go:54 |
| ghost verdict | Score 50(real)/20(ghost);Penalty ≥80 | 半软 | track.go:197 |
| 占用 enter/exit | ±50 | 硬 | state_machine.go |
| bed P 阈 | InBed 0.70(facility)/0.75(home);LeftBed 0.50;Standby |L|<0.5 | 软 | bed_bayesian:49 |
| **RiskLevel 分级** | **Attention / Risk**(浴室日 standing≥8/≥15;夜≥5/≥8;alone≥30/45) | **已是分级软输出** | risk_evaluator.go:39 |
| 速度 ghost 二档 | hard 200cm/s / soft 100cm/s(需 enter 反证) | 硬 | fall_rules_param.go:166 |
| **human-bed fall 总豁免** | Source=Human ∧ AreaBed ∧ **Conf≥99** → 所有 fall 路径不报(跨 silent/bedside/lost 共闸;Conf≥99 切人工 layout vs radar 自学习 95) | 硬总闸(decision veto) | fall_exempt.go:15; bedroom_fall.go:262,341; track_manager.go:1892 |

> τ\* 的 Bayes 形式:$\tau^\*=C_{FP}/(C_{FP}+C_{FN})$。现有 30/70 与 RiskLevel 分级**可直接对接** —— DBN 把散落阈值收敛到一个代价比旋钮。

---

## 7. 标定 / 学习层(C3 —— DBN 可行性的唯一硬前提)

| 机制 | 学到什么 | 触发 | 来源 |
|---|---|---|---|
| **feedback false_alarm** | 坐椅/沙发→AreaSit;Lounge永久→AreaBed;Lounge临时→2h禁报;电磁→GhostCount;pose误差→不学 | 人工反馈 | feedback.go:414 |
| **feedback verified 真摔** | ClearNonHumanLearnedZone + RealFallCount;sticky→LearnBlocked 永久封 | 人工确认 | feedback.go:474 |
| stand-static / region-static 自学 | AreaSit(双cell+90%容忍) | 久站/久坐 | track_manager region/stand-static |
| mirror / static-reflector | AreaDeny(≥3 episode) | 镜像对/金属反射 | mirror_detect, static_reflector |
| auto-deny 15 天门控 | AreaDeny | 长期无人走 | cell.go:175 |
| inside-enter 5 次 | AreaEnter(盲区入口) | 失锁+3s 重生 | cell.go:160 |
| per-device 走速 EWMA | speed-cap(1.5×,30–150) | 走动样本 | belief_adapter.go:37 |
| cell 容忍(FakeAlarm/ToleratedStill) | dwell 放宽 ×[1,2] | 假报/容忍站立 | cell.go:400 |

> **归属**:上表除 `per-device 走速 EWMA`(DBN/track)外,**feedback / stand-static / mirror / auto-deny / inside-enter / cell 容忍 全是 cell engine 内部学习(独立)**,DBN 只读其 AreaType 产物(§11.5)。cell engine 的"长静→sit(用 dwell 不用 pose)/ z 档语义 / 重复+衰减+真摔擦除护栏"属独立 cell-engine spec,不在本 DBN 图展开。
> **C3 不是"没数据" —— feedback 通道一直在产真值标签。** 可行性卡在标定准不准 + 生存函数尾形对不对,可用 P1 oracle 反验。

---

## 8. 上游门(决定哪些 o_t 进网)

| 门 | 效果 | 来源 |
|---|---|---|
| **device_fitness** | Offline/SensorDetached/AngleException/SignalPoor → unfit → 不喂 engine | device_fitness_tracker.go:31 |
| bed event dedup | per-device per-kind 10s 防 spam | bed_event_dedup.go |
| stale message drop | radar 6s / sleepace 30s / monitor 6s | consumers |
| still-box 冻结命门 | StillBoxRunStart≥120s → Conf=0 不更新 | belief_adapter.go:21 |
| ghost jump 剔除 | 现 >200cm/s;**室内-老人口径应降到 ~110–130cm/s**(cd2b 150cm/s 跳变在 200 闸下漏判) | belief_adapter.go:169 ←§10#11 待降标定 |
| 防 loop | 自家 producer(slot fd00:0:fff1::/48)跳过 | consumers |
| **bed-presence 新鲜窗** | sleepad/radar InBed >10min 视为 stale 不计占用;占用 = OR(fresh sleepad, fresh radar)→ 投 N_r | bed_presence_fusion.go:48 |

> **still-box 单源**:track 层每分钟算一次 still-box(50×50 box + StillBoxRunStart),结果**传**给 cell engine(`MarkDwell/MarkLongStill`)和 DBN(fall/M_i)**双消费**;cell engine **不重算**(§2.4 producer/maintainer,避免两套阈值 drift)。传的是 `(cell_xy, is_static, run_duration)`,与 fall 同源。

> 关系 A 类:bathroom 突发丢信号若 firmware 报 **SignalPoor → device_fitness 直接挡**;没报的才落到 door-distance 过滤(§2/§3)。

---

## 9. 现有可复用件(DBN 不是 greenfield)

| 件 | 复用为 | 文件 |
|---|---|---|
| bed_bayesian_scorer | O_b 节点 + 通用 log-odds filter 模板(LR/γ/leak/covers) | bed_bayesian_scorer.go |
| cell belief + decay | **独立 cell engine**(非 DBN 层),DBN 只读其 AreaType 输出 | cell.go |
| belief/ 9态 HMM + likelihood 矩阵 | S_room 层(已标定一版:FirmwareFall×10/ExitRoom→Left×8/NoDetect→Fallen×1.6) | belief/ |
| belief/track.go P1 | T existence 五态(已部署,仅 shadow log) | belief/track.go |
| fall_verify 加性 scorer | fall 读出(baseline+factors→30/70)→ 形式化为标定 log-odds | fall_verify.go |
| RiskLevel 分级 | τ\* 分级对接 | risk_evaluator.go |
| feedback 通道 | C3 标定标签源 | feedback.go |
| leak/decay(bed leak 0.55) | Boyen-Koller mixing 率(保 DBN **快层**因子化滤波有界;**单时间尺度**,慢层不进 DBN) | bed_bayesian |

---

## 10. 未归位 / 待决(gaps —— 本表暴露的)

1. **vital 健康事件无节点**:WeakBio/HR/RR/Apnea 现仅作 realness boost(≥80→Real),但心率/呼吸/呼吸暂停异常**本身是健康事件**,DBN 缺这一列输出(与 fall 正交)。
2. **firmware-fall pose=5 的 LR ×10 与"pose 不可信"矛盾**:likelihood.go:52 + fall_verify 给 firmware fall 极高权,需按混淆矩阵**重标定降权**。
3. **z/pose 退出 fall 压制(round-posonly)**:pose/z 对 fall **只能正向**,绝不负向。待改:(a) **删 likelihood.go:19 kinematics Δz↓**(z=0 噪声,连正向都不算);(b) 收回早期"z=high veto fall";(c) pose=stand/sit 对 fall 中性。fall 压制只走 realness/spatial/sleepad/recapture/human-bed。
4. **radar pose_lying LR=ln4.25**:bed_bayesian 给 radar 躺姿过高权,pose 不可信下应降到 ≈0,bed 靠 sleepad 接触扛。
5. **两个 Bayesian 层 + scorer 待统一**:bed / room-9态 + fall_verify 加性 scorer 收敛到单一因子化框架。**DBN 单时间尺度(快层)**;cell engine 是独立慢层,不在 DBN 联合推断内(早期"两时间尺度 hierarchical + 跨层 Boyen-Koller"已撤)。
6. **循环偏置 = cell engine 内部问题(非 DBN 耦合)**:Z_cell 学自 pose → 继承 70/20 混淆。解偏在 **cell engine 内部**:dwell(长静→sit,不用 pose) + z 档(>80 active/30–80 sit/<30+sleepad bed)+ Source 真值注入 + 跨天重复/衰减/真摔擦除护栏。DBN 这边只按 Source 置信读结果,不卷入学习回路。
7. **多人/count>1**:身份锚定层 **P_id(suite_census)已在跑**(单 resident anchor + sleepad 双锚 + bathroom 锚翻转,§1/§3 已归位);仍缺的是**多 resident 抑制 / room count>1 联合滤波**(belief_dbn §7 #7/#13),待 P5。
8. **Warning 缺席层是否进同一 DBN**:Stay/LeftBed/NightAbsence 是缺席驻留,同族 S_vol 但 anchor 不同,scope 待定。
9. **R_i 静止期的记忆依赖**:摔倒瞬间 v≈0,realness 必须靠摔前走路累积的 L_R(带记忆 filter)—— 现 ghost verdict 是逐帧/累积混合,需确认记忆衰减率。
10. **R_i–S_i 耦合强弱**(belief_dbn 待决):ghost 运动学胎记 ⟹ R_i⊥S_i 不成立,倾向 {R_i,S_i} 同簇联合滤波(4态),需定。
11. **速度闸标定太松(室内-老人)**:`ImpossibleSpeedCm=200 / SuspectSpeedCm=100` 是室外口径;室内空间小+家具,老人达不到 → 应降到 hard ~110–130。cd2b 150cm/s 跳变在 200 闸下只算 suspect 漏掉。空间跳跃应近确定性判 ghost。
12. **still-box 把"真静止"和"冻结伪迹"混了**:`BoxRangeWithinMs≤50cm` 只看范围;真人静止有 ±10–20cm 抖 + pose/z 偶变,ghost/冻结锁死(零方差 + pose/z 恒定)都落在 50cm 内 → 分不开("still-box 卡死")。需加方差/抖动判据:零方差 + pose/z 锁死 = 伪迹(R_i→ghost)。

**故意不收(符合"现态"口径)**:`LeftBedFall*`(weights.go:91–94)、`VitalDerive/DeriveBedStateThreshold`(weights.go:100–105)仅常量无 consumer = 未接线占位,不入本表。

---

## 11.5 cell engine ↔ DBN 接口契约(唯一耦合边)

cell engine 独立(自学习/自衰减/自带护栏),DBN 只读其结果。**唯一一条边:**

| cell engine 输出(read-only) | DBN 怎么用 |
|---|---|
| AreaType(cell) | π(zone):rest/active/bed/toilet → fall 抑制(rest)/escalate(toilet/open) |
| Confidence + Source(Human/Feedback/Learned) | 先验权重(Human 1.0 > Feedback 0.6 > Learned 0.4);human-bed Conf≥99 = fall 总豁免(§6) |
| AreaType → S_vol(t\|zone) 选档 | 驻留生存函数:stand 8min / toilet 15min / rest ∞ |

**反向只有一条 producer→consumer**:track 层 still-box → cell engine(`MarkDwell/MarkLongStill`,§8)。cell engine 内部学习(dwell vs traverse / z 档 / 长静→sit / 护栏)**不在本 DBN 图**,属独立 cell-engine spec。

---

## 11. 定量结论(基于真实 case profile)

标定:发射(pose 正向only / z 三档)+ realness(出生地 + 室内-老人 avg-speed + 逐步跳跃 + 冻结方差 + WeakBio + 记忆)+ π(Z_cell)+ 驻留 $-\ln S_{vol}$。$\tau^\*=C_{FP}/(C_{FP}+C_{FN})$。

### 11.1 可分的(DBN 赢)
- **cabb-0605 真摔(firmware 给 pose=2 lying 52s)**:pose-lying 正向 → 过 $\tau^\*$。✓
- **坐姿/休息区 FP、ghost、门区 exit**:靠 realness / Z_cell rest-zone / door-distance 干净路由出去。

### 11.2 不可分对(信息论残差)——只剩 Z_cell 一个杠杆
- **cabb-0606(firmware 漏判的静止真摔,全程 pose=4)vs cabb-0603(冻结 FP)**:pose=4 / z=0 / real / 无 exit **全等**,pose/z 判别量全 null。唯一分离 = **Z_cell 空间容忍**(水池"可久站" vs 角落)。有 cell 学习则 margin ≈1.5 nat 可分;无则重合,任何 τ\* 分不开。

### 11.3 cd2b(201 bedroom)——冻结 ghost,R_i 是正解(纠正早期"不可约")
- 真相:**跌床 → 雷达丢真人 → track 跳椅子并冻住(ghost)→ 被当人压掉 lost-fall → 漏报**。
- 判别(全 reliable,**不碰 pose/z**):跳速 150cm/s 超室内-老人天花板 + 卡死近零方差 + 距门远 + 连续性断 → **R_i 判 ghost** → 真人 Lost 浮出 → lost-fall 应报。
- **recapture(5.85min 返回)= 跌后自救,不能硬 cancel**(否则抹掉自救型真摔)。
- 结论:**R_i(运动学)能分,DBN 在此有正价值**;下限在"区分冻结 ghost vs 真人静止"要靠方差判据(§10#12),不是现 still-box 的 50cm box。

### 11.4 总判据
**DBN 值得做**:赢在可分多数 + 不是 greenfield。**治不了**:firmware 漏判的静止真摔 vs 冻结 FP(§11.2)只剩 Z_cell 一根杠杆;真正下限要加传感器(多雷达/bedside 垫),非滤波层。**最该先做**:R_i 的运动学/方差判据(§10#11/#12)+ Z_cell 解偏学习(cell engine,§10#6)。
