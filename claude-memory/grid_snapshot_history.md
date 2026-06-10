---
name: Daily Grid Snapshot History (PR1 完成)
description: 每天 11:50 归档 grid 学习状态到 roomengine_grid_snapshot_history 表，保留 365 天，为 playback 历史起点服务
type: project
originSessionId: 3d9a6ccb-8326-4fc0-96ba-8ec961837774
---
PR1 已实施 (2026-05-06)：每天 11:50 local 把 wisefido-sensor 各 room 的 RoomGrid 状态归档到 `roomengine_grid_snapshot_history`（PK=(room_id, snapshot_date)），保留 365 天滚动清理。区别于 36_*.sql 的 live snapshot（5min 全量 dump，仅最新一份）。冒烟过：UPSERT + retain prune 都正确。

**Why:** playback 当前用 `rooms.layout_config`（人工 baseline）从空白 grid 起跑跑 [start, end] 的 track 重新学习——不等于生产环境真实累积学习的 grid。客户演示需要"从某天 → 现在"的真实演化展示。每天 ~21 KB JSONB（实测），24 间房 365 天 ≈ 180 MB，1000 间房 365 天 ≈ 7.5 GB，单机 PG 都行。

**How to apply:** PR2 待做：playback `?from=YYYY-MM-DD` 参数 → `loadLayout` 检测到 from 后从 history 表拉对应日期 snapshot → DecodeSnapshot 灌回 grid Belief+counter → SVG 起点 = 那天的状态。一致性检查：layout_hash 不匹配（中途人工编辑 layout）必须 fallback baseline 或报错。playback 重演会再跑 decay，所以"近似真实生产路径"但不严格相等，演示话术要带这个 caveat。代码入口：`saveAllRoomsHistory()` in engine.go，`HistoryPersister` interface in persist_postgres.go。
