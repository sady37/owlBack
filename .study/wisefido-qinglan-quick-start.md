# wisefido-qinglan 快速开始指南

本目录包含 wisefido-qinglan 服务的独立部署资源。

## 文件说明

- **wisefido-qinglan-deployment.md** - 完整部署文档（配置说明、部署方式、故障排查）
- **docker-compose-qinglan-standalone.yml** - Docker Compose 独立部署配置
- **mosquitto.conf** - MQTT Broker 配置文件
- **start-qinglan-standalone.sh** - Shell 启动脚本

## 快速开始

### 方式1：Docker Compose 部署（推荐）

```bash
# 1. 复制 Docker Compose 配置到项目根目录
cd /path/to/owlBack
cp .study/docker-compose-qinglan-standalone.yml docker-compose.yml
cp .study/mosquitto.conf ./mosquitto.conf

# 2. 创建证书目录（用于HTTPS）
mkdir -p certs
cd certs

# 生成测试证书
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout server.key -out server.crt -days 365 \
  -subj "/CN=qinglan.example.com"

cd ..

# 3. 修改配置（重要！）
# 编辑 docker-compose.yml，将 RADAR_MQTT_SERVER 改为实际IP
# ⚠️ 不能使用 127.0.0.1 或 localhost，设备无法连接

# 4. 启动所有服务
docker-compose up -d

# 5. 查看日志
docker-compose logs -f wisefido-qinglan

# 6. 停止服务
docker-compose down
```

### 方式2：Shell 脚本启动

```bash
# 1. 进入 wisefido-qinglan 目录
cd /path/to/owlBack/wisefido-qinglan

# 2. 配置环境变量（可选）
export DB_HOST=localhost
export REDIS_ADDR=localhost:6379
export MQTT_BROKER=localhost
export RADAR_MQTT_SERVER=10.0.0.30  # 改为实际IP

# 3. 检查依赖
./start-qinglan-standalone.sh check

# 4. 启动服务
./start-qinglan-standalone.sh start

# 5. 查看状态
./start-qinglan-standalone.sh status

# 6. 查看日志
./start-qinglan-standalone.sh logs

# 7. 停止服务
./start-qinglan-standalone.sh stop
```

### 方式3：直接运行

```bash
# 1. 确保依赖服务运行（PostgreSQL, Redis, MQTT）

# 2. 设置环境变量
export DB_HOST=localhost
export REDIS_ADDR=localhost:6379
export MQTT_BROKER=localhost
export RADAR_MQTT_SERVER=10.0.0.30

# 3. 进入目录
cd /path/to/owlBack/wisefido-qinglan

# 4. 启动服务
go run cmd/wisefido-qinglan/main.go
```

## 验证部署

```bash
# 1. 检查端口监听
netstat -tlnp | grep -E '8081|8443'

# 2. 测试HTTP API
curl http://localhost:8081/health

# 3. 查询设备状态
curl "http://localhost:8081/api/v1/radar/devices/status?tenant_id=YOUR_TENANT_ID"

# 4. 测试设备认证（HTTPS）
curl -k -X POST https://localhost:8443/api/v1/radar/auth \
  -H "Content-Type: application/json" \
  -d '{"uid": "BM872266XXXX", "product_id": 88}'

# 5. 检查 Redis Streams
redis-cli -a TeLunSu-36kr XINFO STREAM iot:monitor:stream
```

## 管理工具（可选）

如果使用 Docker Compose 部署，可以启动管理工具：

```bash
# 启动 pgAdmin 和 RedisInsight
docker-compose --profile tools up -d

# 访问管理界面
# pgAdmin:      http://localhost:5050  (admin@example.com / admin)
# RedisInsight: http://localhost:8001
```

## 常见问题

### 1. 设备无法连接 MQTT

**原因**: `RADAR_MQTT_SERVER` 设置为 `127.0.0.1`

**解决**:
```bash
# 修改为设备能访问的IP
export RADAR_MQTT_SERVER=10.0.0.30
# 或在 docker-compose.yml 中修改
```

### 2. HTTPS 证书错误

**解决**: 生成自签名证书
```bash
cd certs
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout server.key -out server.crt -days 365 \
  -subj "/CN=localhost"
```

### 3. 数据库连接失败

**检查**:
```bash
# 测试数据库连接
psql -h localhost -U postgres -d owlrd -c "SELECT 1"

# 检查数据库是否存在
psql -h localhost -U postgres -l | grep owlrd
```

### 4. Redis 认证失败

**检查**:
```bash
# 测试 Redis 连接
redis-cli -h localhost -p 6379 -a TeLunSu-36kr PING

# 查看 Redis Stream
redis-cli -a TeLunSu-36kr XINFO GROUPS iot:monitor:stream
```

## 性能调优

### 1. 数据库连接池

```bash
export DB_MAX_OPEN_CONNS=25
export DB_MAX_IDLE_CONNS=5
```

### 2. Redis Stream 配置

```bash
export STREAM_MONITOR_MAX_LEN=10000
export STREAM_MONITOR_RETENTION=86400  # 24小时
```

### 3. MQTT QoS

```yaml
# config.yaml
mqtt:
  qos: 1  # 0=最多一次, 1=至少一次, 2=仅一次
```

## 监控

### 健康检查

```bash
# HTTP 健康检查
curl http://localhost:8081/health

# 预期返回: {"status":"ok"}
```

### 日志查看

```bash
# Docker 日志
docker logs -f wisefido-qinglan

# Shell 脚本启动的日志
tail -f /tmp/owlback-logs/wisefido-qinglan.log
```

## 详细文档

完整配置说明、故障排查、高可用部署请参考：
- [完整部署文档](./wisefido-qinglan-deployment.md)
- [API 文档](./wisefido-qinglan-api.md)
- [API 测试脚本](./test_qinglan_api.sh)

## 目录结构

```
.study/
├── wisefido-qinglan-deployment.md          # 完整部署文档
├── wisefido-qinglan-quick-start.md         # 本文件（快速开始）
├── docker-compose-qinglan-standalone.yml   # Docker Compose配置
├── mosquitto.conf                          # MQTT配置
├── wisefido-qinglan-api.md                 # API文档
├── test_qinglan_api.sh                     # API测试脚本
└── architecture.md                         # 系统架构

wisefido-qinglan/
├── cmd/wisefido-qinglan/main.go            # 主程序
├── config.yaml                             # 配置文件
├── Dockerfile                              # Docker镜像
└── start-qinglan-standalone.sh             # 启动脚本
```

## 技术支持

如遇到问题，请：
1. 查看日志文件
2. 检查依赖服务状态
3. 参考完整部署文档
4. 提交 Issue
