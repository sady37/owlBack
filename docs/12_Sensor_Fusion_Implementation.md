# wisefido-sensor-fusion 服务实现总结

## ✅ 已完成的工作

### 1. 项目结构 ✅
- ✅ 创建项目基础结构
- ✅ 配置管理 (`internal/config/config.go`)
- ✅ 数据模型 (`internal/models/iot_timeseries.go`)
- ✅ Repository 层 (`internal/repository/card.go`, `internal/repository/iot_timeseries.go`)
- ✅ 融合逻辑 (`internal/fusion/sensor_fusion.go`)
- ✅ 消费者 (`internal/consumer/stream_consumer.go`, `internal/consumer/cache.go`)
- ✅ 服务主逻辑 (`internal/service/fusion.go`)
- ✅ 主程序 (`cmd/wisefido-sensor-fusion/main.go`)

### 2. 核心功能实现 ✅

#### 2.1 Redis Streams 消费者
- ✅ 消费 `iot:data:stream`（标准化后的设备数据）
- ✅ 使用消费者组模式，支持多实例部署
- ✅ 批量处理消息

#### 2.2 设备到卡片映射
- ✅ 实现 `GetCardByDeviceID`：根据设备ID查询关联的卡片
  - 支持设备绑定到 Bed（查询 ActiveBed 卡片）
  - 支持设备绑定到 Room（查询 Location 卡片）

#### 2.3 传感器融合逻辑
- ✅ **HR/RR 融合**：优先 Sleepace，无数据则 Radar
- ✅ **床状态/睡眠状态融合**：优先 Sleepace
- ✅ **姿态数据融合**：合并所有 Radar 设备的 `tracking_id`（不跨设备去重）

#### 2.4 Redis 缓存更新
- ✅ 更新 `vital-focus:card:{card_id}:realtime` 缓存
- ✅ 设置 TTL（默认 5 分钟）
- ✅ JSON 格式存储融合后的实时数据

## 📊 数据流

```
PostgreSQL (iot_timeseries)
    ↓ (通过 wisefido-data-transformer 写入)
Redis Streams (iot:data:stream)
    ↓
wisefido-sensor-fusion 服务
    ├─ 消费 iot:data:stream
    ├─ 根据 device_id 查询关联的卡片
    ├─ 融合卡片的所有设备数据
    └─→ Redis (vital-focus:card:{card_id}:realtime)
```

## 🎯 融合规则

### 1. HR/RR 融合
- **优先 Sleepace**：如果 Sleepace 设备有数据，使用 Sleepace 数据
- **降级 Radar**：如果 Sleepace 无数据，使用 Radar 数据
- **数据来源标记**：`heart_source` 和 `breath_source` 字段标记数据来源

### 2. 床状态/睡眠状态融合
- **优先 Sleepace**：如果 Sleepace 设备有数据，使用 Sleepace 数据
- **降级 Radar**：如果 Sleepace 无数据，使用 Radar 数据（如果有）

### 3. 姿态数据融合
- **来源**：仅来自 Radar 设备
- **合并规则**：合并所有 Radar 设备的 `tracking_id`
- **去重**：不跨设备去重（同一 tracking_id 在不同设备上视为不同的人）
- **结果**：`person_count` 和 `postures[]` 数组

## 📝 缓存格式

### Redis Key
```
vital-focus:card:{card_id}:realtime
```

### Redis Value (JSON)
```json
{
  "heart": 75,
  "breath": 20,
  "heart_source": "Sleepace",
  "breath_source": "Sleepace",
  "sleep_stage": "248233000",
  "bed_status": "370998004",
  "person_count": 2,
  "postures": [
    {
      "tracking_id": "tracking_001",
      "posture_code": "40199007",
      "posture_display": "Sitting"
    },
    {
      "tracking_id": "tracking_002",
      "posture_code": "248220002",
      "posture_display": "Lying"
    }
  ],
  "timestamp": 1234567890
}
```

## 🔧 配置

### 环境变量

```bash
# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=owlrd

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# Streams
STREAM_INPUT=iot:data:stream

# Consumer
CONSUMER_GROUP=sensor-fusion-group
CONSUMER_NAME=sensor-fusion-1

# Cache
CACHE_REALTIME_PREFIX=vital-focus:card:
```

## 🚀 部署

### 启动服务

```bash
cd wisefido-sensor-fusion
go run cmd/wisefido-sensor-fusion/main.go
```

### 构建

```bash
cd wisefido-sensor-fusion
go build -o bin/wisefido-sensor-fusion cmd/wisefido-sensor-fusion/main.go
```

## 📚 相关文件

- `internal/fusion/sensor_fusion.go` - 传感器融合逻辑
- `internal/consumer/stream_consumer.go` - Redis Streams 消费者
- `internal/consumer/cache.go` - Redis 缓存管理器
- `internal/repository/card.go` - 卡片仓库（设备到卡片映射）
- `internal/repository/iot_timeseries.go` - IoT 时序数据仓库

## 🔄 下一步

1. **测试**：测试传感器融合逻辑
2. **实现 wisefido-alarm**：从 `vital-focus:card:{card_id}:realtime` 读取数据，进行报警评估
3. **实现 wisefido-card-aggregator**：聚合卡片数据

