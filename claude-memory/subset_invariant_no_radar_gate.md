---
name: subset-invariant-no-radar-gate
description: 2026-05-28 修复 /88 仅 sleepace 无 radar 时 subset_invariant lift_parent 卡死 InRoom；engine SetRadarPresenceLookup 注入 + drop_no_radar phase + startup force-vacant cleanup
metadata: 
  node_type: memory
  type: project
  originSessionId: bf92cf95-1474-412f-b6f6-532a7f36dc0d
---

card f61e8o 案例：/88 200 只有 Sleepace BM87224601903、零 radar，却显示 "InRoom 59h 5m"。

**死循环**：sleepace InBed → zoneengine `maybeReconcileSubset` lift_parent 抬 /88 200 → Room
合成 ZoneEvent → `lastRoomState[/88 200]` 缓存 → 60s `tickRepublishStates` 重发 → `RadarRoomCountCache.GetZ`
永远 stale（无 radar 喂）→ Z fallback 复读 rs.TotalPeople 自身（8）→ cardagg
`mergeBedField` 见 prev==in 保 prev_ts → state 永远 InRoom + 锚点冻在最早 0→8 那刻。

**Why**：subset_invariant lift_parent 默认信 bed（bed.IsPresent → 抬 room），但 /88 无 radar
时没有任何 Vacant 反向证据通道——radar Z 永 stale、ExtraPeopleInRoom 只加不减、cardagg
N→0 transition 永远不发生 → display.scene_state=1 (InRoom) 钉死。

**How to apply**：任何新增"信 bed 抬 room"/"信 X 抬 Y"路径，先问"反向降回 Vacant 的证据通道
是谁？"——若反向通道依赖某类设备而该 zone 可能不装该设备，必须 gate 或加 force-drop。

实现（[[radar_layout_device_invariant]] / [[v2_cutover_rules]] 同代码债治理风格）：
- `wisefido-sensor/internal/zoneengine/wiring/spatial_cache.go::HasRadarInRoom(/88)` —
  扫 ud.Devices 看 access=true 的 Radar；只看 access 不卡 monitoring（物理存在即视为有通路）
- `wisefido-sensor/internal/zoneengine/engine.go::SetRadarPresenceLookup` 注入；
  `maybeReconcileSubset` + `repairSubsetInvariant` lift 分支 gate 之
- `engine.go::dropOrphanRoomsLocked` repairSubsetInvariant 阶段 3：扫 IsPresent room，
  /88 无 radar 且名下无活体 bed → 强制 TransitionVacant（修历史遗留卡死）
- `cmd/wisefido-sensor/initial_publish.go` startup cleanup：no-radar /88 publish 用真 ts +
  LastExitTs=now（cardagg mergeBedField 触发 N→0 transition）；有 radar /88 仍 ts=0
  避免假活跃（[[startup_clear_and_initial_publish]] 原则保留）

实证：sensor restart 后 5 个 no-radar /88 自动 force-vacant（含 f61e8o）；display
scene_state 1→4（OOB），anchor 用 bed_status_ts 真实离床时间；60s tick 不再合成
/88 200 Room state（engine.states 空）。

**2026-05-28 续 R1 闭环**：`HasRadarInRoom` 升级，DB 有 radar 但 fitness IsFit=false（全 unfit）
也视作"无 radar"。机制：`spatial_cache.go::SetFitness` 注入 DeviceFitnessTracker；wiring/setup.go
启动期 wire `opts.Fitness`。radar 翻 unfit → ≤10s 下次 engine.Tick → drop_no_radar 同路径
触发 → emit TransitionVacant → cardagg 自然清。多 radar /88 任一 fit 即视为有观测（last
writer wins，RadarA 的 per-device 历史由 main.go [[device_status_event_driven_refactor]]
现有 OnDeviceUnfit 回调清）。fitness 未注入时退化 DB-only，启动期兼容。

5 个新 test：`TestEngine_SubsetInvariant_NoRadar_SkipsLift` /
`TestEngine_SubsetInvariant_WithRadar_LiftsAsBefore` /
`TestEngine_RepairSubsetInvariant_NoRadar_SkipsLift` /
`TestEngine_DropNoRadar_DropsOrphanRoom` /
`TestEngine_DropNoRadar_KeepsRoomWithActiveBed`。
