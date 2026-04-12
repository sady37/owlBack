English
-------

Summary
-------
Implement device online/offline detection and cleanup; add Redis stream drain for `config:card:stream`; optimize related logic and pending alarm handling.

Behavior
--------
- Devices are marked online if present in the buffer and their `LastTs` is within 90 seconds.
- Devices with `LastTs` > 90s are removed from the online state and will be pruned from the buffer by `PruneStaleDevices`.

Key changes
-----------
- Add `DeviceOnlineEntry`, `BuildDeviceStatus`, and `ClearCard` to maintain per-card online entries and produce `device_status`.
- Call `AdvancePruneTick` every 4s from `MonitorHandler.RunLoop` (aligned with `PruneFields`) to update online entries.
- Call `PruneStaleDevices` (every 90s in `runDeriveLoop`) to delete device entries that have been inactive > 90s.
- Implement `DrainConfigCardStream` (PEL + NEW phases) with timeout and message-limit protections to preheat `DeviceCardResolver` and `metaCache` before IoT streams start.
- Add `redisPendingAdapter` and `runPendingAlarmScan` for pending alarm handling.
- Remove redundant uses of `ActiveDevicesByCard` to avoid duplicate processing.
- Tune constants: `MonitorFieldTTL = 6000` (ms), `drainTimeout`, `drainMaxMsgs`, `drainBatch`, etc.

Files changed
-------------
- owlBack/wisefido-cardagg/main.go
- owlBack/wisefido-cardagg/internal/service/monitor_buffer.go
- owlBack/wisefido-cardagg/internal/consumer/monitor_handler.go
- owlBack/wisefido-cardagg/internal/consumer/stream_subscriber.go

Notes for reviewers
-------------------
- Verify concurrency between `Write`, `PruneFields`, and `AdvancePruneTick` locks; ensure no races.
- `DrainConfigCardStream` uses `XCLAIM` to handle PEL entries during startup; confirm this behavior is acceptable for your restart/recovery procedures.

中文
----

概要
----
实现设备在线/离线判定与条目清理，在 IoT 流启动前清空 `config:card:stream` 积压，并优化相关逻辑与 pending 报警处理。

行为说明
--------
- 设备在缓冲区存在且其 `LastTs` 距当前时间 ≤ 90 秒时标为在线。
- 当 `LastTs` > 90 秒时从在线表移除，并由 `PruneStaleDevices` 定期彻底删除。

主要改动
--------
- 新增 `DeviceOnlineEntry`、`BuildDeviceStatus`、`ClearCard`，用于维护按卡的在线表并生成 `device_status`。
- 在 `MonitorHandler.RunLoop` 中每 4 秒调用 `AdvancePruneTick`（与 `PruneFields` 对齐）更新在线表。
- 在 `runDeriveLoop` 中每 90 秒调用 `PruneStaleDevices`，删除超过 90 秒无活动的设备条目。
- 新增 `DrainConfigCardStream`（PEL + NEW 两阶段），带超时与条数上限，启动前用于预热 `DeviceCardResolver` 与 `metaCache`。
- 新增 `redisPendingAdapter` 与 `runPendingAlarmScan` 用于 pending 报警处理。
- 移除冗余的 `ActiveDevicesByCard` 调用以避免重复计算。
- 调整常量：`MonitorFieldTTL = 6000`、`drainTimeout`、`drainMaxMsgs`、`drainBatch` 等。

修改文件
--------
- owlBack/wisefido-cardagg/main.go
- owlBack/wisefido-cardagg/internal/service/monitor_buffer.go
- owlBack/wisefido-cardagg/internal/consumer/monitor_handler.go
- owlBack/wisefido-cardagg/internal/consumer/stream_subscriber.go

评审提示
------
- 注意 `Write`、`PruneFields` 与 `AdvancePruneTick` 之间的并发与锁策略，确保无竞态问题。
- `DrainConfigCardStream` 会在启动时使用 `XCLAIM` 认领 pending 条目，请确认该行为与重启/恢复流程兼容。
