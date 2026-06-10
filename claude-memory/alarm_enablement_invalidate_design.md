---
name: alarm-enablement-invalidate-design
description: 2026-05-19 拍板 alarm.cloud_config 变更 invalidate 协议：选项 C 精确 device_addr 列表；producer (wisefido-data) 先做，consumer (sensor + cardagg) 后做
metadata: 
  node_type: memory
  type: project
  originSessionId: 0566c6c8-ddbd-4e2d-b8b6-79fef2a5a715
---

## 拍板

`spatial_config` 行 `config_key='alarm.cloud_config'` 变更时，由 producer `wisefido-data` 在 DB tx commit 之后 **publish 精确受影响 device_addr 列表**（不是 CIDR 也不是 all），两个 consumer（`wisefido-sensor` + `wisefido-cardagg`）的 `AlarmEnablementCache` 订阅同一 channel → 调 `InvalidateDevices([...])`。

**为什么不选 InvalidateAll**：sensor + cardagg 都要做这事，复用同一精确清单避免 over-invalidate / under-invalidate 不一致。

## 协议

- **Redis pub/sub channel**：`config:invalidate:alarm.cloud_config`
- **Payload**（JSON）：`{"devices": ["fd00:0:3:444:3:101::abcd", ...], "reason": "spatial_config update", "ts_ms": 1747...}`
- **Producer 时序**：必须在 DB tx commit **之后** Publish；commit 失败禁止 Publish（防 stale invalidate）
- **Consumer 行为**：subscribe goroutine 收到后调既有 `cache.InvalidateDevices(addrs)`；ctx cancel 时 unsubscribe（防 leak）

## Producer 端要做（wisefido-data — backlog 高优先）

[wisefido-data/internal/service/alarm_cloud_service.go::UpdateAlarmCloudConfig](owlBack/wisefido-data/internal/service/alarm_cloud_service.go#L516) 和 [admin_alarm_handlers.go::saveAlarmDefaults](owlBack/wisefido-data/internal/http/admin_alarm_handlers.go) 写 `spatial_config` 的所有路径都要：

1. 在 tx 内查 `spatial_prefix LPM` 下所有 `devices.device_ipv6`（受影响 /128 device 列表）
2. tx commit 成功后 redis.Publish channel + payload
3. tx 失败 → 不 Publish

注意：`spatial_prefix` 可能是 tenant /48 / branch /64 / unit /80 / room /88 / bed /96 / device /128 多档；LPM 查询要正确处理"上层 prefix 变 → 下游所有 device 受影响"语义。

## Consumer 端要做（sensor + cardagg — producer 完工后）

两边 `AlarmEnablementCache` 已有 `InvalidateDevices([]string)`（0 caller）。加 wiring：

- `cmd/wisefido-sensor/main.go` + `cmd/wisefido-cardagg/main.go` 启动时：
  - `go subscribeAlarmConfigInvalidate(ctx, redisClient, enablementCache, logger)`
- subscriber 解析 payload.devices → `cache.InvalidateDevices(devices)`
- 兼容：subscribe 失败 / channel 不存在不 fatal（producer 未上线阶段 graceful 跳过）

## 现状缺口

- ✅ **producer (wisefido-data)** 已实现：`PublishAlarmDeviceMessage` 按精确 device_addr 写 `config:alarmDevice:stream`（per-device 粒度，无需再加 LPM 计算）
- ✅ **cardagg consumer** 已实现：`AlarmDeviceHandler.Handle` → `enablement.Invalidate(d.DeviceAddr)`
- ✅ **sensor consumer** 2026-05-19 commit `bebcb18` 落地：`AlarmDeviceConfigConsumer` + 6 Tier1 case；独立 consumer group `wisefido-sensor-alarm-device`，与 cardagg `$cons-cardagg` 隔离

完整闭环已就位。spatial_prefix 上层（tenant /48 / branch /64 等）变更扩散到 affected /128 devices 列表的逻辑暂未需要 — 现在所有 producer 路径都已按 per-device 粒度 Publish。

## 关联

- [[feedback_producer_first]] — producer 先彻底完工，downstream 才能做
- [[p4_next_session_prompt]] — FU7 backlog 项
