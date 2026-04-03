# wisefido-qinglan 服务独立部署指南

## 服务概述

**wisefido-qinglan** 是雷达设备网关服务，负责：
- 雷达设备认证（HTTPS端口8443）
- MQTT消息消费与转发
- 设备数据发布到Redis Streams
- 设备控制命令下发（HTTP API端口8081）

## 一、依赖服务

### 1. 必需依赖

| 服务          | 默认端口 | 用途                        | 配置项                |
|---------------|---------|----------------------------|----------------------|
| **PostgreSQL** | 5432    | 存储设备信息、卡片映射        | `DB_HOST`, `DB_PORT` |
| **Redis**      | 6379    | Redis Streams消息队列        | `REDIS_ADDR`         |
| **MQTT Broker**| 1883    | 接收设备上报数据             | `MQTT_BROKER`        |

### 2. 数据库表依赖

wisefido-qinglan 需要访问以下数据库表：

```sql
-- 核心表
devices          -- 设备绑定信息（device_id, device_uid, tenant_id, bound_room_id, bound_bed_id）
device_store     -- 设备库存（device_uid, device_code, device_type, allow_access）
cards            -- 卡片映射关系（card_id, device_id）

-- 辅助表（用于位置解析）
rooms            -- 房间信息
beds             -- 床位信息
units            -- 单元/地址信息
alarm_device     -- 设备告警配置（monitor_config.alarms）
```

### 3. Redis Streams 输出

wisefido-qinglan 向以下 Redis Stream 发布消息：

- `iot:monitor:stream` - 实时监测数据
- `iot:stat:stream` - 统计数据
- `iot:event:stream` - 事件数据
- `iot:alarm:stream` - 告警数据
- `iot:auth:stream` - 设备认证事件

### 4. owl-common 依赖

wisefido-qinglan 依赖 `owl-common` 库（项目根目录的 `owl-common` 模块），需要确保：

```bash
# 确保 go.mod 中包含 owl-common 依赖
replace owl-common => ../owl-common
```

## 二、配置方式

### 方式1：环境变量配置（推荐）

```bash
# ========== 数据库配置 ==========
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd           # 数据库名
export DB_SSLMODE=disable

# ========== Redis 配置 ==========
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=TeLunSu-36kr
export REDIS_DB=0

# ========== MQTT 配置 ==========
export MQTT_BROKER=localhost   # qinglan连接的MQTT Broker地址
export MQTT_PORT=1883
export MQTT_CLIENT_ID=wisefido-qinglan
export MQTT_USERNAME=
export MQTT_PASSWORD=

# ========== 雷达设备 MQTT 配置（返回给设备的配置）==========
# ⚠️ 重要：这是返回给设备的MQTT配置，必须是设备能访问的地址
export RADAR_MQTT_SERVER=10.0.0.30        # 设备访问的MQTT服务器地址（不能是127.0.0.1）
export RADAR_MQTT_PORT=8883                # 设备访问的MQTT端口（默认8883加密）
export RADAR_MQTT_ACCOUNT=wfiot
export RADAR_MQTT_PASSWORD=tt@wf@2025
export RADAR_MQTT_PROTOCOL=1               # 1=不加密, 2=加密
export RADAR_MQTT_PRODUCT_ID=88            # 产品ID
export RADAR_MQTT_TIMEOUT=30
export RADAR_MQTT_KEEPALIVE=60

# ========== HTTP 服务配置 ==========
export HTTP_HOST=0.0.0.0
export HTTP_PORT=8081

# ========== HTTPS 服务配置（设备认证）==========
export HTTPS_PORT=8443
export HTTPS_CERT_FILE=/path/to/server.crt
export HTTPS_KEY_FILE=/path/to/server.key
```

### 方式2：配置文件（config.yaml）

```yaml
# wisefido-qinglan/config.yaml
http:
  port: 8081
  host: "0.0.0.0"

https:
  port: 8443
  cert_file: "/path/to/server.crt"
  key_file: "/path/to/server.key"

mqtt:
  broker: "localhost"
  port: 1883
  username: ""
  password: ""
  client_id: "wisefido-qinglan"

  # 雷达设备MQTT配置（返回给设备）
  radar_device_mqtt:
    server: "10.0.0.30"
    port: 8883
    protocol: "1"
    account: "wfiot"
    password: "tt@wf@2025"
    prefix: ""
    product_id: "88"
    timeout: 30
    keepalive: 60
    client_id_prefix: "radar"

redis:
  address: "localhost:6379"
  password: "TeLunSu-36kr"
  db: 0

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "owlrd"
  sslmode: "disable"
```

## 三、部署方式

### 方式1：直接运行（开发环境）

```bash
cd /path/to/owlBack/wisefido-qinglan

# 1. 确保依赖已安装
go mod download

# 2. 使用环境变量配置
export DB_HOST=localhost
export REDIS_ADDR=localhost:6379
export MQTT_BROKER=localhost
export RADAR_MQTT_SERVER=<your-mqtt-server-ip>  # 设备可访问的IP

# 3. 启动服务
go run cmd/wisefido-qinglan/main.go

# 或者编译后运行
go build -o wisefido-qinglan cmd/wisefido-qinglan/main.go
./wisefido-qinglan
```

### 方式2：Docker 部署

```bash
# 1. 构建镜像
cd /path/to/owlBack/wisefido-qinglan
docker build -t wisefido-qinglan:latest .

# 2. 运行容器
docker run -d \
  --name wisefido-qinglan \
  -p 8081:8081 \
  -p 8443:8443 \
  -e DB_HOST=postgres \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=postgres \
  -e DB_NAME=owlrd \
  -e REDIS_ADDR=redis:6379 \
  -e REDIS_PASSWORD=TeLunSu-36kr \
  -e MQTT_BROKER=mosquitto \
  -e MQTT_PORT=1883 \
  -e RADAR_MQTT_SERVER=10.0.0.30 \
  -e RADAR_MQTT_PORT=8883 \
  -e RADAR_MQTT_PASSWORD=tt@wf@2025 \
  -e HTTPS_CERT_FILE=/certs/server.crt \
  -e HTTPS_KEY_FILE=/certs/server.key \
  -v /path/to/certs:/certs:ro \
  wisefido-qinglan:latest
```

### 方式3：Docker Compose 部署

```yaml
# docker-compose.yml
version: '3.8'

services:
  wisefido-qinglan:
    build: ./wisefido-qinglan
    container_name: wisefido-qinglan
    ports:
      - "8081:8081"
      - "8443:8443"
    environment:
      # 数据库配置
      DB_HOST: postgres
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: owlrd
      DB_SSLMODE: disable

      # Redis配置
      REDIS_ADDR: redis:6379
      REDIS_PASSWORD: TeLunSu-36kr
      REDIS_DB: 0

      # MQTT配置
      MQTT_BROKER: mosquitto
      MQTT_PORT: 1883
      MQTT_CLIENT_ID: wisefido-qinglan

      # 雷达设备MQTT配置
      RADAR_MQTT_SERVER: 10.0.0.30  # 改为实际IP
      RADAR_MQTT_PORT: 8883
      RADAR_MQTT_ACCOUNT: wfiot
      RADAR_MQTT_PASSWORD: tt@wf@2025
      RADAR_MQTT_PRODUCT_ID: 88

      # HTTPS配置
      HTTPS_PORT: 8443
      HTTPS_CERT_FILE: /certs/server.crt
      HTTPS_KEY_FILE: /certs/server.key

    volumes:
      - ./certs:/certs:ro
    depends_on:
      - postgres
      - redis
      - mosquitto
    restart: unless-stopped

  postgres:
    image: timescale/timescaledb:latest-pg15
    environment:
      POSTGRES_DB: owlrd
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    command: redis-server --requirepass TeLunSu-36kr
    ports:
      - "6379:6379"

  mosquitto:
    image: eclipse-mosquitto:2
    ports:
      - "1883:1883"
      - "8883:8883"
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf

volumes:
  postgres_data:
```

## 四、证书配置（HTTPS）

wisefido-qinglan 需要 HTTPS 证书用于设备认证（端口8443）。

### 1. 生成自签名证书（测试环境）

```bash
# 生成私钥
openssl genrsa -out server.key 2048

# 生成证书签名请求
openssl req -new -key server.key -out server.csr \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=WiseFido/CN=qinglan.example.com"

# 生成自签名证书（有效期1年）
openssl x509 -req -days 365 -in server.csr \
  -signkey server.key -out server.crt

# 设置环境变量
export HTTPS_CERT_FILE=/path/to/server.crt
export HTTPS_KEY_FILE=/path/to/server.key
```

### 2. 使用正式证书（生产环境）

```bash
# 从证书颁发机构获取证书后
export HTTPS_CERT_FILE=/etc/ssl/certs/your-domain.crt
export HTTPS_KEY_FILE=/etc/ssl/private/your-domain.key
export HTTPS_PORT=8443
```

## 五、启动验证

### 1. 检查服务状态

```bash
# 检查端口监听
netstat -tlnp | grep -E '8081|8443'

# 预期输出：
# tcp6  0  0 :::8081  :::*  LISTEN  12345/wisefido-qinglan
# tcp6  0  0 :::8443  :::*  LISTEN  12345/wisefido-qinglan
```

### 2. 测试 HTTP API（内部控制）

```bash
# 健康检查
curl http://localhost:8081/health

# 查询设备状态
curl "http://localhost:8081/api/v1/radar/devices/status?tenant_id=YOUR_TENANT_ID"

# 查询设备属性
curl "http://localhost:8081/api/v1/radar/devices/BM872266XXXX/properties?keys=install_height"
```

### 3. 测试 HTTPS API（设备认证）

```bash
# 设备认证接口
curl -k -X POST https://localhost:8443/api/v1/radar/auth \
  -H "Content-Type: application/json" \
  -d '{
    "uid": "BM872266XXXX",
    "product_id": 88
  }'
```

### 4. 验证 Redis Streams

```bash
# 连接 Redis
redis-cli -h localhost -p 6379 -a TeLunSu-36kr

# 查看 Stream 数据
XINFO STREAM iot:monitor:stream
XINFO STREAM iot:auth:stream

# 查看最新消息
XREAD COUNT 1 STREAMS iot:monitor:stream 0-0
```

### 5. 查看日志

```bash
# 如果使用 start-owlback.sh 启动
tail -f /tmp/owlback-logs/wisefido-qinglan.log

# Docker 日志
docker logs -f wisefido-qinglan
```

## 六、常见问题

### 1. 设备无法连接 MQTT

**问题**: 设备认证成功但无法连接 MQTT

**原因**: `RADAR_MQTT_SERVER` 配置为 `127.0.0.1` 或 `localhost`

**解决**:
```bash
# 设置为设备能访问的IP地址
export RADAR_MQTT_SERVER=10.0.0.30  # 或服务器的公网IP
```

### 2. HTTPS 证书错误

**问题**: `❌ Failed to create HTTPS server: open /path/to/server.crt: no such file or directory`

**解决**:
```bash
# 生成测试证书
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout server.key -out server.crt -days 365 \
  -subj "/CN=localhost"

# 设置环境变量
export HTTPS_CERT_FILE=$PWD/server.crt
export HTTPS_KEY_FILE=$PWD/server.key
```

### 3. 数据库连接失败

**问题**: `Failed to connect to database: connection refused`

**检查**:
```bash
# 检查数据库是否运行
pg_isready -h localhost -p 5432

# 检查数据库是否存在
psql -h localhost -U postgres -c "\l" | grep owlrd

# 测试连接
psql -h localhost -U postgres -d owlrd -c "SELECT 1"
```

### 4. Redis 连接失败

**问题**: `Failed to connect to Redis: NOAUTH Authentication required`

**解决**:
```bash
# 检查 Redis 密码
redis-cli -h localhost -p 6379 -a TeLunSu-36kr PING

# 如果密码错误，更新环境变量
export REDIS_PASSWORD=correct-password
```

### 5. owl-common 模块找不到

**问题**: `cannot find package "owl-common/..."`

**解决**:
```bash
cd /path/to/owlBack

# 确保 go.mod 包含 replace 指令
grep "replace owl-common" wisefido-qinglan/go.mod

# 如果没有，添加
cd wisefido-qinglan
go mod edit -replace owl-common=../owl-common
go mod tidy
```

## 七、性能优化

### 1. 数据库连接池

```bash
export DB_MAX_OPEN_CONNS=25
export DB_MAX_IDLE_CONNS=5
export DB_CONN_MAX_LIFETIME=300s
```

### 2. Redis Stream 配置

```bash
# 控制 Stream 最大长度和过期时间
export STREAM_MONITOR_MAX_LEN=10000
export STREAM_MONITOR_RETENTION=86400  # 24小时
```

### 3. MQTT QoS 配置

默认 QoS=1，确保消息至少投递一次。如需修改：

```yaml
# config.yaml
mqtt:
  qos: 1  # 0=最多一次, 1=至少一次, 2=仅一次
```

## 八、安全建议

1. **生产环境必须使用正式 HTTPS 证书**
2. **修改默认密码**（Redis、数据库、MQTT）
3. **限制端口访问**：
   - 8081（HTTP）：仅内网可访问
   - 8443（HTTPS）：设备网关，需公网可访问
4. **启用 PostgreSQL SSL**：
   ```bash
   export DB_SSLMODE=require
   ```
5. **定期轮换 MQTT 密码**

## 九、监控与告警

### 1. 健康检查

```bash
# HTTP 健康检查
curl http://localhost:8081/health

# 预期返回: {"status":"ok"}
```

### 2. Prometheus 监控（可选）

wisefido-qinglan 暴露以下指标（如果启用 Prometheus）：

- `qinglan_mqtt_messages_total` - MQTT 消息总数
- `qinglan_redis_stream_publish_total` - Redis Stream 发布总数
- `qinglan_device_auth_total` - 设备认证总数
- `qinglan_http_requests_total` - HTTP 请求总数

### 3. 日志级别

```bash
# 调试模式
export LOG_LEVEL=debug

# 生产模式
export LOG_LEVEL=info
```

## 十、扩展部署

### 高可用部署

wisefido-qinglan 支持水平扩展（多实例），但需要注意：

1. **MQTT Client ID 唯一性**：每个实例使用不同的 `MQTT_CLIENT_ID`
   ```bash
   # 实例1
   export MQTT_CLIENT_ID=wisefido-qinglan-1

   # 实例2
   export MQTT_CLIENT_ID=wisefido-qinglan-2
   ```

2. **Redis Consumer Group**：多实例自动共享消费负载（由 Redis Streams 保证）

3. **负载均衡**：使用 Nginx/HAProxy 对 8081 端口进行负载均衡
   ```nginx
   upstream qinglan_backend {
       server 10.0.0.10:8081;
       server 10.0.0.11:8081;
   }

   server {
       listen 80;
       location / {
           proxy_pass http://qinglan_backend;
       }
   }
   ```

## 十一、相关文档

- [wisefido-qinglan API 文档](.study/wisefido-qinglan-api.md)
- [API 测试脚本](.study/test_qinglan_api.sh)
- [系统架构图](.study/architecture.md)
