---
name: platform-agent-addressing
description: Platform agent (sensor/cardagg/etc) 首次获得 IPv6 first-class 身份；alarm_events.producer 改 INET；cardagg 永不当 producer；权威 doc owlBack/doc/platform_agent_addressing.md
metadata: 
  node_type: memory
  type: project
  originSessionId: 5c181f1d-cf48-440b-bd17-e5ae711213b7
---

2026-05-15 拍板（本次会话）：解决 alarm_events 字段不完整、parent_span 永远没意义、sensor 把 device_addr 用 ZoneID prefix 兜底等连串问题，**根因 = sensor 没有 first-class 网络身份**，envelope 字段没法填全。

## 一句话总结

Platform agent (sensor/cardagg/cognitive/action/ai-health/iot/data/qinglan/sleepace) 各拿一段保留 IPv6（`fd00:0:fff1::/48` ~ `fd00:0:fff7::/48`），envelope.producer 改为 INET 类型，sensor 自己 fire timer-based alarm，cardagg 永不当 producer。

## 关键决定

1. **Platform slot**：`fd00:0:fff1::/48` (sensor) / `fff2` (cardagg) / `fff3` (data) / `fff4` (iot) / `fff5` (qinglan) / `fff6` (sleepace) / `fff7` (ai-health)；避开业务 tenant `fd00:0:1-9::/48` 占用段
2. **UID = uuid_v5(owl_platform_ns, agent_name + ":" + ipv6)**，确定性派生但**写入 .env 启动只读**（避免改 IP 后 UID 跟历史 trace 对不上）
3. **Sequence**：Redis `<agent>:seq:<ipv6>` INCR，跨重启单调
4. **alarm_events.producer 改 INET**（不加 UID 列；device_uid 已有，且 platform agent UID 是 IPv6 派生的）
5. **device-direct alarm**：producer == device_addr（同一个 /128）
6. **cardagg 永不当 producer**：删 `ScanPendingAlarms` / `cardaggProducer*` / `AddPendingAlarm` 整套；timer fire 移到 sensor
7. **Redis `alarm:pending` Hash 整段删**（连带 Arm/Cancel publish + cardagg A9 分支）；sensor 启动清空 in-memory + Redis
8. **不做 HA pending 恢复**（重启清空，与 cardagg 历史一致——服务重启就清，得不偿失才做 HA）

## 5 战役

- **1**：schema migration（producer→INET 清旧数据）+ owl-common spatial.DerivePlatformUID
- **2**：sensor identity（.env + Redis seq + envelope 填齐）+ BedDeviceLookup（修真 device_addr）+ Fire 用 `p.Trigger.Ts`
- **3**：删 cardagg fire 路径 + 删 Redis pending Hash 整套
- **4**：device-gateway publisher producer 改 INET
- **5**：audit `cardaggProducerAlarmRouter` + memory/doc 收尾

## 权威文档

- [doc/platform_agent_addressing.md](../../../owl/owlBack/doc/platform_agent_addressing.md)（slot 表 / UID 派生 / envelope 契约 / 红线）
- 更新到 [doc/cardagg_sensor_split.md §7](../../../owl/owlBack/doc/cardagg_sensor_split.md)：加 cardagg 不当 producer 红线

## 与既有 memory 的关系

- 配合 [[cardagg_sensor_responsibility_split]]：那里规定 cardagg 是 thin adapter，本文档把 "thin adapter 含义" 收紧到「永不当 producer」
- 扩展 [[envelope_protocol_evolution]]：device-as-host stateless 派生扩展到 agent-as-host
- 承接 [[event_alarm_subscription_gap]]：sensor envelope 字段完备后，event/alarm 订阅链才能 trace 闭环

## 上次审计的 sensor envelope 缺漏（已立 fix 路径）

| 字段 | 之前 | 应为 | 修复位置 |
|---|---|---|---|
| sequence_number | 硬编码 "0" | Redis INCR | alarm_back_channel.go:120 |
| device_addr | ZoneID prefix (zero host) fallback | 真 trigger 设备 /128 (BedDeviceLookup) | alarm_firer.go:77 |
| timestamp | Fire 时刻 time.Now() | `p.Trigger.Ts` (bed→vacant 瞬时) | alarm_firer.go:61 |
| DataValue payload | 只 zone meta | 加 trace[] (SignalEvidence) + evidence{engine_fire_ms/...} | alarm_firer.go:101 |
| producer | 字符串 "wisefido-sensor" | INET fd00:0:fff1::1 | alarm_back_channel.go:118 |

## 实施完成（2026-05-15）

5 战役全部落地：

**战役 1**（schema + helper）
- migration `2026-05-15_alarm_events_producer_to_inet.sql`：producer VARCHAR(100)→INET，DELETE 18 test rows，加 `ae_producer_in_owl` (fd00::/32) + GiST 索引
- `60_alarm_events.sql` canonical 同步更新
- `owl-common/spatial/platform_uid.go`：`PlatformAgentNamespace` (uuid_v5 固定 ns `8d52a4c3-1d28-516a-86df-6e6b1b3f6e6e`) + `DerivePlatformUID(agentName, ipv6)` + `IsPlatformAgentAddr()` + slot 常量 + 6 unit tests（含 snapshot 锁 UID 防 namespace 漂移）

**战役 2**（sensor identity + envelope）
- `.env`: SENSOR_IPV6=fd00:0:fff1::1 / SENSOR_UID=a893ff12-7d72-58dc-a26c-ce724e6007da
- `wisefido-sensor/config.yaml` `identity:` 块 + `IdentityConfig` struct + ENV override
- `sensorAgentIdentity()` 在 main.go 校验 pinned UID == 派生值（不等 warn 不 fail）
- `AlarmBackChannel` 加 `AgentIdentity` 字段；`nextSeq()` Redis INCR `wisefido-sensor:seq:<ipv6>`
- envelope.producer / sequence_number 真值填充
- `BedDeviceLookup` (bed /96 严格 → room /88 兜底，LeftBed/Fall→Radar 优先；3 unit tests)
- `alarm_firer.Fire` triggered_at = `p.Trigger.Ts`；engine_fire_ms 进 evidence；buildTriggerData 带 `p.Trigger.Trace`
- `owl-common/card/alarm_db.go` INSERT SQL 加 `NULLIF($18,'')::INET` cast

**战役 3**（删 cardagg fire + Redis pending Hash）
- `ScanPendingAlarms` 整段删；main.go 删 `runPendingAlarmScan` cron
- `AddPendingAlarm` / `RemovePendingAlarm` / `PurgeDisabledPendingForDevice` / `AddStayPendingIfEnabled` / `TryAddLeftBedPendingAtTrigger` 全改 **no-op stub**（保 callsite 编译）
- `cardagg.alarm_handler` A9 `pending_arm/pending_cancel` 分支改 debug-log drop
- `sensor.alarm_firer.Arm/Cancel` 改 no-op（不再 publish）
- `sensor.wiring.Setup` 启动时 `redis.Del("alarm:pending")` cleanup
- **cardagg.alarm_handler trust-platform-agent**：`if spatial.IsPlatformAgentAddr(m.Producer) { enabled=true; level=envelope.alarm_level }`
- `payload.SequenceNumber = m.SequenceNumber` 漏 copy 修复（parent_span/trace_id 之前永远空）

**E2E 验证**（注入 LeftBed / NightAbsence envelope）
- producer = `fd00:0:fff1::1/128` ✓ INET
- parent_span = `fd00:0:fff1::1.101` ✓ sensor + seq
- trace_id = `fd00:0:fff1::1.101` ✓ 单跳 = parent_span
- device_addr = `fd00:0:3:111:3:301:c3d:53d0/128` ✓ 真 radar
- device_uid = `9D8A326309E7` ✓ LeftBed 主 radar (BedDeviceLookup)
- snapshot：John.Y / 101 / Main / BedA ✓

**战役 4**（device-direct producer）
- `BuildDeviceProducer` 去掉 `"device:"` 前缀，直接返回 device.IPv6
- qinglan / sleepace publisher 不变（已用 BuildDeviceProducer）
- device-direct alarm 的 producer == device_addr（同 /128）

**战役 5.1**（cardagg-internal producer audit）
- `RecordDeviceFailure`：producer 从 `cardaggProducerAlarmRouter` 改 `deviceID`（设备硬件事件 → device 自己 IP）
- `cardaggProducerAlarmRouter` / `cardaggProducerPendingScanner` 常量改空串
- 剩余 backlog：
  - `PersistAlarmFromTrack` → 无 caller，**死代码** 可删
  - `PersistAlarmWithTriggerData` + main.go `runNightAbsenceCheck` → sensor.zonealarm 已覆盖 NightAbsence，cardagg 这条路径冗余可删

## TZ 规则（之前会话已落地，配套此战役）

`wisefido-data/internal/service/alarm_event_service.go`：
- `AlarmEventDTO.TriggeredAt` int64 → **string** (RFC3339 with unit_timezone offset)
- `AlarmEventDTO.HandledAt` *int64 → ***string**
- `applyUnitTimezoneToTimestamps()` 在 ensureUnitTimezone 后 reformat
- alarmRecord API 返回 `"triggered_at": "2026-05-15T17:35:53-07:00"` 形态

## 跟既有 memory/doc 的关系

- 完整契约：[doc/platform_agent_addressing.md](../../../owl/owlBack/doc/platform_agent_addressing.md)
- 关联 [[cardagg_sensor_responsibility_split]]：cardagg=thin adapter，本战役补「永不当 producer」红线
- 关联 [[envelope_protocol_evolution]]：device-as-host stateless 派生扩展到 agent-as-host
- 关联 [[event_alarm_subscription_gap]]：sensor envelope 字段补齐后 trace 链可追

## 已知遗留 bug（不在本战役 scope）

- **cardagg.alarm_enablement.loadDevice IPv6 cast bug**（详 [[cardagg_sensor_responsibility_split]] §当前阻塞）：device-direct alarm 的 enablement gate 仍走旧 UUID cast，导致 device-direct LeftBed/Stay 落库被拦。修复跟本战役无关（platform-agent 路径走 trust bypass，已通）。
- `PersistAlarmFromTrack` / `runNightAbsenceCheck` 冗余路径未删（避免 scope 蔓延）。
