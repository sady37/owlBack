# 所有服务端口配置检查总结

生成时间: 2026-01-11

## 检查结果

### ✅ 已正确配置的服务

1. **wisefido-radar**
   - ✅ start-radar.sh: DB_PORT=5433
   - ✅ internal/config/config.go: 从环境变量读取，默认 5433
   - ✅ exports/config.go: 从环境变量读取，默认 5433

2. **wisefido-sleepace**
   - ✅ start-sleepace.sh: DB_PORT=5433
   - ✅ internal/config/config.go: 从环境变量读取，默认 5433

3. **wisefido-card-aggregator**
   - ✅ internal/config/config.go: 从环境变量读取，默认 5433

4. **wisefido-card-manage**
   - ✅ internal/config/config.go: 从环境变量读取，默认 5433

5. **wisefido-data**
   - ✅ internal/config/config.go: 从环境变量读取（使用 parseInt）

### ❌ 需要修复的服务

1. **wisefido-iot-timeseries**
   - ❌ internal/config/config.go: 硬编码 `cfg.Database.Port = 5432` (Line 37)
   - ❌ start-iot-timeseries.sh: 硬编码 `export DB_PORT=5432` (Line 121)

2. **wisefido-ai**
   - ❌ internal/config/config.go: 硬编码 `cfg.Database.Port = 5432` (Line 46)

3. **wisefido-radar/test-startup.sh**
   - ❌ test-startup.sh: 硬编码 `export DB_PORT=5432` (Line 24)

## 详细检查结果

### wisefido-radar 服务 ✅

**start-radar.sh**:
- 依赖检查端口: 5433 ✅
- 环境变量 DB_PORT: 5433 ✅

**config.go 文件**:
- internal/config/config.go: 从环境变量读取，默认 5433 ✅
- exports/config.go: 从环境变量读取，默认 5433 ✅

**结论**: ✅ 没有硬编码问题

---

### wisefido-iot-timeseries 服务 ❌

**问题**:
- internal/config/config.go Line 37: `cfg.Database.Port = 5432` (硬编码)
- start-iot-timeseries.sh Line 121: `export DB_PORT=5432` (硬编码)

**需要修复**: 改为从环境变量读取，默认 5433

---

### wisefido-ai 服务 ❌

**问题**:
- internal/config/config.go Line 46: `cfg.Database.Port = 5432` (硬编码)

**需要修复**: 改为从环境变量读取，默认 5433

---

### wisefido-radar/test-startup.sh ❌

**问题**:
- test-startup.sh Line 24: `export DB_PORT=5432` (硬编码)

**需要修复**: 改为 5433

---

## 修复建议

### 1. wisefido-iot-timeseries

**config.go**: 参考 wisefido-card-manage 的实现
**start-iot-timeseries.sh**: 将 `export DB_PORT=5432` 改为 `export DB_PORT=5433`

### 2. wisefido-ai

**config.go**: 参考 wisefido-card-manage 的实现

### 3. wisefido-radar/test-startup.sh

将 `export DB_PORT=5432` 改为 `export DB_PORT=5433`

---

## 总结

- ✅ **wisefido-radar 服务**: 所有配置都已正确，没有硬编码问题
- ❌ **其他服务**: 发现 3 个服务/脚本需要修复
