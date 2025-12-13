# 开发进度文档

## 已完成的工作

### 1. 数据库设计 ✅
- ✅ 完成所有表结构设计（24个表）
- ✅ 实现完整性约束（17个触发器）
- ✅ 静态验证通过
- ✅ 修复 `residents.unit_id` 的 ON DELETE 策略问题

### 2. 共享库 (owl-common) ✅
- ✅ `database/postgres.go` - PostgreSQL 连接
- ✅ `redis/client.go` - Redis 客户端
- ✅ `redis/streams.go` - Redis Streams 支持（新增）
- ✅ `mqtt/client.go` - MQTT 客户端
- ✅ `logger/logger.go` - 日志工具
- ✅ `config/config.go` - 配置管理

### 3. wisefido-radar 服务 ✅
- ✅ 项目结构创建
- ✅ 配置管理 (`internal/config/config.go`)
- ✅ 服务主逻辑 (`internal/service/radar.go`)
- ✅ MQTT 消费者 (`internal/consumer/mqtt_consumer.go`)
  - ✅ 订阅雷达数据主题 (`radar/+/data`)
  - ✅ 解析设备标识符（serial_number 或 uid）
  - ✅ 验证设备权限（查询 devices 表）
  - ✅ 发布到 Redis Streams (`radar:data:stream`)
- ✅ 设备仓库 (`internal/repository/device.go`)
  - ✅ `GetDeviceBySerialNumber()` - 根据序列号查询
  - ✅ `GetDeviceByUID()` - 根据 UID 查询
- ✅ 主程序入口 (`cmd/wisefido-radar/main.go`)

---

## 数据流实现

### wisefido-radar 服务数据流

```
1. IoT 雷达设备 → MQTT Broker
   └─ 主题: radar/{device_id}/data
   
2. MQTT Broker → wisefido-radar 服务
   └─ MQTT 消费者订阅: radar/+/data
   
3. wisefido-radar 服务处理
   ├─ 从主题提取设备标识符
   ├─ 查询 devices 表验证设备
   └─ 构建标准化数据
   
4. 发布到 Redis Streams
   └─ Stream: radar:data:stream
   └─ 数据格式:
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

---

## 下一步计划

### 1. wisefido-sleepace 服务 ⏳
- [ ] 创建项目结构（参考 wisefido-radar）
- [ ] 实现 MQTT 消费者
- [ ] 实现设备仓库
- [ ] 发布到 Redis Streams (`sleepace:data:stream`)

### 2. wisefido-data-transformer 服务 ⏳
- [ ] 创建项目结构
- [ ] 实现 Redis Streams 消费者
  - [ ] 消费 `radar:data:stream`
  - [ ] 消费 `sleepace:data:stream`
- [ ] 实现数据转换逻辑
  - [ ] SNOMED CT 编码映射
  - [ ] 数据验证和清洗
  - [ ] FHIR Category 分类
- [ ] 写入 PostgreSQL TimescaleDB (`iot_timeseries` 表)
- [ ] 发布事件到 Redis Streams（触发下游服务）

### 3. wisefido-sensor-fusion 服务 ⏳
- [ ] 创建项目结构
- [ ] 实现 Redis Streams 消费者
- [ ] 实现传感器融合逻辑
  - [ ] HR/RR 融合（优先 Sleepace，无数据则 Radar）
  - [ ] 姿态数据融合（合并所有 Radar 的 tracking_id）
- [ ] 更新 Redis 缓存 (`vital-focus:card:{card_id}:realtime`)

### 4. wisefido-alarm 服务 ⏳
- [ ] 创建项目结构
- [ ] 实现 Redis Streams 消费者
- [ ] 实现报警规则评估
  - [ ] 从 `alarm_cloud` 和 `alarm_device` 读取规则
  - [ ] 应用阈值和级别判断
- [ ] 实现 AI 智能评估（可选）
- [ ] 写入 PostgreSQL (`alarm_events` 表)
- [ ] 更新 Redis 缓存 (`vital-focus:card:{card_id}:alarms`)

### 5. wisefido-card-aggregator 服务 ⏳
- [ ] 创建项目结构
- [ ] 实现 Redis Streams 消费者
- [ ] 实现卡片聚合逻辑
  - [ ] 从 PostgreSQL 读取基础信息（cards, devices, residents）
  - [ ] 从 Redis 读取实时数据 (`vital-focus:card:{card_id}:realtime`)
  - [ ] 从 Redis 读取报警数据 (`vital-focus:card:{card_id}:alarms`)
- [ ] 更新 Redis 缓存 (`vital-focus:card:{card_id}:full`)

### 6. wisefido-data 服务 ⏳
- [ ] 创建项目结构
- [ ] 实现 HTTP API 框架（Gin 或 Echo）
- [ ] 实现认证中间件（JWT）
- [ ] 实现权限过滤
- [ ] 实现 API 端点
  - [ ] `GET /data/api/v1/data/vital-focus/cards`
  - [ ] 其他 API 端点
- [ ] 从 Redis 读取卡片数据并返回

---

## 技术栈

- **语言**: Go 1.21+
- **数据库**: PostgreSQL 15+ with TimescaleDB
- **缓存**: Redis 7+
- **消息队列**: MQTT (Eclipse Mosquitto) + Redis Streams
- **API框架**: 待定（Gin / Echo）

---

## 文件结构

```
owlBack/
├── owl-common/              # 共享库 ✅
│   ├── database/
│   ├── redis/
│   │   ├── client.go ✅
│   │   └── streams.go ✅ (新增)
│   ├── mqtt/
│   ├── logger/
│   └── config/
├── wisefido-radar/          # 雷达服务 ✅
│   ├── cmd/wisefido-radar/
│   ├── internal/
│   │   ├── config/
│   │   ├── consumer/
│   │   │   └── mqtt_consumer.go ✅
│   │   ├── repository/
│   │   │   └── device.go ✅
│   │   └── service/
│   │       └── radar.go ✅
│   └── go.mod
├── wisefido-sleepace/       # 睡眠垫服务 ⏳
├── wisefido-data-transformer/ # 数据转换服务 ⏳
├── wisefido-sensor-fusion/   # 传感器融合服务 ⏳
├── wisefido-alarm/           # 报警处理服务 ⏳
├── wisefido-card-aggregator/ # 卡片聚合服务 ⏳
└── wisefido-data/            # API服务 ⏳
```

---

## 注意事项

1. **MQTT 主题格式**:
   - 雷达数据: `radar/{device_id}/data`
   - 睡眠垫数据: `sleepace/{device_id}/data` (待实现)

2. **Redis Streams 命名**:
   - 雷达数据流: `radar:data:stream`
   - 睡眠垫数据流: `sleepace:data:stream` (待实现)
   - 转换后数据流: `iot:data:stream` (待实现)

3. **Redis 缓存键**:
   - 实时数据: `vital-focus:card:{card_id}:realtime` (TTL: 5分钟)
   - 报警数据: `vital-focus:card:{card_id}:alarms` (TTL: 30秒)
   - 完整卡片: `vital-focus:card:{card_id}:full` (TTL: 10秒)

---

## 更新日志

### 2024-01-XX
- ✅ 完成数据库设计验证
- ✅ 创建 owl-common 共享库
- ✅ 完成 wisefido-radar 服务基础实现

