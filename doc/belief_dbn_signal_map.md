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
慢层(天)   Z_cell 空间语义belief ──参数化──┐         (cell.go, 已存, 带语义遗忘)
                                          ↓
快层(秒)   观测 o_t ──L(o|s)──> 隐节点 ──A──> 隐节点'  ──读出──> 决策 τ*
            pose/z/XY/         R_i  realness         fall / count / risk
            sleepad/vital/     M_i  motion
            door-dist/event    T    existence(5态, belief/track.go 已存 P1)
                               O_b  bed occupancy(bed_bayesian 已存)
                               S_room 9态(belief/ 已存)
                               N_r  room 0/1
标定(慢)   feedback 真值 + 自学习 → 更新 Z_cell / S_vol / speed-cap
上游门     device_fitness / dedup / 防loop → 决定哪些 o_t 进网
```

核心:**已有三个 Bayesian-ish 层(cell belief / bed_bayesian / room 9态 HMM)+ 一个加性 scorer(fall_verify)**。DBN = 收敛它们 + 补 realness/dwell 两段,不是 greenfield(详 §9)。

---

## 1. 节点层(hidden states)

| 节点 | 状态空间 | 时间尺度 | 现有实现 | 现态(硬/软/学) | 来源 |
|---|---|---|---|---|---|
| **R_i** realness | real / ghost | track 生命(秒–分) | ghost_adjudicator 硬 verdict + GhostPenalty 软累积 | 半软 | bathroom_ghost.go, track.go |
| **M_i** motion | moving / static | 秒 | still-box(BoxRangeWithinMs) | 硬 | track.go:350 |
| **posture** | (lying / sit / stand) | 秒 | CorePose 一次性映射 | 硬,**部分不可辨识** | cell.go:57 |
| **O_b** bed 占用 | empty / occupied | 秒–分 | **bed_bayesian_scorer(log-odds HMM)** ✓ | 软 | bed_bayesian_scorer.go |
| **N_r** room 人数 | 0 / 1 (/2+) | 秒 | max(radar_np, bed_np) 硬 | 硬 | stream_publisher.go:426 |
| **Z_cell** 空间语义 | AreaType × conf | **天** | **cell belief + 语义遗忘** ✓ | 学习+软 | cell.go:225,553 |
| **S_room** 9态 | S0–S8 | 秒 | **belief/ HMM forward** ✓ | 软 | belief/state.go, likelihood.go |
| **T** existence | None/Real/Ghost/JustLeft/Lost | 秒 | **belief/track.go(P1 已部署)** ✓ | 软(仅 shadow log) | belief/track.go |

**可辨识性边界(A 类结论)**:pose/z 不可信 ⟹ posture 的 {sit/stand/lying} 三分**不可由 pose 单独辨识**;只有
- moving vs static(XY 可信)、
- lying-on-ground(pose-lying ln18 + immobility + 非休息区 + z-high-veto)、
- in-bed(sleepad 接触)
可辨识。{sit vs stand} 仅在 **z=high(锁 stand)** + **cell 语义(沙发→sit)** + **dwell** 三者下条件可辨识。

---

## 2. 发射层 L(o|s)

每条观测按**真实方向性**标 LR。硬闸做不到方向性,这是软 > 硬的根。

| 观测 | 可信性(A类) | LR 方向 | 喂节点 | 来源 |
|---|---|---|---|---|
| XY 位置 / still-box(50×50 box) | 可信 | moving↔static | M_i, R_i, dwell | track.go:350; StillBoxCm=50 fall_rules_param.go:162 |
| **pose** | 混淆矩阵 | **lying ln18 / sit ln4(特异)/ stand ≈0(模糊)** | posture, S_room | A类round2; RadarPoseToCore cell.go:57 |
| **z = high** | **可信** | upright 确认 / **fall veto** / 锁 stand / cell 解偏 | posture, Z_cell | A类round3 |
| **z = low/0** | **不可信(假低)** | **LR≈1,禁当 fall 确认** | — | A类round3 |
| sleepad 接触(InBed/LeftBed) | 可信,独立模态 | facility ±ln19 / home ±ln9 | O_b | bed_bayesian:31 |
| sleepad vital sustain | 弱 | +ln2 | O_b | bed_bayesian:39 |
| radar vital / pose_lying | 弱(pose 不可信) | 现 ln4.25,**应降到 ≈0** | O_b | bed_bayesian:35 ←§10 待修 |
| **WeakBio(HR/RR/Apnea)** | **可信,独立模态** | **≥80→force Real**;score=max(raw)+5·HR+5·RR+15·Apnea | **R_i + 健康事件** | fall_verify.go:189; aggregator:467 |
| firmware-fall pose=5 | 低(pose 不可信) | belief 现 ×10,**应降** | S_room/fall | likelihood.go:52 ←§10 待修 |
| **door-distance 趋势↓** | **可信 XY** | JustLeft 发射门(替 leftRoom event) | T, lost-fall | A类round5; ExitDistMinCm=30 |
| ObsNoDetect(消失) | — | Fallen ×1.6(贴地遮挡)/ StandWalk ×0.3 | S_room, T | likelihood.go:100 |
| kinematics Δz | **z 不可信** | 现 fall ramp(用 z↓) | S_room | likelihood.go:19 ←§10 待修(z 单边) |
| Kalman 残差 / 隐含速度 | 可信 XY | 跳变→ghost(逐步极值) | R_i | track.go:167; MaxKalmanResidual |
| **avg-speed vs 60–90** | 可信 XY(**仅 moving**) | 带内→real / 0→无信息 / 超→ghost | R_i | A类round4; per-device EWMA belief_adapter:37 |
| mirror 3 不变量 | 可信 XY | 方向<15°+中点共线⊥v+同步<0.4 → ghost | R_i | mirror_detect.go:288 |
| EnterRoom event | 半可信 | +real / S_room +stand ×2 | R_i, S_room | likelihood.go:58 |
| ExitRoom event | **不可靠(L3)** | →Left ×8(**L3 盲区假触发**) | S_room, T | likelihood.go:58; L3 |
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
| recapture→real | birth 重生取消 pending | track_manager:1830 | A_T lost→real | belief_dbn §5#6 |
| 占用 enter/exit | score ±50 | state_machine.go:97 | (或并入 O_b) | state_machine |
| hysteresis 防抖 | 3s(bed)/2s(room) | state_machine | 转移平滑 | state_machine.go:89 |
| enter/leave latch | 10s/2s 窗拒反向 | scorer.go:132 | 转移不应性 | scorer |
| self-contradiction 回滚 | 翻 Occupied 后 15s 内 score<30 → rollback | zoneengine engine.go:368 | 低置信转移撤销 | engine |
| subset-invariant 修复 | bed present→room occupied;stale-bed 5min 降级;orphan-room drop | zoneengine engine.go:512 | 结构约束(节点间一致性) | engine |
| LeavingTimeout | leaving 8s/5s → vacant | state_machine.go:129 | dwell 转移 | state_machine |

---

## 4. 先验层 π

| 先验 | 值 | 喂 | 现态 | 来源 |
|---|---|---|---|---|
| **Z_cell 空间语义** | AreaType+conf,语义半衰期(床9d/沙发7d/卫浴60d/走道14d/即时15min) | posture/fall/dwell | 学习+遗忘 | cell.go:278 |
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

> **C3 不是"没数据" —— feedback 通道一直在产真值标签。** 可行性卡在标定准不准 + 生存函数尾形对不对,可用 P1 oracle 反验。

---

## 8. 上游门(决定哪些 o_t 进网)

| 门 | 效果 | 来源 |
|---|---|---|
| **device_fitness** | Offline/SensorDetached/AngleException/SignalPoor → unfit → 不喂 engine | device_fitness_tracker.go:31 |
| bed event dedup | per-device per-kind 10s 防 spam | bed_event_dedup.go |
| stale message drop | radar 6s / sleepace 30s / monitor 6s | consumers |
| still-box 冻结命门 | StillBoxRunStart≥120s → Conf=0 不更新 | belief_adapter.go:21 |
| ghost jump 剔除 | >200cm/s 单跳 → 剔 | belief_adapter.go:169 |
| 防 loop | 自家 producer(slot fd00:0:fff1::/48)跳过 | consumers |

> 关系 A 类:bathroom 突发丢信号若 firmware 报 **SignalPoor → device_fitness 直接挡**;没报的才落到 door-distance 过滤(§2/§3)。

---

## 9. 现有可复用件(DBN 不是 greenfield)

| 件 | 复用为 | 文件 |
|---|---|---|
| bed_bayesian_scorer | O_b 节点 + 通用 log-odds filter 模板(LR/γ/leak/covers) | bed_bayesian_scorer.go |
| cell belief + decay | Z_cell 慢先验层 | cell.go |
| belief/ 9态 HMM + likelihood 矩阵 | S_room 层(已标定一版:FirmwareFall×10/ExitRoom→Left×8/NoDetect→Fallen×1.6) | belief/ |
| belief/track.go P1 | T existence 五态(已部署,仅 shadow log) | belief/track.go |
| fall_verify 加性 scorer | fall 读出(baseline+factors→30/70)→ 形式化为标定 log-odds | fall_verify.go |
| RiskLevel 分级 | τ\* 分级对接 | risk_evaluator.go |
| feedback 通道 | C3 标定标签源 | feedback.go |
| leak/decay(bed leak 0.55 / cell 语义半衰期) | Boyen-Koller mixing 率(保因子化滤波有界) | bed_bayesian, cell |

---

## 10. 未归位 / 待决(gaps —— 本表暴露的)

1. **vital 健康事件无节点**:WeakBio/HR/RR/Apnea 现仅作 realness boost(≥80→Real),但心率/呼吸/呼吸暂停异常**本身是健康事件**,DBN 缺这一列输出(与 fall 正交)。
2. **firmware-fall pose=5 的 LR ×10 与"pose 不可信"矛盾**:likelihood.go:52 + fall_verify 给 firmware fall 极高权,需按混淆矩阵**重标定降权**。
3. **kinematics Δz ramp 用 z↓**:likelihood.go:19 把 z 骤降当 fall 强证据,违反"z 低不可信",需改成 **z 单边**(高→veto,低→LR≈1)。
4. **radar pose_lying LR=ln4.25**:bed_bayesian 给 radar 躺姿过高权,pose 不可信下应降到 ≈0,bed 靠 sleepad 接触扛。
5. **三个 Bayesian 层未统一**:cell / bed / room-9态 各跑各的 + fall_verify 加性 scorer 第四套 —— 需收敛到单一因子化框架(Boyen-Koller 跨层有界)。
6. **循环偏置**:Z_cell 学自 pose(ActiveType←RadarPoseToCore)→ 继承 70/20 混淆 → 有偏先验×有偏似然。解偏:Source 真值注入(已有)+ LieRetract(已有)+ **用 z=high 学 cell 语义(ai_fall_model 在做,待接)**。
7. **多人/count>1**:v1 单实体(belief_dbn §7 #7/#13),多 resident 抑制 / room count>1 待 P5。
8. **Warning 缺席层是否进同一 DBN**:Stay/LeftBed/NightAbsence 是缺席驻留,同族 S_vol 但 anchor 不同,scope 待定。
9. **R_i 静止期的记忆依赖**:摔倒瞬间 v≈0,realness 必须靠摔前走路累积的 L_R(带记忆 filter)—— 现 ghost verdict 是逐帧/累积混合,需确认记忆衰减率。
10. **R_i–S_i 耦合强弱**(belief_dbn 待决):ghost 运动学胎记 ⟹ R_i⊥S_i 不成立,倾向 {R_i,S_i} 同簇联合滤波(4态),需定。

---

## 11. 下一步(纯定量,不写生产代码)

用 **P1 shadow 的 7-case oracle**(belief_dbn §6.1:cabb/D5F7/D523/MoM/cd2b…,裸存在性层 cabb 0.74 vs D5F7 0.74 分不开),代入本表的标定:
- 发射:pose 混淆(70/20)+ z 单边 veto + ObsNoDetect
- realness:出生地 + 60–90 avg-speed + 逐步连续性 + WeakBio + 记忆衰减
- 先验:Z_cell zone(cabb 那个 cell 有没有被 ToleratedStill 学成"可久站")
- 驻留:开阔地 8min 生存函数 $-\ln S_{vol}(d)$

算每个 case 的 **fall 后验轨迹**,看:
1. cabb(开阔地真人久站 FP)能否被 z-veto / cell-tolerance / realness 压在 $\tau^\*$ 下?
2. D5F7(真摔)能否过 $\tau^\*$?
3. margin 多大 → **DBN 值不值得做的数学判据**。
