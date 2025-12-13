# OwlBack 代码规范

> **目的**: 确保代码质量，使其他 AI 工具（如 ChatGPT）能够准确审查和评估

---

## 📋 代码规范检查清单

### 1. 命名规范 ✅

#### 1.1 包名
- ✅ 小写字母，简短
- ✅ 不使用下划线或混合大小写
- ✅ 示例: `package repository` ✅, `package Repository` ❌

#### 1.2 函数名
- ✅ 驼峰命名，首字母大写（公开函数）
- ✅ 首字母小写（私有函数）
- ✅ 清晰表达功能
- ✅ 示例: `GetCardByDeviceID()` ✅, `getCard()` ⚠️ (不够清晰)

#### 1.3 变量名
- ✅ 驼峰命名，首字母小写
- ✅ 清晰表达含义
- ✅ 避免缩写（除非是通用缩写）
- ✅ 示例: `deviceID` ✅, `dID` ❌

#### 1.4 常量
- ✅ 全大写，下划线分隔
- ✅ 示例: `const MAX_RETRY_COUNT = 3` ✅

---

### 2. 代码结构规范 ✅

#### 2.1 文件组织
```
service/
├── config.go          # 配置
├── service.go          # 服务主逻辑
├── repository/        # 数据访问层
│   └── card.go
├── consumer/          # 消费者
│   └── stream_consumer.go
└── models/           # 数据模型
    └── iot_timeseries.go
```

#### 2.2 导入顺序
```go
import (
    // 1. 标准库
    "context"
    "fmt"
    "time"
    
    // 2. 第三方库
    "github.com/go-redis/redis/v8"
    "go.uber.org/zap"
    
    // 3. 本地包
    "wisefido-sensor-fusion/internal/config"
    "wisefido-sensor-fusion/internal/models"
)
```

#### 2.3 函数长度
- ✅ 单个函数不超过 50 行
- ✅ 复杂逻辑拆分为多个函数
- ⚠️ 当前: 部分函数可能过长，需要重构

---

### 3. 错误处理规范 ✅

#### 3.1 错误返回
```go
// ✅ 正确
func GetCard(id string) (*Card, error) {
    if id == "" {
        return nil, fmt.Errorf("card id cannot be empty")
    }
    // ...
}

// ❌ 错误：忽略错误
func GetCard(id string) *Card {
    // 没有错误处理
}
```

#### 3.2 错误包装
```go
// ✅ 正确：使用 %w 包装错误
if err != nil {
    return nil, fmt.Errorf("failed to query card: %w", err)
}

// ❌ 错误：丢失原始错误信息
if err != nil {
    return nil, fmt.Errorf("query failed")
}
```

#### 3.3 错误日志
```go
// ✅ 正确：记录错误上下文
if err != nil {
    logger.Error("Failed to get card",
        zap.String("card_id", cardID),
        zap.Error(err),
    )
    return nil, err
}
```

---

### 4. 注释规范 ✅

#### 4.1 包注释
```go
// Package fusion 提供传感器融合功能
// 包括 HR/RR 融合、姿态数据融合等
package fusion
```

#### 4.2 公开函数注释
```go
// FuseCardData 融合卡片的所有设备数据
// 
// 融合规则：
// 1. HR/RR：优先 Sleepace，无数据则 Radar
// 2. 姿态：合并所有 Radar 的 tracking_id
//
// 参数:
//   - cardID: 卡片 ID
//
// 返回:
//   - *RealtimeData: 融合后的实时数据
//   - error: 错误信息
func (f *SensorFusion) FuseCardData(cardID string) (*models.RealtimeData, error) {
    // ...
}
```

#### 4.3 复杂逻辑注释
```go
// 优先使用 Sleepace 数据
// 如果 Sleepace 没有数据，使用 Radar 数据
if len(sleepaceData) > 0 {
    // ...
}
```

---

### 5. 性能规范 ⚠️

#### 5.1 避免 N+1 查询
```go
// ❌ 错误：N+1 查询
for _, device := range devices {
    data, err := repo.GetLatestByDeviceID(device.DeviceID, 1)
}

// ✅ 正确：批量查询
deviceIDs := make([]string, len(devices))
for i, device := range devices {
    deviceIDs[i] = device.DeviceID
}
dataMap, err := repo.GetLatestByDeviceIDs(deviceIDs)
```

#### 5.2 使用连接池
```go
// ✅ 正确：使用连接池
db, err := sql.Open("postgres", dsn)
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
```

#### 5.3 避免不必要的内存分配
```go
// ✅ 正确：预分配容量
result := make([]*Card, 0, len(devices))

// ❌ 错误：动态扩容
result := []*Card{}
```

---

### 6. 并发安全规范 ⚠️

#### 6.1 共享资源保护
```go
// ⚠️ 当前问题：多个 goroutine 可能同时更新缓存
// ✅ 建议：使用锁或原子操作
var mu sync.Mutex
mu.Lock()
defer mu.Unlock()
// 更新缓存
```

#### 6.2 Context 使用
```go
// ✅ 正确：传递 context
func (s *Service) Process(ctx context.Context, data Data) error {
    // 使用 ctx 进行超时控制
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    // ...
}
```

---

### 7. 测试规范 ⚠️

#### 7.1 测试文件命名
```
fusion.go          -> fusion_test.go
card_repository.go -> card_repository_test.go
```

#### 7.2 测试函数命名
```go
func TestFuseVitalSigns_PrioritySleepace(t *testing.T) {
    // ...
}

func TestFuseVitalSigns_FallbackRadar(t *testing.T) {
    // ...
}
```

#### 7.3 测试覆盖率
- ✅ 目标: 70%+ 覆盖率
- ⚠️ 当前: 0% 覆盖率

---

### 8. 数据库操作规范 ✅

#### 8.1 使用参数化查询
```go
// ✅ 正确：参数化查询
query := "SELECT * FROM cards WHERE card_id = $1"
err := db.QueryRow(query, cardID).Scan(...)

// ❌ 错误：SQL 注入风险
query := fmt.Sprintf("SELECT * FROM cards WHERE card_id = '%s'", cardID)
```

#### 8.2 事务处理
```go
// ✅ 正确：使用事务
tx, err := db.Begin()
defer tx.Rollback()
// 执行操作
if err := tx.Commit(); err != nil {
    return err
}
```

---

## 📊 当前代码规范评分

| 规范项 | 评分 | 说明 |
|--------|------|------|
| 命名规范 | 8/10 | 基本符合，部分可改进 |
| 代码结构 | 9/10 | 清晰的分层架构 |
| 错误处理 | 7/10 | 基本完善，部分可改进 |
| 注释规范 | 6/10 | 关键函数有注释，但不够详细 |
| 性能规范 | 5/10 | 存在 N+1 查询问题 |
| 并发安全 | 5/10 | 缺少并发保护 |
| 测试规范 | 0/10 | 缺少测试文件 |
| **总体评分** | **6.0/10** | 需要改进 |

---

## 🎯 改进建议

### 高优先级
1. **添加详细注释** - 提高可读性
2. **修复 N+1 查询** - 提高性能
3. **添加单元测试** - 提高代码质量

### 中优先级
4. **添加并发保护** - 确保线程安全
5. **优化错误处理** - 提供更详细的错误信息
6. **添加输入验证** - 提高安全性

---

## 📝 代码审查检查清单

### 提交前检查
- [ ] 代码格式正确 (`go fmt`)
- [ ] 代码规范检查 (`go vet`)
- [ ] 所有函数有注释
- [ ] 错误处理完善
- [ ] 无明显的性能问题
- [ ] 无并发安全问题
- [ ] 有单元测试（至少核心逻辑）

---

**最后更新**: 2024-12-19

