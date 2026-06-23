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

---

## [B·R3] 2026-06-22 — 收 A·R2：B 拥护方案甲 + 交付三道 S0 守卫核实（守卫②有重大发现，影响开关设计）

### B3.0 对 A·R2 的回应

- **拥护 A 强推荐方案甲**：A·R2.1 的量化（方案甲 engine.go 净 −118 行 + 馈送层 reconcile vs 方案乙 ~12k LOC，其中 ~11k 与 fall 正交）与 B 判断一致。等架构师据此最终 go/no-go。
- **B·R1 步骤 + B5.2 seam + B5.3/5.4 A 已采纳**，B 接受 S0 增三道守卫为放行前置。
- **提醒**：A·R2.4 再次点名「馈送层 11 文件逐文件对账清单」——**B 已在 [B·R2] §B2.2 交付**（seam 边界表 + 11 文件 port 判定 + 5 风险旗标），疑与 A·R2 起草交叉。请 A 复查 B·R2，如需补充再点。

### B3.1 守卫① 核实结果 — ✅ CLEAN（馈送层无自发跌倒开火残留）

逐函数审 `track_manager.go`(2442) + ProcessFrame/Tick 路径，全部开火点归类：

| 站点 | 行 | 归类 |
|---|---|---|
| `RecordRadarAlarm`→`PublishAIAlarm` | :815-853 | **FIRMWARE 透传（KEEP）**：转发固件 Fall/SittingOnGround，馈送层零算法判定 |
| `emitGhostVerdict`→`emitAIEvent` | :546-570 | **INFORMATIONAL 非开火（KEEP）**：发 track_verdict 供 cardagg 覆盖，注释明载「不参与 alarm 触发路径」 |
| `emitAIAlarm` helper | :539-543 | **框架 nil-safe**，无自发调用点 |

**v1 算法开火路径已删实证**：注释 :85（stillFallReportCount + v1 fire path 删）/:115（R4 bedside 触发删）/:813（gate-list verifier 删）/:1406 —— grep `stillFallReportCount`/`reportBedFall`/`reportZDrop`/`reportLostFall`/`reportSilentFall` **零 live 调用点**（仅注释残留）。`stillFallTimeoutSec`(:2320) 明载「不再驱动 alarm 发射」，仅作 AreaSit 学习闸。
**结论**：馈送层零自发算法开火，seam swap 后不会与新 DBN 双报。守卫① **通过**。

### B3.2 守卫② 核实结果 — 🔴 重大发现（belief_shadow 文件头自述与实际不符 + 影响开关单源）

- **文件头谎报**：`belief_shadow.go` :17-19 自述「**只 log 不 fire**…**绝不触发任何 alarm**」。**实际**：:878 `if e.dbnSelfFireEnabledFor(unitSuiteID, nowMs)` → :915 `e.PublishAIAlarm(...)` —— **belief_shadow 有真实开火路径**，由 `dbnMode = min(DBN_MODE env, 每-unit 冷启 cap)≥1` 门控（:104-105）。
- **真相**：生产 wisefido-sensor **已内置一代 cut1 DBN**（belief_shadow + `DBN_MODE`/`DBN_COLD_HOURS` env + 每-unit 冷启成熟度 cap，:39/:63/:86-99），默认 `DBN_MODE=0` 静默不发，但开火能力已 wired。它不是纯 shadow。
- **对守卫② 的意义**：A 要求「belief_shadow 降 log-only 或删、dbnMode 收进 DBN」是**真action item 非 no-op**。方案甲 seam swap 必须：**删除 belief_shadow 这条 cut1 开火路径**（:878-928），让 Xsensorv1 的 `belief/engine`（经 `OnRoomFrame` seam）成为**唯一** fire 权威——否则 cut1 DBN 与新 DBN 双报。
- **🔴 对开关设计的连锁影响（B 修正 B5.3）**：生产**已有** `DBN_MODE`(0/1/2 否决固件×DBN自发两正交轴)+`DBN_COLD_HOURS`+每-unit 冷启 cap 这套成熟开关。B·R1/R2 提的新开关 `ROOMENGINE_DBN` 会**与之重复**（违规则 #1.3 单源）。**B 改议**：cutover **复用并演进现有 `DBN_MODE` 语义**接到新 DBN（新 DBN 的 self-fire/veto-firmware 走同一套 dbnMode + 冷启 cap），**不新增 `ROOMENGINE_DBN`**；老 belief_shadow 的 dbnMode 门控逻辑随 seam swap 一并迁去包裹新 DBN。这样开关单源、冷启成熟度 cap（委员会 §6 7d 毕业）这套既有安全语义零损迁移。

### B3.3 守卫③ 核实结果 — 生产 TrackStatusBase 字段合约（待与新 DBN 期望字段逐字段对账）

生产版字段（:971-987）分组：

- **census 身份**：`TrackID`(固件号,会重用) `DeviceAddr`(雷达 UID) `RoomID`
- **pose/z**：`Pose`(0-7) `Z` `RawH/RawV/RawZ`(固件原始,publish 合约)
- **still**：`StillSec`/`StillBoxSec`(30s 50×50 box 抗抖)
- **判定**：`Verdict`(Real/Ghost/Pending/Anchored) `GhostPenalty`(0-100)
- **空间上下文**：`CellAreaType` `EnterTarget` `MoveActive` `TraverseDelta`(PR-5 占位) `SleepadInBed`(房级任一 sleepad InBed)
- **床占用正交**：不在 base，走独立 `BedOccupancyState()` 方法（:750 供 belief shadow P5）

**🔴 缺口（须 S0.5 补齐，呼应 B·R2 风险旗标）**：新 DBN 的 `TrackStatusBase` 还需 `LogicID`（census 身份键,身份统一）`FwAreaID`（固件 area_id,床区路由）`Present`（在帧/coast/blind 路由）`SleepadVitalPresent`（sleepad 生命体征→压 SBed）。守卫③ = S0.5 期把这 4 字段加进生产 base 且语义与新 DBN 消费端逐字段一致。守卫③ **= 待 S0.5 执行项，非阻塞**。

### B3.4 待 A·R3 + 一个开关决策请 A 拍

- **三道守卫核实交付**：①通过 ②真action item(删 cut1 开火路径+迁 dbnMode) ③待 S0.5 补 4 字段。
- **请 A 拍（B3.2 衍生）**：开关复用演进现有 `DBN_MODE`（B 改议,单源,继承冷启 cap 安全语义）vs 新增 `ROOMENGINE_DBN`（B·R1 原议,已自我否决）。B 推荐前者。
- B 仍 HOLD，等架构师最终 go/no-go + A·R3「正式放行 + S0 清单」。

*—— B·R3 提交，90s 后查 A·R3 ——*

---

## [B·R4] 2026-06-22 — 架构师提出"全 copy Xsensor + 仅改 redis 指向"；B 实测纠偏 → 提**方案丙**（Xsensor 当底座，焊回生产输出）

### B4.0 架构师原话

> 为什么不直接 Xsensor 全部 copy 进 wisefido-sensor，平替，仅需要修改 redis 指向，相当于用 Xsensor 替换原来的 DBN，原来的 Xsensor 本来就是源由 Sensor 并裁掉非 DBN 部分。

### B4.1 B 实测纠偏 — 前提 ~80% 对，一个承重缺口

- **对**：Xsensor = wisefido-sensor 馈送层 + 新 DBN，旧裁决（belief_shadow/ghost/fall_exempt）已裁。✓
- **🔴 缺口**：「裁掉非 DBN」时**连生产"输出"一起裁了**，不只裁旧裁决。实测：
  - `tools/Xsensorv1/internal/roomengine/engine.go:941-943` 订阅 `test:iot:monitor/event/alarm:stream`（**只消费**）；`:1061`「不 forward 到 iot:alarm:stream / 不入 alarm_events」；`:1103` runAlarmLoop「仅落账 + Tick，不 verify」。
  - 全仓 grep `PublishAIAlarm`/`XAdd`/`iot:alarm` **发布出口 = 0**，fire 只落 `Xsensor.log`。
- **结论**：Xsensor 现在「只看不喊」。「全 copy + 仅改 redis 指向」= 系统静默分析、**永不报警 = 生产瞎了**。redis **输入**好重指，但**输出**（发报警/card/zone）在 Xsensor 里被删了、不存在，非"改指向"能补。

### B4.2 但架构师方向对 → 方案丙（区别于已否的方案乙）

| 路线 | 底座 | 要做 | 风险 |
|---|---|---|---|
| 方案甲（B·R1 原提） | 旧 ws roomengine | 11 文件 DBN 增量**反向 port 进旧文件**（trackKey/LogicID/evict-purge FN 红线） | 反向 port 微妙改动易错 |
| 方案乙（A 已否 ~12k） | 全新独立 binary | 连 zoneengine/zonealarm/consumer/service **全 copy** 成独立体 | 重写正交管道 |
| **方案丙（架构师方向，B 改推荐）** | **Xsensor roomengine（已验证新代码）** | 生产输出**焊回 Xsensor**，正交子系统**留原地只调它** | 焊 I/O（机械活） |

**方案丙真实工作量**（非 A 的 12k——那是方案乙 copy 正交子系统）：zoneengine(5717)/zonealarm/consumer/service/playback **留 ws 原地、调新 roomengine、零 copy**。仅四步：
1. Xsensor roomengine（belief/engine/adapter + 馈送文件）搬进 `ws/internal/roomengine` 替换旧文件 + import 改写。**直接拿已验证新馈送层，跳过方案甲 11 文件反向对账 = 最大 win**。
2. 焊回生产输出到 engine.go：`PublishAIEvent/PublishAIAlarm` + track_manager 固件 Fall 转发 `iot:alarm:stream`；搬回 5 个生产专属文件 `persist/persist_postgres/room_svg/track_status/feedback`；补回被外部调用的 API（`SetAIPublishConfig`/`RegisterRoom` 生产路径/`SetDailyLayoutReload`…）。
3. 去 `test:` 前缀（= 架构师说的 redis 重指，最简单那步）。
4. roomengine 对外 API 面 = cmd/consumer/zonealarm/zoneengine/service 现有调用一致（外部不改即编译）。外部依赖 API 已勘察（`engine.Configure/NewEngine/RegisterRoom/Run/MapDevice*/OnDeviceUnfit/VetoCell/SetSuiteCensus/SetRoomTenant/SetRadarPresenceLookup/RealPeopleInRoom` 等）。

### B4.3 B 改推荐方案丙（替代 B·R1 的方案甲）

- **理由**：DBN 是难的/已验证/FN 红线；I/O 是机械的。方案甲 = 赌"把 DBN 微妙改动正确搬进旧文件"；方案丙 = 已验证 DBN 原样保住、只焊机械 I/O。**保住引擎整台、只接油门水电 > 把新引擎拆零件塞旧壳**，FN 风险更低。
- **对 B·R2/R3 既有产物的影响**：①B·R2 的 11 文件 seam 边界表/对账清单 → 方案丙下**反向用**：不再"port DBN 进旧"，而是"焊 I/O 进新"，但 seam 边界（裁决核 vs 生产 I/O）这张图**两路通用**仍有效。②B·R3 守卫① CLEAN、守卫② belief_shadow 删 cut1 开火路径、守卫③ 字段合约 → 方案丙下**守卫②自动满足**（Xsensor 本就无 belief_shadow，搬过去时旧 cut1 自然不在），守卫①③仍需核。③开关：方案丙下 Xsensor 自带 `dbnMode`/`DBN_MODE` 语义，B3.2 的"复用 DBN_MODE 单源"结论不变。
- **唯一待 S0 核实可行性点**：5 个生产专属文件（persist/room_svg/track_status/feedback）与旧 engine.go 内部字段耦合度——若重度耦合，焊到 Xsensor engine 需适配。S0 第一验证项，不影响路线选择。

### B4.4 待 A / 架构师拍

- 路线三选一确认：**方案丙（B 改推荐）** / 方案甲（B·R1 原提，回退选项） / 方案乙（A 已否）。
- A·R3 仍未出；本 B·R4 同时回应架构师直接提问 + 把方案丙交 A 评审。请 A 在 A·R3 一并裁路线（甲/丙）+ S0 清单。

*—— B·R4 提交，90s 后查 A ——*

---

## [B·R5] 2026-06-23 — 收 A·R3：GO 方案甲(copy-平替 hybrid)；B 确认 B·R4↔A·R3 收敛，解除 HOLD，进 S0

### B5.0 B·R4 与 A·R3 收敛确认（无冲突）

A·R3 与 B·R4 交叉提交，但**结论收敛**：A·R3 line 141 独立点出 B·R4 同一缺口——「Xsensor 的 `engine.go`/`track_manager.go` 把生产 I/O 裁了，**这两个文件绝不能整文件覆盖**，改 redis 指向也救不回；这两个以生产版为 base merge，其余 copy 平替」。这正是 A·R3.5 的 **hybrid**：
- **copy-平替（≈B·R4 方案丙，Xsensor 当底座）**：3 新子包 + mm.go + layout_hash.go + 8 个 0-diff + 6 个 trivial 档 → 直接拿 Xsensor 已验证版，**零反向 port**。
- **生产版 merge（方案甲）**：只 `engine.go`/`track_manager.go` 两个 I/O 承重文件 → 保 I/O、换裁决核。
- 即 B·R4 的 FN-risk 顾虑（反向 port 11 文件）被 hybrid 缩到**仅 2 文件**，且这 2 文件本就必须生产版 base（Xsensor 删了它们的 I/O）。**B 接受 hybrid，撤回 B·R4 整体改走方案丙的提议。**
- **唯一保留意见（非阻塞）**：S0.b 这 2 文件，"生产版 base 反向 port DBN 增量" vs "Xsensor 版 base 焊回 3 个 I/O 方法(AIPublisher/emitGhostVerdict/RecordRadarAlarm-forward)" 哪个 FN-risk 更低，B 在 S0.b 实做时按证据评估，在 A 要求的 **seam 复审 checkpoint** 一并报（A·R3.6 本就要看 seam diff）。默认按 A·R3.5 生产版 base 执行。

### B5.1 接受 A·R3 全部裁决

- 路线：GO 方案甲 copy-平替 hybrid。✓
- 守卫①结清 / 守卫②升 S0.c 硬删项（删 belief_shadow :878-928 cut1 开火路径 + ghost_adjudicator，dbnMode 门控迁去包裹新 DBN）/ 守卫③ S0.5 补 4 字段。✓
- 开关：复用演进 `DBN_MODE`+冷启 cap，不新增 `ROOMENGINE_DBN`。✓
- 工作流：S0.a→S0.e 按 A·R3.5；**S0.b/c 完成先报 A 复审 seam 切割，再进 S0.e 编译闸**；StageA 前给 A 看 S0.e 编译绿 + S0.b/c seam diff。✓

### B5.2 进 S0（开始执行）

B 解除 HOLD，从 **S0.a（机械 copy-平替 + import 改写）** 起步，并并行做 B·R4 承诺的可行性核实（5 个生产专属文件 persist/persist_postgres/room_svg/track_status/feedback 与旧 engine 内部字段耦合度，定 S0.b merge 适配量）。进展随提交更新本文档。

*—— B·R5 提交，进 S0.a ——*

---

## [B·R6] 2026-06-23 — S0.a 执行中发现：wisefido 已有 belief/ 子包(cut1 一代)，S0.a 须修正为"belief/ 替换非新增" + 守卫② 扩面

### B6.0 实做发现（修正 A·R3.5 S0.a 一处前提）

S0.a copy 时撞出关键事实：**wisefido-sensor/internal/roomengine 已有 `belief/` 子包（19 文件，末次 2026-06-13 commit 301733f）**，是一代 **cut1 DBN**，与 Xsensorv1 的 `belief/`（14 文件，2026-06-22 commit 02b98ac，新验证版）**文件集不同**——仅 3 个文件名相同（model/observation/state，且内容有 diff），其余各异（wisefido 独有 area/belief/calibration/decision_tau/doc/fall_reason/likelihood/survival/track；Xsensor 独有 bed_axis/coupling/decide/emission/filter/floor/joint/mm/neighbor/probe/realness）。

- **A·R3.5 S0.a 写「belief/(14) = 纯新增直接 copy」→ 不准**：belief/ **非纯新增**，它预存为 cut1。直接 copy 会把两代 belief 文件混叠（我已实测撞出并已回滚还原 belief/ 到 pristine）。
- **正确动作 = 替换（REPLACE）而非 copy**：删 wisefido 旧 belief/ 19 文件，drop in Xsensor belief/ 14 文件（import 改写）。
- **与守卫② 同源、扩面**：旧 belief/ 子包**只被** `belief_shadow.go` + `belief_adapter.go` 消费（实测 grep：flat roomengine 外无任何 import，cmd/consumer/service/zonealarm/zoneengine 全零）。这俩正是守卫② 要删的 cut1 文件。故 **守卫② 扩面**：cut1 一代 = belief_shadow.go(开火路径) + belief_adapter.go + belief_cell_contract.go + belief_neighbor.go + **整个旧 belief/ 子包(19 文件)**，一并删，由新 belief/(经新 engine/ OnRoomFrame seam)取代。删后新 belief/ 只被新 engine/ 消费，自洽。

### B6.1 S0.a 进度

- ✅ **engine/(3) + adapter/(3) + mm.go + layout_hash.go**：genuinely 新增（wisefido 无），已 copy + import 改写（`owlBack/tools/Xsensorv1/...`→`wisefido-sensor/...`），零残留 owlBack/tools 引用。**未 commit**（依赖新 belief/ 就位才能编译，遵规则 #1.6 build 绿才提交）。
- ⏸ **belief/ 替换**：待 A 确认修正后的 REPLACE 动作再执行（删 19 旧 + drop 14 新）。
- ⚠️ **工具层障碍**：执行替换需删除旧文件（rm / git rm），本会话 sandbox **拒绝了 rm/git clean/git checkout 等破坏性命令**（cp/sed/mkdir 放行）。请确认是否给 B 放行迁移所需的删除操作（删旧 belief/ 19 文件 + S0.c 删 cut1 flat 文件），否则 replace/删旧裁决无法落地。

### B6.2 待 A 确认（小修正，不改路线）

1. **修正 S0.a**：belief/ 由「copy 新增」改为「REPLACE 替换旧 cut1 一代」——确认？
2. **守卫② 扩面**：删除范围加上「旧 belief/ 子包 19 文件」（同 cut1 一代）——确认？
3. **放行删除操作**：迁移需 rm 旧文件，请放行（或告知 B 用何种方式执行删除）。
4. （内容核对，非阻塞）Xsensor 新 belief/ 是否**功能覆盖** wisefido 旧 belief/ 的能力（旧独有 calibration/survival/decision_tau/fall_reason/likelihood 等是否被新版等价吸收）——B 判断：Xsensor 是"正本"且新验证(NV1-8 绿/cd2b)，旧 belief/ 是 cut1 前身，替换方向正确；但建议 StageA 重放时一并验旧 belief/ 独有路径无回归。

B 暂停 S0.a belief/ 替换，等 A 这 4 点（尤其 #3 放行删除）。engine/adapter/mm/layout_hash 已就位待编译。

*—— B·R6 提交，90s 后查 A ——*

---

## [B·R7] 2026-06-23 — 收 A·R4(改判方案丙,采纳B·R4)+A·R5(路线锁死);确认硬约束 + 交付 S0.a 耦合度(放行前置)

### B7.0 收 A·R4 + A·R5（路线锁定，B 全部确认）

- **A·R4 改判放行方案丙（采纳 B·R4）**：感谢 A 诚实改判。FN 安全论点（保住已验证 DBN 整台、把"静默 FN 回归"风险挪成"编译器可见 wiring 错"）= cutover 最高红线的正解。
- **A·R5 路线锁死，B 确认三约束**：
  1. 🔒 **禁改 `tools/Xsensorv1/` 任何文件**（只读 copy 源 + 冻结验证道，cutover 后作永久回归基线）。
  2. **生产体 = wisefido-sensor**（身份不变，仅 roomengine 内部实现被替换）；**唯一改动地 = `ws/internal/roomengine/`**。
  3. **操作方向 = Xsensor 代码 → 复制进 → ws 容器**；"焊回 Xsensor" 已纠正为 "copy 进 ws 后**在 ws 副本上**焊"。
- **不再用甲/乙/丙命名**，只有一条路：ws 内部 DBN 换成 copy 自 Xsensor 的验证过代码 + ws 侧焊回输出 + 删旧 gate。

### B7.1 B·R6 belief/ 发现的归并（A·R5 已覆盖）

B·R6 的"wisefido 已有 cut1 belief/ 子包需替换"在 A·R5 下**自动归并**进 S0.b「整文件 copy 替换 ws 旧 DBN 实现」：copy Xsensor belief/ 进来时，ws 旧 belief/(19 文件 cut1) 整体被替换（同名覆盖 + 旧独有孤儿文件随旧 DBN 实现一并删）。守卫② 亦自动满足（A·R4.3：Xsensor 本无 belief_shadow，wholesale copy 后 cut1 路径天然不在）。**B·R6 #1/#2 已被 A·R5 吸收，无需单独裁决**；仅 #3（放行删除操作）仍待解（见 B7.3）。

### B7.2 交付 S0.a 耦合度核实（A·R5.3 放行前置）✅ GREEN — 5 文件全 LOW

核 5 个生产专属文件对 copy 进来的新(Xsensor) engine/cell/track 内部符号的耦合：**零缺失符号，全部 LOW weld cost**。

| 文件 | 触及符号 | 新代码是否提供 | weld |
|---|---|---|---|
| `persist.go` | RoomGrid/Cell 30+ 字段、RoomConfig 18 字段、Belief struct、`LayoutHash` | ✅ 全在（Xsensor Cell 字段超集，多 `BedAreaIDs` 不影响） | LOW |
| `persist_postgres.go` | `Persister`/`HistoryPersister` 接口、`SnapshotSchemaVersion` | ✅ 纯 PG 层，零 roomengine 内部依赖 | LOW |
| `room_svg.go` | RoomGrid/Cell 9 字段、`AreaType`/`Source` 枚举、radarutils | ✅ 枚举值逐一相同 | LOW |
| `track_status.go` | `TrackVerdict`/`VerdictName`/`AreaType` | ✅ track.go 同实现 | LOW |
| `feedback.go` | Engine `ApplyToCell`/`RoomForDevice`/`MountForDevice`、Cell 7 方法、`FallSuppressUntilMs` | ✅ 全在 | LOW |

**B 已抽验关键项**（防误报，A 可复核）：`ApplyToCell`/`RoomForDevice`/`MountForDevice` = Xsensor engine.go:455/314/323；`MarkRestZoneByFeedback`/`ClearNonHumanLearnedZone`/`MarkLearnBlocked` = cell.go:475/509/529；`FallSuppressUntilMs` = cell.go:204；`TrackVerdict`/`VerdictName` = track.go。

**🔴 一处必须向 A 澄清的区分（编译耦合 ≠ wiring）**：
- 上表 GREEN = 5 文件**能编译**（它们引用的 Engine/Cell 符号在新代码全在）。
- 但**新 engine 不会主动 CALL 它们**——Xsensor 把 persister/snapshotLoop/dailyReload/PublishAIEvent/Alarm/SetAIPublishConfig/RecordGroundTruth/PublishTrackStatus 这些**调用点裁掉了**。让新 engine **重新调用** persist(snapshot/hydrate)/feedback/track_status publish/room_svg + 焊回 6 个 API = **正是 S0.c 焊接工作**（A·R5.3 S0.c 已涵盖）。
- **结论**：S0.a 放行前置 = **PASS**（焊接成本 LOW，无深耦合、无缺字段需先补）；wiring 是已知的 S0.c 步骤，不构成放行障碍。

### B7.3 待 A（放行 S0.b + 解删除阻塞）

1. **S0.a 放行前置 = PASS**，请 A 确认放行进 **S0.b（整文件 copy Xsensor roomengine → ws 替换旧 DBN 实现 + import 改写）**。
2. **删除操作放行（B·R6 #3 仍未解）**：S0.b 替换旧 belief/ 孤儿 + S0.d 删 belief_shadow/ghost_adjudicator 需 `rm`/`git rm`；本会话 sandbox 拒绝破坏性命令（cp/sed/mkdir 放行、rm/git rm/git checkout 拒）。请放行删除（或告知 B 用何方式执行），否则替换/删旧无法落地。
3. 备注：B 在 S0.a 撞查时已 copy 了 engine/adapter/mm.go/layout_hash.go 到 ws（未 commit），将并入 S0.b 完整 copy 一并处理。

*—— B·R7 提交（S0.a PASS，等 A 放行 S0.b + 解删除阻塞），90s 后查 A ——*
