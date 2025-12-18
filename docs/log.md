# SaaS 日志管理方案

## 架构设计

### 1. 集中式日志管理（推荐：Loki）

```
应用服务 (wisefido-*) 
  ↓ stdout/stderr (JSON格式)
Promtail (日志收集器)
  ↓ 推送日志
Loki (日志聚合服务)
  ↓ 查询接口
Grafana (可视化)
```

### 2. 多租户日志隔离

每条日志必须包含：
- `tenant_id`: 租户标识（用于多租户隔离和查询）
- `service_name`: 服务名称（如 wisefido-data）
- `timestamp`: 时间戳
- 其他业务字段（如 user_id, device_id 等）

### 3. 日志格式

使用结构化 JSON 格式，便于：
- 按租户查询：`{tenant_id="xxx"}`
- 按服务查询：`{service_name="wisefido-data"}`
- 按错误类型查询：`{level="error"}`
- 组合查询：`{tenant_id="xxx", level="warn"}`

## 实现方案

### 方案 A: Loki + Promtail（推荐）

**优点**：
- 轻量级，资源占用少
- 与 Grafana 完美集成
- 支持多租户查询
- 易于扩展

**配置**：
1. 在 docker-compose.yml 中添加 Loki 和 Promtail
2. 配置 Promtail 收集所有服务的日志
3. 在 Grafana 中配置 Loki 数据源

### 方案 B: ELK Stack（Elasticsearch + Logstash + Kibana）

**优点**：
- 功能强大，支持全文搜索
- 成熟稳定

**缺点**：
- 资源占用较大
- 配置复杂

### 方案 C: 云服务（AWS CloudWatch、Azure Monitor等）

**优点**：
- 托管服务，无需维护
- 自动扩展

**缺点**：
- 成本较高
- 供应商锁定

## 日志字段规范

### 必需字段
- `timestamp`: ISO8601 格式时间戳
- `level`: 日志级别（debug/info/warn/error）
- `service_name`: 服务名称
- `tenant_id`: 租户ID（如果适用）

### 业务字段（根据场景）
- `user_id`: 用户ID
- `user_account`: 用户账号（不包含敏感信息）
- `device_id`: 设备ID
- `ip_address`: IP地址
- `user_agent`: 用户代理
- `reason`: 失败原因
- `error`: 错误详情

### 安全注意事项
- **不要记录**：密码、PHI（受保护健康信息）、完整账号信息
- **可以记录**：账号hash、用户ID、设备ID、时间戳

## 查询示例

### Grafana + Loki 查询

```logql
# 查询特定租户的所有错误日志
{tenant_id="xxx"} |= "error"

# 查询特定服务的登录失败日志
{service_name="wisefido-data"} |~ "login failed"

# 查询特定租户的设备连接失败
{tenant_id="xxx", service_name="wisefido-data"} |~ "device connection"

# 按时间范围查询
{tenant_id="xxx"} [1h]
```

## 部署建议

### 开发环境
- 使用 Docker Compose 部署 Loki + Promtail
- 日志保留 7 天

### 生产环境
- 使用 Kubernetes 部署 Loki 集群
- 日志保留 30-90 天（根据合规要求）
- 配置日志备份和归档
- 设置日志告警（如错误率过高）

## 合规性考虑（HIPAA）

1. **日志访问控制**：只有授权人员可以访问日志
2. **日志加密**：传输和存储时加密
3. **审计日志**：记录谁访问了哪些日志
4. **日志保留**：根据合规要求保留一定时间
5. **敏感信息**：日志中不包含 PHI（受保护健康信息）

## 当前日志存储位置

### 开发环境
- **位置**：Docker 容器标准输出（stdout/stderr）
- **查看方式**：
  ```bash
  # 查看 wisefido-data 服务的日志
  docker logs owl-wisefido-data
  
  # 实时跟踪日志
  docker logs -f owl-wisefido-data
  
  # 查看最近 100 行
  docker logs --tail 100 owl-wisefido-data
  
  # 查看特定时间段的日志
  docker logs --since 1h owl-wisefido-data
  ```

### 生产环境（推荐：Loki）
- **位置**：Loki 集中式日志聚合服务
- **收集方式**：Promtail 自动收集所有容器的日志
- **查看方式**：通过 Grafana 界面查询和可视化

## 使用方法

### 1. 启动日志服务

```bash
# 启动基础服务 + 日志服务
docker-compose -f docker-compose.yml -f docker-compose.logging.yml up -d
```

### 2. 访问 Grafana

- **地址**：http://localhost:3000
- **用户名**：admin
- **密码**：admin（首次登录后需要修改）

### 3. 配置服务使用新的 Logger

所有服务需要更新 logger 初始化代码，添加 `serviceName` 参数：

```go
// 旧代码
logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format)

// 新代码（SaaS多租户日志管理）
logger, err := logpkg.NewLogger(cfg.Log.Level, cfg.Log.Format, "wisefido-data")
```

### 4. 日志记录示例

#### 人员登录失败日志

```go
s.Logger.Warn("User login failed: invalid credentials",
    zap.String("tenant_id", tenantID),
    zap.String("user_type", normalizedUserType),
    zap.String("ip_address", getClientIP(r)),
    zap.String("user_agent", r.UserAgent()),
    zap.String("reason", "invalid_credentials"),
)
```

#### 设备登录失败日志

```go
logWarn("Device connection rejected: not allocated",
    zap.String("device_store_id", dsDeviceStoreID),
    zap.String("serial_number", serialNum),
    zap.String("uid", uid),
    zap.String("tenant_id", dsTenantID),  // 如果可用
    zap.String("reason", "device_not_allocated"),
    zap.String("action", "connection_rejected"),
)
```

## 日志查询场景

### 场景 1: 查询特定租户的所有登录失败

```logql
{tenant_id="00000000-0000-0000-0000-000000000001"} 
  |~ "login failed"
```

### 场景 2: 查询所有服务的错误日志

```logql
{level="error"}
```

### 场景 3: 查询特定设备的连接日志

```logql
{service_name="wisefido-data"} 
  |~ "device connection"
  |~ "device_id_here"
```

### 场景 4: 查询特定时间段的登录活动

```logql
{service_name="wisefido-data"} 
  |~ "login"
  [1h]
```

## 配置文件说明

### docker-compose.logging.yml

包含以下服务：
- **loki**: 日志聚合服务（端口 3100）
- **promtail**: 日志收集器
- **grafana**: 日志可视化（端口 3000）

### logging/promtail-config.yml

Promtail 配置，自动收集所有 `owl-` 开头的容器日志。

### logging/grafana-datasources.yml

Grafana 数据源配置，自动连接 Loki。

## 注意事项

1. **日志格式**：确保所有服务使用 JSON 格式输出（`LOG_FORMAT=json`）
2. **服务名称**：每个服务必须使用唯一的 `service_name`
3. **租户ID**：在业务日志中始终包含 `tenant_id` 字段
4. **敏感信息**：不要在日志中记录密码、完整账号等敏感信息
5. **日志级别**：生产环境建议使用 `info` 级别，开发环境可以使用 `debug`

## 故障排查

### 问题：看不到日志

1. 检查容器是否运行：`docker ps`
2. 检查 Promtail 是否连接 Loki：`docker logs owl-promtail`
3. 检查 Loki 是否正常：`curl http://localhost:3100/ready`

### 问题：日志查询慢

1. 检查 Loki 存储空间
2. 考虑增加日志保留时间限制
3. 使用更精确的查询条件

### 问题：Grafana 无法连接 Loki

1. 检查 `logging/grafana-datasources.yml` 配置
2. 确认 Loki 服务名称正确（`loki:3100`）
3. 检查网络连接：`docker network ls`


