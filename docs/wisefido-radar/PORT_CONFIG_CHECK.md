# wisefido-radar 服务端口配置检查报告

生成时间: 2026-01-11

## 检查结果总结

### ✅ 已正确配置

1. **start-radar.sh 脚本**
   - ✅ 依赖检查端口: 5433（已修改）
   - ✅ 环境变量 DB_PORT: 5433（已修改）

2. **internal/config/config.go**
   - ✅ 从环境变量读取 DB_PORT
   - ✅ 默认值: 5433（与 start_owlback.sh 保持一致）

3. **exports/config.go**
   - ✅ 从环境变量读取 DB_PORT
   - ✅ 默认值: 5433（与 start_owlback.sh 保持一致）

## 配置详情

### start-radar.sh

```bash
# 依赖检查（第75-83行）
# 检查 PostgreSQL（端口 5433，与 start_owlback.sh 保持一致）
if nc -zv 127.0.0.1 5433 > /dev/null 2>&1; then
    echo "✅ PostgreSQL (127.0.0.1:5433) is accessible"
fi

# 环境变量设置（第170-176行）
# 数据库配置（使用 127.0.0.1 避免 IPv6 问题，端口 5433 与 start_owlback.sh 保持一致）
export DB_HOST=127.0.0.1
export DB_PORT=5433
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=owlrd
export DB_SSLMODE=disable
```

### internal/config/config.go

```go
// 从环境变量加载（默认值）
cfg.Database.Host = getEnv("DB_HOST", "localhost")
// 默认端口使用环境变量，如果没有则使用 5433（与 start_owlback.sh 保持一致）
if portStr := getEnv("DB_PORT", ""); portStr != "" {
    if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
        cfg.Database.Port = port
    } else {
        cfg.Database.Port = 5433 // 默认使用 5433（与 start_owlback.sh 保持一致）
    }
} else {
    cfg.Database.Port = 5433 // 默认使用 5433（与 start_owlback.sh 保持一致）
}
```

### exports/config.go

与 `internal/config/config.go` 相同，已从环境变量读取端口。

## 结论

✅ **wisefido-radar 服务的所有配置都已正确设置，没有硬编码问题**

- 脚本中的环境变量: 5433 ✅
- 依赖检查端口: 5433 ✅
- config.go 从环境变量读取: ✅
- 默认值: 5433 ✅

所有配置都与 `start_owlback.sh` 保持一致。
