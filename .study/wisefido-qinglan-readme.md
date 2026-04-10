# wisefido-qinglan

基于 MQTT 协议的雷达设备数据采集与处理服务，集成 Redis Stream 和 PostgreSQL 数据库。

## 功能特性

- ✅ **MQTT 协议支持**：支持雷达设备的 6 种主题订阅
- ✅ **Redis Stream 发布**：实时数据发布到 Redis Streams
- ✅ **数据库集成**：设备信息存储到 PostgreSQL
- ✅ **HTTP API**：提供设备管理 RESTful API
- ✅ **设备自动注册**：自动从 device_store 创建设备
- ✅ **位置信息管理**：设备位置信息查询和缓存
- ✅ **命令响应处理**：支持属性读取/设置、功能调用
- ✅ **实时数据订阅**：支持轨迹和生命体征数据订阅

## 架构设计

### 服务依赖总览

| 依赖 | 类型 | 用途 |
|------|------|------|
| **PostgreSQL** (TimescaleDB, `owlrd`) | 基础设施 | 设备注册表、卡片映射查询 |
| **Redis** (`:6379`) | 基础设施 | Redis Streams 数据发布 + 设备身份缓存 |
| **MQTT Broker** (内部, `:1883`) | 基础设施 | 内部服务消息总线（mosquitto） |
| **Radar Device MQTT** (外部, `10.0.0.30:1883`) | 外部设备 | 物理雷达设备原始数据接入 |
| **wisefido-data** (HTTP `:8081`) | 服务调用方 | 调用 qinglan REST API 读写设备属性 |

### 完整架构图

```
  Physical Radar Devices
  (10.0.0.30:1883 MQTT)
         |
         | raw MQTT messages
         | /prop/{productId}/{uid}/post
         | /monitor/{productId}/{uid}/post
         | /stat /event /alarm /func
         v
+---------------------------+
|    wisefido-qinglan       |
|  (Go service :8081)       |
|                           |
|  +-------------------+   |
|  | MQTT Consumer     |   |
|  | (DeviceSubMgr)    |   |
|  +--------+----------+   |
|           |               |
|  +--------v----------+   |       +----------------+
|  | Radar Decoder     |   |       |   PostgreSQL   |
|  | (binary protocol) |   |<----->|  (owlrd DB)    |
|  +--------+----------+   |       | devices table  |
|           |               |       | cards table    |
|  +--------v----------+   |       +----------------+
|  | StreamPublisher   |   |
|  | (Redis Streams)   |   |       +----------------+
|  +--------+----------+   |       |     Redis      |
|           |               |<----->|  (:6379)       |
|  +--------v----------+   |       | device cache   |
|  | ConfigSubscriber  |   |       +----------------+
|  | (card config)     |   |
|  +-------------------+   |
|                           |
|  +-------------------+   |
|  | HTTP Server       |   |
|  | REST API :8081    |   |
|  +-------------------+   |
+----------+----------------+
           |
           | Redis Streams PUBLISH
           |
    +------+-------+-------+-------+-------+--------+
    |              |       |       |       |        |
    v              v       v       v       v        v
iot:monitor  iot:stat  iot:event iot:alarm iot:auth iot:card
  :stream      :stream   :stream   :stream  :stream  :stream
    |              |       |       |       |
    +-------+------+-------+-------+-------+
            |
     consumed by downstream services
            |
    +-------+-------+----------+
    |               |          |
    v               v          v
wisefido-data  wisefido-     wisefido-
               cardagg         ai

   ^
   | HTTP REST API calls
   | GET/PUT /api/v1/radar/devices/{uid}/properties
   |
wisefido-data
(radar_install_service)

config:card:stream
  (wisefido-data PUBLISH --> wisefido-qinglan CONSUME)
  [card binding/unbinding config changes]
```

### Redis Streams 说明

**生产 (PUBLISH)**：

| Stream | 数据内容 | 保留时间 |
|--------|----------|----------|
| `iot:monitor:stream` | 实时数据（轨迹、生命体征） | 30 秒 |
| `iot:stat:stream` | 统计数据（睡眠统计，1分钟/次） | 5 分钟 |
| `iot:event:stream` | 事件数据（进出、离床等） | 24 小时 |
| `iot:alarm:stream` | 告警数据（跌倒、呼叫等） | 24 小时 |
| `iot:auth:stream` | 设备认证事件 | 24 小时 |
| `iot:card:stream` | 卡片 IoT 数据 | 24 小时 |

**消费 (CONSUME)**：

| Stream | 来源服务 | 用途 |
|--------|----------|------|
| `config:card:stream` | wisefido-data | 卡片绑定/解绑配置变更，触发设备订阅刷新 |

### 关键设计说明

- **双 MQTT 连接**：qinglan 同时维护两条 MQTT 连接——一条连接外部物理雷达设备（`10.0.0.30:1883`），一条连接内部 broker（`:1883`）
- **qinglan 是雷达数据唯一入口**：所有下游服务（wisefido-data、cardagg、AI）通过 Redis Streams 消费，不直接接触雷达设备
- **wisefido-data 双向依赖**：既通过 HTTP 调用 qinglan（设备属性读写），也通过 Redis Streams 消费 qinglan 产出的数据
- **设备身份解析**：qinglan 查询 PostgreSQL 将 `device_uid` 解析为 `(tenant_id, branch_id, unit_id, card_id)`，并附加到每条 Stream 消息

### 支持的 MQTT 主题
1. `/prop/{productId}/{uid}/post` - 属性响应
2. `/monitor/{productId}/{uid}/post` - 实时数据
3. `/func/{productId}/{uid}/post` - 功能响应
4. `/stat/{productId}/{uid}/post` - 统计数据
5. `/event/{productId}/{uid}/post` - 事件数据
6. `/alarm/{productId}/{uid}/post` - 告警数据

## 快速开始

### 1. 环境要求
- Go 1.19+
- PostgreSQL 12+
- Redis 6.2+
- MQTT Broker (EMQX/Mosquitto)

### 2. 配置数据库
```sql
-- 确保以下表存在（来自 owlRD 项目）
-- devices 表
-- device_store 表
-- rooms, beds, units, buildings, branches 表
-- iot_timeseries 表（TimescaleDB 超表）
```

### 3. 启动服务

#### 方式一：使用启动脚本（推荐）
```bash
# 启动服务
./start-qinglan.sh

# 停止服务
./stop-qinglan.sh
```

#### 方式二：手动设置环境变量
```bash
# 设置环境变量（与现有 owlBack 环境一致）
export DB_HOST=127.0.0.1
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable

export REDIS_ADDR=127.0.0.1:6379
export REDIS_PASSWORD=TeLunSu-36kr
export REDIS_DB=0

export MQTT_BROKER=127.0.0.1
export MQTT_PORT=1883
export MQTT_CLIENT_ID=wisefido-qinglan

export RADAR_MQTT_PREFIX=
export RADAR_MQTT_PRODUCT_ID=88

export HTTP_HOST=0.0.0.0
export HTTP_PORT=8081

export LOG_LEVEL=info
export LOG_FORMAT=json

# 启动服务
go run cmd/wisefido-qinglan/main.go
```

#### 方式三：使用配置文件
复制 `config.yaml.example` 为 `config.yaml` 并修改配置：

```yaml
# 数据库配置（与现有环境一致）
database:
  host: "127.0.0.1"
  port: 5433
  user: "postgres"
  password: "postgres"
  dbname: "owlrd"
  sslmode: "disable"

# Redis配置（与现有环境一致）
redis:
  address: "127.0.0.1:6379"
  password: "TeLunSu-36kr"
  db: 0

# MQTT配置（wisefido-qinglan 连接 MQTT）
mqtt:
  broker: "127.0.0.1"
  port: 1883
  username: ""
  password: ""
  client_id: "wisefido-qinglan"
  
  # 雷达设备MQTT主题配置（与 wisefido-radar 一致）
  radar_device_mqtt:
    prefix: ""
    product_id: "88"

http:
  port: 8081
  host: "0.0.0.0"
```

### 4. 构建和运行
```bash
# 构建
go build -o wisefido-qinglan cmd/wisefido-qinglan/main.go

# 运行
./wisefido-qinglan
```

### 5. Docker 运行
```bash
docker build -t wisefido-qinglan .
docker run -p 8081:8081 -v ./config.yaml:/app/config.yaml wisefido-qinglan
```

## HTTP API 文档

### 设备管理

#### 1. 获取设备属性
```http
GET /api/v1/radar/devices/{uid}/properties?keys=key1,key2
```

#### 2. 设置设备属性
```http
PUT /api/v1/radar/devices/{uid}/properties
Content-Type: application/json

{
  "properties": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

#### 3. 订阅实时数据
```http
POST /api/v1/radar/devices/{uid}/subscribe
Content-Type: application/json

{
  "content": 0,  // 0-同时订阅，1-订阅轨迹，2-订阅呼吸心率
  "duration": 3600  // 订阅时长（秒），最大3600
}
```

#### 4. 调用设备功能
```http
POST /api/v1/radar/devices/{uid}/function
Content-Type: application/json

{
  "dev": 0  // 0-重启雷达和主控，1-只重启雷达，2-只重启主控
}
```

#### 5. 获取设备信息
```http
GET /api/v1/radar/devices/{uid}/info
```

#### 6. 获取设备位置信息
```http
GET /api/v1/radar/devices/{uid}/location
```

#### 7. 获取租户设备列表
```http
GET /api/v1/tenants/{tenantId}/devices
```

### 健康检查
```http
GET /health
```

## 数据结构

### Redis Stream 消息格式
```json
{
  "device_id": "uuid",
  "device_uid": "device_uid",
  "device_type": "Radar",
  "tenant_id": "tenant_id",
  "timestamp": 1672531200,
  "topic_type": "monitor",
  "category": "track",
  "data_value": {
    "person_index": 0,
    "coordinate_x": 100,
    "coordinate_y": 200,
    "coordinate_z": 0,
    "posture": 1,
    "event": 0
  },
  "branch_id": "branch_id",
  "building_id": "building_id",
  "unit_id": "unit_id",
  "room_id": "room_id",
  "bed_id": "bed_id"
}
```

### 数据库表结构
参考 `owlRD/db/` 目录下的 SQL 文件：
- `devices.sql` - 设备信息表
- `device_store.sql` - 设备存储表
- `iot_timeseries.sql` - 时间序列数据表（TimescaleDB）

## 开发指南

### 项目结构
```
wisefido-qinglan/
├── cmd/                          # 可执行文件入口
│   └── wisefido-qinglan/
│       └── main.go              # 主程序入口
├── internal/                     # 内部包
│   ├── config/                  # 配置管理
│   ├── consumer/               # MQTT 消费者
│   ├── decode/                 # 数据解码
│   ├── http/                   # HTTP API
│   ├── mqtt/                   # MQTT 客户端
│   ├── repository/             # 数据仓库
│   ├── service/                # 业务服务
│   └── models/                 # 数据模型
├── config.yaml                  # 配置文件
├── go.mod                       # Go 模块定义
├── start-qinglan.sh            # 启动脚本
├── stop-qinglan.sh             # 停止脚本
└── README.md                    # 本文档
```

### 添加新的数据解码器
1. 在 `internal/decode/` 中添加新的解码器
2. 在 `internal/consumer/mqtt_consumer.go` 中集成解码器
3. 更新 Redis Stream 发布逻辑

### 扩展 API
1. 在 `internal/http/api.go` 中添加新的处理器
2. 在 `internal/service/radar_service.go` 中添加业务逻辑
3. 更新路由注册

## 监控和日志

### 日志级别
- INFO: 服务启动、停止、重要事件
- DEBUG: 详细数据处理日志
- ERROR: 错误和异常

### 健康检查
- HTTP: `GET /health`
- 检查 MQTT 连接状态
- 检查数据库连接状态
- 检查 Redis 连接状态

## 故障排除

### 常见问题

1. **MQTT 连接失败**
   - 检查 MQTT Broker 地址和端口
   - 检查用户名和密码
   - 检查网络连接

2. **数据库连接失败**
   - 检查 PostgreSQL 服务状态
   - 检查连接参数
   - 检查数据库权限

3. **Redis 连接失败**
   - 检查 Redis 服务状态
   - 检查地址和端口
   - 检查密码和数据库编号

4. **设备无法自动创建**
   - 检查 device_store 表中是否有设备信息
   - 检查数据库权限
   - 检查设备 UID 格式

### 日志查看
```bash
# 查看服务日志
tail -f /tmp/wisefido_qinglan.log

# 查看错误日志
grep ERROR /tmp/wisefido_qinglan.log
```

## 性能优化

### 缓存策略
- 设备信息缓存 5 分钟
- 位置信息缓存 5 分钟
- 使用 sync.Map 实现线程安全缓存

### 连接池
- 数据库连接池：最大 25 个连接
- Redis 连接池：默认配置
- MQTT 连接：持久连接

### 批量处理
- Redis Stream 批量发布
- 数据库批量插入
- 异步处理非关键任务

## 安全考虑

### 数据安全
- 使用 UUID 作为设备标识
- HIPAA 合规的数据处理
- 敏感信息不记录日志

### 访问控制
- MQTT 认证和授权
- 数据库访问控制
- API 访问日志

### 网络安全
- 使用 TLS/SSL 加密通信
- 防火墙规则配置
- 定期安全更新

## 部署指南

### 生产环境部署
1. 使用 systemd 管理服务
2. 配置日志轮转
3. 设置监控告警
4. 定期备份数据

### Kubernetes 部署
```yaml
# deployment.yaml 示例
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wisefido-qinglan
spec:
  replicas: 2
  selector:
    matchLabels:
      app: wisefido-qinglan
  template:
    metadata:
      labels:
        app: wisefido-qinglan
    spec:
      containers:
      - name: wisefido-qinglan
        image: wisefido-qinglan:latest
        ports:
        - containerPort: 8081
        env:
        - name: DB_HOST
          value: "127.0.0.1"
        - name: DB_PORT
          value: "5433"
        - name: DB_USER
          value: "postgres"
        - name: DB_PASSWORD
          value: "postgres"
        - name: DB_NAME
          value: "owlrd"
        - name: REDIS_ADDR
          value: "127.0.0.1:6379"
        - name: REDIS_PASSWORD
          value: "TeLunSu-36kr"
        - name: MQTT_BROKER
          value: "127.0.0.1"
        - name: MQTT_PORT
          value: "1883"
        - name: RADAR_MQTT_PRODUCT_ID
          value: "88"
```

## 许可证

MIT License

## 支持

如有问题，请提交 Issue 或联系开发团队。