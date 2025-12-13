# Sleepace 数据流 v1.5（方案 B）

## 📋 概述

v1.5 中，Sleepace 数据采用**方案 B**：统一经过 `wisefido-data-transformer` 处理，保持数据流的一致性。

## 🔄 数据流

```
Sleepad 设备
    ↓
Sleepace 厂家服务（第三方，独立 DB + HTTP API + MQTT）
    ↓
Sleepace 厂家 MQTT Broker (mqtt://47.90.180.176:1883)
    ↓
wisefido-sleepace 服务（v1.5）
    ├─ MQTT 订阅（订阅 Sleepace 厂家 MQTT，保持 v1.0 方式）
    ├─ 查询设备信息（验证设备权限）
    └─→ Redis Streams (sleepace:data:stream) ✅ 新增
        ↓
wisefido-data-transformer 服务
    ├─ 消费 sleepace:data:stream
    ├─ 数据标准化（SNOMED CT映射）
    ├─ FHIR Category 分类
    └─→ PostgreSQL TimescaleDB (iot_timeseries) ✅ 统一格式
    └─→ Redis Streams (iot:data:stream) ✅ 触发下游服务
```

## ✅ 关键变化

### v1.0 → v1.5

| 项目 | v1.0 | v1.5 |
|------|------|------|
| **数据存储** | 直接写入 MySQL (`sleepace_realtime_record`) | 经过 transformer → PostgreSQL (`iot_timeseries`) |
| **数据格式** | 独立表，未标准化 | 统一格式，标准化 |
| **数据流** | MQTT → wisefido-sleepace → MySQL | MQTT → wisefido-sleepace → Streams → transformer → PostgreSQL |
| **MQTT 订阅** | ✅ 保持（订阅 Sleepace 厂家 MQTT） | ✅ 保持（订阅 Sleepace 厂家 MQTT） |

## 🎯 优势

1. **数据统一化**：Sleepace 和 Radar 数据都存储在 `iot_timeseries` 表，格式统一
2. **架构一致性**：所有设备数据都经过 `wisefido-data-transformer` 标准化
3. **代码复用**：转换逻辑集中在 transformer，便于维护
4. **扩展性好**：新增设备类型只需新增转换器
5. **数据一致性**：统一的 SNOMED 映射和 FHIR Category 分类

## 📊 数据格式

### wisefido-sleepace 发布到 Redis Streams 的格式

```json
{
  "device_id": "uuid",
  "tenant_id": "uuid",
  "serial_number": "device_code",
  "uid": "device_uid",
  "device_type": "Sleepace",
  "raw_data": {
    "breath": 20,
    "heart": 75,
    "bedStatus": 0,
    "sleepStage": 2,
    "turnOver": 0,
    "bodyMove": 1,
    "sitUp": 0,
    "initStatus": 1,
    "signalQuality": 95
  },
  "timestamp": 1234567890,
  "topic": "sleepace/realtime"
}
```

### wisefido-data-transformer 转换后的格式

转换后的数据写入 `iot_timeseries` 表，格式与 Radar 数据一致：
- SNOMED CT 编码映射
- FHIR Category 分类
- 单位标准化
- 数据验证和清洗

## 🔧 实现细节

### 1. wisefido-sleepace 服务

**功能**：
- 订阅 Sleepace 厂家 MQTT（保持 v1.0 方式）
- 查询设备信息（验证设备权限）
- 发布数据到 Redis Streams (`sleepace:data:stream`)

**处理的数据类型**：
- `realtime` - 实时数据（主要）
- `sleepStage` - 睡眠阶段
- `connectionStatus` - 连接状态
- `alarmNotify` - 报警通知

### 2. wisefido-data-transformer 服务

**新增功能**：
- 消费 `sleepace:data:stream`
- 实现 `SleepaceTransformer`
- 转换 Sleepace 数据为标准格式

**SleepaceTransformer 转换逻辑**：
- 生命体征：心率、呼吸率（过滤无效值 0/255）
- 床状态：0=在床 → SNOMED "370998004", 1=离床 → SNOMED "424287000"
- 睡眠阶段：0=清醒, 1=浅睡眠, 2=深睡眠, 3=REM睡眠 → SNOMED 编码
- 行为事件：坐起、翻身、体动等
- FHIR Category：根据数据内容自动分类

## 📝 配置

### wisefido-sleepace 配置

```yaml
mqtt:
  broker: "mqtt://47.90.180.176:1883"
  username: "wisefido"
  password: "env(MQTT_PASSWORD)"
  client_id: "wisefido-sleepace"

sleepace:
  topic: "sleepace-57136"  # Sleepace 厂家 MQTT 主题
  stream: "sleepace:data:stream"  # Redis Streams 输出流
```

### wisefido-data-transformer 配置

```yaml
transformer:
  streams:
    radar: "radar:data:stream"
    sleepace: "sleepace:data:stream"  # 新增
    output: "iot:data:stream"
```

## 🚀 部署

1. **启动 wisefido-sleepace 服务**
   ```bash
   cd wisefido-sleepace
   go run cmd/wisefido-sleepace/main.go
   ```

2. **启动 wisefido-data-transformer 服务**
   ```bash
   cd wisefido-data-transformer
   go run cmd/wisefido-data-transformer/main.go
   ```

## 📚 相关文档

- [开发计划更新](./03_Development_Plan_Updated.md)
- [数据转换服务实现](./06_Data_Transformer_Implementation.md)
- [Sleepace v1.0 架构分析](./09_Sleepace_v1.0_Architecture_Analysis.md)

