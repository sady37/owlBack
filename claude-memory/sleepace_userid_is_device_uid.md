---
name: sleepace-userid-is-device-uid
description: Sleepace 厂家 API userId 全栈统一 = device_factory_meta.device_uid (logMAC)；2026-05-23 收口 3 处分歧 + 9 sleepad 重 bind 实证
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 36f14658-2d5a-470c-b075-31726dab854a
---

Sleepace 厂家 API `data.userId` 在 wisefido 这边的唯一映射 = `device_factory_meta.device_uid` (logMAC，如 `BM87224601335`)，**不是** `devices.device_id` UUID，**绝不是** `device_addr` IPv6。

**Why:** 2026-05-23 [[device_to_card_one_to_one]] sleepad 上数据排错发现：9 个 unit-bound sleepad 只有 4 个出 realtime。根因 `device_monitor_settings_service.pushSleepadToHardware` 把 IPv6 device_addr 当 userId 传给 Sleepace gateway；其它路径（device_store_service / sleepace_realtime_interval_scheduler / report_service）已是 device_uid，三处分歧 silent drift。修完调 `/api/v1/sleepace/device/initialize` 给 9 个全重 bind → 8 个立即出 sleepStage/InBed/LeftBed（剩 1 = Danna 空 unit 无人）。

**How to apply:**
- 任何 `Sleepace*Gateway` / `SleepaceAPI` 参数命名禁用 `deviceID`（歧义），用 `deviceUID`；payload 字段名仍 `userId`（厂家契约）
- bindInfo / initialize / bind / SetRealtimeInterval / SetLeaveSensibility / SetBedParameters / SetReportUploadType / get24HourDailyWithMaxReport / GetAlarmConfig / UpdateAlarmConfig — 所有这些 userId 必须 = device_uid
- 改 sleepad-相关 API callsite 必先 grep userId/userID/deviceID 看变量来源，不要假定参数名
- bindInfo 调试：`curl POST /api/v1/proxy/sleepace/bindInfo {"userId":"<device_uid>"}` 返回 `data:null` = 厂家无 binding，调 initialize 即可
- 重 bind 是幂等的（厂家 user_device 表 PK userId+deviceId upsert），不破坏 in-flight 数据

不要在 `wisefido-data/internal/service/sleepace_gateway_client.go` 里再写 UUID 注释；旧的 "userId=devices.device_id (UUID)" 注释统统已删；类似 Phase 2.1 TODO 注释一并清掉（决策已落，不留 TODO 漂移）。
