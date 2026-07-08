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

1. **area 自查用 raw 坐标**（非 Kalman），在 **new-tid / enter2out / np>0** 触发点算，**newer-covers-老值**（按 ts 单调取最新）。area 自查是**双用途**：floor 豁免 + 喂 ghost 判定（新轨落声明床 + sleepad 占用 → 大概率真睡者，不该判 phantom）。
2. **Kalman 保留、继续每 tick 跑**（velocity 硬依赖它，且仅几百 flop）。但**位置消费者除速度相关外一律改读 raw**；**4 个 ghost 检测器优先**。

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

**wiring**：`base.CellAreaType = ts.AreaEff`（替 [track_manager.go:1108] 现在的每帧 QueryAreaType(Kalman位)）。下游 floor 豁免 / emission / ghost 都吃这个单源。

**开放点**（实现时定）：
- 触发点是否补 **InBed 事件** 和 **stillbox-start**？（floor 关键是 coast 起点 area；但在床者几乎总以 new-tid 重生，new-tid 可能已覆盖——先不加，replay 看够不够。）
- `np>0` 遍历成本可忽略（np 变化频率低）；只在**同 radar 系**内，用 raw。

---

## Phase 2 — ghost 检测器 Kalman→raw（治 nr=0 正面战场）

逐个把 `ts.Kalman.Position()` 换成 `ts.History[len-1]`（raw 末点）：

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
2. **Phase 1**（area 自查）→ replay 看 bed 豁免 / floor。
3. Phase 3 清理。
4. 每步 `go vet && go build`；Xsensor 验证过再镜像 wisefido-sensor。

## 相关 memory

[[unified_ghost_score_lid_lifecycle_fp_root]] [[cd2b_present_coast_evict_purge]] [[firmware_areaid_enter_event_latch_geom_fallback]] [[track_identity_logicid_ghost_track2_scope]] [[boundary_truncation_box_replaces_silent_drop]]
