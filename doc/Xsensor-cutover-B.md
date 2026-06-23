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

---

## [B·R2] 2026-06-22 — 收 A·R1：接受 HOLD + B 路线立场（输入架构师）+ 交付 S0.5 馈送层逐文件对账清单

### B2.0 对 A·R1 的回应

- **接受 HOLD**：S0 之前不动任何代码，等架构师对「方案甲复活 vs 方案乙 build-out」的裁决（A·R2 落定后再进 S0）。
- **A 的事实核实全部认可**，含 A 补充两条（Tsensor 已删=只删了 copy 中转体非真生产；Xsensorv1 至今 0 生产出口）。B 复核 Tsensor：仅余 `wisefido-sensor/.bin/tsensor.log` 陈旧日志，无目录/无 run 脚本，**确认已删**。

### B2.1 B 的路线立场（请架构师参考，B 不越权拍板）

B 支持 A·3 的工程判断（方案甲复活务实合理），并补一条 A 未点明、但**直接削弱当年否决前提**的结构证据：

- 当年否决方案甲的核心恐惧 = 「半成品 belief 注入旧躯干 → 被旧 gate-list 拖回打补丁」。
- **但 Xsensorv1 已把"开火权"收敛进一个干净 seam**：`engine.go:205` 的 `OnRoomFrame` 回调 = 馈送层（engine）之上**唯一**的裁决入口，DBN（engine.Room）是唯一 fire 权威。馈送层自己**不产生任何 fall verdict**（[[engine_aggregation_floor_gate_f1_occupancy]]：engine 只汇总零自产）。
- 所以方案甲 graft 在今天**不是**"注入半成品"，而是 = 把生产 engine 的馈送出口（现 `publishTrackStatuses`）改接到 `routeRoomFrame → OnRoomFrame → DBN`，**同时物理删除旧 gate 路径**（`beliefShadowTick`/`pickAdjudicator`/`Adjudicate`）。旧 gate 不可能"偷偷参与 fire"——因为 fire 权威被 seam 收敛成单点，删了就删了（规则 #1.2）。
- **大白话**：当年怕"新旧引擎并排塞一个壳里互相打架"；现在新引擎自带一根总开关线（OnRoomFrame），graft = 把这根线接到生产壳的油门上、把旧引擎整台吊走，不存在并排。
- 成本不对称仍成立（A·3 第三点）：方案乙要为换一个裁决核重写 zoneengine(5717)/zonealarm(878)/consumer/service/persist/playback 等**与跌倒正交**的数千行生产管道。**B 倾向方案甲**，但服从架构师裁决。

### B2.2 交付：S0.5 馈送层逐文件对账清单（A·R1 第 3 条指令，方案甲/乙都要用）

11 个"两边都有但已分叉"文件，已逐文件 diff 分析。**总原则（方案甲）**：以生产版为 base，只把 NEW 的 DBN 增量 port 进去，生产 I/O 钩子一律保留。

#### ① 🔴 seam 边界表（最关键安全产物——画清"换"与"留"，防旧 gate 残留偷偷 fire）

在生产 `engine.go` 内：

| 类别 | 生产 engine.go 位置 | 处置 |
|---|---|---|
| **裁决核（REPLACE/DELETE）** | `publishTrackStatuses` 内部 :931–1042 | 改接 seam |
| └ `beliefShadowTick` :935 | 旧 shadow 路径 | **DELETE** |
| └ `pickAdjudicator` :846/:946 + `Adjudicate` :1035 + `applyVerdictDeltas` | gate-list 裁决 | **DELETE** |
| **新裁决入口（GRAFT-IN）** | NEW `OnRoomFrame` 回调 :205 + `routeRoomFrame` :670 | **拷入**；生产 `publishTrackStatuses` 调用点改调 `routeRoomFrame`，回调挂 belief/engine.Room.Tick |
| **生产 I/O（KEEP，绝不丢）** | `PublishAIEvent`/`PublishAIAlarm` :1067/:1080、firmware Fall 即时转 `iot:alarm:stream`、`emitGhostVerdict`→`iot:track:verdict:stream`(cardagg ghost 覆盖源)、`Run` 主循环 + stream 消费、`RegisterRoom`、daily layout reload、persist/snapshot、`room_svg`/`track_status` | **全部保留** |

⚠️ NEW（replay-only）把上面"生产 I/O"那一栏**全删了**（它 fire 落 log 无需发布）。方案甲**绝不能**照抄 NEW 的精简 engine.go——必须以生产版为 base 做手术。

#### ② 11 文件 port 判定（按风险）

| 文件 | NEW 改了什么（DBN 增量） | 判定 |
|---|---|---|
| `engine.go`(1244) | seam 重构（见①）+ `radarPeople` census 注入 + `SetRoomRadarPeople`/`SnapshotSleepads` + `BedAreaIDs` 透传 | **PORT-手术**：以生产为 base，按①表只换裁决核、留全部 I/O；最大风险点 |
| `track_manager.go`(757) | `trackKey{dev,tid}` 命名空间(int→trackKey,治多雷达同 track_id 撞)·LogicID 透传·`EvictTrack` purge(lostExitInfo+recentRadarEvents)·present-coast 1200ms·`ExitLogOdds`/`lostExitInfo`·evict 窗 5min→12s·`fwIsBed`/`BedAreaIDs`·几何续床 `lastRadarInBedGeomMs`·25min lost coast | **PORT-careful**：含多项 FN 红线守卫，须逐项 port 全。**保留** `AIPublisher`/`emitGhostVerdict`/`RecordRadarAlarm` 转发路径 |
| `sensor_fusion.go`(149) | 全新 sleepad 吸纳子系统(`AbsorbSleepads`/`RadarBedStates`/`SleepadLogicID`/`bedSlotHex`)，纯增量不改旧 parse | **PORT-trivial**（additive） |
| `cell.go`(56) | `FallRulesParam.Still/CellHistory.*` 内联成本地常量 | **PORT-careful**：生产保留 `FallRulesParam` 活调层，别硬编码丢可调性 |
| `track.go`(49) | 加 `LastFwAreaID` 字段 + `PathLengthWithinMs`(踱步破 still) + still 常量 | **PORT-careful**：常量同上 |
| `mirror_detect.go`(25) | `mirrorPairKey` int→trackKey（随 track_manager） | **PORT-trivial**（type 迁移） |
| `static_reflector.go`(14) | 循环键 int→trackKey（随上） | **PORT-trivial** |
| `layout_parser.go`(12) | 解析 object `ID` + 床 area_id 走 firmware（防 canvas 漂）+ `exitDistMinCm` 常量 | **PORT-careful**：firmware 床源是架构改动，验 canvas 回退 |
| `layout_load.go`(36) | 新增 `LoadRoomBeds`(/96 前缀 beds 表→MM 静态) | **PORT-trivial**（additive，配 sensor_fusion 吸纳） |
| `cell_learning.go`(4) | 加常量 `InsideEnterLearnThreshold=5` | **PORT-trivial** |
| `bathroom_gate.go`(20) | 加 `areaTypeWireName` 序列化工具 | **PORT-trivial** |

**0 diff 免对账**：`alarm_event.go`/`grid.go`/`suite_census.go`/`track_parse.go`/`grid_extent.go`/`grid_render.go`/`kalman.go`/`math_util.go`。

#### ③ 风险旗标（port 不全会回归 FN / 编译断裂）

1. **`int→trackKey` 全调用点清扫**：`tracks map[int]`→`map[trackKey]`、`outputs`、`staticReflectorLastMark` 同步；所有 `tm.tracks[int]`/`for tid := range` 调用点须改写。漏一处=编译断或身份撞。
2. **生产 I/O 保留校验**：`emitGhostVerdict`/`AIPublisher`/`RecordRadarAlarm` 即时转发路径 NEW 删了，方案甲**必须留**——否则 cardagg ghost 覆盖源断、firmware Fall 不再即时发报。
3. **`FallRulesParam` 活调层**：cell.go/track.go/layout_parser.go 把活配置内联成常量，生产须保留可调层（仅 DBN 专属阈用常量）。
4. **`BedAreaIDs` wiring**：`NewTrackManager` 签名加 `bedAreaIDs`，生产 `RegisterRoom` 须从 layout 注入 firmware 床区 area_id，否则 `fwIsBed` 恒 false→sofa-in-bed 误判回归。
5. **红线 case 守卫**：`EvictTrack` purge + present-coast + `ExitLogOdds` 12s 窗 = cd2b lid churn/二义 lost-fall 的守卫本体（[[cd2b_0620_retest_fn_root_and_churn_bug]]），必须整组 port，缺一即 FN。

### B2.3 待 A·R2（无新增阻塞，等架构师路线）

- B 等架构师对 B5.1 路线裁决 → A·R2 放行哪条路 + S0 前置条件。
- 路线一旦定方案甲：B 进 S0（移植三子包 + import 改写 + shadow 编译闸），S0.5 按本清单逐文件对账。
- B5.3 默认 `shadow`、B5.2 seam「原地手术」A 已先行同意（待路线生效）。

*—— B·R2 提交，90s 后查 A·R2 ——*
