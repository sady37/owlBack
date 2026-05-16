# cardagg vs wisefido-sensor 职责分工（终态契约）

记录日期：2026-05-15  
拍板：用户原话 "cardagg 只是薄薄一层适配层 / fall 也必须经过 sensor，因为雷达会误报"

---

## 1. 划分原则

**cardagg = 即时 + 已确定 + 可信源 的薄 adapter**  
职责：envelope → DB / Redis hash 翻译；不做任何推断、时间窗、跨源融合。

**sensor = 不可信源（radar）+ 推导 + 跨源融合 + 时间窗 的判定层**  
职责：所有需要"算"的事；判定后通过 `alarm_back_channel` 回流给 cardagg 落库。

### 可信 vs 不可信

| 源 | 可信度 | 理由 |
|---|---|---|
| Sleepad firmware（native alarm 151 等） | ✅ 可信 | 设备端已 gate；时序 / 阈值由 firmware 完成 |
| Device 健康类（Offline/SignalPoor/AngleException/SensorDetached/DeviceFailure） | ✅ 可信 | 物理事实，无需推断 |
| Radar firmware（Fall/SittingOnGround/Stay/InBed/LeftBed/etc） | ❌ 不可信 | radar 容易误报；必须 sensor verifier 跑一遍 |
| sensor 派生（任何 producer="wisefido-sensor"） | ✅ 可信（已 gate） | sensor 内部已做 enablement + verifier + 时间窗 |

---

## 2. cardagg 应处理（持久化 + 直发 FE/push）

### 2.1 Sleepad 设备直发 alarm

| Alarm | 来源 device |
|---|---|
| `InBed` / `LeftBed`（即时型，duration_sec=0） | sleepace native alarm 151 |
| `BedSitUp` / `SuspectedBedSitUp` | sleepad |
| `AbnormalBodyMovement` / `NoBodyMove` / `NoTurnOver` | sleepad（radar 不支持） |
| `ApneaHypopnea` | sleepad vital |
| `WeakBiometricSignal` | sleepad vital |
| `HeartRateAlertHigh/Low` / `RespRateAlertHigh/Low` | sleepad vital 阈值 |
| `HeartRateNormal` / `RespiratoryRateNormal` | sleepad vital recovery |
| `PressureSensor` | sleepad device status |

### 2.2 设备健康类（sleepad + radar 共有，即时无需推断）

| Alarm | 设备类型 |
|---|---|
| `AlarmTypeOffline` / `AlarmTypeOfflineRecover` | 双 |
| `AlarmTypeDeviceFailure` / `AlarmTypeDeviceRecover` | 双 |
| `SignalPoor` / `SingalPoorRecover` | radar |
| `AngleException` / `AngleExceptionRecover` | radar |
| `SensorDetached` / `SensorDetachedRecover` | sleepad |

### 2.3 Sensor 回流的"已 gate" alarm（producer="wisefido-sensor"）

| Alarm | sensor 派生路径 |
|---|---|
| `Fall` / `SuspectedFall` | roomengine.track_manager → emitAIAlarm（lost/silent/bedside/still）；以及 RecordRadarAlarm verifier 通过的 firmware Fall |
| `SittingOnGround` / `SuspectedSittingOnGround` | 同上路径 |
| `Stay` | zonealarm 时间窗 10min |
| `LeftBed`（30min timer） | zonealarm 时间窗 |
| `NightAbsence` / `BedNightAbsence` | zonealarm 时间窗 + 21-7 night window |

cardagg 看到这些时**信任已 gate**，不再二次 enablement check（理想态退化成 `if level != ""`）。

---

## 3. sensor 应处理（推导 + 验证 + 派生）

### 3.1 Radar firmware 直发 alarm（不可信，必须 sensor verifier）

| Firmware Alarm | sensor 动作 |
|---|---|
| `Fall` / `SuspectedFall`（producer="device:..." device_type=Radar） | RecordRadarAlarm 跑 verifier；通过转发为 producer="wisefido-sensor" 的 Fall；ghost 则 emit cancel 或不转发 |
| `SittingOnGround` / `SuspectedSittingOnGround` | 同上 |
| `WarningArea`（radar） | sensor 验 |

### 3.2 Radar event（zone engine / roomengine 输入）

| Event | sensor 处理路径 |
|---|---|
| `EnterRoom` / `ExitRoom` / `NumberPeople` | zoneengine.adapter_radar → ZoneEngine → RedisAdapter 写 RoomState/BathRoomState |
| `InBed` / `LeftBed`（radar 来源） | zoneengine.adapter_radar → ZoneEngine → RedisAdapter 写 BedState |
| `Activity`（radar walk/stand 时长） | sensor 推导（state_service 等价物） |
| `Track` 帧（pose / position） | roomengine 跑 fall 判定 + ghost verdict |

### 3.3 Sensor 派生 alarm（推导 / 融合 / 时间窗）

| Alarm | sensor 内位置 | 触发条件 |
|---|---|---|
| Fall (`lost_fall`) | roomengine.track_manager scanLostFall | track 突然消失 + 无 ExitRoom + 无 number_people=0 兜底 |
| Fall (`silent_fall`) | roomengine.track_manager scanSilentFallLeftBed | sleepad LeftBed + radar 仍在 Bed 邻域（双源融合） |
| Fall (`bedside_fall`) | roomengine.track_manager scanBedsideFall | LeftBed 后床边 100cm 内 ≥ 15min 静止（夜间） |
| Fall (`still_fall`) | roomengine 卫生间路径 | bathroom 内 stand 静止超时（Stay alarm 启用） |
| `Stay` 10min | zonealarm.supervisor | bathroom→Occupied 持续 10min |
| `LeftBed` 30min | zonealarm.supervisor | bed→Vacant 持续 30min |
| `NightAbsence` | zonealarm.supervisor | room→Vacant 持续 30min（21-7） |
| `BedNightAbsence` | zonealarm.supervisor | bed→Vacant 持续 30min（21-7） |
| ghost verdict | roomengine.emitGhostVerdict | birth-score / Kalman / cell history 判 ghost |

### 3.4 跨源融合（必须归 sensor，cardagg 无资格）

| 融合 | sensor 内位置 |
|---|---|
| sleepad LeftBed + radar 仍在床区 → silent_fall | scanSilentFallLeftBed |
| sleepad InBed + radar InBed ±15s 双源确认 → RadarInBedConfirmedMs gate | bedSession 入场门控 |
| sleepad sleep_stage + radar track 验证 | （目前在 cardagg.routeSleepStageEvent，应迁 sensor） |

---

## 4. 数据链 — Gateway 端分流（2026-05-15 拍板）

**核心原则：stream 语义 = 信任级别。Gateway 端就分好流，下游订阅者无需知道"谁可信"。**

| Stream | 含义 | 订阅者 |
|---|---|---|
| `iot:alarm:stream` | "明确的、可信的、可直接落库" | **只 cardagg** |
| `iot:event:stream` | "原始信号 / 待判定 / 需推断" | **只 sensor** |

### Gateway 分流规则（device-gateway 自己拍板发哪个 stream）

| Source | Alarm/Event | 发送 stream | 理由 |
|---|---|---|---|
| sleepace (sleepad) | 所有 native alarm（InBed/LeftBed/BedSitUp/vital/SensorDetached/Offline/etc） | `iot:alarm:stream` | sleepad firmware 已 gate，可信 |
| qinglan (radar) | **Fall / SuspectedFall / SittingOnGround / SuspectedSittingOnGround** | `iot:event:stream` | radar firmware 容易误报，必须 sensor verifier |
| qinglan (radar) | **Vital 阈值类**：HeartRateAlertHigh/Low / RespRateAlertHigh/Low / ApneaHypopnea / WeakBiometricSignal | `iot:alarm:stream` | "测量+阈值"型，与 sleepad vital 同性质，可信。注：radar HR 精度问题在 FE/load 层处理（[radar_hr_no_critical](../memory/radar_hr_no_critical.md)：CRITICAL→WARNING 收窄），不影响 gateway 分流 |
| qinglan (radar) | **设备健康类**：Offline / OfflineRecover / SignalPoor / SingalPoorRecover / AngleException / AngleExceptionRecover / SensorDetached / SensorDetachedRecover / DeviceFailure / DeviceRecover | `iot:alarm:stream` | 硬件物理事实，与 sleepace 设备健康同性质，可信 |
| qinglan (radar) | EnterRoom / ExitRoom / NumberPeople / InBed / LeftBed / Activity / Track / SleepStage / WarningArea | `iot:event:stream` | 已在 event stream，不动 |
| sensor 派生（producer="wisefido-sensor"） | Fall / Stay / NightAbsence / BedNightAbsence / LeftBed-30min / SittingOnGround | `iot:alarm:stream` | sensor 已 gate，可信 |

### 数据链示意

```
                                     ┌──► InBed/LeftBed/BedSitUp/vital/Offline/SensorDetached/etc
sleepace ──► iot:alarm:stream ◄──────┤
                                     │
qinglan(radar) ──► iot:alarm:stream ◄┤
                                     │   ┌──► HeartRateAlertHigh/Low / RespRateAlertHigh/Low
                                     │   ├──► ApneaHypopnea / WeakBiometricSignal
                                     │   │       (Vital 阈值类 = 测量+阈值)
                                     │   │
                                     │   ├──► Offline / OfflineRecover
                                     │   ├──► SignalPoor / SingalPoorRecover
                                     │   ├──► AngleException / AngleExceptionRecover
                                     │   └──► SensorDetached / SensorDetachedRecover
                                     │           (设备健康类 = 硬件事实)
                                     │
sensor (verdict 通过 + zonealarm 派生)
   └─► alarm_back_channel ──► iot:alarm:stream
                                     │
                                     ▼
                              cardagg.alarm_handler ─► alarm_events + card_status (sleepad 字段)


qinglan(radar) ──► iot:event:stream ◄┐
                                     │   ┌──► Fall / SuspectedFall              (待 sensor verifier)
                                     │   ├──► SittingOnGround / SuspectedSittingOnGround  (待 sensor verifier)
                                     │   ├──► EnterRoom / ExitRoom / NumberPeople (zone engine 输入)
                                     │   ├──► InBed / LeftBed (radar)            (zone engine 输入)
                                     │   ├──► Activity / Track / SleepStage      (sensor 处理)
                                     │   └──► WarningArea                        (sensor 验)
                                     │
                                     ▼
                              sensor consumes:
                                  ├─► roomengine: Fall verifier (RecordRadarAlarm) → 通过则 emit alarm
                                  ├─► zoneengine: presence 三投影写 card_status
                                  └─► zonealarm: 时间窗派生 (Stay/LeftBed30min/NightAbsence/BedNightAbsence) → emit alarm
```

### 跨 device class 融合 — 全归 sensor

sensor 同时订阅 `iot:event:stream` 和 `iot:alarm:stream`（前者为 radar 输入，后者为 sleepace 输入做跨源融合）：
- silent_fall = sleepace LeftBed (alarm stream) + radar 仍在床区 (event stream)
- bedSession 入场门控 = sleepace InBed + radar InBed ±15s

cardagg 不做任何跨源融合，永不订阅 `iot:event:stream`。

---

## 5. 落地步骤（Gateway 分流方案）

### Phase 1 — qinglan publisher 分流

[`wisefido-qinglan/internal/consumer/mqtt_consumer.go`](../wisefido-qinglan/internal/consumer/mqtt_consumer.go) `case 2:` 段（Fall/SittingOnGround）：
- 当前：`PublishAlarm` → `iot:alarm:stream`
- 改为：`PublishEvent` → `iot:event:stream`
- 设备健康类（Offline/SignalPoor/AngleException/SensorDetached）保持 `PublishAlarm`，不动

### Phase 2 — sensor 订阅 event stream 上的 radar Fall

sensor 现在已经订阅 `iot:alarm:stream` 跑 RecordRadarAlarm verifier，需要补上对 `iot:event:stream` 上 radar Fall 的处理：
- [`wisefido-sensor/internal/roomengine/engine.go`](../wisefido-sensor/internal/roomengine/engine.go) runAlarmLoop / runEventLoop：在 event 流增加 Fall/SittingOnGround 路由到 RecordRadarAlarm
- 等 Phase 1 上线后再切走 alarm 流的 RecordRadarAlarm 订阅（双订过渡）

### Phase 3 — sensor 派生 enablement gate

- [`wisefido-sensor/internal/roomengine/track_manager.go::emitAIAlarm`](../wisefido-sensor/internal/roomengine/track_manager.go) 加 enablement check：spatial_config 里 Radar.Fall.is_enabled=false 则不 emit
- [`wisefido-sensor/internal/zonealarm/rules.go`](../wisefido-sensor/internal/zonealarm/rules.go) 4 条 rule 加 spatial_config 启动时灌入 + 运行时按 enablement 决策 fire / 不 fire

### Phase 4 — cardagg 退订 event stream + 简化 alarm gate

- cardagg 退订 `iot:event:stream`（连带 EventHandler 删除/退役 — radar event 完全归 sensor）
- [`wisefido-cardagg/internal/consumer/alarm_handler.go`](../wisefido-cardagg/internal/consumer/alarm_handler.go) 砍 `case alarm.Fall, alarm.SittingOnGround` 分支（这些不再走 alarm stream 的 firmware 直发，只接 sensor producer="wisefido-sensor" 的回流）
- 简化 enablement gate：trust producer="wisefido-sensor" 的 alarm，不二次 check（cardagg.alarm_enablement IPv6 cast bug 影响范围缩小到 sleepad 类 + device 健康类）

### 阻塞当前 Fall 测试的 bug

- [`alarm_enablement.go:139`](../wisefido-cardagg/internal/service/alarm_enablement.go#L139) `device_id = $1::uuid` 在 v2 IPv6 cutover 后传入 IPv6 字符串导致 cast 失败；同时 `loadDevice` 完全跳过 spatial_config，只用硬编码 defaults。
- 终态：影响范围只剩 sleepad alarm 类 + device 健康类（radar 业务 alarm 不再走 cardagg gate）
- 当前需先修才能让现有 Fall 测试落库 — 即使 Gateway 分流方案没落地，sleepad 设备 enablement 也是断的

---

## 6. 与既有北极星的关系

- 与 [`agent_pipeline_north_star.md`](../memory/agent_pipeline_north_star.md) 一致：sensor=Layer 1（rules/sensor 层），cardagg 是适配/持久化层
- 与 [`AI_Iot_stream.md`](AI_Iot_stream.md) 一致：sensor 是 producer，cardagg 是 consumer
- 与 [`alarm_back_channel.go`](../wisefido-sensor/internal/consumer/alarm_back_channel.go) 注释 **冲突** — 旧注释说"不绕过 cardagg 的 enablement gate"，按本文档应改为"sensor 自 gate，cardagg trust"
- 与 [`AI_fall_detect.md`](AI_fall_detect.md) §17 部分冲突 — 旧文档说"cardagg 现有 case alarm.Fall handler 自动落 alarm_events"，按本文档 cardagg 只接受 producer="wisefido-sensor" 的 Fall

修改这两份冲突时引用本文档作为权威。

---

## 7. 不要做的事

- ❌ cardagg 加任何"算法"代码（推导 / 时间窗 / 多源融合）
- ❌ sensor 写 alarm_events 表（PR1 A11 红线，仍然有效）
- ❌ cardagg 处理 radar producer 的 Fall（必须经 sensor verifier）
- ❌ "为了过渡期保 fail-safe" 在 cardagg 留 radar Fall 直处理 — 终态目标必须明确，过渡期可以软落地但不要忘
