---
name: device_status 事件驱动重构（已落地 2026-05-05）
description: cardagg 撤掉 b.online 推导，device:status 改 alarm/event/monitor 流事件驱动 + 看门狗 180s fail-safe；默认不在线
type: project
originSessionId: b1d43ed2-c37b-4fc4-9663-3580c50fb0c9
---
2026-05-05 设计对齐 + 当日 cardagg 落地。取代旧"monitor 推导 b.online → DeriveDeviceOnlineOnly 1s 全字段覆写"架构。一并关闭 [cardagg_b_online_leak_todo.md](cardagg_b_online_leak_todo.md) 和 [cardagg_dedup_and_device_status_todo.md](cardagg_dedup_and_device_status_todo.md) 里"device_status 独立化"两条 backlog。

**2026-05-21 收尾**：drs 表 DROP + Redis 成为 device status 唯一真相源 + FE 端打通（data/qinglan SQL 不再 JOIN drs，service 层用 FillDeviceStatusFromCardagg 拿 status）；详 [[drs_retired_status_redis_only]]。下半"上线后实测验证"清单本日 FE 实测得到证据：9 online / 6 offline 混合状态正确显示（之前永远 15/15 offline 因为 drs.online 永不刷 TRUE）。

## 已落地变更

- 删除 `MonitorBuffer.online/prevEntries/DeviceOnlineEntry/AdvancePruneTick/BuildDeviceStatus/KeepAlive/ClearCard`
- 删除 `StateService.DeriveDeviceOnlineOnly/SetCardOffline`
- 删除 `card.Writer.WriteDeviceStatus`（owl-common）
- monitor_handler.go Handle 末尾 + event_handler.go Handle 末尾各加一处 `state.SetDeviceOnline(true)` —— 正向维护 last_seen_ms+offline=0
- alarm_handler.go：Offline 仍调 `SetDeviceOnline(false)`（patch offline=1，不动 last_seen）；OfflineRecover/DeviceRecover 调 `SetDeviceOnline(true)`；删 KeepAlive 续命路径
- `StateService.SetDeviceOnline` 签名去掉 cardID（仅日志用，已无意义）
- 新增 `internal/service/device_watchdog.go`：30s 周期 SCAN `device:status:*`，last_seen_ms idle > 180s → patch offline=1
- main.go 启动 `deviceWatchdog.Run`；`runDeriveLoop` 撤掉 `DeriveDeviceOnlineOnly`/`ClearCard`/`SetCardOffline` 三处调用，仅保留 prevTargets GC

## audit-only 双轨设计（2026-05-05 同日落地）

设备健康类报警走 **alarm 流落库审计 + 不污染 unhandled count / popAlarm** 双轨：

- `AlarmDef.SkipUnhandledCount=true` 标 7 个设备类（Offline/OfflineRecover/SignalPoor/AngleException/SensorDetached/DeviceFailure/DeviceRecover）
- `InsertAlarmAndUpdateCard`：SkipUnhandledCount 时仅 INSERT alarm_events（active），跳 `updateAlarmCount(+1)` + `setPopAlarmIfHigher`，返回 `SkippedNotify=true` 让调用方短路 writeAlarmState + push
- `AutoResolveDeviceAlarms`：拆分 countable/skipped——所有 active 行 UPDATE→auto_resolved（保留审计），但只对 countable 做 `-count`
- `countActiveAlarms` / `findTopActiveAlarm`（recalc 路径）：SQL 加 `event_type <> ALL(skipUnhandledCountTypes)`
- 5 个 cardagg 调用点 + 2 个 sensor 调用点全部加 `if result.SkippedNotify { return nil }` 短路
- alarm_device.monitor_config 给所有 device 补 5 项设备健康配置（`scripts/migrations/2026-05-05-add-device-class-alarms.sql`）

效果（实测验证）：alarm_events 表设备类 audit 落库 ✅；cards.unhandled_alarm_N 不增 ✅；card.pop_alarm 不写 ✅；signal_poor 等 alarm flag 不被 monitor patch 冲掉 ✅；前端 Overview 卡片现有 sleepace-offline.svg / radar-offline.svg 图标继续生效（device:status hash 独立通道）。

## parseMonitorConfig 隐藏 bug（同日修复）

`alarm_enablement.go::parseMonitorConfig` 原逻辑 "defaults 为基础 + override 套用"——**defaults 列表里没有的项即使 monitor_config 里配了也被丢弃**。
DefaultAlarmSetting.Sleepad/Radar 没有 Offline / SignalPoor / AngleException / DeviceFailure（只有 SensorDetached），导致前 4 项即便 SQL 补在 alarm_device 里也对 cache 不可见 → IsEnabled() 返回 false → silent skip。
修法：parseMonitorConfig 改为 union(defaults, overrides)——defaults 没有但 overrides 配了的项也保留。

## 模型：默认不在线 + 事件驱动写入

| 来源 | offline | last_seen_ms | flags |
|---|---|---|---|
| monitor 流（每条 device 数据） | → 0 | → now | — |
| event 流（device 直发的事件） | → 0 | → now | — |
| alarm Offline | → 1 | 不动 | — |
| alarm OfflineRecover | → 0 | → now | — |
| alarm SignalPoor/AngleException/SensorDetached | 不动 | 不动 | flag=1 |
| alarm xxxRecover | 不动 | 不动 | flag=0 |
| 看门狗（fail-safe） | last_seen_ms 超 180s → 1 | 不动 | — |
| 默认（key 不存在） | = 1（不在线）| = 0 | = 0 |

关键语义：
- **alarm Offline 不刷 last_seen_ms**：last_seen_ms 是设备最后活着的时刻，"刚知道它死了"不算它活着
- **写入全部 HSET 部分字段（patch）**，不再 HMSET 全字段覆写——治愈 [state_service.go:879 SyncDeviceStatusFromActiveAlarms](../../../owl/owlBack/wisefido-cardagg/internal/service/state_service.go#L879) 注释里的"WriteDeviceStatus 历史把 alarm flags 重置为 0"旧伤
- **monitor 流回来时 → offline=0 + last_seen_ms=now（选项 A）**：sensor data is positive proof，覆盖过期 alarm 状态。要求 qinglan/sleepace 在重新看到设备时也补发 OfflineRecover 关闭 alarm_events 表 active 行（transition dedup 已天然做到）

## 看门狗 TTL = 180s

理由：
- 2 × sleepace OfflineRecover 80s 周期 + 余量（容忍一次心跳延迟）
- qinglan radar 1Hz monitor 流，180s 永远摸不到它
- gateway 真挂时前端最长 3 min 内反映出离线

90s 不行：== sleepace 80s + 10s 余量，一次 MQTT 抖动 → 95s 心跳到达 → 90s 看门狗已经标 offline → 紧接着 OfflineRecover 翻回 0 → UI 闪烁+假推送。

心跳源盘点：
- qinglan monitor：~1Hz（雷达运行时）
- qinglan health_check OfflineRecover：10 min，transition-only，健康态静默（OK，因为 monitor 流持续）
- sleepace monitor：~1Hz（mode=0 在床期）
- sleepace OfflineRecover：**80s 永不 dedup**（mode=1 离床期唯一心跳，[wisefido-sleepace/internal/subscriber/health_check.go:100](../../../owl/owlBack/wisefido-sleepace/internal/subscriber/health_check.go#L100) 注释）

## 上线后实测验证（待做）

- [ ] BM87224700978 sleepace 离线 → device:status `offline=1` 在 5s 内出现，且不被刷回 0（原 b.online leak 复现场景）
- [ ] sleepace mode=1 离床期 1h + 80s OfflineRecover 正常到达：last_seen_ms 跟随跳动，offline 始终 0
- [ ] sleepace 真离线 ≥ 5min（80s 心跳停）：看门狗 180s 内 patch offline=1
- [ ] qinglan radar 24h 在线：offline=0，last_seen_ms 跟随 monitor 流刷新
- [ ] alarm SignalPoor 来后 monitor 持续刷：signal_poor 仍=1（验证 patch 模式不被冲掉，历史回归点）
- [ ] cardagg 重启 bootstrap：`SyncDeviceStatusFromActiveAlarms` 重建 flags 正常

## 双 TTL 解耦

- `PruneStaleDevices`（90s）：内存 GC
- 看门狗（180s）：fail-safe offline 判定

不再纠缠（旧架构两者合一所以必须 sleepace 80s < TTL 90s 配齐，新架构解耦后看门狗自由取值）。
