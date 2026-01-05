# 雷达 Fall 报警完整流程

## 📋 概述

本文档描述当雷达设备检测到 Fall（跌倒）并上报 MQTT 后，卡片弹出报警的完整数据流和处理流程。

---

## 🔄 完整流程

### 阶段 1：设备上报（MQTT）

```
雷达设备检测到 Fall
    │
    ▼
MQTT Broker (主题: radar/{device_id}/data)
    │
    ▼
wisefido-radar 服务 (MQTT Consumer)
```

**代码位置**：`wisefido-radar/internal/consumer/mqtt_consumer.go`

**处理逻辑**：
1. 订阅 MQTT 主题：`radar/{device_id}/data`
2. 接收 MQTT 消息（包含 `event_type: "Fall"` 或 `posture: "Fall"`）
3. 从主题中提取设备标识符（serial_number 或 uid）
4. 查询设备信息（如果不存在，从 `device_store` 自动创建）
5. 构建标准化数据：
   ```json
   {
     "device_id": "xxx",
     "tenant_id": "xxx",
     "serial_number": "xxx",
     "uid": "xxx",
     "device_type": "Radar",
     "raw_data": {
       "event_type": "Fall",  // 或 posture: "Fall"
       "posture": "Fall",
       "tracking_id": "xxx",
       ...
     },
     "timestamp": 1234567890,
     "topic": "radar/{device_id}/data"
   }
   ```
6. 发布到 Redis Streams：`radar:data:stream`

---

### 阶段 2：数据转换和持久化

```
Redis Streams: radar:data:stream
    │
    ▼
wisefido-data-transformer 服务 (Stream Consumer)
```

**代码位置**：`wisefido-data-transformer/internal/consumer/stream_consumer.go`

**处理逻辑**：
1. 消费 `radar:data:stream` 消息
2. 解析原始设备数据（`RawDeviceData`）
3. 根据设备类型选择转换器（`RadarTransformer`）
4. 转换数据：
   - **如果包含 `event_type: "Fall"`**：
     - 调用 `transformEvent` → 设置 `EventType = "Fall"`
     - 查询 SNOMED 映射 → 设置 `EventSNOMEDCode = "161898004"`
     - 设置 `Category = "safety"`
     - 设置 `DataType = "alarm"`（如果是报警事件）
   - **如果包含 `posture: "Fall"`**：
     - 调用 `transformPosture` → 设置 `PostureSNOMEDCode = "161898004"`
     - 设置 `Category = "activity"`
     - 设置 `DataType = "observation"`（姿态数据）
5. 写入 PostgreSQL `iot_timeseries` 表：
   ```sql
   INSERT INTO iot_timeseries (
     tenant_id, device_id, timestamp,
     event_type, event_snomed_code,  -- 如果 event_type 存在
     posture_snomed_code,             -- 如果 posture 存在
     category, data_type,
     raw_original, ...
   ) VALUES (...)
   ```
6. 发布到输出 Stream：`iot:data:stream`
   ```json
   {
     "iot_timeseries_id": 123,
     "device_id": "xxx",
     "tenant_id": "xxx",
     "device_type": "Radar",
     "timestamp": 1234567890,
     "data_type": "alarm" | "observation",
     "category": "safety" | "activity"
   }
   ```

---

### 阶段 3：数据融合

```
Redis Streams: iot:data:stream
    │
    ▼
wisefido-sensor-fusion 服务 (Stream Consumer)
```

**代码位置**：`wisefido-sensor-fusion/internal/consumer/stream_consumer.go`

**处理逻辑**：
1. 消费 `iot:data:stream` 消息
2. 根据 `device_id` 查询关联的卡片（`cards` 表）
3. 查询该卡片的所有设备数据（从 `iot_timeseries` 表）
4. 融合数据：
   - 如果 `event_type = "Fall"` → 保留在融合结果中（用于报警评估）
   - 如果 `posture = "Fall"` → 添加到 `Postures` 数组：
     ```json
     {
       "postures": [
         {
           "tracking_id": "xxx",
           "posture_code": "161898004",
           "posture_display": "Fall"
         }
       ]
     }
     ```
5. 更新 Redis 实时数据缓存：
   ```
   Key: vital-focus:card:{card_id}:realtime
   Value: {
     "heart": ...,
     "breath": ...,
     "postures": [...],  // 包含 Fall 姿态
     "timestamp": 1234567890
   }
   ```

---

### 阶段 4：报警评估

```
Redis: vital-focus:card:{card_id}:realtime
    │
    ▼
wisefido-alarm 服务 (轮询评估，每 10 秒)
```

**代码位置**：`wisefido-alarm/internal/consumer/cache_consumer.go`

**处理逻辑**：
1. **轮询触发**（每 10 秒）：
   - 获取所有卡片 ID（从 PostgreSQL `cards` 表）
   - 批量处理（每批 10 张卡片）
2. **读取实时数据**：
   - 从 Redis 读取：`vital-focus:card:{card_id}:realtime`
3. **评估报警规则**：
   - **方式 1：设备直接上报 Fall 事件**
     - 如果 `realtimeData` 中包含 `event_type: "Fall"`（从 `iot_timeseries` 融合）
     - 直接生成报警事件（`Fall`，`ALERT` 级别）
   - **方式 2：从姿态数据检测 Fall**
     - 检查 `realtimeData.Postures` 数组
     - 如果存在 `posture_code = "161898004"`（Fall）
     - 生成报警事件（`Fall`，`ALERT` 级别）
4. **生成报警事件**：
   ```go
   event := &models.AlarmEvent{
     EventID:     uuid.New().String(),
     TenantID:    tenantID,
     DeviceID:    deviceID,
     CardID:      cardID,
     EventType:   "Fall",
     Category:    "safety",
     AlarmLevel:  "ALERT",  // 1
     AlarmStatus: "active",
     TriggeredAt: time.Now(),
     TriggerData: {
       "posture_code": "161898004",
       "posture_display": "Fall",
       "tracking_id": "xxx"
     }
   }
   ```
5. **写入 PostgreSQL**：
   ```sql
   INSERT INTO alarm_events (
     event_id, tenant_id, device_id, card_id,
     event_type, category, alarm_level, alarm_status,
     triggered_at, trigger_data, ...
   ) VALUES (...)
   ```
6. **更新 Redis 报警缓存**：
   ```
   Key: vital-focus:card:{card_id}:alarms
   Value: {
     "alarms": [
       {
         "event_id": "xxx",
         "event_type": "Fall",
         "alarm_level": "ALERT",
         "alarm_status": "active",
         "triggered_at": 1234567890
       }
     ]
   }
   TTL: 30 秒
   ```

---

### 阶段 5：前端获取和显示

```
PostgreSQL: alarm_events 表
Redis: vital-focus:card:{card_id}:alarms
    │
    ▼
wisefido-data 服务 (HTTP API)
    │
    ▼
前端 (Vue)
```

#### 5.1 获取卡片列表（包含报警统计）

**API 端点**：`GET /admin/api/v1/monitors/cards`

**代码位置**：
- 后端：`wisefido-data/internal/http/monitor_handler.go`
- 前端：`owlFront/src/api/monitors/monitor.ts`

**处理逻辑**：
1. 前端轮询调用（每 2 秒）
2. 后端查询 `cards` 表，计算未处理报警统计：
   ```sql
   SELECT 
     c.*,
     COUNT(CASE WHEN ae.alarm_level = '0' THEN 1 END) as unhandled_alarm_0,
     COUNT(CASE WHEN ae.alarm_level = '1' THEN 1 END) as unhandled_alarm_1,
     ...
   FROM cards c
   LEFT JOIN alarm_events ae ON (
     c.card_id = ae.card_id 
     AND ae.alarm_status = 'active'
     AND ae.alarm_level IN ('0', '1', '2', '3', '4')
   )
   WHERE c.tenant_id = $1
   GROUP BY c.card_id
   ```
3. 返回卡片列表（包含 `unhandled_alarm_0` 到 `unhandled_alarm_4` 字段）
4. 前端根据 `unhandled_alarm_0`（EMERG）或 `unhandled_alarm_1`（ALERT）显示报警图标

#### 5.2 获取卡片详情（包含最新报警）

**API 端点**：`GET /admin/api/v1/monitors/cards/:card_id`

**处理逻辑**：
1. 查询 `alarm_events` 表，获取该卡片的最新报警：
   ```sql
   SELECT * FROM alarm_events
   WHERE card_id = $1
     AND alarm_status = 'active'
     AND alarm_level <= pop_alarm_emerge  -- 默认 0 (EMERG)
   ORDER BY triggered_at DESC
   LIMIT 1
   ```
2. 返回卡片详情（包含 `alarms` 数组，只包含最新的一个报警）
3. 前端根据 `alarm_level` 和 `alarm_status` 决定是否弹出报警

#### 5.3 前端显示逻辑

**代码位置**：`owlFront/src/views/monitoring/overview/Overview.vue`

**显示逻辑**：
1. **报警图标**：
   - 如果 `unhandled_alarm_0 > 0`（EMERG）→ 显示红色报警图标
   - 如果 `unhandled_alarm_1 > 0`（ALERT）→ 显示橙色报警图标
   - 如果 `unhandled_alarm_2 > 0`（CRIT）→ 显示黄色报警图标
2. **报警弹出**：
   - 如果 `alarms[0].alarm_level <= pop_alarm_emerge`（默认 0）
   - 且 `alarms[0].alarm_status = 'active'`
   - → 弹出报警模态框（Modal）
3. **报警内容**：
   - 显示 `event_type`（如 "Fall"）
   - 显示 `alarm_level`（如 "ALERT"）
   - 显示 `triggered_at`（触发时间）
   - 显示 `trigger_data`（触发数据，如姿态信息）

---

## ⏱️ 时间线

### 当前版本（v1.5+）

| 阶段 | 服务 | 延迟 | 说明 |
|------|------|------|------|
| 1. MQTT 上报 | 雷达设备 → wisefido-radar | ~0.1 秒 | 设备本地检测到 Fall，立即上报 |
| 2. 数据转换 | wisefido-radar → wisefido-data-transformer | ~0.5 秒 | Redis Streams 异步处理 |
| 3. 数据融合 | wisefido-data-transformer → wisefido-sensor-fusion | ~0.5 秒 | Redis Streams 异步处理 |
| 4. 报警评估 | wisefido-sensor-fusion → wisefido-alarm | **0-10 秒** | **轮询间隔 10 秒，最坏情况延迟 10 秒** |
| 5. 前端显示 | wisefido-alarm → 前端 | **0-3 秒** | **前端轮询间隔 3 秒** |

**总延迟**：约 **1-13 秒**（取决于报警评估和前端轮询时机）

### v1.0 版本（wisefido-backend）

**代码位置**：
- Fall 事件处理：`wisefido-backend/wisefido-radar/modules/handler_impl.go:183-199`
- Fall 事件接收：`wisefido-backend/wisefido-radar/socket/server.go:615-647`
- 报警查询：`wisefido-backend/wisefido-data/modules/data_service.go:120-124`

| 阶段 | 服务 | 延迟 | 说明 |
|------|------|------|------|
| 1. 设备上报 | 雷达设备 → wisefido-radar (TCP Socket) | ~0.1 秒 | 设备本地检测到 Fall，通过 TCP Socket 上报 |
| 2. 事件处理 | wisefido-radar → processEventAsAlarm | **~0 秒** | **事件驱动，立即处理，无轮询延迟** |
| 3. 报警创建 | processEventAsAlarm → AddAlarmRecord | ~0.1 秒 | 立即写入数据库 `device_alarm` 表 |
| 4. 前端显示 | wisefido-data API → 前端 | **0-1 秒** | **前端轮询间隔 1 秒**（雷达数据） |

**总延迟**：约 **0.2-1.2 秒**（主要是前端轮询延迟）

**关键发现**：
- v1.0 中 Fall 报警是**事件驱动**的，不是轮询的
- 当雷达设备发送 `EVENT_TYPE_FALL_CONFIRMED` 事件时，`processEventAsAlarm` 函数立即处理
- 立即创建报警记录并写入数据库，**没有轮询延迟**
- 前端通过 HTTP API 查询 `device_alarm` 表获取报警数据

### 版本对比

| 项目 | v1.0 (wisefido-backend) | 当前版本（v1.5+） |
|------|------------------------|------------------|
| **报警触发机制** | **事件驱动**（立即处理） | **轮询**（每 10 秒） |
| **报警评估延迟** | **~0 秒**（事件驱动） | **0-10 秒**（轮询间隔） |
| **前端轮询间隔** | **1 秒** | **3 秒** |
| **前端显示延迟** | 0-1 秒 | 0-3 秒 |
| **总延迟** | **0.2-1.2 秒** | **1-13 秒** |
| **弹出警报时间** | **最快 0.2 秒，最慢 1.2 秒** | **最快 1 秒，最慢 13 秒** |

**关键差异**：
- **v1.0**：Fall 报警是**事件驱动**的，雷达设备检测到 Fall 后立即通过 TCP Socket 上报，`wisefido-radar` 服务立即处理并创建报警记录，**几乎没有延迟**
- **当前版本**：Fall 报警是**轮询评估**的，`wisefido-alarm` 服务每 10 秒轮询一次，**可能有 0-10 秒的延迟**

**代码证据**：
- v1.0 Fall 处理：`wisefido-backend/wisefido-radar/modules/handler_impl.go:183-199`（事件驱动）
- v1.0 前端轮询：`wisefido-frontend/wisefido-platform-vue/src/store/radar/radarData.ts:194`（1 秒）
- 当前版本报警评估：`owlBack/wisefido-alarm/internal/consumer/cache_consumer.go:47`（10 秒轮询）
- 当前版本前端轮询：`owlFront/src/views/monitoring/overview/Overview.vue:577`（3 秒）

---

## 🔑 关键点

### 1. 两种 Fall 检测方式

- **方式 1：设备直接上报 Fall 事件**
  - 雷达设备在 MQTT 消息中包含 `event_type: "Fall"`
  - 经过转换后，`data_type = "alarm"`
  - 报警评估时直接识别为 Fall 事件

- **方式 2：从姿态数据检测 Fall**
  - 雷达设备在 MQTT 消息中包含 `posture: "Fall"`
  - 经过转换后，`posture_snomed_code = "161898004"`
  - 报警评估时从 `Postures` 数组检测 Fall 姿态

### 2. 报警评估延迟

- **轮询间隔**：10 秒
- **最坏情况延迟**：10 秒（如果 Fall 发生在轮询间隔末尾）
- **平均延迟**：约 5 秒

### 3. 前端显示延迟

- **当前版本（v1.5+）前端轮询间隔**：3 秒
  - **代码位置**：`owlFront/src/views/monitoring/overview/Overview.vue:577`
  - **代码**：`const intervalTime = 3 * 1000 // 3 seconds`
  - **最坏情况延迟**：3 秒（如果报警发生在轮询间隔末尾）
  
- **v1.0 版本前端轮询间隔**：1 秒（雷达数据）
  - **代码位置**：`wisefido-frontend/wisefido-platform-vue/src/store/radar/radarData.ts:194`
  - **代码**：`setInterval(() => { refreshRadarData(radarId) }, 1 * 1000)  // 1 秒间隔`
  - **最坏情况延迟**：1 秒（如果报警发生在轮询间隔末尾）

- **报警评估机制对比**：
  - **v1.0**：**事件驱动**（`wisefido-backend/wisefido-radar/modules/handler_impl.go:183-199`）
    - 雷达设备检测到 Fall → 通过 TCP Socket 上报 → `processEventAsAlarm` 立即处理 → 创建报警记录
    - **延迟：~0 秒**（事件驱动，无轮询）
  - **当前版本**：**轮询评估**（`owlBack/wisefido-alarm/internal/consumer/cache_consumer.go:47`）
    - 每 10 秒轮询一次 Redis 实时数据 → 评估报警规则 → 创建报警记录
    - **延迟：0-10 秒**（取决于轮询时机）

- **总延迟对比**：
  - **当前版本**：报警评估延迟（0-10秒）+ 前端轮询延迟（0-3秒）= **1-13 秒**
  - **v1.0 版本**：报警评估延迟（~0秒，事件驱动）+ 前端轮询延迟（0-1秒）= **0.2-1.2 秒**

**结论**：v1.0 版本的弹出警报时间明显更快（0.2-1.2秒 vs 1-13秒），因为：
1. **报警评估是事件驱动的**，没有轮询延迟
2. **前端轮询间隔更短**（1秒 vs 3秒）

---

## 📊 数据流图

```
┌─────────────────────────────────────────────────────────────────┐
│ 1. 雷达设备检测到 Fall                                           │
│    MQTT: radar/{device_id}/data                                 │
│    { "event_type": "Fall" } 或 { "posture": "Fall" }            │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 2. wisefido-radar 服务                                          │
│    - 接收 MQTT 消息                                             │
│    - 查询设备信息                                               │
│    - 发布到 Redis Streams: radar:data:stream                    │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 3. wisefido-data-transformer 服务                               │
│    - 消费 radar:data:stream                                     │
│    - 转换数据（SNOMED 映射）                                    │
│    - 写入 PostgreSQL: iot_timeseries                           │
│    - 发布到 Redis Streams: iot:data:stream                     │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 4. wisefido-sensor-fusion 服务                                  │
│    - 消费 iot:data:stream                                       │
│    - 查询关联的卡片                                             │
│    - 融合卡片的所有设备数据                                     │
│    - 更新 Redis: vital-focus:card:{card_id}:realtime            │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 5. wisefido-alarm 服务（轮询，每 10 秒）                        │
│    - 读取 vital-focus:card:{card_id}:realtime                   │
│    - 评估报警规则（检测 Fall）                                  │
│    - 写入 PostgreSQL: alarm_events                              │
│    - 更新 Redis: vital-focus:card:{card_id}:alarms             │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 6. wisefido-data 服务（HTTP API）                               │
│    - GET /admin/api/v1/monitors/cards                           │
│    - 查询 cards 表 + alarm_events 表（统计未处理报警）          │
│    - 返回卡片列表（包含 unhandled_alarm_0 到 unhandled_alarm_4）│
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│ 7. 前端（Vue，轮询，每 2 秒）                                   │
│    - 调用 getVitalFocusCardsApi()                               │
│    - 根据 unhandled_alarm_0/1 显示报警图标                     │
│    - 如果 alarm_level <= pop_alarm_emerge，弹出报警模态框       │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔍 相关代码文件

### 后端

1. **MQTT 接收**：`wisefido-radar/internal/consumer/mqtt_consumer.go`
2. **数据转换**：`wisefido-data-transformer/internal/transformer/radar.go`
3. **数据融合**：`wisefido-sensor-fusion/internal/fusion/sensor_fusion.go`
4. **报警评估**：`wisefido-alarm/internal/evaluator/evaluator.go`
5. **报警事件写入**：`wisefido-alarm/internal/repository/alarm_events.go`
6. **HTTP API**：`wisefido-data/internal/http/monitor_handler.go`

### 前端

1. **API 调用**：`owlFront/src/api/monitors/monitor.ts`
2. **卡片显示**：`owlFront/src/views/monitoring/overview/Overview.vue`
3. **报警模型**：`owlFront/src/api/monitors/model/monitorModel.ts`

---

## ⚠️ 注意事项

1. **报警评估延迟**：由于轮询间隔为 10 秒，Fall 报警的检测延迟可能在 0-10 秒之间
2. **前端显示延迟**：前端轮询间隔为 2 秒，总延迟可能在 2-12 秒之间
3. **报警缓存 TTL**：Redis 报警缓存 TTL 为 30 秒，确保前端能及时获取最新报警
4. **报警去重**：`wisefido-alarm` 服务会检查最近 5 分钟内的相同报警，避免重复创建

