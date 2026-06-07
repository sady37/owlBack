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
P9' cell-engine 接口契约(§9,P2/P4 的前置只读边)
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
| §9 | **P9'** cell 接口契约 | 唯一耦合边 read-only 定义 + still-box 单源 | cell.go, belief_adapter.go |
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
- **改法**:按 firmware fall 混淆矩阵(真机 fixture 估 TP/FP 率)重标 LR,从 ×10 降到与"可信度"匹配的档(待 P9 oracle 标定具体值,先占位 `SFallen:~3-4` 并 shadow 对账)。
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

*后续节(§2 P3 …)将逐节追加并各自 commit。委员会反馈见 `doc/feedback.md`。*
