# wisefido-ai 单元测试指南

## ✅ 测试状态

### 当前测试结果

- **总测试数**: 15
- **通过**: 15 ✅
- **失败**: 0
- **总体覆盖率**: ~30%

### 各模块覆盖率

| 模块 | 覆盖率 | 状态 |
|------|--------|------|
| `internal/config` | 100.0% | ✅ 完全覆盖 |
| `internal/consumer` | 29.9% | ⚠️ 部分覆盖 |
| `internal/evaluator` | 11.5% | ⚠️ 待完善 |
| `internal/repository` | 19.0% | ⚠️ 待完善 |
| `internal/service` | 0.0% | ⚠️ 待测试 |

## 🚀 运行测试

### 方式 1：使用测试脚本（推荐）

```bash
cd /Users/sady3721/project/owlBack/wisefido-ai
bash scripts/run_tests.sh
```

### 方式 2：直接运行

```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./internal/config -v
go test ./internal/consumer -v
go test ./internal/evaluator -v
go test ./internal/repository -v

# 显示覆盖率
go test ./... -cover

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📋 已实现的测试

### 1. Config 模块 (`internal/config/config_test.go`)

✅ **3个测试全部通过**
- `TestLoad_DefaultValues` - 默认值测试
- `TestLoad_EnvironmentVariables` - 环境变量测试
- `TestGetEnv` - 环境变量获取测试

**覆盖率**: 100% ✅

### 2. Consumer 模块 (`internal/consumer/cache_manager_test.go`)

✅ **5个测试全部通过**
- `TestCacheManager_GetRealtimeData_Success` - 读取实时数据
- `TestCacheManager_GetRealtimeData_NotFound` - 数据不存在
- `TestCacheManager_UpdateAlarmCache_Success` - 更新报警缓存
- `TestStateManager_SetState_GetState` - 状态管理
- `TestStateManager_ExistsState` - 状态存在性检查

**覆盖率**: 29.9%

### 3. Evaluator 模块 (`internal/evaluator/alarm_event_builder_test.go`)

✅ **3个测试全部通过**
- `TestAlarmEventBuilder_BuildAlarmEvent` - 报警事件构建
- `TestBuildTriggerData` - 触发数据构建
- `TestBuildTriggerData_WithNilValues` - 空值处理

**覆盖率**: 11.5%

### 4. Repository 模块 (`internal/repository/card_test.go`)

✅ **4个测试全部通过**
- `TestGetCardByID_Success` - 获取卡片成功
- `TestGetCardByID_NotFound` - 卡片不存在
- `TestGetCardDevices_Success` - 获取设备列表
- `TestGetAllCards_Success` - 获取所有卡片

**覆盖率**: 19.0%

## 🛠️ 测试工具

### 使用的库

- **go-sqlmock** - 数据库 Mock
  - 用于 Repository 层测试
  - 模拟 PostgreSQL 查询

- **miniredis** - Redis Mock
  - 用于 Consumer 层测试
  - 模拟 Redis 操作

- **testify** - 断言库
  - `assert` - 断言（失败不中断）
  - `require` - 必需断言（失败中断）

## 📝 测试最佳实践

### 1. 测试结构

```go
func TestFunctionName_Scenario(t *testing.T) {
    // 1. 设置（Setup）
    // 2. 执行（Execute）
    // 3. 验证（Assert）
}
```

### 2. 使用 Mock

```go
// 数据库 Mock
db, mock, repo := setupMockDB(t)
defer db.Close()

// Redis Mock
mr := miniredis.RunT(t)
redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
```

### 3. 测试覆盖

- ✅ 正常流程
- ✅ 错误情况
- ✅ 边界条件（空值、nil、无效输入）

## ⏳ 待实现的测试

### 高优先级

1. **Repository 层**
   - [ ] `alarm_cloud.go` 测试
   - [ ] `alarm_device.go` 测试
   - [ ] `alarm_events.go` 测试
   - [ ] `device.go` 测试
   - [ ] `room.go` 测试

2. **Evaluator 层**
   - [ ] `event1_bed_fall.go` 测试
   - [ ] `event2_sleepad_reliability.go` 测试
   - [ ] `event3_bathroom_fall.go` 测试
   - [ ] `event4_sudden_disappear.go` 测试
   - [ ] `evaluator.go` 测试（主评估器）

3. **Consumer 层**
   - [ ] `cache_consumer.go` 测试

### 中优先级

4. **Service 层**
   - [ ] `alarm.go` 测试（需要依赖注入重构）

## 🎯 覆盖率目标

- **当前**: ~30% 总体覆盖率
- **目标**: 70%+ 总体覆盖率
  - Config: 100% ✅
  - Consumer: 70%+
  - Evaluator: 70%+
  - Repository: 70%+
  - Service: 50%+

## 🔗 相关文档

- `TEST_SUMMARY.md` - 测试总结
- `TESTING_GUIDE.md` - 测试指南
- `QUICK_START.md` - 快速启动指南

