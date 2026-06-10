---
name: drs-retired-status-redis-only
description: "2026-05-21 device_runtime_state 表整张退役 — firmware_version 进 dfm；online/heartbeat/rssi/battery/sensor_detached 只在 Redis device:status:{ipv6}"
metadata: 
  node_type: memory
  type: project
  originSessionId: 6596aa1c-6420-4198-9805-2c0620c5d97c
---

`device_runtime_state` (drs) 表彻底 DROP。原 9 列去向：

| 原 drs 列 | 新归宿 |
|---|---|
| `firmware_version` | `device_factory_meta.firmware_version` (持久，OTA 写) |
| `online`, `last_heartbeat_at` | Redis `device:status:{ipv6}` (`offline`, `last_seen_ms`) |
| `rssi`, `current_angle_deg`, `battery_pct`, `sensor_detached` | Redis `device:status:{ipv6}` (`signal_poor`, `angle_abnormal`, `sensor_detached`) |
| `last_position_addr`, `last_error`, `updated_at` | 删除 |

**Why:** drs 是误放表 — runtime 状态本该全在内存/Redis；drs.online 历史上从来没人 UPDATE 成 TRUE，导致 FE 永远显 offline。dfm 跟 devices 都是持久层；status 是事件驱动的瞬态。原 1:1 关系下 drs 整张多余。

**How to apply:**
- 任何代码要读 device 状态 (online/offline/signal_poor/angle_abnormal/sensor_detached/last_seen) → 用 `card.Reader.ReadDeviceStatusByDeviceAddrs` 或 helper `FillDeviceStatusFromCardagg(ctx, stateReader, deviceIPv6s, logger)`（在 wisefido-data service 层）
- 任何代码要读 firmware_version → `device_factory_meta.firmware_version`
- **禁止**新加 `JOIN device_runtime_state`、`drs.*` SELECT、`UPDATE drs SET online`、`INSERT INTO drs` 等 —— 该表 schema 已不存在
- `card_static_service.fillDevicesV3` 是参考模板：SQL 拉设备列表带 status='offline' 占位，然后 Redis fill-pass 覆盖
- qinglan `deviceSelectV2` 同理：status SQL 占位 'offline'，service 层用 Redis 覆盖（目前 qinglan service 层未实现覆盖，按需补）

**已知未实现的 status 路径**（drs 删后 SQL 退化）：
- `wisefido-data/postgres_devices.go ListDevices` —— online/offline filter + sort_by_online_status 都已删，FE 想按 status 过滤需 client-side 或新接口
- `wisefido-qinglan postgres_device.go SearchDevices` status 条件 — 同上
- `qinglan CountDevicesByStatus` — 退化为返全 offline，等 caller 出现再用 Redis SCAN 实现

详 [[device_status_event_driven_refactor]]（事件驱动设计源）+ [[feedback_schema_review_via_dbv2.md]]（schema 流程教训：本次 drs 退役 ALTER + DROP 先于 dbv2 CREATE 文件同步，事后才补 21_*.sql/删 22_*.sql，下次 schema 改动顺序要反过来）
