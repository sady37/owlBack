# #3 Neighbor wire —— 生产-gate 验证执行 spec（待 unit201 数据 turnkey）

状态：shadow-first 已落地并验收（feedback 审查63）+ N-7 撤回（审查64）。本 spec 为**生产-gate 前置**的验证执行手册——**当前 blocked on test data**（无 unit201 整单元 redis 数据），数据一到即可逐项跑。doc-first，未实现 harness 扩展。

---

## 0. 前置发现（harness gap）→ ✅ 已修复（bReplayUnit 落地）

现有 replay 载体 B `bReplay(t, dir, file)`（belief_b_replay_test.go:99）**只 RegisterRoom 单房**（:106 `e.RegisterRoom(cfg)`，单 roomID + 单 radar + 单 sleepad）。⟹ `neighborHandoff` 在 `e.rooms` 里**找不到兄弟房** → ObsNeighbor 在单房 replay 下**永不 fire**。`allObsKinds` 审计（:214 含 "Neighbor"）在单房 replay 必判 Neighbor「未 populate」——这是**单房结构使然，非 wire 死管**。

**故 Neighbor 的真验必须先扩 harness 为整单元（multi-room，同 suite）**——这正是委员会反复说的「redis-replay 须整单元重放（非单设备）」。

**✅ 已落地** `bReplayUnit`（belief_neighbor_replay_test.go，审查65 sequencing 建议 → 用户拍 B → 委员会执行）：多房注册同 suite + census seed + per-device 路由 + 按 ts 全局合并喂**生产** handleMessage/handleEventMessage（保真硬条件守）。本房 lost-track 经真 beliefShadowTick 涌现（合成 radar 帧：tid=1 走动后丢轨 + frame-keeper tick 触发 lost-sweep），邻房 hand-off 经真 handleEventMessage 落 lastEnterMs。**V1-V4 逻辑保真已绿**（见 §1）。recall（§2）仍 blocked on 真数据。

### harness 扩展设计（`bReplayUnit`，✅ 已按此实现）
保真硬条件不变（只喂 raw record 进生产 `handleMessage`/`handleEventMessage`，禁手搓 bases/禁 fork）。在 `bReplay` 基础上：
1. **多房注册**：每个 device 的房各 `ParseLayoutConfig` + `e.RegisterRoom(cfg)`；`e.roomSuiteID[roomID]=suiteID`（同一 unit 的所有房 → 同 suiteID）。
2. **per-device 路由**：`e.deviceRoom[addr]=roomID`、`e.deviceMounts[addr]=cfg.Radar` 按设备各设；radar/sleepad UID→addr 映射同 bReplay（靠 position_x 区分 radar vs sleepad.track）。
3. **census 注入**：`e.suiteCensus` seed 该 unit 的 resident/visitor（从 unit201 绑定关系或测试指定）；Neighbor 的 N-3 门读 `SoleResidentRecaptureState(suiteID)`。
4. **合并喂入**：所有 device 的 record 按 `timestamp` 全局升序合并后逐条喂（跨房时序真实交织——hand-off 的「先走后到」靠真实 ts 涌现）。
5. **采日志**：observer 捕全房 `belief_shadow_*`，按 roomID 分组。

unit201 = 3 设备（CD2B radar + sleepad1641 + 333B radar）；CD2B/333B 各自房 + sleepad 归其床房；suiteID = unit201 的 /80。

---

## 1. 要验的项（V1-V4 ✅ 合成 fixture 已绿；recall 待真数据）

逻辑/接线保真用合成 multi-room fixture 经真路径已验（`TestNeighborReplayUnit_V1..V4`）；recall（真摔召回，§2）合成证不了，仍 blocked。

| # | 项 | 判据 | 日志/断言 | 状态 |
|---|---|---|---|---|
| V1 | Neighbor populated（活管） | 某房 lost-track 期间，兄弟房窗内 fresh 有向 hand-off → ObsNeighbor fire | `belief_shadow_neighbor_handoff` 恰 1 次（n3_lost_seen_ms/conf/src 对） | ✅ V1 |
| V2 | stale_corr no-silent-caps | 兄弟房 durable 占用但相关度 stale → 不压 + LOG 不静默 | handoff 0 次 + `belief_shadow_neighbor_stale_corr` 1 次 | ✅ V2 |
| V3 | **N-6 单合并** | room∧bed 同窗命中 → OR-合并**单条**（conf=max），非双似然 | (a) bed-only→src=bed 证注册；(b) room+bed→handoff 恰 1 次 src=room | ✅ V3 |
| V4 | R5 多-resident gate-OFF | census >1 resident → lost-track + 兄弟房占用 → **不** fire | 多-resident 无 handoff、无 stale_corr | ✅ V4 |
| V5 | 回归绿 | build/vet/belief 绿 + R5-lock 绿 + roomengine 9 红 0 新增 | CI | ✅ |

**合成保真边界（诚实标注）**：N-6「单合并」的**权威**保证是结构性的——`neighborHandoff` 返单条 + 消费侧 append 单条（构造即不可能喂 2），V3 在真路径上佐证「双占用→单次 fire」。`damp==0.7 非 0.49` 的 belief-trace 量化检留待真数据（合成 conf 固定，量化意义有限）。

## 2. DBN recall 真摔集（铁律：recall 从未验证）

聚焦验证集（审查㊼用户定）= **2 精度(FP) + 3 recall(真摔，含 2 firmware 漏报)**。对每真摔案：
- **核 Neighbor 不误压真摔**：真摔案里若本房 lost-track 同期兄弟房有占用，确认 ObsNeighbor **未把该真摔压成 Vacant/Left**（`belief_shadow_trace` 末态 `p_fallen` 仍主导 / `belief_shadow_fall` 仍 log）。
- **★ N-7 visitor 噪声扣偏（审查64 降级 caveat）**：兄弟房的 fresh hand-off 若来自 **visitor**（census visitor 候选）而非 resident，则该次 ObsNeighbor 命中是**概率噪声**，不是真消歧——recall 统计时**标注剔除**，别把噪声当系统偏（审查64：visitor 巧合是 measurement noise 非 defect）。判别：hand-off 命中时刻兄弟房 enter 的归属（若 census 该时段 unit 有 visitor 且 resident 在本房真摔 → 标 noise）。
- **不改设计**：N-7 不要求 headcount 硬门（净负，废掉有 caregiver 的单元）；若数据显示 visitor 噪声显著，可选 plan-B 软 down-weight（headcount>1 降 conf 非硬 off）。

## 3. 放行 bar（生产-gate）
V1–V5 全过 + recall 真摔集 0 误压（扣 N-7 噪声后）+ N-6 终验无过压 + 铁律 R0/R1/R5/R7 守。通过后方可议 shadow→production cutover（ObsNeighbor 进 firmware∧shadow 漏报-safe gate）。

## 4. 当前 blocked / 待用户
- **V1-V5 ✅ 已绿**（`bReplayUnit` + 合成 multi-room fixture，逻辑/接线保真）。harness 已落地，不再 block on 数据。
- **唯一仍 blocked = recall 真摔集（§2）**：需 **unit201 整单元 redis 真数据**（CD2B+1641+333B 的 monitor/event 流窗口导出）——谁跑导出、哪个时间窗（需含至少一段「人从一房走到邻房 + 另房 lost-track」+ 一段**真摔**对照）。合成数据证不了召回率。
- 真数据到 → 复用 `bReplayUnit`（喂真 record 替合成）逐跑 recall + N-6 的 damp 量化（§1 边界）。
