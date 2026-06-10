---
name: sensor-stream-subscriptions
description: sensor 该订阅哪些流 + 每条流的职责 + 与 cardagg 的分工 + Producer 防 loop 规则；含 fall/sit/vital/sleep/activity 落点
metadata: 
  node_type: memory
  type: reference
  originSessionId: 0566c6c8-ddbd-4e2d-b8b6-79fef2a5a715
---

## sensor 订阅清单（producer 4 wire 设计前提）

| 流 | sensor 订阅？ | 用途 | 备注 |
|---|---|---|---|
| `iot:event:stream` | ✅ 是 | activity stat → LastActive/Standing；pose → Standing reset；fall / SitOnGround → roomengine 验证；sleep stage → BedState.SleepStage 融合 | 多 category 共流；详 category 见下 |
| `iot:alarm:stream` | ✅ 是 | WeakBio 累加（HR/RR/ApneaH/WeakBio raw）+ **vital/sleepace alarm 旁路读**：alarm 由 cardagg 直接处理，但 sensor 仍 subscribe，因为 alarm 里某些信息会影响后续 zone/room engine 判定（如 sleepace InBed→engine 抑制 lost_fall pending；vital 异常→fall verifier 风险放大） | 不重复处理 alarm 派发，只**旁路读**；持久化由 iot 模块负责 |
| `iot:monitor:stream` | ❌ **不订阅** | 1Hz raw track（XY/HR/RR）；zoneengine track_manager 自家走原始链路，不进 aggregator | aggregator 关心分钟级派生量，不消费 raw 1Hz |
| `sensor:derived:stream` | ❌ 自家产物 | bed/room/target state；只发不订 | 订阅了就 loop |

## iot:event:stream category 落点

| Category | Producer | sensor 关心？ | 用法 |
|---|---|---|---|
| `activity` | qinglan radar（per minute stat） | ✅ S2 | walk_distance≥2m/walk_duration≥6s → LastActive 60s 节流；stand_duration≥55s → StandingContinuousMin 累+1 封顶 8 |
| `pose` (FieldPose type=2) | qinglan radar | ✅ | 姿态切换 → Standing reset / 跌倒判定辅助 |
| `Fall` / `SuspectedFall` | qinglan radar pose=5/6 → event 流（不是 alarm） | ✅ | roomengine fall verifier 接消费 |
| `SittingOnGround` / `SuspectedSittingOnGround` | qinglan radar pose=7/8 → event 流 | ✅ | 同上 |
| `EnterRoom` / `ExitRoom` / `InBed` / `LeftBed` / `NumberPeople` | qinglan radar / sleepace | ✅ | zoneengine 现有消费链路（不在 aggregator 范围）|
| `SleepStage` | sleepace + radar 双源 | ✅ S4 | FU1 vital 融合 — 加权（Sleepad=8 / Radar=4）写 BedState.SleepStage |

**关键修正**：sensor 文档/memory 早期把 Fall/SitOnGround 当 "alarm" 描述是不准确的——qinglan 把它们发**event 流**（带 EventItem.track_id 等上下文），只有 firmware 已经"判定终态"的 vital/HR/RR/ApneaH/WeakBio_raw + 设备类（Offline/SignalPoor/AngleException/SensorDetached）才走 alarm 流。

## iot:alarm:stream — sensor 旁路读用途

| Alarm type | Producer | sensor 用途 |
|---|---|---|
| `WeakBiometricSignal` | qinglan publishStatSleep | WeakBio score 累加（raw max） |
| `HeartRateAlert` / `RespRateAlert` | qinglan publishStatSleep | WeakBio score 累加（±5） |
| `ApneaHypopnea` | qinglan publishStatSleep | WeakBio score 累加（±15） |
| `InBed` / `LeftBed`（sleepace 来源）| wisefido-sleepace | engine 床事件协调（旁路读，不重复派发） |
| `Offline` / `SignalPoor` / `AngleException` / `SensorDetached` | qinglan device status | engine device fitness 判断（旁路读） |

## Producer 防 loop 规则（必须遵守）

**envelope 字段**：`IoTStreamMessage.Producer` (owl-common/redis/message_types.go:51)
- 设备直发：`canonical /128 IPv6`（来自 `BuildDeviceProducer(addr)`）
- sensor agent 派生：覆盖为 `"sensor.<name>"`（如 `"sensor.caregiver01"`，roomengine/engine.go:1115）

**订阅时第一关过滤**：
```go
if strings.HasPrefix(msg.Producer, "sensor.") {
    return nil // 自家产的消息，跳过；防 escalation loop
}
```

**为什么必须做**：sensor 既是 producer（发 sensor:derived:stream 也可能派生 alarm/event 进 iot 流）又是 consumer（订阅 iot:event/alarm）；如果不过滤 Producer，sensor 派生的 alarm 会被自己消费→再派生→无限 loop（已经在 alarm_back_channel 里见过这个 pattern）。

**做在哪**：每个新 subscriber 入口（`handleAlarmEvent` / `handleEventFrame` 等）第一行 Producer 前缀检查；aggregator 的 `AlarmEventSnapshot.Producer` 字段已预留，但当前消费链路没接通——S1/S2 wire 时必须把过滤一并加上。

## sensor vs cardagg alarm 分工

| 关心字段 | Sensor | Cardagg |
|---|---|---|
| **派发 alarm** | ❌ 不重复发 | ✅ AlarmRouter 唯一派发入口（持久化 + 路由） |
| **读 alarm 影响内部状态** | ✅ 旁路读（roomengine / aggregator）| - |
| **WeakBio 累加 score** | ✅ aggregator | - |
| **派生新 alarm**（如 fall confirmed）| ✅ roomengine verifier→发 iot:alarm | ❌ 不派生 |

cardagg 不把自家收到的 alarm 反向回流 iot:alarm（CLAUDE.md 规则 #1.3 单源真相）；sensor 派生 alarm 时带 Producer="sensor.xxx" 让 cardagg 正常处理但自己订阅时 skip。
