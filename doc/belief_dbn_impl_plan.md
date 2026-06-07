# DBN 跌倒检测 — 分节施工计划(P-task 路线图)

状态:**施工计划(委员会审计中)**。本文把 `belief_dbn_signal_map.md`(§0–§11)与
`belief_dbn_proposal.html`(委员会评审稿)拆成**可施工、可验收、可灰度**的 P-task。
每节一次 commit&push;委员会每 ~10min 从 github 审,反馈写 `doc/feedback.md`,
施工方每节开工前先 pull 读 feedback,有反馈先消化再推进。

> **本文只规划,不写生产代码。** 代码动作待委员会签字单节计划后,另起 commit。

---

## §0 总纲与施工铁律

### 0.1 目标(一句话)
把散落、互斗的 ~100 条硬阈值规则,收敛成**单一因子化 DBN(白盒确定性滤波)**:
每条信号按真实可信方向贡献 log-LR,跌倒判定从"15 个互斗的闸"变成"一个按代价比
可调的阈值 $\tau^\*=C_{FP}/(C_{FP}+C_{FN})$"。**不是 greenfield** —— 系统已有三层
贝叶斯雏形(bed_bayesian / room 9 态 HMM / belief P1 T-existence shadow),本计划是
统一它们 + 补 realness/dwell 两段。

### 0.2 施工铁律(每个 P-task 必须遵守,违反先改本节)

| # | 铁律 | 出处 |
|---|---|---|
| R0 | **shadow-first**:所有 DBN 新逻辑先进 `belief_shadow.go` 旁路,只 log 不 fire;与 gate-list 离线对账过关再谈 canary。绝不直接接 alarm。 | belief_shadow.go; 记忆[belief_state_rule_engine_reframe] |
| R1 | **不碰 alarm 决策路径**:派生/趋势/聚合信号不进 verdict/severity/routing/dedup/timing。DBN 后验在 shadow 期只读出到 log。 | 记忆[feedback_no_dynamic_threshold_modulation] |
| R2 | **DBN 单时间尺度(快层)**:慢层(cell 语义跨天学习)归独立 cell engine,**不进 DBN 联合推断**。唯一耦合边见 §9。 | signal_map §0/§11.5 |
| R3 | **cell engine 只读**:DBN 只读 `AreaType+Conf+Source`,不写不卷入其学习回路。 | signal_map §11.5 |
| R4 | **producer-first**:跨服务/跨层字段,producer(写入/publish)先彻底完工,downstream/FE 再做。 | 记忆[feedback_producer_first] |
| R5 | **pose/z 对 fall 只正向**:lying→+fall;stand/sit/z↓→中性(LR=1),永不压 fall。 | signal_map §2/§10#3 |
| R6 | **改前读 design、改后枚举边界复查**;完成态断言前先 grep 核查,不说"已 done"未验证。 | 记忆[feedback_read_design_before_modify / feedback_no_unverified_claims] |
| R7 | **常量化**:事件名/类别字面量唯一来源 `owl-common/observation`;标定参数集中,不散落字面量。 | CLAUDE.md §1.1 |

### 0.3 验收单模板(DoD,每个 P-task 套用)
1. **范围**:动哪些文件、新增哪些 shadow 字段。
2. **标定来源**:每条 LR/转移/生存函数的数值出处(A 类对话标定 / B 类代码现存 / 真机 fixture)。
3. **shadow log 字段**:旁路输出什么,供对账。
4. **oracle case**:用 §8 的哪个 fixture 验证,期望 margin 方向。
5. **回归**:`go vet ./... && go build ./... && go test ./internal/roomengine/...` 全绿。
6. **边界枚举**:列穷举失败场景(信号丢失/多人/ghost/offline)的预期行为。
7. **灰度门**:shadow 对账通过率 / canary 范围 / 可秒关开关。

### 0.4 P-task 依赖 DAG(施工顺序)
```
P2 发射标定 ──┬─> P3 realness ──┐
(emission)    │   (R_i)         ├─> P9 oracle 定量验收(go/no-go 闸)
              ├─> P4 dwell ─────┤
              │   (S_vol)        │
P5 bed O_b ───┤                 │
P6 room/N_r ──┴─> P7 decision τ* ┘
P8 health(与 fall 正交,可并行)
P0 cell-engine 接口契约(§9,P2/P4 的前置只读边)
```
**关键路径**:P2 → {P3,P4} → P7 → P9。P9 oracle 是**go/no-go 数学闸**:
margin 不够则 DBN 不值得继续(signal_map §11.4),止于 shadow。

### 0.5 编号约定
- `P<n>` = 大施工阶段(对应本文 §1–§9 各节)。
- `P<n>.<m>` = 该阶段下可独立 commit 的子任务。
- 编号只增不复用;废弃用删除线保留占位。

### 0.6 本计划的节(每节一次 commit&push)
| 节 | P-task | 一句话 | 关键文件 |
|---|---|---|---|
| §1 | **P2** 发射层标定 | pose 正向only + z 三档 + 降 firmware/radar-pose 权 + 删 Δz | belief/likelihood.go |
| §2 | **P3** realness R_i | 室内速度天花板 + 逐步跳跃近确定 ghost + 冻结方差 + 记忆 L_R | belief_adapter.go, track.go |
| §3 | **P4** dwell HSMM | 散落硬阈 → $-\ln S_{vol}(d\mid zone)$ ramp | fall_rules_param.go, belief/ |
| §4 | **P5** bed O_b 统一 | radar pose_lying 降权 + bed_bayesian 并入 DBN 模板 | bed_bayesian_scorer.go |
| §5 | **P6** room/N_r | S_room 9 态 + N_r + multi-resident 联合(P_id) | belief/, suite_census.go |
| §6 | **P7** decision τ\* | 30/70 + RiskLevel → 单一 $\tau^\*$ 代价比旋钮 | fall_verify.go, risk_evaluator.go |
| §7 | **P8** health 节点 | WeakBio/HR/RR/Apnea 正交健康事件输出 | aggregator, fall_verify.go |
| §8 | **P9** oracle 验收 | 7-case fixture + margin + go/no-go | doc/cases/, belief_replay_test.go |
| §9 | **P0** cell 接口契约 | 唯一耦合边 read-only 定义 + still-box 单源 | cell.go, belief_adapter.go |
| §10 | 里程碑/灰度门 | shadow→canary→cutover 排序 + 风险登记 | — |

---

## §1 P2 — 发射层标定 L(o|s)(keystone)

**目标**:把 `belief/likelihood.go` 的发射似然全部对齐 signal_map §2 的"真实方向性"原则:
**pose/z 对 fall 只正向**(R5),删伪正向,降不可信源权。全程 shadow,只改 likelihood 标定,
不动判决路径。

**现状审计(belief/likelihood.go,已逐行核)**:
| 现行为 | 行 | 与新标定的冲突 |
|---|---|---|
| `ObsKinematics` z↓→`SFallen:1+7f` | likelihood.go:19-27 | **违 R5/§10#3**:z 骤降当 fall 强正向证据;z=0 是噪声,连正向都不算 → **删** |
| `ObsFirmwareFall` →`SFallen:10` | likelihood.go:52-55 | **§10#2**:pose=5 不可信却给 ×10 极权 → 按混淆矩阵降权 |
| `poseLikelihood` stand→`SFallen:0.4` / walk→`0.3` / sit→(无 fall 项但 walk 压 0.8) | likelihood.go:137,145,147 | **违 R5**:pose=stand/walk **压** fall = pose 负向;firmware 误把"摔后"标成 walk 就会抹真摔 → 压 fall 项一律提到中性 1.0 |
| pose=fallen OpenFloor→`SFallen:10` / InBed→`1.5` | likelihood.go:149-156 | 正向,保留(lying/fallen 是允许的正向) |
| z 三档无独立发射(仅折进 Geom) | — | §2 新增 z>80→stand / 30–80→sit / <30→LR=1,且 z **不进 fall** |

### P2.1 — 删 kinematics Δz fall 正向(§10#3a)
- **动**:`belief/likelihood.go:19-27` `ObsKinematics` case。
- **改法**:整 case 删除(或 return `lk(nil)`);上游 `belief/observation.go` 停产 `ObsKinematics`(z↓ 信号)。
- **标定来源**:A 类 round-z3/round-posonly —— z 骤降=环境噪声,非 fall 证据。
- **shadow 字段**:对账日志记 `p2_1_kinematics_dropped=true`,对比删前/删后 Fallen 后验差。
- **oracle**:cabb-0603 冻结 FP —— 删 Δz 后该 case Fallen 后验应**下降**(不再被 z=0 抬)。
- **边界**:真摔(cabb-0605)靠 pose-lying 正向扛,不依赖 Δz → 后验不应塌。
- **DoD**:`grep ObsKinematics` 全栈 0 命中(producer+consumer 同删,R4 producer-first)。

### P2.2 — pose 对 fall 改正向 only(§10#3c / R5)
- **动**:`poseLikelihood` 各 case 的 **fall 压制项**:walk `SFallen:0.3`→`1.0`、stand `SFallen:0.4`→`1.0`、走动门口 SLeft 保留(那是 S_room 路由非 fall 压制)。
- **保留正向**:lying/fallen/suspected-fall/sit-ground 的 `SFallen>1` 不动。
- **posture vs fall 双表分离**:pose 对 **posture 节点**仍可做混淆矩阵区分(sit 特异/stand 模糊),但对 **fall 读出**只正向 —— 用两条 LR 通道,别让 posture 判别污染 fall 压制。
- **标定来源**:A 类 round-posonly;混淆矩阵 70/20。
- **oracle**:构造"firmware 误标 walking 的静止真摔" —— 改后 walk 不再压 Fallen,真摔后验不被抹。
- **DoD**:`poseLikelihood` 内 `SFallen` 取值 ∀ <1.0 的条目清零为 1.0(grep 自查)。

### P2.3 — z 三档发射(posture,§2 新增)
- **动**:`belief/observation.go` 新增 `ObsZBand`(或扩 ObsPose 带 z);`likelihood.go` 加 case。
- **标定**:`z>80→+stand`(`SStandWalk:↑`)、`30–80→+sit`(`SSit:↑`)、`<30→lk(nil)`(假低,LR=1)。**z 不写任何 SFallen 项**(R5:不确认不否决)。
- **来源**:A 类 round-z3。
- **依赖**:z 档同时是 cell engine 解偏输入(§9 接口,但 DBN 这侧只读 posture 用)。
- **oracle**:cabb-0606(全程 pose=4/z=0)—— z<30 段 LR=1,不污染;验证 z 不参与 fall(§11.2 残差对仍只剩 Z_cell 杠杆,符合预期)。

### P2.4 — firmware-fall pose=5 降权(§10#2)
- **动**:`likelihood.go:52-55` `ObsFirmwareFall` 的 `SFallen:10`。
- **改法**:按 firmware fall 混淆矩阵(真机 fixture 估 TP/FP 率)重标 LR,从 ×10 降到与"可信度"匹配的档。**shadow 期先取保守档 `SFallen:≤2`(委员会#5)**—— pose=5 是 pose 派生(R5 域),即便有 firmware 升级限定仍是强正向,不宜占位过高;待 P9 用 firmware_fall 真机 TP/FP 率标定后再定终值。
- **来源**:§10#2;firmware_fall_qualification 真机统计。
- **风险**:firmware pose=5 仍是现网 Device_ALARM 直发路径(记忆[firmware_fall_qualification]),**本改只动 belief shadow 的 LR,不动 firmware 直发**(R1)。
- **oracle**:对比降权前后,firmware-fall FP fixture 的后验是否落到 τ\* 下而真 fall 仍过。

### P2.5 — enter/exit 正向 + 缺失由 door-distance 补(已入 signal_map 3da5dfe)
- **现状**:`likelihood.go:58-65` 已是正向(EnterRoom→StandWalk×2 / ExitRoom→SLeft×8)。
- **本 task**:**编码"absence≠状态"** —— event 不在时**不**喂反向证据;信号丢失时由 `ObsReachableExit`(door-distance,likelihood.go:116-125 已存)补缺失退场。核对二者不重复计 SLeft。
- **来源**:signal_map §2 enter/exit 行 + door-distance 行。
- **oracle**:cd2b —— 丢真人时无 ExitRoom event,door-distance 远(未近门)→ 不误判 Left → 真人 Lost 浮出(配合 P3)。

### P2.6 — 发射标定集中化(R7)
- **动**:把散落在 `likelihood.go` 的 LR 数值抽到一张**标定常量表**(单文件/单 struct),附每条 `来源` 注脚(A 类 round-x / fixture)。
- **目的**:P9 oracle 调灵敏度时单点改,不猎散落字面量;委员会可审"每个数从哪来"。
- **DoD**:likelihood.go 内裸数值字面量 → 命名常量;对照表与 signal_map §2 行一一对应。

**P2 验收闸**:shadow 重放 §8 全 fixture,逐 case 记 fall 后验**改前/改后** delta;
要求(a) 真摔 case 后验不降破 τ\*;(b) 冻结/误标 FP case 后验下降。delta 报告进 P9。

---

## §2 P3 — realness R_i(运动学 + 方差 + 记忆)

**目标**:R_i 是 cd2b 类"冻结 ghost 压掉真人 lost-fall"的正解(§11.3)。把 realness 判据从
室外口径收紧到**室内-老人**,补**冻结方差**与**记忆衰减**两段,全部走 reliable 量(XY/速度/方差),
**不碰 pose/z**(R5)。

**现状审计**:
| 现行为 | 行 | 与新标定的冲突 |
|---|---|---|
| `isGhostJump` 用 `ImpossibleSpeedCm=200` | belief_adapter.go:169-175 | **§10#11**:200 是室外口径;cd2b 150cm/s 跳变在 200 闸下漏判 → 降到 ~110–130 |
| per-device speed EWMA cap(已存,P2.1)`1.5×EWMA∈[30,150]` | belief_adapter.go:184-214 | 复用为 R_i 个性化天花板;ceil 150 偏高,室内复核 |
| ghost verdict:`GhostPenalty≥80`→Ghost;Score 50/20 | track.go:177-179,197 | 逐帧/累积混合,**§10#9 需明确记忆衰减率**(摔倒瞬间 v≈0 靠摔前 L_R) |
| still-box `BoxRangeWithinMs≤50cm` 当静止 | track.go:350; §10#12 | **把真静止与冻结伪迹混了**:真人 ±10–20cm 抖,ghost 零方差 → 需方差判据 |
| `MaxKalmanResidual` 峰值残差 | track.go:163-167 | 已有,复用为跳变→ghost 的 reliable 量 |
| 速度类信号曾混标(委员会 §2 自纠 8cdc4ad) | signal_map §2:80/82/83 | **三个不同失效探测器,非等价**:空间跳跃(逐步 raw teleport,isGhostJump)/ Kalman 残差(逐步 model-relative,track.go:167)/ 隐含速度(全程平均 reachability,track.go:168 MaxImpliedSpeedFromBirth,单跳会被抹平)→ P3.1 须三者分立 |

### P3.1 — 室内-老人速度天花板 + 三探测器分立(§10#11;委员会 §2 自纠对齐)
- **动**:`fall_rules_param.go` `ImpossibleSpeedCm` 200→**~120**(待 P9 用 cd2b/真机走速分布定);`SuspectSpeedCm` 100 复核;`isGhostJump`(belief_adapter.go:169)随之收紧。
- **三个失效探测器分立喂 R_i(signal_map §2:80/82/83,委员会自纠不可混标)**:
  1. **空间跳跃**(逐步 raw,isGhostJump,belief_adapter.go:169):单帧 Δ/dt > 室内天花板 → `P(瞬移|real)≈0`,**近确定 ghost**(直接强 ghost LR,不再"需 enter 反证"软档);
  2. **Kalman 残差**(逐步 model-relative,track.go:167 MaxKalmanResidual):观测偏离匀速预测 → raw 速度漏的方向/加速度异常;匀速直线快走残差低(**不误判正常快走**);
  3. **隐含速度**(全程平均 reachability,track.go:168 MaxImpliedSpeedFromBirth):dist(now,birth)/age 超人极限 → firmware track_id 拼接两反射(birth-incoherence);**平均量,单跳被抹平**,与 1 互补。
- **关键**:三者仅在干净 teleport 重叠,不等价 —— 各喂独立 LR,别再捆一行(委员会 8cdc4ad 已在 signal_map §2 拆开)。
- **标定来源**:A 类 round-室内;cd2b 150cm/s fixture;per-device EWMA 实测走速上沿。
- **shadow 字段**:`p3_1_jump_cmps`、`p3_1_kalman_resid`、`p3_1_implied_speed`、`p3_1_verdict_delta`。
- **oracle**:cd2b —— 150cm/s 跳变(探测器1)在新闸下判 ghost → 真人 Lost 浮出 → lost-fall 应报;隐含速度(探测器3)兜 birth-incoherence 拼接。
- **边界**:真人快走/小跑(门口)匀速 → 探测器2 残差低不误判;天花板取走速分布上沿 + per-device cap 个性化兜底。

### P3.2 — 冻结伪迹**复合签名**(§10#12,新增;**委员会#2 阻塞:非单纯方差**)
- **真机反例(委员会#2)**:bedroom201-cd2b 那个冻结椅子 ghost 位置在 `(-170,330)/(-150,340)/(-120,300)` 间晃 **~50cm,不是零方差** —— 静止反射体带多径抖动。**单方差阈既漏(ghost 有抖)又误(真人微抖)**。
- **动**:`track.go` 新增 per-track 位置散布 + 出生/约束/距门统计;belief_adapter 加 `ObsFrozenArtifact` 发射。
- **门控签名(委员会第3轮:近必要前置 + 佐证,取代裸 3/4)**:
  - **近必要前置(A,二者居一)**:① **不可能跳变出生**(birth 跳速超室内天花板/瞬移,track-swap)**∨** ② **cell=AreaDeny**(常驻反射已被 cell static_reflector→AreaDeny 学到,§9 只读)。
  - **佐证(B,≥2 命中)**:③ **pose/z 锁死**(`pose=4 ∧ z=0` 连续 N tick)④ **位置钉死小区**(~50cm 散布且不漂移/无偶发大位移)⑤ **距门远**(非门区)。
  - **判 ghost = A ∧ (B≥2)**。
- **为何门控而非裸 3/4(委员会第3轮:补漏报反引真人久站 FP)**:第2轮放宽到"无跳变靠 ③④⑤=3/4"会误判 **真人远角久站**(看窗外):恰好 ③pose/z锁死 + ④钉死小区 + ⑤距门远 = 3/4,缺的也正是跳变出生 → 被当 ghost(undercount)。**根因**:③ pose/z 锁死**判别力弱** —— firmware 给真人站立也是 `pose=4` 恒定数分钟,非 ghost 专属。
  - 故把"跳变出生缺失"的唯一合法补位收紧为 **cell=AreaDeny**(常驻反射本就走 static_reflector→AreaDeny 兜底);**"无跳变 ∧ 无 AreaDeny"路径直接不判 ghost** → 真人久站(无跳变出生、cell 非 Deny)落此路 → 保 real。
  - 经典打地鼠规避:补 #2 常驻反射漏报**不**靠放宽佐证,而靠 cell 先验(AreaDeny)这条**正交强证据**。
- **关键**:**不是** zero-variance;真人久站 ③④⑤ 全中也**因缺 A 而不判**(③ 判别力弱不能独立定 ghost)。
- **标定来源**:A 类 round-cd2b + cd2b 实测坐标散布;真人远角久站反例。
- **shadow 字段**:`p3_2_jump_birth`、`p3_2_zcell_deny`、`p3_2_pose_z_locked_ticks`、`p3_2_pos_spread_cm`、`p3_2_door_dist_cm`、`p3_2_gateA(bool)`、`p3_2_corrobB(命中数)`、`p3_2_frozen=bool`。
- **oracle**:
  - cd2b 椅子 ghost(A①跳变出生 + B:③④⑤)→ A∧B≥2 判 ghost;
  - **常驻反射**(无跳变,但 A②cell=AreaDeny + B:③④)→ 判 ghost(cell 先验补位);
  - **真人远角久站**(无跳变 + cell 非 Deny,B③④⑤=3 全中)→ **A 不成立 → 不判,保 real**(关键反例);
  - cabb-0605 真摔躺地(无跳变 + 非 Deny + B≤1)→ 保 real。
- **DoD**:`ObsFrozenArtifact` = `A ∧ (B≥2)`,A=(跳变出生 ∨ cell=AreaDeny);**无单方差阈分支、无裸 3/4 分支**;fixture **必含"真人远角久站"反例并验 ≤判定阈(不判 ghost)**;与 cell static_reflector→AreaDeny 对账覆盖常驻反射。

### P3.3 — realness 记忆 filter L_R(§10#9)
- **动**:把 ghost verdict 的逐帧/累积混合(track.go GhostPenalty)显式化为 **log-odds 带遗忘递归**:$L_R^t=\gamma L_R^{t-1}+\sum\ln\mathrm{LR}_k$(proposal 时间反馈①)。
- **机理**:摔倒瞬间 v≈0 无当下运动学证据,realness 必须靠**摔前走路段累积的 L_R** 带进静止段。
- **待定参数**:$\gamma$(记忆衰减率)—— 既是记忆深度也是 Boyen-Koller mixing 率(proposal §4.1)。用 fixture 标:走路段建立的 L_R 要能"存活"到倒地后判定窗。
- **shadow 字段**:`p3_3_LR_walk`(摔前)、`p3_3_LR_atfall`(摔时)、`p3_3_gamma`。
- **oracle**:cabb-0605(摔前有走动)L_R 带进倒地段,realness 维持 real → 不被当 ghost 压掉。
- **依赖**:与 P2.6 标定集中化共用 log-odds 模板(bed_bayesian leak 0.55 是同款,§9 复用)。

### P3.4 — recapture 软恢复(非硬 cancel,§11.3 / signal_map §3)
- **动**:`track_manager.go:1830` birth 重生取消 pending 的逻辑。
- **改法**:recapture(人返回)**不能硬 cancel** lost-fall —— 返回可能是**跌后自救**;真摔自救应留**低 severity** 事件(对齐记忆[silent_leftbed_fall_recovery_window_gap])。
- **shadow 字段**:`p3_4_recapture_ms`、`p3_4_would_cancel`、`p3_4_self_rescue_candidate`。
- **oracle**:cd2b recapture(5.85min 返回)—— 不抹真摔,留低 severity。
- **R1 注意**:shadow 期只 log "若发会发什么 severity",不真发。

**P3 验收闸**:cd2b fixture 在 P3 全开后:冻结 ghost 被判 ghost(P3.1+P3.2)→ 真人 Lost 浮出 → lost-fall 后验过 τ\*;且 cabb-0605 真静止真摔**不**被 P3.2 误判 ghost。两者同时成立才算过(§11.3 下限)。

---

## §3 P4 — dwell HSMM 生存函数 S_vol(t|zone)

**目标**:把 signal_map §5 散落的**硬阈值驻留**收敛成**显式半马尔可夫驻留** $d$ +
平滑生存函数 ramp:$\ln\mathrm{LR}_{fall}(d)=-\ln S_{vol}(d\mid zone)$。硬阈悬崖 → 软斜坡,
zone 选档只读 cell engine(R3)。

**现状清单(signal_map §5,全硬阈)**:
| 上下文 | 现生存尾 | 来源 |
|---|---|---|
| stand 开阔地 | 8min | aggregator:302 StandingMin cap 8 |
| toilet/shower | 10(风险)/12min | bathroom_fall.go:46 |
| bed/sofa RestZone | ∞ | cell.go:326 |
| deny zone | 5min | DenyZoneSec |
| leftBed 床边 | 15min(夜)/窗3min | bedroom_fall.go:41 |
| lost-fall wait | bed60/walk5/deny5 min | fall_rules_param.go:157 |
| moving precondition | 60s | MovingPreconditionMs:165 |
| 缺席:Stay/LeftBed/NightAbsence | 45/30 · 30 · 30min | zonealarm rules.go:111/122/133 |

### P4.1 — 生存函数 ramp 取代硬阈(核心)
- **动**:新增 `belief/survival.go`(或 likelihood 内)`fallLRFromDwell(d, zone) = -ln S_vol(d|zone)`。
- **改法**:每个 zone 一条生存尾参数(尺度 + 形状),`d` 由 HSMM 驻留状态累积。**输出连续 LR**,不再是"≥阈值就报"的二元悬崖。
- **标定来源**:A 类"老人站立静止 >8min 几乎不可能"= 生存尾在 8min 处快速衰减;各 zone 尾形用 §8 fixture 标。
- **shadow 字段**:`p4_1_dwell_sec`、`p4_1_zone`、`p4_1_fall_LR`。
- **oracle**:开阔地真人久站 FP(cabb 类)—— ramp 下 8min 前 LR 温和,配合 Z_cell tolerance 压在 τ\* 下;真摔躺地 LR 随 d 快速累积过 τ\*。
- **依赖**:替代 fall_unified 的 still-fall 硬阈表(silent/lost/moving 三类的"等阈值")。
- **⚠️ 样本不足(委员会#3)**:7-case fixture **标不出生存函数尾形**,尤其 bedside benign-外出尾(cd2b 两例都 ~5.5–6min 返回,**n=2**)。8min 站立尾、bedside 尾**先用粗档**(尺度取 A 类先验 + 现有硬阈点,形状用单参指数/Weibull 占位),**标注"待样本收紧"**;P9 报告(§8 P9.2/P9.4)须把**样本量列为 margin 置信的限制项** —— 尾形不确定 → margin 给区间不给点值。线上 canary 期持续收集 dwell 样本回标尾形。

### P4.2 — zone 选档(cell engine 只读边,§9 前置)
- **动**:`fallLRFromDwell` 的 `zone` 入参 = cell engine 输出 `AreaType`(read-only)。
- **映射**:rest/bed→尾∞(不报)、toilet→15min 尾、stand 开阔→8min 尾、deny→5min 尾。
- **R3**:DBN 不写 cell,只读 `AreaType+Conf`;Conf 低时尾形向"中性 zone"blend(对齐 likelihood.go:14-18 已有的 GeomConf blend 思路)。
- **oracle**:cabb-0606 vs cabb-0603 残差对(§11.2)—— 唯一分离杠杆是 Z_cell zone 容忍,选档正确则 margin≈1.5nat 可分。

### P4.3 — risk-time 夜间尾形切换
- **动**:`math_util.go:16` risk-time 22:00–06:30 → S_vol 夜间用更短尾(夜间久静更可疑)。
- **改法**:生存尾尺度参数按 risk-time 二档(昼/夜),折进 P4.1 参数表,不另开闸。
- **来源**:signal_map §4 risk-time;zonealarm 昼夜阈(45/30 等)。

### P4.4 — per-cell 自适应放宽(读 cell tolerance,§9)
- **现状**:`cell.go:400` toleranceFactor ×[1,2.0] by FakeAlarm/ToleratedStill —— **cell engine 内部学习**。
- **本 task**:DBN 只**读** tolerance factor,乘到 S_vol 尾尺度(久站被容忍的 cell 尾拉长)。**不在 DBN 学**(R2/R3)。
- **oracle**:cabb 那个 cell 若被 ToleratedStill 学成"可久站",尾拉长 → 久站不报;未学则正常 8min 尾。

### P4.5 — 缺席驻留同族归并(§10#8 scope 决议)
- **现状**:Stay(独处45/30)/LeftBed(离床30)/NightAbsence(离室30,21–07)是**同族 S_vol 不同 anchor**(独处/离床/离室)。
- **决议项(待委员会)**:是否纳入同一 DBN dwell 框架,还是保留 zonealarm 独立。**建议**:纳入 S_vol 框架统一参数化(anchor 当 zone 维度),但**发射仍走 zonealarm**(producer 不变,R4)—— DBN 这侧只做 shadow 对账,验证同族尾形一致性。
- **shadow 字段**:`p4_5_absence_anchor`、`p4_5_dwell_sec`、`p4_5_LR`。

### P4.6 — moving precondition 并入 A_T(与 P3 接)
- **动**:`MovingPreconditionMs=60s`(still≥60s→不进 lost)= signal_map §3 的 static→lost vs moving→lost 分叉判据。
- **改法**:作为 HSMM 进入 lost 态的**前置驻留**,与 P3.1/P3.2 的 realness 联合(消失前 prev-state 由驻留时长定)。
- **oracle**:走着突然消失(moving→lost,moving-fall)vs 静止后消失(static→lost)分得开。

**P4 验收闸**:全 fixture 重放,验证(a) ramp 输出连续无悬崖;(b) §11.2 残差对在 zone 选档正确时 margin≥~1.5nat;(c) 缺席三类尾形参数自洽。硬阈→ramp 的等价性:旧阈值点处新 LR≈ln(odds at τ\*)。

---

## §4 P5 — bed O_b 统一(降 radar-pose 权 + 并入 DBN 模板)

**目标**:`bed_bayesian_scorer` 已是 log-odds HMM(O_b 节点雏形,§9 复用件)。本阶段
(a) 修 radar pose_lying 过高权(§10#4);(b) 把它形式化为 DBN 通用 log-odds 模板,
供 R_i/O_b/缺席同款复用;(c) sleepad 接触为主、radar 为辅的模态权重对齐。

**现状审计(bed_bayesian_scorer.go)**:
| 现行为 | 行 | 与新标定的冲突 |
|---|---|---|
| `lrRadarPoseLying=ln4.25` | bed_bayesian:35 | **§10#4**:pose 不可信,radar 躺姿给 O_b 过高权 → 降到 ≈0,bed 靠 sleepad 接触扛 |
| `lrRadarVital=ln1.75` / `lrSleepadVitalOnly=ln2` | bed_bayesian:36,39 | radar vital 弱,复核;sleepad 为主 |
| LR sleepad facility ±ln19 / home ±ln9 | bed_bayesian:31,32 | 接触模态强,保留 |
| leak 0.55 / γ 三相(0/0.5/0)/ maintain timeout 120s | bed_bayesian:60-76 | log-odds 模板核心,抽出复用(R_i/缺席同款) |
| 夜/日 prior +1.39/−0.85 | bed_bayesian:63-66 | π,保留;与 §4 先验表对齐 |
| 多源 LeftBed sticky(60s 双 LeftBed) | bed_bayesian:142 | 记忆[bed_stale_leftbed_vetoes_radar_inbed]相关,复核衰减 |

### P5.1 — radar pose_lying 降权(§10#4)
- **动**:`bed_bayesian_scorer.go:35` `lrRadarPoseLying` ln4.25 → ≈0(LR≈1)。
- **理由**:pose 不可信(R5);床占用靠 **sleepad 接触**(独立模态)定,radar 躺姿不该独立抬 O_b。
- **shadow 字段**:`p5_1_Ob_with_radarpose`、`p5_1_Ob_without`(对比 O_b 后验差)。
- **oracle**:无 sleepad 的纯 radar 房,radar 误报 lying 不应假抬 O_b → 床态由 radar vital/event 弱证据 + 夜 prior 定,不被 pose 拉偏。
- **边界**:有 sleepad 的床不受影响(sleepad ±ln19 主导)。

### P5.2 — bed leak 模型遗忘率复核(记忆衔接)
- **动**:`leakFactor=0.55`、γ 三相(0–60s:1.0 / 60–120s:0.5 / ≥120s:0)。
- **复核点**:记忆[bed_stale_leftbed_vetoes_radar_inbed]——陈旧 sleepad LeftBed 永久否决 fresh radar InBed 的 bug;本 task 确认 leak/γ 让陈旧 LeftBed 指数遗忘(L*=0.55/min leak 回中性),|L|<0.5→Standby8。
- **shadow 字段**:`p5_2_L`、`p5_2_leak_applied`、`p5_2_standby`。
- **oracle**:sim_decay 真实数据(19:27 进 Standby8 → 19:55:58 radar InBed → P0.80 InBed)—— 验证 leak 让陈旧 LeftBed 不再永久压。
- **R4**:此项是 bed scorer producer 侧,已部分落地待 port shadow;DBN O_b 只读其 P 输出。

### P5.3 — log-odds 模板抽象(§10#5 部分,§9 复用)
- **动**:把 bed_bayesian 的 `(LR 表 / γ tempering / leak / covers 权重 / cap±5)` 抽成**通用 log-odds filter 模板**。
- **复用为**:R_i 记忆 filter(P3.3 同款 $\gamma L_{t-1}+\sum\ln LR$)、缺席驻留(P4.5)、O_b 自身。
- **理由**:signal_map §9 列 bed_bayesian = "通用 log-odds filter 模板(LR/γ/leak/covers)";避免四套滤波各写一遍(§10#5)。
- **DoD**:R_i/O_b/缺席三处共用同一模板函数,leak/cap/γ 入参化;`go test` 覆盖模板单测。

### P5.4 — O_b → S_room 投影对齐(likelihood.go:34)
- **动**:`ObsBedOccupied` 投影 `S_BedLying(1+5p)/S_Fallen(1−0.7p)` 用 O_b 的 P(InBed)=p。
- **核对**:O_b(P5)输出的 p 与 likelihood.go:34 消费口径一致(嵌套 bed 贝叶斯 → S_room 发射),避免双口径 drift。
- **oracle**:床占用高时 S_room 抬 BedLying 压 Fallen(床上不误报倒地);O_b 低时不抑制地板 fall。
- **❓ 边界:O_b 抑制 fall 必须 fresh∧高(委员会#4)**:`S_Fallen(1−0.7p)` 压制**仅当 O_b 新鲜且高** —— **陈旧 O_b 不得压 fall**。否则"leftBed→床边静止晕倒"窗内,残留的旧 O_b 会压掉 bedside-fall(漏报)。要求:
  1. 消费 `ObsBedOccupied` 时带新鲜度闸(对齐 §8 bed-presence 10min 窗 / bed_bayesian vital window 35s);stale → p 视为低,不压 Fallen。
  2. **leftBed 后 O_b 必须及时落**(对齐 P5.2 leak/γ:LeftBed event → L 快速回中性,|L|<0.5→Standby)—— 验证 leftBed 到床边静止的过渡窗内 O_b 已低。
- **oracle 补**:bedroom leftBed→床边 15min 静止真摔 —— 过渡窗内 O_b 已落 → S_Fallen 不被旧床占用压 → bedside-fall 后验过 τ\*。

**P5 验收闸**:(a) radar pose_lying 降权后纯 radar 房 O_b 不被 pose 拉偏;(b) leak 模型让陈旧 LeftBed 遗忘(sim_decay 通过);(c) log-odds 模板被 R_i/O_b/缺席三处复用,单测绿。

---

## §5 P6 — room S_room / N_r / multi-resident(P_id)

**目标**:S_room 9 态 HMM(belief/,已部署 forward)是房级读出层。本阶段
(a) 校 likelihood 矩阵中 firmware/exit 的过权(§10#2);(b) N_r 人数从硬 max 软化;
(c) 接 P_id 身份层(suite_census 已在跑,§1/§3),为 multi-resident 联合滤波铺路(§10#7/#10)。

**现状审计**:
| 现行为 | 行 | 与新标定的冲突 |
|---|---|---|
| `ObsFirmwareFall→SFallen:10` | likelihood.go:52 | §10#2 与 P2.4 同源,room 侧同降 |
| `ObsEnterExit ExitRoom→SLeft:8` | likelihood.go:58-63 | 已改正向(3da5dfe);核 absence≠还在 |
| `N_r = max(radar_np, bed_np)` 硬 | stream_publisher.go:426 | 硬 max,multi-resident 下需软化 |
| `P_id` suite_census(resident/visitor anchor) | suite_census.go:97 | 已在跑(§1 节点),未进 S_room 联合 |
| `S_room` 9 态 forward | belief/state.go | 已部署软层,复用 |

### P6.1 — S_room likelihood 矩阵再标定(承 P2.4)
- **动**:`belief/likelihood.go` room 侧 `ObsFirmwareFall`(:52)、`ObsNoDetect`(:100,Fallen×1.6)、`ObsReachableExit`(:116)核对。
- **改法**:firmware ×10 随 P2.4 降;NoDetect/ReachableExit 已是软发射,验证与 P3(realness)、P4(dwell)不重复计 Fallen。
- **oracle**:firmware 漏判静止真摔(cabb-0606)—— 不靠 firmware-fall(它没报),靠 dwell+pose-lying 累积;firmware FP 不靠 ×10 误抬。

#### P6.1a — ObsNoDetect→Fallen 必须门控(**阻塞项,委员会#1**)
- **根因**:`ObsNoDetect→SFallen:1.6`(likelihood.go:100,113)是"消失→抬 fall",= **用 absence 当正向证据**,正是历史 dropout-FP 来源。P3 还没判出 ghost,no-detect 已先把 Fallen 抬起来。
- **门控(与 P3/P2.5 联锁)**:ObsNoDetect 抬 Fallen **仅当**
  1. `R_i=real`(P3:ghost 消失 → **不抬**,ghost 闪灭是伪迹不是倒地);**且**
  2. 非门区消失(`ObsReachableExit`/door-distance **未↓** → 不是从门口走出;门区消失 → **不抬**,走 SLeft)。
- **实现(软形,委员会第2轮#1)**:ObsNoDetect 的 Fallen 因子从"固定 1.6"改为**连续边缘化形** `1 + 0.6·P(R_i=real)·(1−P(door-exit))` —— **不用硬 𝟙**。R_i 判错时退化平滑(P→0/1 才趋同硬闸),与 §4.3 ghost 融合方程同构;避免"软网里塞硬阈"重新引入"realness 判错就全 0/全 1"的脆性(局部回退硬阈范式 = drift 风险)。ghost(P(real)→0)或门区(P(door-exit)→1)时因子→1.0(中性),退场由 SLeft/SEmpty 仲裁(沿用 likelihood.go:106-107 方向仲裁,但**前置 realness 边缘化权**)。
- **shadow 字段**:`p6_1a_nodetect_raised`、`p6_1a_Ri`、`p6_1a_door_exit`。
- **oracle**:cd2b —— 真人被冻结 ghost 顶替时,**那个 ghost 的消失**(若有)不抬 Fallen;**真人 track 的消失**(R_i=real,非门区)才抬 → 配合 P3 判出 ghost 后,真人 Lost 的 no-detect 正确抬 lost-fall。D5F7/D523 边缘门区丢轨 → door-distance↓ → 不抬(防 dropout-FP)。
- **DoD**:`grep ObsNoDetect` 确认 Fallen 因子受 R_i + door-distance 双闸;无任何"裸 absence 抬 Fallen"路径。

### P6.2 — N_r 软化(承 §8 bed-presence)
- **动**:`N_r` 从 `max(radar_np, bed_np)` 硬 → 软 count 后验(radar_np 不可信,bed occupancy 较可信)。
- **改法**:O_b(P5)+ radar_np + P_id anchor 联合估 N_r;radar_np=0 是 corroboration 非 substitution(记忆[number_people_zero_exitroom_fallback])。
- **shadow 字段**:`p6_2_radar_np`、`p6_2_bed_np`、`p6_2_Nr_posterior`。
- **oracle**:镜面 ghost 致 radar_np 虚高 → 不假抬 N_r(R_i 判 ghost 后该 track 不计数)。

### P6.3 — P_id 身份接入 S_room(§10#7 铺路)
- **动**:`suite_census` 的 resident/visitor anchor + AnchorRoomType → S_room 的人数/身份维度。
- **scope(待委员会)**:v1 仍单实体;本 task 只把 P_id 作为 **N_r 的证据之一**(单 resident anchor 确认"房里有人"),**不做** multi-resident 联合滤波(留 §10#7 P-later)。
- **oracle**:bathroom 锚翻转(§3 转移)后,fall dispatch 走 bathroom 尾(P4.2)—— 验证 P_id→zone→S_vol 链路。

### P6.4 — R_i–S_i 耦合决议(§10#10,纯设计)
- **决议项**:ghost 运动学胎记 ⟹ $R_i\perp S_i$ 不成立 → 倾向 `{R_i, S_i}` 同簇**联合滤波(4 态)**还是保持因子化 + 消息传递。
- **建议**:shadow 期先**因子化 + 期望值消息传递**(proposal L2:track→bed→room 逐级期望传递),实现简单;若 oracle 显示 cd2b 类必须联合(R_i 与 S_room Fallen 强耦合分不开)再升 4 态联合。**本 task 只出决策依据,不写码**。
- **oracle 判据**:cd2b 在因子化下能否分出"冻结 ghost + 真人 Lost";不能则需联合。

**P6 验收闸**:(a) firmware/exit 过权校正后 room FP 降;(b) N_r 软后验不被 ghost/镜面虚高;(c) P_id→zone→S_vol 链路在 bathroom 锚翻转 fixture 上通;(d) R_i–S_i 耦合决议有 oracle 依据。

---

## §6 P7 — decision τ\*(代价比单旋钮读出)

**目标**:把散落判决阈值(fall 30/70、ghost 50/20、bed P、RiskLevel 分级)收敛到
**一个代价比阈值** $\tau^\*=C_{FP}/(C_{FP}+C_{FN})$。后验过 τ\* → 告警;reason 路由由
各节点后验来源决定。**shadow 期只读出到 log,不接 alarm**(R1)。

**现状清单(signal_map §6)**:
| 决策 | 现阈 | 来源 |
|---|---|---|
| firmware-fall 验证 | 50→30(ghost)/70(real) 加性 scorer | fall_verify.go:54 |
| ghost verdict | 50(real)/20(ghost);Penalty≥80 | track.go:197 |
| bed P 阈 | InBed 0.70/0.75;LeftBed 0.50;Standby\|L\|<0.5 | bed_bayesian:49 |
| RiskLevel 分级 | Attention/Risk(浴室日8/15 夜5/8 alone30/45) | risk_evaluator.go:39 |
| 速度 ghost 二档 | hard200/soft100 | fall_rules_param.go:166 |
| human-bed fall 总豁免 | Human∧AreaBed∧Conf≥99 | fall_exempt.go:15 |

### P7.1 — fall 后验 → τ\* 判决
- **动**:DBN 输出 `P(Fallen)` 后验 → 与 τ\* 比;`fall_verify.go` 加性 scorer 形式化为标定 log-odds(signal_map §9)。
- **改法**:`baseline 50→30/70` 映射成 odds:`τ\* = C_FP/(C_FP+C_FN)`,30/70 对应两个代价比工作点(suspect/confirm)。
- **shadow 字段**:`p7_1_P_fallen`、`p7_1_tau`、`p7_1_decision`(vs gate-list 实际)。
- **oracle**:全 fixture —— 真摔 P_fallen>τ\*,FP<τ\*;margin 报告进 P9。

### P7.2 — τ\* 代价比参数化(单旋钮)
- **动**:把 30/70 与 RiskLevel 的多档阈值,统一成 `τ\*(context)` —— 不同 zone/risk-time 用不同 C_FP/C_FN(夜间/浴室 C_FN 高 → τ\* 低 → 更敏感)。
- **理由**:signal_map §6 注 "DBN 把散落阈值收敛到一个代价比旋钮"。
- **DoD**:risk_evaluator 的 Attention/Risk 分级表 → τ\* 二档函数;委员会可审"每个工作点的代价假设"。

### P7.3 — reason 路由(读出可解释性)
- **动**:fall 告警的 `reason` 由"哪个节点后验主导过 τ\*"决定(realness-lost / dwell-still / pose-lying / bathroom-still …),取代 fall_unified 的 6+ 散落 reason 函数。
- **映射**:silent(dwell 主导)/ lost(R_i+消失主导)/ moving(prev-moving+消失)三类 + human-bed 豁免。
- **oracle**:每 fixture 的 reason 与 fall_unified 现有 reason 对账一致(语义不丢)。

### P7.4 — human-bed 豁免接入 τ\*(承 §6 决策)
- **动**:`fall_exempt.go` Conf≥99 人工床 → fall 后验**前置短路**(不进 τ\* 判决)。
- **位置**:作为 τ\* 读出前的 veto gate(decision 层,非发射层 —— 区别于 P4.2 的 zone 软抑制)。
- **R7**:Conf≥99 切人工 layout vs radar 自学习 95,常量复用 fall_exempt。

### P7.5 — ghost/bed 判决并入(可选,scope 待定)
- **决议项**:ghost verdict(50/20)、bed P(0.70/0.75)是否也用 τ\* 统一,还是各保留(它们是节点内判决,非 fall 读出)。
- **建议**:保留节点内判决(O_b/R_i 各自 τ),fall 读出层只统一 fall 的 τ\*;避免过度统一伤可解释性。

**P7 验收闸**:(a) fall 判决 = 单一 τ\* 比较,30/70 等价复现;(b) reason 路由与现有语义对账无丢;(c) human-bed 豁免短路正确;(d) shadow 判决 vs gate-list 对账过关率达门(具体率 P9 定)。

---

## §7 P8 — health 正交节点(WeakBio/HR/RR/Apnea)

**目标**:§10#1 暴露的缺口 —— WeakBio/HR/RR/Apnea 现仅作 realness boost(≥80→Real),
但**心率/呼吸/呼吸暂停异常本身是健康事件**,与 fall 正交。补这一列输出。**可与 P2–P7 并行**
(不在 fall 关键路径)。

**现状审计**:
| 现行为 | 行 | 缺口 |
|---|---|---|
| WeakBio score=max(raw)+5·HR+5·RR+15·Apnea,cap100 | aggregator:467-486 | 只喂 R_i realness(≥80→Real),无独立健康输出 |
| ≥80→force Real | fall_verify.go:189 | 健康信号被当 fall 辅助,本身不告警 |
| WeakBio 30min 窗 | aggregator:446 | 窗口口径,复用 |

### P8.1 — health 节点定义(与 fall 正交)
- **动**:DBN 加 `H` 健康节点(或独立读出列),输入 = WeakBio/HR/RR/Apnea 事件。
- **输出**:健康事件分级(与 fall τ\* 正交的独立判决),**不**进 fall 决策路径(R1 + 记忆[feedback_no_dynamic_threshold_modulation])。
- **关键**:WeakBio 现喂 realness 的用法**保留**(≥80→Real 是 realness 证据);P8 是**新增**正交输出,不改 realness 用法。
- **shadow 字段**:`p8_1_health_score`、`p8_1_hr/rr/apnea_count`、`p8_1_health_level`。

### P8.2 — 输出去向(producer-first,R4)
- **决议项**:健康事件走 `iot:event:stream`(持久化 event_log)还是 alarm。
- **建议**:参考记忆[device_class_alarm…]与 zonealarm 模式 —— 健康异常是 actionable → 可走 alarm 流,但**非 medical-grade 诊断**(care-not-treatment 原则,ai_health Phase0)。具体归属待委员会 + 与 wisefido-ai-health 服务边界对齐。
- **R4**:producer(sensor 或 ai-health)先定,downstream/FE 再做。

### P8.3 — 与 ai-health 服务边界(避免重复)
- **核对**:记忆[ai_health_phase0] —— wisefido-ai-health 已有 realtime_liveness_alert(非 medical)+ daily ETL。本 P8 是 **sensor 实时层**的健康事件,与 ai-health 批处理层分工:sensor 出实时异常 event,ai-health 做趋势/cohort。
- **DoD**:明确 sensor-realtime vs ai-health-batch 的字段/职责切割,不双写。

**P8 验收闸**:(a) health 节点输出与 fall 完全正交(grep 确认无 health 信号进 fall verdict/severity/routing);(b) 输出去向与 ai-health 边界无重叠;(c) WeakBio→realness 旧用法不回归。

---

## §8 P9 — oracle 定量验收(go/no-go 数学闸)

**目标**:signal_map §11 的定量结论落成**可重放、可量化 margin 的 oracle 测试**。
这是整个 DBN 的 **go/no-go 闸**:margin 不够则止于 shadow,不进 canary(§11.4)。
依赖 P2–P7 全部 shadow 落地。

**现有基建**:`belief_replay_test.go`、`doc/cases/`(export_case.sh 导出)、记忆中的 fixture
(cabb-*、d5f7、cd2b、D523、MoM)。

### P9.1 — 7-case fixture 套件落定
- **动**:汇集/补齐 §11 的 case profile 到 `doc/cases/`,每 case 带真值标签(verified fall / FP / ghost):
  | case | 类型 | 期望判决 | 关键判别量 |
  |---|---|---|---|
  | cabb-0605 | 真摔(firmware pose=2 lying 52s) | fall>τ\* | pose-lying 正向 |
  | cabb-0606 | firmware 漏判静止真摔(全程 pose=4) | fall>τ\*(难) | 仅 Z_cell 杠杆(§11.2) |
  | cabb-0603 | 冻结 FP | fall<τ\* | Z_cell tolerance + 删 Δz |
  | cd2b | 跌床→冻结 ghost 压真人 | lost-fall>τ\* | R_i 方差+跳变(§11.3) |
  | D5F7 | 浴室真摔 | fall>τ\* | dwell+pose |
  | D523 | 边缘丢轨 | 不误报 lost | door-distance+d_fall |
  | MoM | (多人/移动场景) | 按真值 | N_r/P_id |
- **DoD**:每 case fixture 可被 `belief_replay_test.go` 加载重放。

### P9.2 — margin 量化(信息论)
- **动**:每 case 算 **fall 后验轨迹** + 与 τ\* 的 margin(nat):`margin = |ln(P/(1−P)) − ln(τ\*/(1−τ\*))|`。
- **报告**:per-case margin 表 + 改前(gate-list)/改后(DBN)对比。**margin 给区间不给点值**:尾形/标定样本不足(委员会#3,如 bedside 尾 n=2)→ margin 带置信区间,样本量作显式限制项列出。
- **判据**(§11.4):
  - 可分多数(cabb-0605 真摔 / 坐姿 FP / ghost / 门区 exit)→ margin>0 干净 → **DBN 赢**。
  - §11.2 残差对(cabb-0606 vs cabb-0603)→ 只剩 Z_cell 一杠杆:有 cell 学习 margin≈1.5nat 可分,无则重合。
  - cd2b → R_i 运动学能分(§11.3),下限在冻结 ghost vs 真静止方差判据(P3.2)。

### P9.3 — shadow vs gate-list 对账
- **动**:`belief_shadow.go` 旁路输出 vs 现网 gate-list alarm 逐 case/逐 tick 对账。
- **指标**:一致率、DBN 多报(潜在 FP)、DBN 漏报(潜在 FN)、DBN 修正(gate-list 错 DBN 对,如 cd2b 漏报)。
- **DoD**:对账报告 + 每条分歧归因(标定问题 / 真值问题 / DBN 缺陷)。

### P9.4 — go/no-go 判据 + 报告
- **产出**:`doc/belief_dbn_oracle_report.md` —— margin 表 + 对账 + 结论:
  - **go**(进 canary):可分多数 margin 干净 + cd2b 类修正成立 + 无新增 FP。
  - **no-go-but-shadow**(止于 shadow):§11.2 残差对治不了(只剩 Z_cell)→ 需 cell 解偏(§9 cell engine,§10#6)或加传感器(多雷达/bedside 垫),非滤波层能解。
- **诚实边界**(§11.4):firmware 漏判静止真摔 vs 冻结 FP 的下限要加硬件,DBN 不夸大。

### P9.5 — 良性残口验证(委员会第4轮记录)
- **残口**:全新装机、cell 尚未学到 AreaDeny(<3 episode)**且**无跳变出生的**常驻反射** → P3.2 门控 A 两臂皆不成立 → 暂不判 ghost。
- **为何良性**:纯反射体不会"先成为人再 lost",**不进 lost-fall pending** → 不触发漏报。
- **P9 须验**:用全新装机 fixture(cell 空 AreaDeny)确认此 gap **不致误**(既不误报、也不因漏判 ghost 而漏真摔);并记录 cell 学满 AreaDeny(≥3 episode)前的过渡期行为。
- **类别**:良性残口(非阻塞),仅作 oracle 覆盖项,不改 P3.2 设计。

**P9 验收闸(总闸)**:oracle report 出 go/no-go 结论 + 每个 P2–P7 改动的 margin 贡献归因 + P9.5 良性残口验证。**这是"DBN 值不值得继续做"的数学判据**,委员会据此决定是否进 canary。

---

## §9 P0 — cell engine ↔ DBN 接口契约(唯一耦合边)

**目标**:落定 signal_map §11.5 的**唯一耦合边**:cell engine 独立(自学习/自衰减/自带护栏),
DBN 只读其结果(R2/R3)。这是 P2.3/P4.2/P4.4 的**前置只读边**,须先冻结契约再施工那几节。

### P0.1 — 正向只读边(cell → DBN)
- **契约**(signal_map §11.5):
  | cell 输出(read-only) | DBN 怎么用 | 消费节 |
  |---|---|---|
  | `AreaType(cell)` | π(zone):rest/bed→fall 抑制;toilet/open→escalate | P4.2 |
  | `Confidence + Source(Human/Feedback/Learned)` | 先验权重(Human1.0>Feedback0.6>Learned0.4);human-bed Conf≥99=fall 总豁免 | P4.2/P7.4 |
  | `AreaType → S_vol(t\|zone)` 选档 | 驻留生存尾:stand8/toilet15/rest∞ | P4.2 |
- **DoD**:DBN 侧定义只读 accessor(无任何 cell 写路径);`grep` 确认 DBN 不调 cell 学习/promote 函数。

### P0.2 — 反向单源边(track still-box → cell)
- **契约**(signal_map §8 注):track 层每分钟算一次 still-box(50×50 + StillBoxRunStart),**传** `(cell_xy, is_static, run_duration)` 给 cell engine(`MarkDwell/MarkLongStill`)和 DBN(fall/M_i)**双消费**;cell engine **不重算**(§2.4 producer/maintainer,避免两套阈值 drift)。
- **DoD**:still-box 单一计算点;cell engine 与 DBN 消费同一份;`grep` 确认无第二处 still-box 计算。

### P0.3 — 边界纪律(防 §10#6 循环偏置回流)
- **铁律**:cell 内部学习(dwell→sit / z 档语义 / 长静→sit / 重复+衰减+真摔擦除护栏)**不在 DBN 图**,属独立 cell-engine spec。
- **解偏归属**:§10#6 循环偏置(Z_cell 学自 pose→继承 70/20 混淆)是 **cell engine 内部问题**,解在 cell 内(dwell 不用 pose / z 档 / Source 真值注入 / ai_fall_model z=high 学语义),DBN 不卷入学习回路。
- **DoD**:文档化 cell-engine spec 边界(单独 doc 链接),本计划只管 DBN 侧只读消费。

**P0 验收闸**:(a) 只读边 accessor 无写路径;(b) still-box 单源双消费无重算;(c) cell 学习回路与 DBN 完全解耦(grep 双向无越界调用)。**此节须在 P2.3/P4.2/P4.4 之前冻结。**

---

## §10 里程碑 / 排序 / 灰度门

### 10.1 施工顺序(承 §0.4 DAG)
```
阶段 0(前置):  P0 接口契约冻结 ─────────────┐
阶段 1(发射):  P2 ──────────────────────────┤
阶段 2(并行):  P3(realness) ‖ P4(dwell) ‖ P5(bed) ‖ P8(health)
阶段 3(房级):  P6 ───────────────────────────┤
阶段 4(读出):  P7(τ*) ───────────────────────┤
阶段 5(验收):  P9 oracle ── go/no-go 闸 ──────┘
```
- P8(health)与 fall 链正交,任意阶段并行。
- 每阶段内子任务(P<n>.<m>)各自 commit;跨阶段守 producer-first(R4)。

### 10.2 灰度门(shadow → canary → cutover)
| 门 | 进入条件 | 可秒关 |
|---|---|---|
| **shadow** | 代码进 belief_shadow,只 log 不 fire | `beliefShadowEnabled` |
| **canary** | P9 oracle go + 单房/单 unit 小流量对账过关率达门 | per-device 灰度开关 |
| **cutover** | canary 期零新增 FP + 修正 cd2b 类漏报 + 委员会签字 | 整套停/启(记忆[v2_cutover_rules]) |
- **绝不**:shadow 未过关直接接 alarm(R0/R1)。

### 10.3 风险登记
| 风险 | 缓解 |
|---|---|
| §11.2 残差对治不了(只剩 Z_cell) | P9 诚实判 no-go-but-shadow;转 cell 解偏 / 加硬件,不夸大滤波层 |
| 标定数值拍脑袋 | 每条 LR/τ\* 带来源(A 类 round-x / fixture);P9 oracle 反验 |
| 改 likelihood 影响现网 | 全程 shadow,firmware 直发路径不动(R1) |
| cell↔DBN 边界泄漏(循环偏置回流) | P0 契约冻结 + grep 双向越界检查 |
| multi-resident 未覆盖 | v1 单实体明示;P_id 已铺路,联合滤波留 P-later |

### 10.4 交付物清单
- `doc/belief_dbn_impl_plan.md`(本文,施工计划)
- `doc/belief_dbn_oracle_report.md`(P9 产出,go/no-go)
- cell-engine spec 边界 doc(P0.3)
- 各 P-task 的 shadow 代码 + 单测(committee 签字后另起 commit)

---

## §11 委员会反馈消化记录(倒序)

### 第 4 轮 — `14ede79`(b20f412..3ade655 审,裁决:第3轮消化合格,P3.2 收敛稳定)
委员会核验第3轮 ❓ 闭合 ✅(P3.2 门控形比建议更干净);**本轮无新阻塞/refine**(明示"不为挑而挑")。P3.2 四轮收敛轨迹:zero-variance→4-way AND→加权3/4→门控 A∧(B≥2),FP/漏报两侧都封。
| # | 类型 | 委员会意见 | 消化 | commit |
|---|---|---|---|---|
| — | 良性残口(记入 P9) | 全新装机 cell 未学 AreaDeny 且无跳变出生的常驻反射 P3.2 暂不判;但纯反射不进 lost-fall pending 故良性 | 新增 P9.5 良性残口验证项(全新装机 fixture 验不致误 + 记录过渡期);不改 P3.2 设计 | `<本commit>` |
| — | 裁决 | 整体 P0–P9 经四轮审无悬置阻塞;**建议进入 P0 接口契约冻结 + P2 发射标定代码施工**(仍 shadow-first,逐 P-task 单独 commit + 委员会签字) | 计划侧已就绪;代码施工待用户 go-ahead(铁律 R0 shadow-first + 逐 P-task 签字) | — |

### 第 3 轮 — `7340b2a`(98a33a7..b20f412 审,裁决:第2轮消化合格,可继续 P0→P2)
委员会核验第2轮 2 refine + §2 对齐全 ✅;给 1 个新 ❓(refine#2 副作用 = 漏报修复引入 FP)。
| # | 类型 | 委员会意见 | 消化 | commit |
|---|---|---|---|---|
| 1 | ❓非阻塞(P3.2 编码前须定) | 裸"3/4 无跳变出生"误判**真人远角久站**(pose=4 恒定非 ghost 专属,③ 判别力弱)= 新 FP | P3.2 改**门控形** `A∧(B≥2)`:A=(跳变出生 ∨ cell=AreaDeny)近必要,B=③pose/z锁死④钉死小区⑤距门远;"无跳变∧无 AreaDeny"不判→真人久站保 real;常驻反射靠 cell=AreaDeny 正交补位。DoD 必含"真人久站"反例验不判 | `828ca7e` |

### 第 2 轮 — `8cdc4ad`(f0bb43f..98a33a7 审,裁决:首轮消化合格,可继续 P0→P2)
委员会核验首轮 5 项全 ✅(看实际改法非声明);给 2 个 refine(非阻塞)+ signal_map §2 自纠。
| # | 类型 | 委员会意见 | 消化 | commit |
|---|---|---|---|---|
| 1 | ⚠️refine | P6.1a 硬 𝟙 在软网里塞硬阈 = drift(realness 判错全 0/1 脆性) | 改连续边缘化 `1+0.6·P(real)·(1−P(door-exit))`,与 §4.3 ghost 融合同构 | `ec89464` |
| 2 | ❓refine | P3.2 全 AND 漏判**常驻**反射体(无跳变出生) | 改加权任 ≥3/4 即疑;常驻反射另由 cell static_reflector→AreaDeny 兜底(§9 只读边) | `050352e` |
| — | 自纠 | signal_map §2 速度类错捆一行/误标(委员会本轮 8cdc4ad 修) | P3.1 拆三探测器分立:空间跳跃(raw teleport)/Kalman 残差(model-relative)/隐含速度(全程平均) | `217a8f7` |

### 第 1 轮 — `f84e8a9` 首审(3da5dfe..f0bb43f,裁决:计划通过可开工)
| # | 类型 | 委员会意见 | 消化 | commit |
|---|---|---|---|---|
| 1 | ⚠️阻塞 | ObsNoDetect→Fallen×1.6 = 用 absence 当正向(dropout-FP 根因),须 realness+door-distance 门控 | 新增 P6.1a:Fallen 因子改 `1+0.6·𝟙[R_i=real]·𝟙[¬door-exit]`,与 P3/P2.5 联锁 | `af3b3c4` |
| 2 | ⚠️阻塞 | P3.2 不能只靠零方差(cd2b 椅子 ghost 实测晃 ~50cm) | P3.2 改 4 条 AND 复合签名(跳变出生∧pose/z锁死∧钉死小区∧距门远),无单方差阈 | `40d1a2a` |
| 3 | ⚠️注意 | S_vol 尾形样本不足(bedside 尾 n=2) | P4.1 先粗档+待收紧;P9 margin 给区间不给点值,样本量列限制项 | `48a0934` |
| 4 | ❓注意 | O_b→S_Fallen 抑制勿掩 bedside-fall | P5.4 加边界:O_b 压 fall 仅 fresh∧高;leftBed 后 O_b 须及时落 | `c895c6b` |
| 5 | ❓注意 | firmware-fall 占位 ×3-4 偏高 | P2.4 shadow 期先 `SFallen:≤2`,待 P9 真机率标定 | `d22a769` |
| — | 命名 | P9'(前置带撇号)改编号 | P9'→P0(P1 已被 T-existence 占,P0=前置阶段0) | `0e798ae` |
| — | 一致 | P6.4 先因子化后按 oracle 升联合 ✓ | 无需改,与委员会一致 | — |

**阻塞项 1/2 已落为计划约束**(直接关系 cd2b 漏报与 dropout-FP 两类核心错误);P-task 落地时作硬门。

---

> **计划完结(§0–§10)+ 第 1 轮反馈消化(§11)。** 委员会反馈见 `doc/feedback.md`;施工方按 5min 节奏 pull 消化。
> 代码施工待委员会逐 P-task 签字后另起 commit,守 shadow-first / 不碰 alarm 决策路径。
