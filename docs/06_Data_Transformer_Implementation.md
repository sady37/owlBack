# wisefido-data-transformer 服务实现总结

## ✅ 已完成

### 1. 项目结构 ✅

```
wisefido-data-transformer/
├── cmd/wisefido-data-transformer/
│   └── main.go                    # 主程序入口
├── internal/
│   ├── config/
│   │   └── config.go             # 配置管理
│   ├── consumer/
│   │   └── stream_consumer.go    # Redis Streams 消费者
│   ├── models/
│   │   └── stream_message.go     # 数据模型
│   ├── repository/
│   │   ├── snomed.go             # SNOMED 映射仓库
│   │   └── iot_timeseries.go     # IoT 时序数据仓库
│   ├── service/
│   │   └── transformer.go        # 服务主逻辑
│   └── transformer/
│       └── radar.go              # 雷达数据转换器
└── go.mod
```

### 2. 核心功能 ✅

#### 2.1 Redis Streams 消费者 ✅
- 消费 `radar:data:stream` 和 `sleepace:data:stream`
- 使用消费者组模式
- 批量处理消息

#### 2.2 数据转换 ✅
- **SNOMED CT 映射**: 姿态值、事件类型映射到标准编码
- **FHIR Category 分类**: 自动确定数据分类
- **单位转换**: dm → cm
- **数据验证**: 验证数据完整性

#### 2.3 PostgreSQL 写入 ✅
- 写入 `iot_timeseries` 表
- 自动更新位置信息（unit_id, room_id）
- 保留原始数据（raw_original 字段）

#### 2.4 下游触发 ✅
- 发布事件到 `iot:data:stream`（触发下游服务）

---

## 📊 数据流

```
[已实现] Redis Streams (radar:data:stream, sleepace:data:stream)
    ↓
wisefido-data-transformer 服务
    ├─ 解析原始设备数据
    ├─ SNOMED CT 映射（查询 snomed_mapping 表）
    ├─ FHIR Category 分类
    ├─ 单位转换（dm → cm）
    ├─ 数据验证和清洗
    └─→ PostgreSQL TimescaleDB (iot_timeseries 表)
    └─→ Redis Streams (iot:data:stream) - 触发下游服务
```

---

## 🔧 配置

### 环境变量

```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=owlrd
DB_SSLMODE=disable

# Redis 配置
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=

# Stream 配置
STREAM_RADAR=radar:data:stream
STREAM_SLEEPACE=sleepace:data:stream
STREAM_OUTPUT=iot:data:stream

# 消费者配置
CONSUMER_GROUP=data-transformer-group
CONSUMER_NAME=data-transformer-1
```

---

## 📝 实现细节

### 1. 数据转换流程

#### 输入：原始数据（Redis Streams）
```json
{
    "device_id": "uuid",
    "tenant_id": "uuid",
    "device_type": "Radar",
    "raw_data": {
        "posture": 3,
        "position_x": 100,  // dm
        "heart_rate": 75,
        "breath_rate": 18
    },
    "timestamp": 1234567890
}
```

#### 处理：数据标准化
1. **SNOMED CT 映射**
   - `posture: 3` → `posture_snomed_code: "109030009"` (Lying position)
   - 查询 `snomed_mapping` 表获取映射

2. **FHIR Category 分类**
   - 生命体征 → `category: "vital-signs"`
   - 姿态/运动 → `category: "activity"`

3. **单位转换**
   - `position_x: 100` (dm) → `radar_pos_x: 1000` (cm)

4. **数据验证**
   - 验证必填字段
   - 过滤无效数据

#### 输出：标准化数据（PostgreSQL）
```sql
INSERT INTO iot_timeseries (
    tenant_id,
    device_id,
    timestamp,
    data_type,              -- 'observation'
    category,               -- 'vital-signs' or 'activity'
    posture_snomed_code,    -- "109030009"
    posture_display,        -- "Lying position"
    radar_pos_x,            -- 1000 (cm)
    heart_rate,             -- 75 (bpm)
    respiratory_rate,       -- 18 (次/分)
    raw_original            -- 原始数据（JSONB）
) VALUES (...)
```

### 2. SNOMED 映射

- **姿态映射**: 查询 `snomed_mapping` 表（mapping_type = 'posture'）
- **事件映射**: 查询 `snomed_mapping` 表（mapping_type = 'event'）
- **固件版本支持**: 支持固件版本特定的映射

### 3. 位置信息更新

- 从 `devices` 表查询设备位置
- 通过 `bound_bed_id` 或 `bound_room_id` 获取 `room_id` 和 `unit_id`
- 更新 `iot_timeseries` 表的冗余字段

---

## ⚠️ 待完善

### 1. SleepPad 转换器 ⏳
- 当前只实现了 Radar 转换器
- SleepPad 转换器需要单独实现

### 2. 错误处理 ⏳
- 需要更完善的错误处理和重试机制
- 消息确认（ACK）机制

### 3. 性能优化 ⏳
- 批量插入优化
- SNOMED 映射缓存

### 4. 监控和日志 ⏳
- 处理统计
- 错误监控
- 性能指标

---

## 🚀 运行

```bash
cd wisefido-data-transformer
go run cmd/wisefido-data-transformer/main.go
```

---

## 📚 相关文档

- [数据转换服务目的和作用](./05_Data_Transformer_Purpose.md)
- [iot_timeseries 表结构](../../owlRD/db/14_iot_timeseries.sql)
- [snomed_mapping 表结构](../../owlRD/db/19_snomed_mapping.sql)

