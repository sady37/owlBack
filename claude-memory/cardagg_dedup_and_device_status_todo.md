---
name: cardagg dedup + device_status 独立化 backlog
description: cardagg 侧 DedupWhileActive 仍待做（qinglan 侧源头去重已 2026-05-03 落地）+ card:state.DeviceStatus 拆到 device:status:{deviceID}；详细设计在 doc/TODO.md
type: project
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---
两个已对齐方向、部分实施的改造，2026-05-03 写定，下次提到任一个时切到 doc/TODO.md 看完整设计。

## 1. 设备健康类去重（两层）

**Layer 1 — gateway 侧源头 transition dedup（✅ 2026-05-03 已实施）**：

- **qinglan**: [wisefido-qinglan/internal/subscriber/health_check.go `publishHealthIfChanged`](../../../owl/owlBack/wisefido-qinglan/internal/subscriber/health_check.go) — 复用早就声明但没接线的 `prevHealth` 缓存，按 (deviceUID, fieldKey) 记录上次发布值。0↔1 跳变才下发；首次观测=0 静默初始化（不发"全清"假恢复）；unsubscribe / disabled 时 `delete(prevHealth, deviceUID)`。Offline/SignalPoor/AngleException 同走这条。
- **sleepace**: [wisefido-sleepace/internal/consumer/mqtt_consumer.go `shouldPublishDeviceHealth`](../../../owl/owlBack/wisefido-sleepace/internal/consumer/mqtt_consumer.go) — `prevConnFault`+`prevSensorFailure` 两张 map，串到 `connectionStatus`(Offline) 和 `deviceSenSor`(SensorDetached) 两个 case。同样的 transition 语义。
- **统一行为**：服务重启 → cache 清空 → 设备处于故障态时再次 onset 一次（这是想要的，cardagg 端可借此重建 active）；处于健康态时静默。

**Layer 2 — cardagg 端 onset dedup（✅ 2026-05-03 已实施 = Phase C P0）**：
- `AlarmDef.DedupWhileActive` 加在 [owl-common/alarm/alarm.go](../../../owl/owlBack/owl-common/alarm/alarm.go)，5 个设备类（Offline/SignalPoor/AngleException/SensorDetached/DeviceFailure）置 true
- `InsertAlarmAndUpdateCard` 在 [owl-common/card/alarm_db.go](../../../owl/owlBack/owl-common/card/alarm_db.go) 加守卫：BeginTx 前查 `idx_alarm_events_device_active`，命中 active 行返回 `AlarmInsertResult{Deduped:true, EventID:existing}`，不开 tx
- 7 个 caller 全改：cardagg `alarm_service.go` 5 处 + wisefido-sensor 2 处（`evaluator.go` / `alarm_event_service.go`），统一 `if result.Deduped { return nil }` 短路 push 通知
- 仅设备类受影响；Fall/SuspectedFall/NightAbsence/Stay/InBed/LeftBed 等事件型保持每次独立成行

## 2. device_status 独立维护（已落地 2026-05-05）

device:status:{deviceID} 已从 card:state 拆出（之前 Phase A 完成），且 cardagg 写入路径在
2026-05-05 重构为事件驱动（monitor/event 正向、alarm Offline 负向、看门狗 fail-safe），
不再有 b.online 推导导致的状态冲掉问题。详见 [device_status_event_driven_refactor.md](device_status_event_driven_refactor.md)。

## 共同约束

- OfflineRecover 周期发**必须保留**（KeepAlive 软心跳，见 alarm_handler.go:271-277）
- sleepace 已有非对称 dedup（onset 翻转/recover 永发）作为参考实现
- 用 alarm_events.device_id 做 dedup 查询比经 card:state 转一道更直接

## 相关 case

D523 fall 测试 doc/cases/d523-fall-13-52/raw.tsv：实测 angle 全天稳定 -63.89（超阈值 4°），正是要被 dedup 的典型周期触发场景。

## 完整设计 + 验收 + 迁移策略

→ [owlBack/doc/TODO.md](../../../owl/owlBack/doc/TODO.md)
