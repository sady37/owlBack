---
name: device_config_vs_spatial_config_split
description: 双表分工 — 跟位置走的配置进 spatial_config(prefix key+继承)，跟设备走的进 device_config(uid key)
metadata: 
  node_type: memory
  type: project
  originSessionId: 0aded70b-2158-4b22-9fe3-ae67182b384e
---

2026-06-05 拍板+落地:配置存储**按"跟谁走"分两表**。

- **spatial_config**(PK `(spatial_prefix, config_key)`,longest-prefix-match 继承):跟**位置/scope**走的配置——tenant/unit/room/bed 默认与 override。这些 prefix 是结构地址,设备 rebind **不变**。当前实际只有 2 行 tenant `alarm.cloud_config`(/48)。
- **device_config**(PK `(device_uid, config_key)`,无 spatial_prefix,FK→device_factory_meta ON DELETE CASCADE):跟**那台物理设备**走、rebind 要跟着搬的配置——`alarm.device_config`、per-device sleepace 等。

**Why:** device 级配置曾按 `spatial_prefix=device_addr` 存进 spatial_config,但 device_addr 由绑定编码、**rebind 就变 → 配置孤儿**(实测 31 行里 17 行孤儿)。用 device_uid(logMAC 终身不变)作 key 根治,且**不需要任何"rebind 时追着改 prefix"的迁移**(那条旧 todo 因此废除)。

**How to apply:**
- device_config **不存 spatial_prefix**——"当前绑在哪"永远从 `devices` 表查(device_uid→device_addr→mask),devices 是绑定单源,再存一份=又 stale+双写 drift。
- 消费端持 device_addr 时:`device_config dc JOIN devices d ON d.device_uid=dc.device_uid WHERE d.device_addr=$1`(接口不变,内部解析 uid)。
- 2-layer alarm 解析 = device_config[uid] override 叠 spatial_config 按 tenant prefix 取的默认。
- 失效/publish 仍按 device_addr 广播(="这台变了"),消费端收到后按 uid reload。
- producer=`device_monitor_settings_service.go`;readers=device_store_service/sleepace_realtime_interval_scheduler + cardagg/qinglan/sensor 三份 alarm_enablement.go。
- 迁移 `owlRD/dbv2/migrations/2026-06-06_device_config_split.sql`;canonical `22_device_config.sql`。

关联 [[alarm_config_2layer_snapshot]](tenant+device 两层快照)、[[feedback_store_raw_not_derived]]、[[device_ipv6_single_trip]]。
