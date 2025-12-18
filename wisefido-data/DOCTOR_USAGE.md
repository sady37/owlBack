# Doctor (诊断功能) 使用指南

Doctor 是 wisefido-data 服务的诊断功能，提供健康检查、就绪检查和性能分析工具。

## 启用方式

### 1. 通过环境变量启用（默认启用）

```bash
# 启用诊断功能（默认：true）
export DOCTOR_ENABLED=true

# 启用 pprof 性能分析（默认：false）
export DOCTOR_PPROF=true

# 运行服务
go run cmd/wisefido-data/main.go
```

### 2. 禁用诊断功能

```bash
export DOCTOR_ENABLED=false
go run cmd/wisefido-data/main.go
```

## 功能说明

### 1. 健康检查端点

**端点：**
- `GET /health`
- `GET /healthz` (Kubernetes 兼容)

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

**状态码：**
- `200 OK` - 所有服务健康
- `503 Service Unavailable` - 有服务不健康

### 2. 就绪检查端点

**端点：**
- `GET /ready`
- `GET /readyz` (Kubernetes 兼容)

**响应示例：**
```json
{
  "ready": true,
  "checks": {
    "redis": true,
    "database": true
  }
}
```

**用途：**
- Kubernetes liveness/readiness probes
- 负载均衡器健康检查

### 3. pprof 性能分析

**启用方式：**
```bash
export DOCTOR_PPROF=true
```

**可用端点：**
- `GET /debug/pprof/` - 性能分析主页
- `GET /debug/pprof/profile` - CPU 性能分析
- `GET /debug/pprof/heap` - 堆内存分析
- `GET /debug/pprof/goroutine` - Goroutine 分析
- `GET /debug/pprof/allocs` - 内存分配分析
- `GET /debug/pprof/block` - 阻塞分析
- `GET /debug/pprof/mutex` - 互斥锁分析
- `GET /debug/pprof/trace` - 执行追踪

**使用示例：**
```bash
# 查看性能分析主页
curl http://localhost:8080/debug/pprof/

# 生成 30 秒的 CPU 分析报告
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# 查看堆内存使用
go tool pprof http://localhost:8080/debug/pprof/heap

# 查看 goroutine 信息
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

## 完整配置示例

```bash
# HTTP 服务配置
export HTTP_ADDR=:8080

# 数据库配置
export DB_ENABLED=true
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd

# Redis 配置
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=

# Doctor 配置
export DOCTOR_ENABLED=true
export DOCTOR_PPROF=true

# 运行服务
go run cmd/wisefido-data/main.go
```

## 安全注意事项

⚠️ **生产环境建议：**

1. **pprof 功能应限制访问**
   - 仅在内网环境启用
   - 使用反向代理限制访问
   - 或通过防火墙限制访问源 IP

2. **健康检查端点可以公开**
   - `/health` 和 `/ready` 不包含敏感信息
   - 适合用于监控和负载均衡

3. **禁用 pprof（生产环境默认）**
   ```bash
   export DOCTOR_PPROF=false
   ```

## Kubernetes 集成示例

```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: wisefido-data
    image: wisefido-data:latest
    env:
    - name: DOCTOR_ENABLED
      value: "true"
    - name: DOCTOR_PPROF
      value: "false"  # 生产环境禁用
    livenessProbe:
      httpGet:
        path: /healthz
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /readyz
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
```


