# Sleepace 数据上报链路分析

> 分析时间：2026-04-08

## 完整数据流

```mermaid
flowchart LR
    DEV[睡眠板硬件] -->|私有协议| JAVA[sleepace-service Java :8090]
    JAVA -->|MQTT publish| MQ[Mosquitto\ntopic: sleepace-57136]
    MQ -->|Subscribe QoS=1| WS[wisefido-sleepace\nmain.go:148]
    WS -->|XADD| R1[iot:monitor:stream\niot:event:stream\niot:alarm:stream]
    R1 -->|XREADGROUP\ncardagg-group| CA[wisefido-cardagg]
    CA -->|HSET + XADD| R2[card:state:{card_id} Hash\ncard:realtime:stream\ncard:status:stream]
    R1 -->|XREADGROUP\niot-timeseries-group| IOT[wisefido-iot → PostgreSQL]
    R2 -->|XREADGROUP| DATA[wisefido-data HTTP API]
```

## 各阶段说明

### Stage 1：设备 → sleepace-service (Java) → MQTT

- 睡眠板硬件通过私有协议与外部 Java 中间件 `sleepace-service`（:8090）通信
- Java 服务负责设备注册、鉴权，并将实时数据作为 MQTT 消息发布
- MQTT Broker：`owl-mqtt` 容器（eclipse-mosquitto:2.0），端口 1883/8883/9001
- **Topic：`sleepace-57136`**（来自 `sleepace-dev.yaml:mqtt.topic_id`）

### Stage 2：wisefido-sleepace — MQTT 订阅 + 消息转换

**入口：** `wisefido-sleepace/cmd/wisefido-sleepace/main.go`

- 订阅单一 topic `sleepace-57136`，QoS=1，所有 dataKey 复用同一 topic
- MQTT 回调仅将原始消息投入 `msgCh (cap=1000)` 缓冲通道
- 10 个 goroutine worker 异步消费，调用 `handleMessage()`
- 通过 PostgreSQL 解析 `deviceId → (tenantID, cardID, deviceUID, deviceCode...)`

#### dataKey 路由规则

| dataKey | 解析结构体 | 目标 Stream | 备注 |
|---|---|---|---|
| `realtime` | `RealtimeData` | `iot:monitor:stream` | heart/breath=255 过滤；bedStatus 变化额外写 event |
| `inBedStatus` | `InBedStatusData` | `iot:event:stream` | inbedStatus=8 丢弃 |
| `sleepStage` | `SleepStageData` | `iot:event:stream` | |
| `pressureSenSor` | `SensorData` | `iot:event:stream` | |
| `analysis` | `AnalysisData` | `iot:event:stream` | 同时触发异步 goroutine 下载报告 |
| `upgradeProgress` | `UpgradeProgressData` | `iot:event:stream` | |
| `connectionStatus` | `ConnectionStatusData` | `iot:alarm:stream` | 0=offline→AlarmTypeOffline |
| `deviceSenSor` | `SensorData` | `iot:alarm:stream` | 0=detached |
| `alarmNotify` | `AlarmNotifyData` | `iot:alarm:stream` | 通过 mqttAlarmMap 映射 |

#### alarmNotify 类型映射

| MQTT type | 内部 EventName |
|---|---|
| `alarmLeftBed` | `alarm.LeftBed` |
| `alarmHeartRateFast` | `alarm.HeartRateAlertHigh` |
| `alarmHeartRateSlow` | `alarm.HeartRateAlertLow` |
| `alarmBreathRateFast` | `alarm.RespRateAlertHigh` |
| `alarmBreathRateSlow` | `alarm.RespRateAlertLow` |
| `alarmBreathRatePause` | `alarm.ApneaHypopnea` |
| `alarmBodymove` | `alarm.AbnormalBodyMovement` |
| `alarmNoBodymove` | `alarm.NoBodyMove` |
| `alarmNoTurnOver` | `alarm.NoTurnOver` |
| `alarmBedSitup` / `alarmSitup` | `alarm.BedSitUp` |
| `alarmInBed` / `alarmOnBed` | `alarm.InBed` |
| `alarmSensorFall` | `alarm.SensorDetached` |

### Stage 3：Redis Streams（由 wisefido-sleepace 写入）

Stream 名称定义于 `owl-common/redis/stream_names.go`。

| Stream | 类型 | MaxLen | 保留时长 | 消息格式 |
|---|---|---|---|---|
| `iot:monitor:stream` | Stream | 1000 | 30s | `IoTStreamMessage{TopicType:"monitor"}` |
| `iot:event:stream` | Stream | 500 | 24h | `IoTStreamMessage{TopicType:"event"}` |
| `iot:alarm:stream` | Stream | 500 | 24h | `IoTStreamMessage{TopicType:"alarm"}` |

**XADD 字段：** `device_uid`, `device_id`, `device_type`, `card_id`, `tenant_id`, `timestamp`(wall clock ms), `topic_type`, `category`, `dataValue`(JSON)

> 注意：`timestamp` 使用 `time.Now().UnixMilli()`（服务器时钟），设备原始时间戳保存在 `dataValue.event_since` 中。

### Stage 4：wisefido-cardagg — 流聚合 & 状态机

- Consumer group：`cardagg-group`，`XREADGROUP BLOCK 2s COUNT 10`
- 订阅全部 6 个流：`iot:monitor/event/alarm:stream` + 3 个 config 流
- `MonitorHandler` 累积实时数据到 `MonitorBuffer`
- 独立 `runDeriveLoop()` goroutine 每 1s tick 推导状态
- `card.Writer.WriteCardStatus()` 使用 `TxPipeline` 原子写：`HSET` + `XADD`

### Stage 5：最终 Redis 状态（由 wisefido-cardagg 写入）

| Redis Key | 类型 | 写入函数 | 消费方 | 保留 |
|---|---|---|---|---|
| `card:state:{card_id}` | Hash | `Writer.WriteCardStatus()` | `wisefido-data` HTTP 读 | 无过期 |
| `card:realtime:stream` | Stream | `Writer.PublishMonitor()` | `wisefido-data-consumer` | 6s / MaxLen 5000 |
| `card:status:stream` | Stream | `Writer.WriteCardStatus()` | `wisefido-data-consumer` | 12h / MaxLen 2000 |

## 所有 Redis Key 汇总

| Redis Key | 类型 | 写入服务 | 消费方 | 保留 |
|---|---|---|---|---|
| `iot:monitor:stream` | Stream | wisefido-sleepace | cardagg-group, iot-timeseries-group | 30s / MaxLen 1000 |
| `iot:event:stream` | Stream | wisefido-sleepace | cardagg-group, iot-timeseries-group | 24h / MaxLen 500 |
| `iot:alarm:stream` | Stream | wisefido-sleepace | cardagg-group, iot-timeseries-group | 24h / MaxLen 500 |
| `card:state:{card_id}` | Hash | wisefido-cardagg | wisefido-data HTTP | 无过期 |
| `card:realtime:stream` | Stream | wisefido-cardagg | wisefido-data-consumer | 6s / MaxLen 5000 |
| `card:status:stream` | Stream | wisefido-cardagg | wisefido-data-consumer | 12h / MaxLen 2000 |

## 风险与建议

### 🔴 Critical

1. **`card:realtime:stream` 保留仅 6 秒**：`wisefido-data` 重启期间数据永久丢失。Consumer group 模式下 PEL 无 XAUTOCLAIM 清理，长期运行 PEL 会无限增长。

2. **`deviceId` 字段二义性**：MQTT payload 中的 `deviceId` 可能是 `device_uid` 或 `device_code`（固件版本差异，见 `mqtt_consumer.go:213` 注释）。解析失败时 `device_id` 为空，`wisefido-iot` 的 PostgreSQL 写入会静默失败。

### 🟡 Warning

1. **`sleepace-57136` 硬编码业务 ID**：Topic 内嵌 appId `57136`，不支持多 channel 场景，无通配符订阅。

2. **设备未注册时静默丢弃**：`deviceID==""` 时多处 `dispatch()` 直接 return，仅打 DEBUG 日志，排查困难。

3. **`iot:monitor:stream` 30s 保留过短**：`wisefido-iot` DB 写入慢时会丢失时序数据。

4. **`analysis` dataKey 触发无限制 goroutine**：无并发上限，突发时可能产生大量并发 HTTP 调用 sleepace-service。

### 🟢 Good Practices

1. **Stream 名称单一来源**：`owl-common/redis/stream_names.go` 统一定义，服务通过常量引用，无字符串重复。

2. **时间戳纪律**：服务器时钟写入 envelope，设备原始时间保存于 `dataValue.event_since`，防止下游误判过时数据。

3. **MQTT 回调解耦**：回调仅投队列，10 worker 异步处理，MQTT 交付延迟与业务处理解耦。

4. **边沿触发事件**：`lastBedStatus` 互斥锁保护的状态图，仅在状态转换时写 event stream，减少下游噪声。

5. **原子 Hash+Stream 写入**：`TxPipeline (MULTI/EXEC)` 保证 `card:state` Hash 与 `card:status:stream` 永远同步。
