# wisefido-data Docker 快速启动

## 启用 Doctor 功能的 Docker 运行方式

### 1. 启动所有服务（包括 wisefido-data）

```bash
cd /Users/sady3721/project/owlBack
docker-compose up -d
```

### 2. 查看服务状态

```bash
docker-compose ps
```

### 3. 查看日志

```bash
# 查看 wisefido-data 日志
docker-compose logs -f wisefido-data

# 查看所有日志
docker-compose logs -f
```

### 4. 测试 Doctor 端点

```bash
# 健康检查
curl http://localhost:8080/health

# 就绪检查
curl http://localhost:8080/ready
```

### 5. 启用 pprof（可选，用于性能分析）

修改 `docker-compose.yml` 中的环境变量：

```yaml
environment:
  - DOCTOR_PPROF=true  # 改为 true
```

然后重启服务：

```bash
docker-compose up -d --build wisefido-data
```

### 6. 停止服务

```bash
# 停止 wisefido-data
docker-compose stop wisefido-data

# 停止所有服务
docker-compose down
```

## Doctor 功能说明

- **DOCTOR_ENABLED=true**（默认）：启用健康检查和就绪检查端点
- **DOCTOR_PPROF=false**（默认）：禁用 pprof，生产环境推荐
- **DOCTOR_PPROF=true**：启用 pprof 性能分析，仅用于开发/调试

## 健康检查端点

- `GET /health` 或 `/healthz` - 健康检查
- `GET /ready` 或 `/readyz` - 就绪检查

## 详细文档

查看 `DOCKER.md` 获取更多详细信息。


