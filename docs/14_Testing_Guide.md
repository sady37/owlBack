# OwlBack 测试指南

## 📋 概述

本文档提供 OwlBack 项目的测试策略、测试方法和测试工具使用指南。

---

## 🎯 测试策略

### 测试金字塔

```
        /\
       /  \      E2E Tests (少量)
      /____\
     /      \    Integration Tests (适量)
    /________\
   /          \  Unit Tests (大量)
  /____________\
```

### 测试类型

1. **单元测试** (Unit Tests)
   - 测试单个函数/方法
   - 快速、隔离、可重复
   - 目标覆盖率: 70%+

2. **集成测试** (Integration Tests)
   - 测试组件间交互
   - 使用真实数据库和 Redis
   - 目标覆盖率: 50%+

3. **端到端测试** (E2E Tests)
   - 测试完整数据流
   - 使用完整测试环境
   - 目标覆盖率: 20%+

---

## 🛠️ 测试工具

### Go 测试框架

```bash
# 标准库
go test ./...

# 测试框架
go get github.com/stretchr/testify
```

### 测试数据库

```bash
# 使用 Docker 启动测试数据库
docker run -d \
  --name test-postgres \
  -e POSTGRES_PASSWORD=test \
  -e POSTGRES_DB=owlrd_test \
  -p 5433:5432 \
  postgres:15
```

### 测试 Redis

```bash
# 使用 Docker 启动测试 Redis
docker run -d \
  --name test-redis \
  -p 6380:6379 \
  redis:7
```

---

## 📝 单元测试示例

### 示例 1: 传感器融合逻辑测试

```go
// internal/fusion/sensor_fusion_test.go
package fusion

import (
    "testing"
    "wisefido-sensor-fusion/internal/models"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock Repository
type MockCardRepository struct {
    mock.Mock
}

func (m *MockCardRepository) GetCardDevices(cardID string) ([]DeviceInfo, error) {
    args := m.Called(cardID)
    return args.Get(0).([]DeviceInfo), args.Error(1)
}

func TestFuseVitalSigns_PrioritySleepace(t *testing.T) {
    // 准备测试数据
    sleepaceData := []*models.IoTTimeSeries{
        {
            HeartRate: intPtr(75),
            RespiratoryRate: intPtr(20),
            DeviceType: "Sleepace",
        },
    }
    radarData := []*models.IoTTimeSeries{
        {
            HeartRate: intPtr(80),
            RespiratoryRate: intPtr(18),
            DeviceType: "Radar",
        },
    }
    
    result := &models.RealtimeData{}
    fusion := &SensorFusion{}
    
    // 执行测试
    fusion.fuseVitalSigns(sleepaceData, radarData, result)
    
    // 验证结果
    assert.Equal(t, 75, *result.Heart)
    assert.Equal(t, 20, *result.Breath)
    assert.Equal(t, "Sleepace", result.HeartSource)
    assert.Equal(t, "Sleepace", result.BreathSource)
}

func TestFuseVitalSigns_FallbackRadar(t *testing.T) {
    // 测试 Sleepace 无数据时，使用 Radar 数据
    sleepaceData := []*models.IoTTimeSeries{}
    radarData := []*models.IoTTimeSeries{
        {
            HeartRate: intPtr(80),
            RespiratoryRate: intPtr(18),
            DeviceType: "Radar",
        },
    }
    
    result := &models.RealtimeData{}
    fusion := &SensorFusion{}
    
    fusion.fuseVitalSigns(sleepaceData, radarData, result)
    
    assert.Equal(t, 80, *result.Heart)
    assert.Equal(t, 18, *result.Breath)
    assert.Equal(t, "Radar", result.HeartSource)
    assert.Equal(t, "Radar", result.BreathSource)
}

func intPtr(i int) *int {
    return &i
}
```

### 示例 2: Repository 测试

```go
// internal/repository/card_test.go
package repository

import (
    "database/sql"
    "testing"
    
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/assert"
)

func TestGetCardByDeviceID_BedBinding(t *testing.T) {
    // 创建 mock 数据库
    db, mock, err := sqlmock.New()
    assert.NoError(t, err)
    defer db.Close()
    
    repo := NewCardRepository(db, nil)
    
    // 设置期望的 SQL 查询
    rows := sqlmock.NewRows([]string{"card_id", "tenant_id", "card_type", "bed_id", "unit_id"}).
        AddRow("card-123", "tenant-456", "ActiveBed", "bed-789", nil)
    
    mock.ExpectQuery(`WITH device_info`).
        WithArgs("device-001").
        WillReturnRows(rows)
    
    // 执行测试
    card, err := repo.GetCardByDeviceID("device-001")
    
    // 验证结果
    assert.NoError(t, err)
    assert.Equal(t, "card-123", card.CardID)
    assert.Equal(t, "ActiveBed", card.CardType)
    
    // 验证所有期望都被满足
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

---

## 🔗 集成测试示例

### 示例: 传感器融合服务集成测试

```go
// tests/integration/sensor_fusion_test.go
package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/go-redis/redis/v8"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    "wisefido-sensor-fusion/internal/config"
    "wisefido-sensor-fusion/internal/service"
)

func TestSensorFusionIntegration(t *testing.T) {
    // 1. 设置测试环境
    cfg := &config.Config{
        Database: config.DatabaseConfig{
            Host:     "localhost",
            Port:     5433,
            User:     "postgres",
            Password: "test",
            Database: "owlrd_test",
        },
        Redis: config.RedisConfig{
            Addr: "localhost:6380",
        },
    }
    
    // 2. 初始化服务
    logger := zap.NewNop()
    fusionService, err := service.NewFusionService(cfg, logger)
    require.NoError(t, err)
    
    // 3. 准备测试数据
    // 插入设备、卡片、时序数据到测试数据库
    
    // 4. 触发融合逻辑
    ctx := context.Background()
    go fusionService.Start(ctx)
    
    // 5. 等待处理
    time.Sleep(1 * time.Second)
    
    // 6. 验证结果
    redisClient := redis.NewClient(&redis.Options{
        Addr: "localhost:6380",
    })
    
    result, err := redisClient.Get(ctx, "vital-focus:card:card-123:realtime").Result()
    assert.NoError(t, err)
    assert.NotEmpty(t, result)
    
    // 7. 清理
    fusionService.Stop(ctx)
}
```

---

## 🚀 运行测试

### 运行所有测试

```bash
# 运行所有单元测试
go test ./...

# 运行特定包的测试
go test ./internal/fusion/...

# 运行测试并显示覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 运行集成测试

```bash
# 设置测试环境变量
export DB_HOST=localhost
export DB_PORT=5433
export REDIS_ADDR=localhost:6380

# 运行集成测试
go test -tags=integration ./tests/integration/...
```

---

## 📊 测试覆盖率目标

| 模块 | 目标覆盖率 | 当前覆盖率 |
|------|-----------|-----------|
| fusion | 80% | 0% |
| repository | 70% | 0% |
| consumer | 60% | 0% |
| service | 50% | 0% |
| **总体** | **70%** | **0%** |

---

## 🔍 测试检查清单

### 单元测试检查清单

- [ ] 测试正常流程
- [ ] 测试错误情况
- [ ] 测试边界条件
- [ ] 测试空值/空数组
- [ ] 测试并发安全（如适用）

### 集成测试检查清单

- [ ] 测试完整数据流
- [ ] 测试数据库交互
- [ ] 测试 Redis 交互
- [ ] 测试错误恢复
- [ ] 测试性能（如适用）

---

## 🐛 调试测试

### 使用 delve 调试器

```bash
# 安装 delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试测试
dlv test -- -test.run TestFuseVitalSigns
```

### 使用日志

```go
// 在测试中启用详细日志
logger := zap.NewDevelopment()
```

---

## 📚 参考资源

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [SQLMock Documentation](https://github.com/DATA-DOG/go-sqlmock)
- [Go Testing Best Practices](https://golang.org/doc/effective_go#testing)

---

## ✅ 下一步

1. **创建测试基础设施**
   - [ ] 设置测试数据库和 Redis
   - [ ] 创建测试工具函数
   - [ ] 创建 mock 对象

2. **编写核心测试**
   - [ ] 传感器融合逻辑测试
   - [ ] Repository 测试
   - [ ] Consumer 测试

3. **添加 CI/CD 集成**
   - [ ] 在 CI 中运行测试
   - [ ] 生成覆盖率报告
   - [ ] 测试失败时阻止合并

