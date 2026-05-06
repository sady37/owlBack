---
name: wisefido-sensor 无 Wall 时用 Boundary 兜底
description: layout 没画 Wall 时 cfg.WallPolygon 用 radar.BoundaryVertices；让 InRoom 至少 == InFOV
type: feedback
originSessionId: c2d0b96a-ef02-4ee6-9673-3a25811aadfc
---
wisefido-sensor 的 [layout_parser.go](owlBack/wisefido-sensor/internal/roomengine/layout_parser.go) 在客户没画 Wall（顶点 < 2）时，`cfg.WallPolygon` fallback 到 `radarutils.BoundaryVertices(cfg.Radar)`，而不是 nil。

**Why**：之前 fallback 是 nil → engine.go 的 `if len(cfg.WallPolygon) >= 3` 跳过 StampRoomPolygon → 所有 cell `InRoom=false`。下游所有 `InRoom`-依赖逻辑全部退化失效：
- [cell_learning.go](owlBack/wisefido-sensor/internal/roomengine/cell_learning.go) 的 `if !c.InRoom || !c.InFOV` 跳过所有 cell，**完全不学习**
- [track_manager.go](owlBack/wisefido-sensor/internal/roomengine/track_manager.go) `isOutdoorButInFOV := !cell.InRoom` 永真，birth scoring 走错分支
- [room_svg.go](owlBack/wisefido-sensor/internal/roomengine/room_svg.go) 整张图染成 "Out"（深灰）

owl-common 里 `InFOV = SignalReachable = inside BoundaryVertices`（[signal.go](owlBack/owl-common/radarutils/signal.go)），所以用 BoundaryVertices 兜底等价于让 `InRoom = InFOV`，下游所有判定按"在 FOV 内 = 在屋内"工作，符合常识。

**How to apply**：

1. **不要把 WallPolygon nil 当成正常情况处理**——下游全失效。要么用户画 Wall，要么走 boundary 兜底。
2. **boundary 兜底是退而求其次**：boundary 是设备配置的探测矩形（前端 IoTSave 时写的 leftH/rightH/frontV/rearV），不一定贴合物理墙。如果客户配的 boundary 跨过实际墙（伸进隔壁房间），墙外 ghost 不会被 `!InRoom` 拦截。要彻底防 ghost 还是得画 Wall。
3. **engine 层不要再加 InRoom nil 检查**——layout_parser 保证 WallPolygon 永不为 nil（最次也是 boundary 4 顶点），下游可以放心 stamp。
4. **改 boundary 兜底逻辑时小心 layout 真值不变量**：参考 [radar_layout_device_invariant.md](radar_layout_device_invariant.md)，layout 的 boundary 必须 == device firmware；boundary 兜底链路在这个一致性保证下才有意义。
