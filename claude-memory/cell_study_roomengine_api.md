---
name: cell-study-roomengine-api
description: cell study(RoomLearn)回放工具 = roomengine-api dev server；v1→v2 cutover 丢了已恢复+迁 owl_v2 monitor_stream/event_log
metadata: 
  node_type: memory
  type: project
  originSessionId: bfa121c6-2088-4ca1-a827-9ad91206bc28
---

2026-06-01：前端 WaveMonitor "RoomLearn" 模式（用户口中的 "cell study 模块"，生成 cell 学习过程 SVG/HTML）的后端 = `wisefido-sensor/cmd/roomengine-api`（dev HTTP server :7788，前缀 `/roomengine/*` 避开 owlBack `/api/*` auth；vite proxy 在 owlFront/vite.config.ts:122）。前端 [WaveMonitor.vue:799-842] Generate→`/roomengine/playback?...&format=json`、files 列表、Play 开 `/roomengine/files/<name>` 新 tab。渲染库 `internal/playback`(Run/WriteHTML) + `internal/roomengine.BuildRoomSVG`（cell 着色 aiDeny/AreaSit）。

**丢失根因**：sensor v1→v2 cutover，commit `b77f7c2`(sensor-v1 整目录退役)删了 `cmd/roomengine-api` + `cmd/roomengine-playback` + start/stop 脚本；`internal/playback` 搬进 v2 但成孤儿。已从 `b77f7c2~1` 恢复两个 cmd + 脚本。

**铁律：playback 是 /128 设备级，layout 不走房级合并**（2026-06-01 修，commit ffee4ca）。layout 按 /128 device 编辑、playback 按单 device 回放(QueryMonitorTracks 只取该 device_addr 的 track)，所以 `LoadDeviceRoomConfig` 只读该 device 自己的 /128 canvas + `ParseLayoutConfig` 单块解析。**不要**改回复用 live engine 的房级合并(`LoadRoomCanvases`+`BuildRoomConfigFromCanvases`)——那是给"一房一引擎吃所有雷达"的房级引擎用的；多雷达房各 device-local 帧未配准，合并会把同房其他雷达的家具/床错位叠进 SVG(实证：E598A2ACD523 删了自己那台的 Bed，但合并把第二台 3263:9e7 的 Normal Bed 当蓝色 Lying 区叠回来)。

**v2 数据迁移已做**：cell-study 读时序数据。v1 在 owlrd 单表 `iot_timeseries`(device_uid/topic_type/category/data_value/timestamp ms)；v2 在 **owl_v2 拆两表**：`monitor_stream`(61_，device_addr INET，stream_type='radar.track'，payload，ts timestamptz)+`event_log`(62_，event_kind)。payload 数组形状与 v1 data_value 完全一致(position_x/y/z 无后缀)→ ParseRadarTracks/ParseRadarTrackEvents 不动。改了 `internal/playback/db.go`：OpenDB 默认 owl_v2；LookupDeviceAddr(device_uid→host(device_addr))；LoadDeviceRoomConfig(复用 roomengine.LoadRoomCanvases+BuildRoomConfigFromCanvases 读 room_visual_layout /128 canvas 按 /88 room 合并，**不再** rooms.layout_config)；QueryMonitorTracks/QueryEvents 按 device_addr 查两表。删了 LookupTenantID/LookupRoomLayout/Options.TenantID/Row.ID。实测 9923003AB197(LivingRoom)出 25 帧 6.5M HTML，引擎跑出 area_sit_auto_learned_region/ghost/real_fall。

**feedback_ingest.go 也已迁 v2**(PR-9.1 历史人工反馈 overlay)：alarm_events 列改 `notes→handler_notes`、`trigger_data->>'event_payload'→payload`、select `host(device_addr)`；lookbackPositionFromIoT 改读 `monitor_stream`(device_addr ::inet + stream_type='radar.track' + ts timestamptz，弃 `iot_timeseries ::uuid`)。live 实测 9923003AB197 success=1/total=3(那条 ☑Verbal/☑Bleeding verified→MarkRealFall)。CLI `--feedback` 默认恢复 true；api 路径仍不开 overlay(与 v1 parity)。LookupHistorySnapshot(FromDate)也迁了：v2 表列 `room_id→spatial_prefix inet`、`snapshot_date→archive_date`。

整个 internal/playback 已无 v1 schema 残留(grep iot_timeseries/topic_type/data_value/::uuid/trigger_data 全空)。代码全 uncommitted（owlBack 非 git 根，wisefido-sensor 是 git repo）。详见 [[radar_measure_flow_design]] [[grid_snapshot_history]]。
