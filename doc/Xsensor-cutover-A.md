# Xsensor Cutover — 审核记录 (A 侧)

> 角色分工
> - **A（审核员，本文件作者）**：每 360s 检查 `doc/Xsensor-cutover-B.md`，在此文件给出审核意见。
> - **B（项目组）**：在 `doc/Xsensor-cutover-B.md` 写变更/问题/待审批项，每次提交后 90s 检查本文件回复。
>
> 任务：用 `tools/Xsensorv1` 的逻辑替换当前 `wisefido-sensor`。

## 约定（A 提请 B 确认）

1. **文件位置**：`doc/` 解析为 `owlBack/doc/`（Xsensorv1 源码、§7.7 契约、全部委员会 feedback 文档均在此）。若 B 期望其他路径请在 B 文件指出。
2. **审核范围**：cutover 的正确性/安全性/回归风险。重点关注 FN（漏报跌倒）红线——历史多个 case（cd2b 床边摔、09e7 双雷达、二义 lost-fall）的根因与守卫不得在迁移中丢失。
3. **验证铁律**：禁止 unit test / 替身 harness；验证 = 真实 case 重放（tools/export → tools/replay → 看 xsensor_xray 日志）。A 审核时会据此要求 B 给出真 case 实证而非自测断言。

---

## 审核日志

### [A·R0] 2026-06-22 — 初始化，等待 B 首次提交

- 当前 `doc/Xsensor-cutover-B.md` **尚不存在**，B 未开始提交。
- 已勘察现状，供 B 起步参考：
  - 新逻辑：`tools/Xsensorv1`（belief/roomengine/adapter，方案乙自成完整 sensor）。
  - 待替换：`owlBack/wisefido-sensor`（含 consumer/zonealarm/zoneengine/evaluator/playback 等 Xsensorv1 暂未覆盖的子系统）。
  - 近期已落地的契约工作见 git log：§7.7 hand-off 矩形核、GainedReal 升承重、durable purge 等。
- **A 提请 B 在 B 文件先给出**：
  1. cutover 的**总体步骤拆解**（建议沿用历史"零回归闸"：StageA 单房 cd2b → StageB 多房 Unit → StageC 删旧 sensor）。
  2. Xsensorv1 **尚未覆盖**而 wisefido-sensor 现有的子系统清单（zonealarm/zoneengine/evaluator/consumer/repository/service/playback…），以及各自的迁移/保留/删除决策。
  3. 切换的**可逆开关**与回滚路径。
- A 将每 360s 复查 B 文件；B 一旦写入即开始逐条审核。

---

### [A·R1] 2026-06-22 — 审 B·R1：质量过关，但 B0/B5.1 触发**架构级路线冲突**，A 不自批，升级架构师

**总评**：B·R1 质量高——勘察扎实、文件级 delta 图清晰、可逆开关三档设计到位、红线 case 守卫意识正确。事实声明 A 已逐条核实**基本属实**（见下）。但 B0 范围收窄 + B5.2 推荐撞上一条已被架构师**明文取消**的路线，这超出 A 的审核权限，A 标注阻塞并请架构师裁决，**不自行批准也不否决**。

#### A·1 事实核实（B·R1 声明 → A 复核结论）

| B 声明 | A 复核 | 结论 |
|---|---|---|
| engine.go ws 2333 / xs 1303，diff≈1244 | ✅ `wc -l` 实测 2333 / 1303 | 属实 |
| belief 14 / engine 3 / adapter 3 文件 | ✅ 实测一致 | 属实 |
| cmd/xsensor 无生产 I/O（不发 alarm/card/zone，fire 落 Xsensor.log） | ✅ `main.go:42` OutputPaths=Xsensor.log；grep 无 XADD/Publish-alarm/zoneengine 出口 | 属实 |
| evaluator/models/repository 空目录 | （B 已自验）暂采信 | 采信 |

**A 补充两条 B 未点明、但决定路线判断的事实**：
- **Tsensor 已不存在**（`find` 无 tsensor 目录、无 `run-tsensor.sh`，只剩 `run-xsensor.sh`）。即历史方案乙路线图 §9.3 的「Stage C 删 Tsensor」**已发生**——但 Tsensor 只是 wisefido-sensor 的**精简克隆/copy 源中转**，删它没动到真生产体。
- **Xsensorv1 至今 0 生产出口**：`main.go:2` 自述「fire 落 Xsensor.log」，internal/ 下无 alarm/card/zoneengine 发布。cmd/xsensor 多处注释写「**cutover 后服务 zoneengine total_people**」「去掉已剥离的 AI publish」——证明它的设计预期是**将来要 build-out 接回生产 I/O**（方案乙），现在只是半成品。

#### A·2 🔴 阻塞发现：B0 收窄 + B5.2「原地手术」= 已被架构师取消的**方案甲**

B0 把 cutover 收窄到「roomengine 裁决核，其余 KEEP」，B5.2 推荐「在 ws engine.go **原地手术、保留生产 I/O、只换裁决核 seam**」。这与 2026-06-15 架构师拍板的终局方向**直接冲突**。铁证（`doc/DBN-wire-roadmap-p6.md` §9.1，原文）：

> **9.1 终局变更：方案甲 → 方案乙**
> §0 原"验完走方案甲（**注入 Tsensor 换 belief**）归档"**取消**。
> 改：Xsensorv1 自 build-out 成消费 test:* 的 replay-重算 sensor，**删 Tsensor**。**DBN 作顶层唯一决策**…馈送层之上只有 DBN。

- **方案甲 = 注入现有 sensor、原地换 belief**（= B5.2 推荐的「ws engine.go 原地手术」），**已明文取消**。
- **方案乙 = Xsensor 自己长成完整 sensor、删旧**（钦定终局）。

**大白话**：B 提的「在老房子里把旧引擎抠出来换上新引擎、其余管线照用」，正是架构师一周前否掉的那条路；钦定的是「把新引擎那栋楼盖完整（自带水电=生产 I/O），再拆老楼」。

#### A·3 A 的工程判断（给架构师，非 A 拍板）

A 不机械挡 B——否决方案甲的**原始理由今天可能已消解**，须架构师重判：

- **当年否决方案甲的语境**：belief 四轴**尚未成熟**，怕「半成品注入旧躯干 → 被旧 gate-list 拖回打补丁」（floor-strip 血泪、"在打补丁不是建框架"）。
- **今天**：belief 四轴已独立验证成熟（NV1-8 绿、cd2b 0.5203 零回归、四轴全内化零 gate）。**把一个已成熟的引擎 graft 进生产躯干，与当年"半成品注入"的风险性质不同**——否决前提大幅减弱。
- **两条路的真实成本不对称**：方案乙 build-out 要给 Xsensor **重建全部生产 I/O**——alarm/card/zone/persist/room_svg/track_status/feedback/playback/consumer/service，**外加 zoneengine(5717 loc)/zonealarm(878 loc) 这些与跌倒裁决正交的子系统**。为换一个裁决核而重写数千行无关生产管道，风险与工期都远大于 B 的原地手术。

**A 倾向**：B 的收窄（方案甲复活/原地手术）在今天是**务实合理**的，因为①否决方案甲的"半成品污染"前提已随 belief 成熟消解；②方案乙重写正交子系统成本不对称。**但因其推翻 2026-06-15 钦定方向，A 无权自批**——须架构师**显式重新批准这次路线回归**。

#### A·4 对 B5.2 / B5.3 / B5.4 的先行意见（待路线定调后生效）

- **B5.2 seam 策略**：若架构师批准收窄，A 同意 B 推荐的「ws engine.go 原地手术、保住生产 I/O、只换裁决核 seam」——前提是 S0.5 馈送层 reconcile 必须**逐文件**对账（engine.go diff 1244 是最大风险点），且 seam 边界要在文档画清「哪些是裁决核（换）、哪些是生产 I/O（留）」，防止旧 gate 残留偷偷参与 fire。
- **B5.3 开关默认值**：A 同意默认 `shadow`。算力开销在单房/单 unit 重放道可忽略，持续收 diff 的价值 > 开销。生产线若上 `on` 前必经 `shadow` 比对窗。
- **B5.4 cmd/xsensor 去留**：StageA/B 期间**必须保留**作并行验证道（它是当前唯一跑通 DBN 顶层裁决的入口）。StageC 是否删 `tools/Xsensorv1`，取决于路线——若走方案甲，移植完成后 Xsensorv1 与生产码重复，可删；若走方案乙，Xsensorv1 **就是**未来生产体，不删反而要扶正。**此问的答案绑死在 A·3 的路线裁决上**，现在不预判。

#### A·5 给 B 的指令

1. **B HOLD 在 S0 之前**，勿动码。B5.1 是阻塞项，且已升级为架构级路线裁决（非 A 能批）。
2. A 已就「方案甲复活 vs 方案乙 build-out」请架构师拍板（见本轮对话）。架构师裁决落定后，A 会在 A·R2 明确放行哪条路 + S0 前置条件。
3. 在等裁决期间，B 可**并行**做一件零风险、对两条路都有用的事：把 S0.5 的「11 个已分叉馈送层文件」**逐文件 diff 成对账清单**（每个文件标注：Xsensor 侧改了什么生产钩子有没有丢），尤其 engine.go / track_manager.go 两个大头。这份清单方案甲方案乙都要用。

*—— A·R1 提交。已就路线冲突请架构师裁决；B 请 HOLD 等 A·R2。——*

---

### [A·R2] 2026-06-22 — 架构师令 A 深评工作量；A 量化两路，**强推荐方案甲(B 收窄)** + 三道 S0 守卫

架构师裁决：「先让 A 深评工作量再定」。A 已对两条路做生产 I/O 出口盘点 + engine.go seam 边界勘察，量化如下。

#### A·R2.1 工作量量化对比（实测 LOC + seam 勘察）

| 维度 | 方案乙（钦定：Xsensor build-out 成完整 sensor 删旧） | 方案甲（B 收窄：ws engine.go 原地换裁决核） |
|---|---|---|
| **新增/改动 LOC** | **~12,000 LOC** 新增/搬运 | engine.go **净 −118 行**（删 ~133 加 ~15）+ 馈送层 reconcile |
| **其中服务 fall 裁决** | ~1,000 LOC（且已在 Xsensor 内） | 全部都是裁决核 |
| **与 fall 正交的被迫搬运** | **~11,000 LOC**：zoneengine 5717 / zonealarm 878 / consumer 1533 / service 1610 / playback 1081 / persist 新写 ~200 | **0**（zoneengine 等正交子系统零触碰） |
| **生产 I/O 出口** | 13 个出口现 Xsensor **缺 12 个**（只有 census/RealPeopleInRoom 共享），全要重接 | 13 个出口**原样保留**，不碰 |
| **裁决核 seam** | 全新接线 | **唯一且清晰**：`engine.go:867-905 applyVerdictDeltas`「唯一 verdict 写点」+ `:1035 Adjudicate`，seam 清晰度 8/10 |
| **双路 fire 结构风险** | 无（全新） | **无**——silent_leftbed/lost_fall/engine_z_drop 已删或内化为 belief_shadow 的 log-only，engine.go 无旧 gate 直接 fire |
| **可逆/回滚** | 整进程切换，回滚 = 重启旧进程 | 配置开关一步回滚（StageC 前永远配置级） |

#### A·R2.2 A 的判断与强推荐

**A 强推荐：批准方案甲（B 的收窄）。** 三条硬理由：

1. **成本悬殊且不对称**：方案乙 ~12k LOC 里 **~11k 是与跌倒裁决正交的稳定生产管道**（zoneengine 等当前线上跑得好好的）。为换一个裁决核把整套下游管道拆进 Xsensor 再删原版 = "为换引擎把整辆车拆了重装"，纯增风险零裁决收益。方案甲净改动是它的零头。
2. **方案乙的唯一架构收益（"馈送层之上只有 DBN"）方案甲也能等价达成**：裁决核 seam 已是**单一写点**（applyVerdictDeltas）、无双路 fire。方案甲换的就是这个唯一 seam——换完后生产版的裁决纯净度与方案乙等价，只是复用了现成生产 I/O 而非重写。
3. **否决方案甲的原始前提已消解**：当年否决是因 belief 半成品、怕被旧 gate 拖回打补丁；今天 belief 四轴成熟（NV1-8 绿、cd2b 0.5203 零回归），且勘察证实 engine.go **已无旧 gate 双路结构**——grafting 一个成熟引擎进一个已清干净的 seam，风险性质与当年完全不同。

#### A·R2.3 批准方案甲的三道 S0 强制守卫（针对勘察暴露的三个风险点）

方案甲低风险但非零风险。seam 勘察暴露三处必须在 S0 闸前核实，A 列为放行硬条件：

- **守卫① 防新旧双报**：核实 `track_manager.go` / `ProcessFrame` 内**无隐含 fall 判定残留**（~50-150 行待核）。若有，必须拆除——否则换核后旧路径与新 DBN 双报。
- **守卫② belief_shadow 归位**：旧 `belief_shadow.go`（~1070 行 log-only）+ 其 dbnMode 冷启动档位，必须明确**降为 log-only 或删除**，档位逻辑收进 DBN，不得偷偷参与 fire。
- **守卫③ 合约一致性**：`TrackStatusBase` 的 census/bed/exit 字段在新旧裁决核间的语义必须逐字段验证一致（DBN 消费的字段 = 生产馈送层产出的字段）。

#### A·R2.4 给 B 的更新指令

- 路线**待架构师据本量化做最终 go/no-go**（A 已强推荐方案甲）。一旦批准：
  1. B 的 cutover 步骤拆解（B·R1 §B1）**整体采纳**，但 S0 增加上述三道守卫为放行前置。
  2. B5.2 seam 策略**采纳 B 推荐**（ws engine.go 原地手术，只换 applyVerdictDeltas/Adjudicate seam，生产 I/O 原样不动）。
  3. B5.3 开关默认 `shadow`、B5.4 cmd/xsensor 保留至 StageC 后评估——均按 A·R1 §A4。
- B 仍 HOLD 一步：等架构师对 A·R2 的最终批复，A 会在 A·R3 落「正式放行 + S0 清单」。
- 等待期 B 可并行做的零风险活不变（馈送层 11 文件逐文件对账清单），现在再加一项同样零风险且方案甲直接需要的：**核实守卫① —— grep track_manager.go/ProcessFrame 的隐含 fall 判定残留**，产出清单。

*—— A·R2 提交。已带量化数据请架构师做最终 go/no-go；B 继续 HOLD，可并行做对账清单 + 守卫①核实。——*
