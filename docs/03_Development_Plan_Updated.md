# 开发计划更新

## 📋 背景

已有其他同事实现了 **radar-server 之间的 MQTT 通信**，因此我们不需要重新实现 MQTT 客户端部分。

## ✅ 已由其他同事实现的部分

1. **MQTT 客户端封装** - 已有实现
2. **wisefido-radar 服务** - MQTT 订阅和数据处理

**数据流（已实现）**:
```
Radar 设备 → MQTT Broker（直接）
    ↓
wisefido-radar 服务 → Redis Streams (radar:data:stream)
```

## ⚠️ Sleepace 数据说明（已更新为方案 B）

**Sleepace 数据流（v1.5，方案 B - 统一数据流）**：
```
Sleepad 设备 → Sleepace 厂家服务（第三方，有独立数据库和 HTTP API）
    ↓
Sleepace 厂家服务 → MQTT Broker（厂家提供的 MQTT）
    ↓
wisefido-sleepace 服务（我们的服务，v1.5 格式）
    ├─ MQTT 订阅（订阅 Sleepace 厂家 MQTT，保持 v1.0 方式）
    └─→ Redis Streams (sleepace:data:stream) ✅ 新增
        ↓
wisefido-data-transformer 服务
    ├─ 消费 sleepace:data:stream
    ├─ 数据标准化（SNOMED CT映射）
    └─→ PostgreSQL TimescaleDB (iot_timeseries) ✅ 统一格式
```

**关键点**：
- Sleepace 厂家服务是第三方服务，有独立的数据库和 HTTP API
- wisefido-sleepace 服务订阅 Sleepace 厂家的 MQTT，处理数据
- **数据发布到 Redis Streams，由 wisefido-data-transformer 统一处理**（方案 B）

**影响**：
- ✅ `wisefido-data-transformer` **需要消费** `sleepace:data:stream`
- ✅ 实现 `SleepaceTransformer` 转换器
- ✅ Sleepace 和 Radar 数据统一存储在 `iot_timeseries` 表

## 🎯 我们需要实现的部分

### 1. wisefido-data-transformer 服务 ⏳ (当前任务)

**输入**: Redis Streams
- `radar:data:stream` - 雷达设备数据
- `sleepace:data:stream` - Sleepace 设备数据（方案 B）

**处理**:
- 数据标准化（SNOMED CT 映射）
- 数据验证和清洗
- FHIR Category 分类

**输出**:
- PostgreSQL TimescaleDB (`iot_timeseries` 表)
- Redis Streams 事件（触发下游服务）

---

### 2. wisefido-sensor-fusion 服务 ✅

**输入**: 
- Redis Streams (`iot:data:stream`) - 标准化后的设备数据

**处理**:
- 消费 `iot:data:stream`
- 根据 `device_id` 查询关联的卡片
- 融合卡片的所有设备数据
  - HR/RR：优先 Sleepace，无数据则 Radar
  - 床状态/睡眠状态：优先 Sleepace
  - 姿态：合并所有 Radar 的 `tracking_id`（不跨设备去重）

**输出**:
- Redis `vital-focus:card:{card_id}:realtime` (TTL: 5分钟)

---

### 3. wisefido-alarm 服务 ⏳

**输入**:
- Redis `vital-focus:card:{card_id}:realtime`
- PostgreSQL `alarm_cloud`, `alarm_device`（报警规则）

**处理**:
- 传统规则评估
- AI 智能评估（可选）

**输出**:
- PostgreSQL `alarm_events` 表
- Redis `vital-focus:card:{card_id}:alarms` (TTL: 30秒)

---

### 4. wisefido-card-aggregator 服务 ⏳

**输入**:
- PostgreSQL `cards`, `devices`, `residents` 表（基础信息）
- Redis `vital-focus:card:{card_id}:realtime`（实时数据）
- Redis `vital-focus:card:{card_id}:alarms`（报警数据）

**处理**:
- 聚合所有数据

**输出**:
- Redis `vital-focus:card:{card_id}:full` (TTL: 10秒)

---

### 5. wisefido-data 服务 ⏳

**输入**:
- HTTP 请求（JWT Token）
- Redis `vital-focus:card:{card_id}:full`

**处理**:
- 权限过滤（tenant_id, role, caregiver_id）
- Focus 过滤（users.preferences.vitalFocus.selectedCardIds）

**输出**:
- HTTP 响应（VitalFocusCard[] + filter_counts）

---

## 📊 数据流（完整）

```
[已实现] IoT 设备 → MQTT Broker
    ├─ Radar → wisefido-radar → Redis Streams (radar:data:stream)
    └─ Sleepace → wisefido-sleepace → Redis Streams (sleepace:data:stream) ✅ 方案 B

[待实现] Redis Streams → wisefido-data-transformer
    ├─ 数据标准化（SNOMED CT映射）
    └─→ PostgreSQL TimescaleDB (iot_timeseries)
    └─→ Redis Streams (iot:data:stream)

[已实现] Redis Streams (iot:data:stream) → wisefido-sensor-fusion
    ├─ 消费标准化数据
    ├─ 多传感器融合
    └─→ Redis (vital-focus:card:{card_id}:realtime) ✅

[待实现] Redis → wisefido-alarm
    ├─ 传统规则评估
    ├─ AI智能评估
    └─→ PostgreSQL (alarm_events) + Redis (alarms缓存)

[待实现] Redis → wisefido-card-aggregator
    ├─ 聚合卡片数据
    └─→ Redis (vital-focus:card:{card_id}:full)

[待实现] Redis → wisefido-data (API)
    └─→ HTTP Response (前端)
```

---

## 🔍 需要确认的事项

### 1. Redis Streams 数据格式

需要确认现有实现发布到 Redis Streams 的数据格式：

```go
// 可能的格式
{
    "device_id": "...",
    "tenant_id": "...",
    "serial_number": "...",
    "uid": "...",
    "device_type": "Radar",
    "raw_data": {...},
    "timestamp": 1234567890,
    "topic": "radar/xxx/data"
}
```

### 2. Stream 名称

- `radar:data:stream` - 雷达数据流
- `sleepace:data:stream` - 睡眠垫数据流

### 3. 数据标准化规则

需要了解：
- SNOMED CT 映射规则
- FHIR Category 分类规则
- 数据验证规则

---

## 📝 下一步行动

1. **了解现有实现** ⏳
   - 查看 Redis Streams 数据格式
   - 确认 Stream 名称和数据结构

2. **实现 wisefido-data-transformer 服务** ⏳
   - 消费 Redis Streams
   - 数据标准化
   - 写入 PostgreSQL

3. **实现其他下游服务** ⏳
   - sensor-fusion
   - alarm
   - card-aggregator
   - data (API)

---

## 🎯 优先级

1. **高优先级**: wisefido-data-transformer（数据流的关键节点）
2. **中优先级**: wisefido-sensor-fusion, wisefido-alarm
3. **低优先级**: wisefido-card-aggregator, wisefido-data

