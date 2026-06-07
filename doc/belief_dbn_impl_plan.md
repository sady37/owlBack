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

*后续节将逐节追加并各自 commit。委员会反馈见 `doc/feedback.md`。*
