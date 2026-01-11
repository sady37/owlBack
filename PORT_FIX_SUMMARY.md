# 端口配置修复总结

生成时间: 2026-01-11

## 修复完成 ✅

所有服务的端口硬编码问题已修复。

### 修复的服务

#### 1. wisefido-iot-timeseries ✅

**文件**: `internal/config/config.go`
- ✅ 添加 `strconv` 导入
- ✅ 从环境变量读取 `DB_PORT`
- ✅ 默认值: 5433

**文件**: `start-iot-timeseries.sh`
- ✅ 依赖检查端口: 5432 → 5433
- ✅ 环境变量 `DB_PORT`: 5432 → 5433

#### 2. wisefido-ai ✅

**文件**: `internal/config/config.go`
- ✅ 添加 `strconv` 导入
- ✅ 从环境变量读取 `DB_PORT`
- ✅ 默认值: 5433

#### 3. wisefido-radar/test-startup.sh ✅

**文件**: `test-startup.sh`
- ✅ 环境变量 `DB_PORT`: 5432 → 5433

---

## 之前已修复的服务

### wisefido-radar ✅
- ✅ `start-radar.sh`: 依赖检查端口 5433，环境变量 DB_PORT=5433
- ✅ `internal/config/config.go`: 从环境变量读取，默认 5433
- ✅ `exports/config.go`: 从环境变量读取，默认 5433

### wisefido-sleepace ✅
- ✅ `start-sleepace.sh`: 依赖检查端口 5433，环境变量 DB_PORT=5433
- ✅ `internal/config/config.go`: 从环境变量读取，默认 5433

### wisefido-card-aggregator ✅
- ✅ `internal/config/config.go`: 从环境变量读取，默认 5433

### wisefido-card-manage ✅
- ✅ `internal/config/config.go`: 从环境变量读取，默认 5433

### wisefido-data ✅
- ✅ `internal/config/config.go`: 从环境变量读取，默认值改为 5433

---

## 修复模式

所有服务统一使用以下模式：

```go
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

---

## 验证结果

所有服务的配置都已统一：
- ✅ 从环境变量读取 `DB_PORT`
- ✅ 默认值: 5433（与 `start_owlback.sh` 保持一致）
- ✅ 启动脚本中的环境变量: `DB_PORT=5433`
- ✅ 依赖检查端口: 5433

---

## 总结

✅ **所有硬编码问题已修复**

所有服务现在都：
1. 从环境变量读取数据库端口
2. 默认使用 5433 端口
3. 与 `start_owlback.sh` 的配置保持一致
