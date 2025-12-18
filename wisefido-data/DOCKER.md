# wisefido-data Docker 运行指南

## 快速启动

### 1. 构建并启动所有服务（包括 wisefido-data）

```bash
cd /Users/sady3721/project/owlBack
docker-compose up -d
```

### 2. 仅启动 wisefido-data 服务

```bash
# 确保依赖服务已启动
docker-compose up -d postgresql redis

# 启动 wisefido-data
docker-compose up -d wisefido-data
```

### 3. 查看日志

```bash
# 查看 wisefido-data 日志
docker-compose logs -f wisefido-data

# 查看所有服务日志
docker-compose logs -f
```

### 4. 停止服务

```bash
# 停止 wisefido-data
docker-compose stop wisefido-data

# 停止所有服务
docker-compose down
```

## Doctor 功能配置

### 默认配置（已启用）

在 `docker-compose.yml` 中，Doctor 功能默认已启用：

```yaml
environment:
  - DOCTOR_ENABLED=true
  - DOCTOR_PPROF=false
```

### 启用 pprof 性能分析

如果需要启用 pprof 性能分析（开发/调试环境），修改 `docker-compose.yml`：

```yaml
environment:
  - DOCTOR_ENABLED=true
  - DOCTOR_PPROF=true  # 启用 pprof
```

然后重启服务：
```bash
docker-compose up -d --build wisefido-data
```

⚠️ **注意**：生产环境建议将 `DOCTOR_PPROF` 设置为 `false`。

### 禁用 Doctor 功能

```yaml
environment:
  - DOCTOR_ENABLED=false
```

## 测试 Doctor 端点

### 健康检查

```bash
# 在容器内
docker exec owl-wisefido-data wget -qO- http://localhost:8080/health

# 从主机
curl http://localhost:8080/health
```

**响应示例：**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "redis": "healthy",
    "database": "healthy"
  }
}
```

### 就绪检查

```bash
curl http://localhost:8080/ready
```

### pprof 性能分析（如果启用）

```bash
# 在主机上使用 go tool pprof
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# 或者查看堆内存
go tool pprof http://localhost:8080/debug/pprof/heap
```

## 环境变量配置

可以在 `docker-compose.yml` 中修改以下环境变量：

| 环境变量 | 默认值 | 说明 |
|---------|--------|------|
| `HTTP_ADDR` | `:8080` | HTTP 服务监听地址 |
| `DB_ENABLED` | `true` | 是否启用数据库 |
| `DB_HOST` | `postgresql` | 数据库主机（容器内使用服务名） |
| `DB_PORT` | `5432` | 数据库端口 |
| `DB_USER` | `postgres` | 数据库用户 |
| `DB_PASSWORD` | `postgres` | 数据库密码 |
| `DB_NAME` | `owlrd` | 数据库名称 |
| `REDIS_ADDR` | `redis:6379` | Redis 地址（容器内使用服务名） |
| `DOCTOR_ENABLED` | `true` | 启用 Doctor 诊断功能 |
| `DOCTOR_PPROF` | `false` | 启用 pprof 性能分析 |
| `LOG_LEVEL` | `info` | 日志级别 |
| `SEED_SYSADMIN` | `true` | 是否自动创建系统管理员 |

## 重建镜像

如果修改了代码，需要重新构建镜像：

```bash
# 重新构建并启动
docker-compose up -d --build wisefido-data

# 或者先构建再启动
docker-compose build wisefido-data
docker-compose up -d wisefido-data
```

## 进入容器调试

```bash
# 进入容器
docker exec -it owl-wisefido-data sh

# 在容器内查看进程
ps aux

# 在容器内测试端点
wget -qO- http://localhost:8080/health
```

## 查看服务状态

```bash
# 查看所有服务状态
docker-compose ps

# 查看 wisefido-data 服务详情
docker-compose ps wisefido-data
```

## 常见问题

### 1. 服务启动失败

**检查依赖服务是否运行：**
```bash
docker-compose ps
```

**检查日志：**
```bash
docker-compose logs wisefido-data
```

### 2. 数据库连接失败

确保 PostgreSQL 服务已启动并且健康：
```bash
docker-compose ps postgresql
```

### 3. Redis 连接失败

确保 Redis 服务已启动：
```bash
docker-compose ps redis
```

### 4. 端口冲突

如果 8080 端口被占用，可以修改 `docker-compose.yml` 中的端口映射：
```yaml
ports:
  - "8081:8080"  # 主机端口:容器端口
```

## 生产环境建议

1. **禁用 pprof**：设置 `DOCTOR_PPROF=false`
2. **使用环境变量文件**：创建 `.env` 文件管理敏感信息
3. **配置日志级别**：生产环境使用 `LOG_LEVEL=warn` 或 `error`
4. **使用健康检查**：Kubernetes 或其他编排工具可以使用 `/healthz` 和 `/readyz` 端点


