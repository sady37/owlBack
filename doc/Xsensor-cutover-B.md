# Xsensor Cutover — 变更/提案记录 (B 侧)

> 角色：**B（项目组）**。在本文件写变更/问题/待审批项；每次提交后 90s 检查 `doc/Xsensor-cutover-A.md` 的 A 回复。
> 任务：用 `tools/Xsensorv1` 的逻辑替换当前 `wisefido-sensor`。
> 约定确认（回 A·R0 三条约定）：① `doc/` = `owlBack/doc/`，**确认**；② 审核范围聚焦 FN 红线，**确认**；③ 验证铁律 = 真 case 重放、禁 unit test，**确认**。

---

## [B·R1] 2026-06-22 — 现状勘察 + cutover 范围界定 + 回 A·R0 三问

### B0. 关键范围澄清（请 A 先确认这条，它决定后面一切）

A·R0 把待替换对象写成"整个 `wisefido-sensor`（含 consumer/zonealarm/zoneengine/evaluator/playback）"。勘察后我认为**这是范围误判，需要收窄**：

- `wisefido-sensor` 是**生产服务**，roomengine 只是其中一块。各子系统 LOC：
  - `roomengine` 15253 loc — 雷达/跌倒引擎（**← cutover 真正发生在这里**）
  - `zoneengine` 5717 loc — zone 占用状态引擎（独立子系统）
  - `zonealarm` 878 loc — zone 报警规则（Stay rule 等）
  - `consumer` 1533 loc — 流消费者（monitor/alarm/activity/sleepstage/config）
  - `service` 1610 loc — MonitorBuffer / AlarmEnablementCache / FitnessTracker
  - `playback` 1081 loc — roomengine 回放（cmd/roomengine-playback）
  - `config` 438 loc
  - `evaluator` / `models` / `repository` = **0 loc（空目录，无需处理）**
- `tools/Xsensorv1` **只覆盖 roomengine 的跌倒裁决核**（belief/engine/adapter 三个新子包 + 自 copy 的馈送层），外加一个 replay-only 的 `cmd/xsensor`（消费 test:*、fire 落 Xsensor.log、**无任何生产 I/O**：不发 alarm、不发 card、不接 zone）。

**结论（大白话）**：Xsensor 不是"另一个完整的 sensor"，它是"把跌倒怎么判这件事重写了一遍并在重放道验证过"。所以 cutover ≠ 换掉整个 sensor 进程，而是 = **把 wisefido-sensor 里"怎么判跌倒"的那段旧逻辑，换成 Xsensor 的 DBN 四轴引擎，其余生产管道（消费流、发报警、发 card、zone、playback）原样保留**。

**B 推荐**：cutover 范围 = `wisefido-sensor/internal/roomengine` 内部的**裁决核**；zoneengine/zonealarm/consumer/service/playback/config **全部 KEEP，本战役不动**。请 A 确认是否同意收窄；若 A 坚持"整进程替换/方案乙长成完整 sensor 再删 wisefido-sensor"，请在 A 文件指出 —— 那是更大、更高风险的另一条路线，需要重新拆解。

---

### B1. 回 A·Q1 — cutover 总体步骤拆解（沿用零回归闸）

前置阶段（Stage0，纯准备，不改生产行为）：

- **S0. 移植 + 编译闸**：把 Xsensorv1 的三个新子包（`belief/` 14 文件、`engine/` 3 文件、`adapter/` 3 文件）+ `mm.go` + `layout_hash.go` 搬进 `wisefido-sensor/internal/roomengine`（module 从 `owlBack/tools/Xsensorv1` 改写为 `wisefido-sensor/internal/...`，import 路径机械改写）。此阶段只让它**能编译、被引用、走 shadow**（与旧裁决并行算、只 log diff，不参与 fire）。
- **S0.5 馈送层对账**：reconcile 11 个"两边都有但已分叉"的馈送层文件（见 B4），逐文件确认 Xsensor 侧的改动（covers 所有权、M×N firmware area_id、LogicID 统一、evict purge/coast）是**增量正确**而非丢了 wisefido-sensor 的生产钩子（persist / room_svg / track_status / feedback）。

正式闸（沿用历史"零回归闸"）：

- **StageA — 单房 cd2b**：DBN 接成顶层裁决（开关 ON，仅 cd2b 房），重放 cd2b/09e7/二义 lost-fall 三个 FN 红线 case，看**机制**（lost 登记 / hand-off 查询 / 跨雷达对账 / SLeft 注入 / 床占用解耦 / 床边摔不被归床），不看 fire 阈（规则 #3）。通过判据 = 三个历史 FN 根因守卫**机制仍在**。
- **StageB — 多房 Unit**：扩到整 unit（多雷达 + sleepad 融合），重放 unit 级 case，验 census/realness/N_r 分组、per-device covers 路由、同房双雷达不互相滤、床耦合 κ 无回归。
- **StageC — 删旧裁决**：StageA/B 全绿且 A 批准后，删除旧裁决文件（belief_shadow / ghost_adjudicator / fall_exempt / fall_rules_param / belief_neighbor / belief_adapter / belief_cell_contract），删 shadow 并表，关可逆开关（规则 #1.2 不留双路）。

### B2. 回 A·Q2 — 子系统覆盖清单 + 迁移/保留/删除决策

| 子系统 | Xsensor 是否覆盖 | 决策 | 理由（大白话） |
|---|---|---|---|
| roomengine 裁决核（belief/engine/adapter） | ✅ 新逻辑本体 | **MIGRATE（替换）** | 这就是要换的东西 |
| roomengine 馈送层（track_manager/cell/sensor_fusion/layout/mirror…11 文件） | ⚠️ 已分叉 copy | **RECONCILE（对账合并）** | 同源但 DBN 战役改过，须合并不能直接覆盖（会丢生产钩子） |
| roomengine 旧裁决（belief_shadow/ghost_adjudicator/fall_exempt/fall_rules_param/belief_neighbor/belief_adapter/belief_cell_contract） | ❌ 被新逻辑取代 | **DELETE（StageC）** | 被 DBN 四轴取代，规则 #1.2 不留兼容层 |
| roomengine 生产专属（persist/persist_postgres/room_svg/track_status/feedback） | ❌ replay 道用不上 | **KEEP** | 持久化/SVG/track API/反馈学习是生产必需，Xsensor 没有 ≠ 该删 |
| zoneengine（5717 loc） | ❌ 未覆盖 | **KEEP，不动** | 独立 zone 状态引擎，与跌倒裁决正交 |
| zonealarm（878 loc） | ❌ 未覆盖 | **KEEP，不动** | zone 报警规则（Stay rule 等） |
| consumer（1533 loc） | ❌ 未覆盖 | **KEEP，不动** | 生产流消费/发报；Xsensor 的 replay cmd 不能替代 |
| service（1610 loc） | ❌ 未覆盖 | **KEEP，不动** | MonitorBuffer/EnablementCache/FitnessTracker |
| playback（1081 loc） | ❌ 未覆盖 | **KEEP，需回归** | 回放走 roomengine，引擎换核后须验回放不挂 |
| config（438 loc） | 部分（Xsensor 有自己精简 config） | **KEEP wisefido-sensor 的**，新增 DBN 开关字段 | 生产 config 字段多，不替换只追加 |
| evaluator / models / repository | — | **空目录，N/A** | 无文件 |
| cmd/xsensor（replay 道） | replay-only | **本战役不并入生产**；StageA/B 期间留作并行验证道，StageC 评估去留 | 它是验证工具不是生产入口 |

### B3. 回 A·Q3 — 可逆开关 + 回滚路径

- **可逆开关**：env / config 单开关 `ROOMENGINE_DBN`（建议默认 `shadow`）三档：
  - `off` — 纯旧裁决（当前生产行为，回滚目标态）。
  - `shadow` — 旧裁决 fire + DBN 并行算只 log diff（S0/StageA 初期，零生产影响）。
  - `on` — DBN 顶层裁决 fire，旧裁决退场（StageA 单房可加房间白名单 `ROOMENGINE_DBN_ROOMS=cd2b,...` 灰度）。
  - 历史范式：[[dbn_cutover_state]] 的 `DBN_FIRE` union 开关 + 4-tag silent log，沿用。
- **回滚路径**：StageA/B 期间任何 case 暴露 FN → 开关回 `off` 或把房间移出白名单，**一步回滚到旧裁决**，无需 revert 代码。StageC（删旧码）之前回滚永远是配置级；StageC 之后回滚 = git revert（故 StageC 必须 A 显式批准 + 红线 case 全绿）。
- **安全红线（继承历史，迁移中绝不丢）**：cd2b 床边摔 FN（[[cd2b_0620_retest_fn_root_and_churn_bug]]）、09e7 双雷达 FN（[[two_radar_fn_firmware_areas_via_qinglan]]）、二义 lost-fall 三锁（[[ambiguous_lost_fall_coverage_conditional]]）、realness 绝不否决 fall（[[realness_never_vetoes_fall]]）。每个的守卫机制在 StageA/B 必须用真 case 实证仍在。

---

### B4. roomengine 文件级 delta 图（供 A 核对工作量/风险）

移植两边 roomengine 非 test 文件 diff（已忽略 import 路径噪声）：

- **NEW（Xsensor 独有，直接搬入）**：`belief/`(14) `engine/`(3) `adapter/`(3) `mm.go` `layout_hash.go`
- **DELETE（仅 wisefido-sensor，StageC 删）**：`belief_shadow.go` `belief_adapter.go` `belief_cell_contract.go` `belief_neighbor.go` `ghost_adjudicator.go` `fall_exempt.go` `fall_rules_param.go`
- **KEEP（仅 wisefido-sensor，生产必需）**：`persist.go` `persist_postgres.go` `room_svg.go` `track_status.go` `feedback.go`
- **RECONCILE（两边都有，按 diff 行数排序，越大越重）**：
  - `engine.go` diff≈1244（ws 2333 行含全部生产 I/O；xs 1303 行精简版委托给 belief/engine）—— **最大风险点**：不能直接覆盖，须保留 ws 的 PublishAIEvent/Alarm/RegisterRoom/Run/daily reload，仅置换裁决核 seam
  - `track_manager.go` diff≈757（LogicID 统一 / covers / evict purge / coast）
  - `sensor_fusion.go` diff≈149
  - `cell.go` 56 · `track.go` 49 · `layout_load.go` 36 · `mirror_detect.go` 25 · `bathroom_gate.go` 20 · `static_reflector.go` 14 · `layout_parser.go` 12 · `cell_learning.go` 4
  - **0 diff（完全相同，免对账）**：`alarm_event.go` `grid.go` `suite_census.go` `track_parse.go` `grid_extent.go` `grid_render.go` `kalman.go` `math_util.go`

### B5. 待 A 审批 / 决断的开放问题

1. **【范围】** 同意 B0 把 cutover 收窄到 roomengine 裁决核（zoneengine/zonealarm/consumer/service/playback KEEP 不动）吗？还是坚持整进程替换？—— **这条是阻塞项，A 不拍 B 不动手。**
2. **【seam 策略】** engine.go（diff 1244）reconcile 走"**在 ws engine.go 原地手术、保留生产 I/O、只换裁决核**"（B 推荐，安全）还是"**搬 xs 精简 engine.go 再补回生产 I/O**"？前者改动面小、保住生产管道；后者更接近 Xsensor 原貌但要重接所有 alarm/card/zone 钩子，风险高。
3. **【开关默认值】** `ROOMENGINE_DBN` 默认 `shadow` 还是 `off`？B 推荐 `shadow`（零 fire 影响但能持续收 diff），A 若担心 shadow 算力开销可定 `off` 手动开。
4. **【cmd/xsensor 去留】** StageA/B 期间保留 replay 道做并行验证；StageC 是否一并删除 `tools/Xsensorv1`（它与生产代码重复后即冗余）？

B 在等 A 对 B5.1（范围）拍板后再进 S0。其余 B5.2–4 可并行给意见。

*—— B·R1 提交，90s 后查 A 回复 ——*
