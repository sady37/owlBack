# area 自查 + ghost 检测器改 raw —— 实现 spec

> 本文是实现工单。先在 **Xsensorv1** 改+replay 验证，再逐字镜像 wisefido-sensor（[[sensor_xsensor_one_to_one_env_differs]]）。
> 验证用 case：`doc/cases/case-cd2b-0707-02280243/`（含 xsensor.log；生产真火 02:42:04，stillbox 起 02:30:03）。

## 背景 / 根因（cd2b-0707 复盘定论）

真人在床睡觉，被 floor 兜底误火。根因链：

1. **radar 抓不住静止睡者** → 反复丢-重抓（01:00~02:30 np 在 0↔1 横跳 6 次，多数"空"无 ExitRoom = 人没走是 radar 丢）。
2. 每次重抓是条**没有 EnterRoom 门第的 new tid** → census 判 ghost/反射 → 生产 `nr=0`（两轨都非真人）。
3. 现有 ghost 检测器（mirror/static-reflector/teleport/split）读的是 **Kalman 平滑位**，把"瞬移跳变/反射几何/出生即不动"这些**它们要抓的高频签名抹平了** → 误判助推。
4. `obs.AreaType` 靠固件 InBed/area_id 事件定，而 ghost 轨**无此事件**（area_id=255）→ area 落默认 → floor 无床豁免。
5. sleepad 连续读数在生产因 `vital_confidence=0` 落 `bed_reading=NoReport` → coast 期无持续床占用 → 12min floor 误火。

## 锁定的决策

1. **area 自查用 raw 坐标**（非 Kalman），在 **new-tid / enter2out / np>0** 触发点算，**newer-covers-老值**（按 ts 单调取最新）。**Option A（2026-07-07 定）**：`AreaEff` 是 **latch 私有旁路**——只喂 Phase 1.5 的 provenance/confine 判定，**不替 floor 主路的 `CellAreaType`**。
   - **floor 主路 CellAreaType drift 已收敛（7-7）**：曾 Xsensorv1=每帧 QueryAreaType / sensor=StillBoxCellArea 锁定。已把 **Xsensorv1 收敛到 sensor**：box 起点锁定 `StillBoxCellArea`（+chair 锁）、still→锁定值 / move→AreaUnknown（躲逐帧跨格抖动误报）/ lost coast→raw 末点重算；三分支两引擎逐字一致，cd2b replay fire=0 无回退。`AreaEff`（latch）与 CellAreaType（floor）正交——latch 只压 ghost 分类不压 floor。
2. **Kalman 保留、继续每 tick 跑**（velocity 硬依赖它，且仅几百 flop）。但**位置消费者除速度相关外一律改读 raw**；**4 个 ghost 检测器优先**。
3. **real-by-provenance latch（挂 lid、单调、永久 real）**：出生判定有 2 个评估时刻（出生瞬时门距 + census 5s 窗），**任一时刻**满足 provenance → born ghost score 归 0、**以后一律不再做 ghost 判断**；唯一解闩 = swap / split-group 重分配。provenance = 门第（EnterRoom 或 raw 自查落 AreaEnter）**或** 床（raw 自查落声明床 ∧ sleepad 占用）。详见 Phase 1.5。

## Kalman 必须集（唯一硬依赖 = 速度）

- `Kalman.Velocity()` → [track_manager.go:1581] `MarkOccupancy(vx,vy)` + 慢走老人 Speed>20cm/s 兜底判 Move；[:1734] card VX/VY。**保留。**
- 其余 `Kalman.Position()` 消费者全部候选改 raw（见 Phase 2/3）。

**raw 源** = `ts.History[len-1].{X,Y}`（末次真观测的 canvas 点，与 Kalman.Position() 同坐标系，drop-in）。丢轨期 History 冻结不变 = 正是 ghost/反射要的"最后真位"。

---

## Phase 1 — area 自查（raw + 触发 + newer-covers）

**目标**：`obs.AreaType` 不再等固件事件，改由"raw 坐标自查 + 事件"按 ts 取最新。

**TrackState 新增**：`AreaEff AreaType` + `AreaEffTs int64`（有效区型及其时间戳，单源）。

**触发点**（每处：对该 radar 下**所有 tid**用各自 `History末点` 自查 `QueryAreaType`，`ts=now`）：
1. **new tid**（出生钩子）— 核心，catch 无门第 ghost。
2. **enter2out 事件**（EnterRoom/ExitRoom，RecordRadarEvent）— 事件带 areaID，也一并纳入 merge。
3. **np>0 事件**（NumberPeople，np>0）— np 无坐标 → 遍历当前各 tid 的 raw 自查。

**merge（newer-covers-老值）**：`AreaEff` 只被 ts 更新的值覆盖。
- 固件事件的 area（enter2out / InBed 的 area_id）带事件 ts；自查带 now。
- 谁 ts 新用谁。语义：自查早于事件 → 后到的事件覆盖；自查晚于事件 → 前事件是瞬态，用当前自查。

**wiring（Option A）**：`AreaEff` **不进** floor 主路（`CellAreaType` 保持各引擎原样）；仅 `evalProvenance`（provenance 判定）与 `realLatchActive`（confine）读它。下游 floor/emission 的 area 源不变。

**开放点**（实现时定）：
- 触发点是否补 **InBed 事件** 和 **stillbox-start**？（floor 关键是 coast 起点 area；但在床者几乎总以 new-tid 重生，new-tid 可能已覆盖——先不加，replay 看够不够。）
- `np>0` 遍历成本可忽略（np 变化频率低）；只在**同 radar 系**内，用 raw。

---

## Phase 1.5 — real-by-provenance latch（治"重生睡者被全判 ghost"根因）

**目标**：一条轨一旦有 provenance 证据（进门 / 落声明床占用），**永久判 real、彻底退出 ghost 判定**，且这份记忆**跨 new-tid churn 持续**——这是 cd2b 根因（真睡者反复丢-重抓、每条重生轨被判 ghost → nr=0）的正解。

### 现状（本节是升级，不是新建）

已有 `EnterBorn`（[track.go:30]）：出生有 EnterRoom 配对 → `MaxGhost` 钳 ≤0.5，`LidRebound`(=swap) 解禁；`confirmedMirrorResidual`([track_manager.go:2125]) 已用它护进门真人。本节把它**广义化 + 硬化**，按规则 1.3 单源——**复用/改造 `EnterBorn`，禁止新开并行 latch 字段**。

### 2 个 born-ghost 评估时刻（"任一命中即 latch"）

| # | 位置 | 现在算什么 |
|---|---|---|
| ① 出生瞬时 | [track_manager.go:2086] `③ born-ghost 门距` | `85*d/150 > 65` ⟺ 出生 >~115cm 离门 |
| ② 出生 5s 窗 | [census.go:214] `birth_verdict` | 窗末冻结 `p_real<0.5 → ghost` |

**OR 语义**：①②**任一时刻**满足下面 provenance → 立即 latch，另一时刻不必再看。

### provenance 入口（两条通道，OR）

1. **门第**：`EnterRoom` 事件 **或** raw 自查 `AreaEff==AreaEnter`。
   - `AreaEff` 来自 Phase 1（这也是 Phase 1.5 硬依赖 Phase 1 先落地的原因；Phase 1 之前只有固件 EnterRoom 事件，拿不到"落 Enter 区"）。
2. **床**：raw 自查 `AreaEff∈{AreaBed, AreaMonitorBed}`（有 sleepad 的真床，非沙发 AreaLying）**∧** 同房 sleepad 占用 present。
   - **缺 2 则 cd2b 不治**：重生睡者无 EnterRoom、不在 Enter 区（area_id=255，落床），门第通道对他永不触发。床通道是他唯一的 provenance。
   - **sleepad flaky 张力（replay 必查）**：cd2b 根因 #5 = coast 期 `vital_confidence=0 → bed_reading=NoReport`，sleepad 占用会掉。latch 是 lid 单调，只需在**任一**重生时刻（~6 次）逮到一次"落床 ∧ sleepad 占用"即永久置位。replay 要确认这 6 次里**至少一次** sleepad 仍报占用；若全 NoReport，则 sleepad AND-gate 太严，需退到"近期占用窗"或另议（不在本次改，先由 replay 证伪）。

### latch 语义（挂 lid、单调、硬化）

- **载体 = `LogicID`（lid），不是 tid**。tid 每次重抓换新，挂 tid = 每次 new-tid 清零 = 等于无 latch。census 已按透传 lid 做键（[[logicid_unified_census_refs_tm]]），latch 随 lid 走才带得过 churn（出生继承 `parent.RealProven`）。
- **`RealProven` 置位 = 单调**：出生/present 帧任一时刻 provenance 满足即置真，一旦置真只有 split 重分配能清（见解闩）。
- **force-real 生效 = 条件（confine）**：消费端一律读 `realLatchActive() = RealProven ∧ AreaEff∈rest 区`（Option B）。active 时 → census `t.rt.ForceReal()`（保 nr≥1）+ SetTrackPReal 强制 PReal=1 且不累积 MaxGhost + 出生软压 stamp 跳过 + residual 驱逐豁免。不 active（飘出 rest 区）→ ghost 检测器照常跑。
- **`EnterBorn`/`LidRebound` 退役**：`EnterBorn` 广义化并改名 `RealProven`；`LidRebound`（每次 churn 置位的软钳解禁）删除，其职责由 confine 取代。

### 解闩策略 = Option B「继承保留 + 空间 confine」（架构师 2026-07-07 定；唯一 FP 安全阀 —— 命门）

**方向辨明**：ghost 判定压的是 FP。把一条轨 latch 成永久 real = 它永不被 ghost 压 → floor/fall 一定放行。

- latch 真人 → 治 FN（睡者不再被全判 ghost）。✅
- ghost **冒用了 real lid** → 带 real-latch 永不被压 → **floor FP 直接放行**。🔴

**发现的冲突（原"swap 解闩"行不通）**：现有 `LidRebound` 在**每次无门第 `nearestAliveTrack` 继承**就置位（[track_manager.go:1514]），而 cd2b 真睡者重抓走的正是这条路径 → 若 `LidRebound` 当解闩，latch 每次 churn 被清 → cd2b 不治。根子=`LidRebound` 把"延续churn(同一真人)"与"真冒用"混成一个信号。

**Option B 定稿**：
1. **latch 挂 lid、跨 churn 继承**（出生 `nearestAliveTrack` 继承 `parent.RealProven`，如现 EnterBorn 继承）。**删 `LidRebound`**——不再靠它解闩。
2. **confine 应用（读时判，非清 latch）**：latch 的 force-real 只在**当前 `AreaEff ∈ rest 区**时生效。`realLatchActive() = RealProven ∧ AreaEff∈{AreaBed,AreaMonitorBed,AreaSit,AreaLying,AreaEnter}`。
   - 真睡者赖床上 → AreaEff=Bed → latch active → 永不判 ghost + 床 floor 豁免 → **cd2b 治**。
   - 冒名 phantom 飘到开阔区 → AreaEff=Active/Deny → latch **不** active → ghost 检测器照常跑 → **FP 口堵**。
   - 真人真摔在开阔地板 → AreaEff=Active → latch 不 active → floor 照常 fire（**要的**，latch 只压 ghost 分类不压真摔）。
3. **硬清 latch**：split-group 重分配（组内 lid 重新分派成员）时 `RealProven=false`。

**关键辨析**：latch 的作用是**反 ghost 分类**（保 nr≥1、不被当 phantom purge），**不是压 floor**。压 cd2b 睡者 floor 的是**床区 floor 豁免**（Phase 1 把 AreaEff 定成 Bed）。两者正交、协同治 cd2b。

### 验证判据（规则 #3，看机制非 fire）

cd2b-0707 replay 后：重生睡者的 lid 是否在①或②命中床 provenance 而 latch → 后续重生轨 census **不再判 ghost**（`nr≥1`）；swap/split 发生时 latch 是否正确清除（不留冒用口子）。

### ✅ 2026-07-07 replay 实证（Xsensorv1，case-cd2b-0707-02280243 @8x）

- **睡者 `CD2B02800853`**：02:29:59 经**床通道**（AreaEff∈{Bed,MonBed} ∧ sleepad InBed）latch，`p_real` 1.00 贯穿；census `ForceReal` 粘住 → 全程 real，非 ghost。→ 治根因 #2（原 census 判 ghost/nr=0）。
- **床通道命中 = AreaEff=Bed**：反证 Phase 1 自查把重生轨落到了声明床（治根因 #4，floor 拿床豁免）。
- **sleepad flaky 张力已解**：`bed_reading=InBed` 全程（02:30~02:43）held，即便 `vital` 02:36 掉 False；`sleepadOccupied()` 取 `InBed` 非 vital → 02:42:40 fresh 重生 `CD2B04240772` **立即再 latch**（rp_true=80/80）。→ 治根因 #5，且确认"用 InBed 不用 vital"是对的。
- **confine/FP**：全房全轨 `fire=0`（原生产 02:42:04 floor 误火消除）；两雷达兄弟 333b 无误火。
- 注：`real_proven`（xray，来自 GhostSignals=tm 侧裸 RealProven）在 tm 轨 coast 驱逐后读 false 属正常（GhostSignals found=false），census 侧 `p_real` 仍 1.00 粘住，占用不丢。

---

## Phase 2 — ghost 检测器 Kalman→raw（治 nr=0 正面战场）

逐个把 `ts.Kalman.Position()` 换成 `ts.History[len-1]`（raw 末点）。**前置**：已 latch(Phase 1.5) 的 lid 在入口 short-circuit，下列检测器对它一律跳过。

| 文件:行 | 检测器 | 换 raw 的收益 | 风险/注意 |
|---|---|---|---|
| mirror_detect.go:294 | 镜像反射候选 | raw 保真反射几何 | 抖动 → 反射轴判定要看是否需去抖窗 |
| static_reflector.go:50,140 | 静态反射（出生即在此+从不移走）| 静态反射 raw 位恒定，Kalman 无增益 | BirthPos 对比要同用 raw |
| teleport_interference.go:150 | 瞬移伪迹 | **Kalman 平滑藏瞬移**，raw 才见跳变 | 收益最大 |
| split_ghost.go:54(kinematicGlued),106 | 离锚点净距 | raw 是真位移 | 三档阈（GLUED50/HOLD/WALKOUT）标定基于哪种位？换 raw 后需 replay 复核 |

**验证判据**（规则 #3，看机制非 fire）：cd2b-0707 replay 后，**`nr` 是否≥1**（睡者不再被全判 ghost）、new-tid 重生轨的 `p_real` / `is_refl` 走向、以及 census `birth_verdict`。

## Phase 3 — 其余位置消费者（可选清理）

- witnessNearby(2173,2194) 半径检查 → raw（容差大，低风险）。
- 关联 460（跨 gap，只对丢轨）→ raw-last（损失小；预测仅对移动断档略有用）。
- 输出快照 1733 位置 → raw（UI 平滑仅美观，非必须；VX/VY 仍 Kalman）。

---

## 执行顺序建议

1. **Phase 2 先做**（teleport + static_reflector 风险最低，直接见 nr 变化）→ replay 看 nr 是否回正。
2. **Phase 1**（area 自查）→ replay 看 bed 豁免 / floor。**Phase 1.5 硬依赖 Phase 1**（要 `AreaEff`），必须排在 Phase 1 之后。
3. **Phase 1.5**（real-by-provenance latch）→ replay 看重生睡者是否 latch、census 是否不再判 ghost、swap/split 解闩是否干净。
4. Phase 3 清理。
5. 每步 `go vet && go build`；Xsensor 验证过再镜像 wisefido-sensor。

## 相关 memory

[[unified_ghost_score_lid_lifecycle_fp_root]] [[cd2b_present_coast_evict_purge]] [[firmware_areaid_enter_event_latch_geom_fallback]] [[track_identity_logicid_ghost_track2_scope]] [[boundary_truncation_box_replaces_silent_drop]]
