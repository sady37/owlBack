# DBN 四轴集成路线图（build order ③，方案乙）— A 草拟，待 C 复审

> ground truth：`DBN-Zone-Room.md`（含 §A neighbor）。本图 = 把 belief 单元已独立验证的四隐轴
> wire 进真实 Xsensorv1 roomengine 的施工图。
>
> **⚠️ 2026-06-15 新指令修订**：架构师拍"Xsensor 替换 Tsensor，所有重放接入 Xsensor"——§0 终局
> 方案甲取消、改方案乙 build-out 删 Tsensor；§4 copy 源/范围扩展；§5 realness 契约已被取代。
> **§1–§8 = W3.1–W3.5 历史记录（四轴 wire 进 engine，已基本完成）；新方向与偏离统一收进 §9（权威）。**
> 本轮已动码（用户授权，非抢跑）：feed 库剥离 green，详 §9.7。

## 0. 目标与边界

- **目标**：四隐轴（S 人态 / B 床占用 / realness ghost / neighbor 跨房）从 belief 单元 wire 进一个真 roomengine（**方案乙**：新建 roomengine + copy 非DBN包），端到端验证全轴融合；~~验完走方案甲（注入 Tsensor 换 belief）归档~~ **【2026-06-15 取消，改方案乙 build-out + 删 Tsensor，见 §9.1】**。
- **边界**：Xsensorv1 = **replay 重算道（非生产）**；不碰 wisefido-sensor 生产码。**【生产=wisefido-sensor 线上实跑；Xsensor 喂 test:* 仿 iot:* 把重放当生产重算，边界仍守无反转，见 §9.0】**
- **铁律守**：四轴**内化进 joint 前向滤波、不加 gate**（DBN 根本目的，[[fall_detection_risk_stratified_design]]）；fall_data artificial 不标定安全阈（[[fall_data_is_artificial_test]]）；两态 sleepad（在线 OR 没有）；bed_id 时间窗绑定禁坐标反推（§27）。

## 1. 现状 gap（survey 2026-06-16 坐实）

| 件 | 状态 |
|---|---|
| belief 四轴 | ✅ 单元验完：joint/bed_axis/coupling/emission/decide/probe（S/B）+ realness（GH/RV）+ neighbor（NV1-8）|
| adapter | 只译 **S/B**：raw→`Observation`→`filter.Step(logPsi, logPhi)`。**未产 `RealnessObs` / `SiblingHandoff`** |
| filter.Step | 签名 `(nowMs, online, logPsi, logPhi)` = **只 S/B**。realness（`PRoomHasReal`）、neighbor（`GateBlindRow` on Blind 行）**未入 Predict/Correct** |
| roomengine | **无**。`replay` = 单房 cd2b harness，无多房 census / track 生命周期 / device 路由 |

## 2. 四轴融合契约（③ 核心架构决定，先定再 wire）

避免 gate = 四轴全经 `filter.Step` 的 Predict/Correct 融合，无独立否决分支：

- **neighbor → Predict**：ρ_xroom 仅 lost-track 激活，`GateBlindRow` 改 T_S 的 Blind 行（→Fallen 整流入 →Left）= **转移先验**（§A.2；ρ=0 行不变 = lost-fall 安全默认，非 gate）。
- **realness → Correct（emission 调制）**：per-track `RealnessTrack` 维护 P(real)；`PRoomHasReal` 喂 fall 后验——真人 track 消失（P(real) 高）→ Blind→Fallen ramp 不被"无 track"抑制（共生律）；ghost 消失（P(real) 低）→ 不喂 SFallen（fall×P(real)）。
- **ghost→neighbor 接口（§A.3①）**：`SiblingHandoff.GainedReal` 吃兄弟房**去 ghost** 占用 = realness 喂 neighbor。
- **正确性 oracle（关键）**：realness+neighbor 取中性（ρ=0、P(real) 使发射恒等）→ 应**逐 tick 等价**现 S/B-only。这是 wire 没接错的零回归闸。

## 3. filter.Step 签名扩展（最小侵入）

- 现：`Step(nowMs, online, logPsi, logPhi)`。
- 扩：加 `rhoXroom float64`（lost-track 时 >0，Predict 吃它做 GateBlindRow）+ `realness RealnessSummary`（`PRoomHasReal` + per-state fall 调制，Correct 吃它调 fall 发射）。
- 中性值（rhoXroom=0、realness 恒等）→ 回退现行为（§2 oracle）。

**C §42 裁定 = ①折进 logPhi（定）**：独立调制步 = filter.Step 外另起修正 = 本质**软 gate**，违背"四轴全经 filter.Step 融合"；折进 logPhi（fall 发射项 ×P(real)）走同一 Correct 路径才是真内化。realness 成为发射的一部分，不另起步骤。

## 4. copy 清单（方案乙，源 = `wisefido-sensor/internal/roomengine/`）

> **【2026-06-15 修订，见 §9.2】**：源改 **Tsensor**（已精简克隆）；范围扩 **consumer(test:* 前端) + 床融合(zoneengine)**；bathroom_gate"不copy"实暂留（§9.5）。下表 = 原 roomengine 范围。

| 处置 | 文件/能力 | 理由 |
|---|---|---|
| **copy** | kalman + track parse (track.go) | device frame → RadarTrack |
| **copy** | grid / grid_extent / layout_load | canvas → Rect/entrance/wall（喂 realness 出生地·近墙 + adapter Beds）|
| **copy** | cell / cell_learning | AreaDeny → realness Static 先验（cell 三时标**只读**，[[cell_dbn_timescales_stillbox_single_source]]）|
| **copy（Gap1 补）** | **mirror_detect / static_reflector** | realness 的 `CoexistRho`/`IsReflection`（镜面几何）+ Static 签名（困 BirthPos·近墙）的**几何输入源**——几何事实，区别于该废的硬结论 `ghost_adjudicator` |
| **copy/改造（Gap3 补）** | **track_manager 连续指标层**（`updateContinuousIndicators`/`StillBoxRunStart`/`StillSec`）| **still-box 单源产出处**（[[cell_dbn_timescales_stillbox_single_source]]：全系统单算一次）——`CrossedStillPeriod` 的数据源，必须保留；只去掉 gate-list 部分 |
| **copy** | fall_rules_param / risk | Census → decide C_FN |
| **改造非 copy** | engine 主循环 | 旧 `beliefShadowTick` → 新 four-axis `filter.Step`；去 gate-list/belief_shadow |
| **改造非 copy** | suite_census | 喂 `SiblingHandoff`（复用 P_id 跨区账**语义**不照搬实现，§A.3②）|
| **不 copy** | belief_*.go / ghost_adjudicator / bathroom_gate / fall_exempt / **track_manager 的 gate-list 残余**（engine_z_drop/silent_leftbed/lost_fall）| **被 DBN 四轴取代**（删即删，#1.2）。注：track_manager **拆**——连续指标层 copy（上行），gate-list 残余删 |

## 5. adapter 译入契约（raw → 各轴 Obs）

> **【2026-06-15 已被取代，见 §9.4】**：RealnessObs 契约 = realness 轴 2026-06-14 重定前版；现 engine.Room 从 census/walls(IsReflection) **内部涌现**（Real vs Mirror，Static 删）。本节作历史保留，非现行接线契约。

- **RealnessObs（每 track）** ← 出生档案 + cell/grid：
  `bornNearEntrance`←entrance geom + 出生 XY；`inAreaDeny`←cell；`Displaced`←相对 BirthPos 位移；`ConfinedNearWall`←static_reflector 近墙+困 BirthPos（Gap1 源）；`AgeLongStatic`←寿命×静止；`CoexistRho`/`IsReflection`←mirror_detect 配对共动+镜面几何（Gap1 源）；`CrossedStillPeriod`←**still-box**（由 §4 copy 的 track 连续指标层**单源**产 `StillSec`/`StillBoxRunStart`，Gap3 源已定，[[cell_dbn_timescales_stillbox_single_source]]）。
- **SiblingHandoff（lost-track 时，每兄弟房）** ← 跨房 census：
  `ArrivalDeltaMs`=兄弟房 +1 track ts − 本房 lost ts（守恒重现 P6.5）；`CAttr`←源型（sleepad 0.9/room-enter 0.8/radar-only 0.2）；`GainedReal`←兄弟房新增 real track 去 ghost 占用；W=`HandoffWindowFor(base, publicness)`、D=`DelayWindowFor(stillVanish, margin, coverage)`，unit_property/coverage 来源 spatial/device config。
- **Observation（S/B）**：维持现状。

## 6. wire 顺序（每步独立可验，防大爆炸）

1. **W3.1 filter 融合**（纯 belief 包，不碰 roomengine）🚩**C 复审 gate①**：扩 Step 签名接 realness(折 logPhi)+neighbor(进 Predict GateBlindRow)；测试 = 中性零回归 oracle / realness 共生律端到端 / neighbor lost-track 整流。低风险可先起。
2. **W3.2 roomengine 单房骨架**🚩**C 复审 gate②（floor-strip 教训正在此守）**：copy 最小包（track 连续指标/kalman/grid/layout）；单房 engine 主循环 driving `filter.Step`（接 S/B/realness，neighbor=空）；**cd2b 单房 replay 端到端复现 belief 单元结果**（零回归闸——belief 单元能否在真 roomengine 复现的第一道关）。
3. **W3.3 realness 接通**：adapter 译 RealnessObs（接 mirror_detect/static_reflector/cell/still-box 源）；GH/RV 等价 case 在 roomengine 端到端（ghost 不喂 fall / 真人摔不被滤 RV4）。
4. **W3.4a roomengine 多房编排**（Gap2 展开——最重 wiring，独立成步防大爆炸）：engine 持多房 map；`device_addr→room` 路由（IPv6 /prefix 解析，[[config_double_path_env_silently_ignored]] 坑）；suite census = 跨房 track 账（P_id 跨区，产 lost ts / arrival ts / per-room 去 ghost 占用）；**跨房读出契约** = 兄弟房去 ghost 占用 + track 守恒（+1 track 匹配丢失）。验 = 多房 fixture census 账平（track 进出账目正确）。
5. **W3.4b neighbor 接通**🚩**C 复审 gate③**：adapter 译 SiblingHandoff（吃 W3.4a census 读出）；多房 hand-off replay（挪去邻房压 phantom / lost-fall 安全默认 / NV-等价）。
6. **W3.5 全轴端到端**：cd2b + 多房 fixture 全四轴跑；decide 55%三分；§7 验收全过。

**三道 C 复审 gate（§42 定）**：W3.1（融合零回归 oracle）/ W3.2（cd2b 单房零回归，floor-strip 教训守此）/ W3.4b（neighbor 接通）。

## 7. 验收点

- **零回归 oracle**：realness+neighbor 中性 → 逐 tick 等价 S/B-only（W3.1/W3.2 闸）。
- cd2b 端到端 P(SFallen) 仍涌现 fire（vs belief 单元 0.9992）。
- realness：ghost 消失不喂 SFallen；真人摔静止消失仍 fire（RV4 端到端）。
- neighbor：多房 fresh hand-off 压 phantom；无 hand-off/stale 保 lost-fall。
- **无 gate 证**：grep 无 gate-list 残余；四轴全经 filter.Step 融合。
- decide 55%三分 + Λ 不可判默认不报。

## 8. C §42 复审结论（已裁，3 gap 已补本图）

- **裁1 融合契约**：realness 折进 logPhi（①）—— §3 已锁（独立步=软 gate，违内化）。
- **裁2 copy 边界**：基本对 + **Gap1** mirror_detect/static_reflector 补入 §4（realness 几何源）。
- **裁3 复审 gate**：三道 = W3.1 / **W3.2 cd2b 单房零回归（C 加，floor-strip 守此）** / W3.4b —— §6 已标 🚩。
- **Gap2**：W3.4 多房编排展开为 W3.4a（独立成步）—— §6 已改。
- **Gap3**：`CrossedStillPeriod` 源 = still-box，载体 track_manager **拆**（连续指标层 copy 作单源 / gate-list 残余删）—— §4/§5 已补。

**图已补全 3 gap + 锁 3 裁定。下一步**：A 起 **W3.1**（纯 belief 包，扩 Step 签名接 realness 折 logPhi + neighbor 进 Predict，中性零回归 oracle），过 gate① 交 C 复审。

---

## 9. 新指令（2026-06-15 架构师拍）：Xsensor 替换 Tsensor，所有重放接入 Xsensor — A 草拟，待 C 复审

> §1–§8 = "四轴 wire 进 engine"（W3.1–W3.5，engine.Room/Unit + NV1-8 已基本完成）。本节 = 其**之上新增的一段**：把已验证的 DBN engine 包成消费 test:* 的 replay-重算 sensor，替换 Tsensor。新方向与 §0–§5 偏离处以本节为准。

### 9.0 定位澄清（消解"现状 vs roadmap"gap）
- **生产 = wisefido-sensor**（线上实跑、发真告警给护理端）。本任务**完全不碰** → §0"不碰 wisefido-sensor 生产码"边界**仍守、无反转**（我先前喊"定位反转"=夸大，收回）。
- **Xsensor / Tsensor = replay 重算道（非生产）**：喂 `test:*`（仿 `iot:*`），sensor"**以为**在跑生产"实把重放当生产**重算**。"替换 Tsensor"全程在此非生产道内换引擎。

### 9.1 终局变更：方案甲 → 方案乙（同道换实现，非边界反转）
- §0 原"验完走方案甲（注入 Tsensor 换 belief）归档"**取消**。
- 改：Xsensorv1 自 build-out 成消费 test:* 的 replay-重算 sensor，**删 Tsensor**。**DBN 作顶层唯一决策**（非旧 belief_shadow 跑在 gate-list/zoneengine 旁的 log-only 子集）；馈送层之上只有 DBN，Decision/fire = 最终输出。

### 9.2 copy 范围/源扩展（补 §4）
- **源**：`wisefido-sensor/internal/roomengine/` → **Tsensor**（已精简克隆；不够再回 wisefido-sensor）。
- **扩两摊**（§4 未列）：**consumer（test:* 前端）** + **床融合（zoneengine `bed_bayesian_scorer` + `adapter_sleepace/radar/vital` + service/`bed_presence_fusion`）**——roadmap 原 FrameInput/bed reading 手搭，现要**真 track_manager 派生** track/ghost/still-box（业务逻辑一致、比 §6 file-replay 更 production-faithful）。

### 9.3 dbn_mode 三档语义须 DBN 顶层重实现
- 旧 `belief_shadow.go`（删）持 dbn_mode（env `DBN_MODE`，可逆，正交轴=否决 firmware×DBN 自发）：**0**=firmware 直发·DBN log-only / **1**=union firmware∨DBN 自发不否决 / **2**=DBN 否决 firmware 误火；effectiveMode=min(全局,unitCap 启动 1)。
- firmware fall 解析（`alarm_event.go`）三档都留（用户拍）。union/veto 语义须在新 engine.Room/Unit **DBN 顶层重实现**（roadmap 未提的 cutover 迁移）。

### 9.4 §5 realness 译入契约已被取代
- §5 RealnessObs（bornNearEntrance/inAreaDeny/…）= realness 轴 2026-06-14 重定**前**的契约；现 engine.Room 从 census/walls(`IsReflection`) **内部涌现**（Real vs Mirror，Static 删，主职 N_r 排 ghost）。§5 历史保留，非现行接线契约。

### 9.5 偏离待裁
- **bathroom_gate**：§4 列"不copy"，feed 剥离实**留**且 `routeRoomFrame` 活跃调用（census 流量控制）。**暂留到 Stage B**（bathroom case）再裁：census 流量(留) vs 已被 DBN realness 取代的裁决(删)。
- **suite_census**：§4 "改造非copy"，实**整 copy**。单房不影响；多房 Stage B 回"改造"口径。
- **gate-list fall 参数骸骨（C §74 揪出，架构师 2026-06-15 拍"留独立 gate 到 Stage B"）**：`FallRulesParam.Lost`(lostFallParam) + `BedsideFallConfig`/`bedsideFallCfg`/`SetBedsideFallConfig` + track.go:142/204 保留字段 = 旧 gate-list Lost Fall / R4 床边晕倒参数，**触发已删**（重点1 证 DBN 顶层唯一裁决），DBN 子包不引用（grep=0）、track_manager 只存不消费。**架构师裁定：bathroom 保留独立 fall gate 可能（PR-10/11），骸骨留到 Stage B 随 bathroom_gate 决策一起裁**——非孤儿死码，scope 已锁 Stage B（触发=Stage B bathroom 设计）。**保留对照**：`FallRulesParam.Still/CellHistory` = cell.go:360/372/374 still-box 阈值馈送层，DBN realness 消费，**永久留对**（非骸骨）。

### 9.6 新阶段施工 — Stage A/B/C（分步 + 零回归闸）
1. **Stage A 单房**：consumer(test:*) + feed 派生 → `Engine.OnRoomFrame` 构造 `adapter.FrameInput` → `engine.Room.Tick`（DBN 顶层）→ fire 落 log。**零回归闸** = replay 灌 cd2b → fire 复现 belief 单元 0.9992（= §7/W3.2 闸的端到端版）。
2. **Stage B 多房**：`engine.Unit` + `device_addr→room` 路由 + suite_census 改造（unit201 hand-off）。
3. **Stage C**：删 Tsensor + `run-tsensor.sh`→`run-xsensor.sh` + doc。

### 9.7 当前状态（2026-06-15）
- 分支 `feat/xsensor-replace-tsensor`；Xsensorv1 升 module（redis/pq/zap/yaml/owl-common + replace owl-common）。
- `config` 包 copy；**roomengine feed 库剥离 green**（21 文件，`go build ./...` + `go vet ./...` 绿）：
  - 留：Engine 消费循环（test:* 三流）+ track_manager 全生命周期（kalman/census/still-box/`ProcessFrame`/`SnapshotTrackStatuses`）+ layout/grid/cell/mirror/static 几何。
  - 删：`belief_shadow`/`belief_adapter`/`ghost_adjudicator`/`fall_exempt`/`persist`/`room_svg`/`feedback`/`track_status` 投影 + track_manager 里 gate-list fall 发射（engine_z_drop/silent_leftbed/lost_fall）+ 全部 redis 推流/快照/AI 输出。
  - **seam** = `Engine.OnRoomFrame(roomID string, bases []TrackStatusBase, nowMs int64)` 回调（`engine.go:1167`，原 `beliefShadowTick` 出口）；Engine **不 import** 任何 belief 包 = DBN 解耦在上层。
- **待**：Stage A 剩 bootstrap（registerAllRooms/mapDevicesToRooms）+ OnRoomFrame 构造 FrameInput + `cmd/xsensor` main + cd2b 零回归闸。
