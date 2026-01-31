# 启动时全量 Card 创建逻辑

## 实现位置

### 1. 轮询模式（Polling Mode）

**文件**: `internal/service/aggregator.go`

**函数**: `startPollingMode`

**代码**:
```go
func (s *AggregatorService) startPollingMode(ctx context.Context) error {
	interval := time.Duration(s.config.Aggregator.Polling.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	
	s.logger.Info("Starting polling mode",
		zap.Duration("interval", interval),
	)
	
	// 首次执行一次全量创建
	if err := s.createAllCards(ctx); err != nil {
		s.logger.Error("Failed to create all cards on startup", zap.Error(err))
	}
	
	// 定时轮询
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.createAllCards(ctx); err != nil {
				s.logger.Error("Failed to create cards", zap.Error(err))
			}
		}
	}
}
```

**执行流程**:
1. ✅ **启动时立即执行**：调用 `s.createAllCards(ctx)`（第 127 行）
2. ✅ **定时轮询**：每 `CARD_POLLING_INTERVAL` 秒（默认 3600 秒）执行一次

### 2. 事件驱动模式（Event-Driven Mode）

**文件**: `internal/service/aggregator.go`

**函数**: `startEventDrivenMode`

**代码**:
```go
func (s *AggregatorService) startEventDrivenMode(ctx context.Context) error {
	s.logger.Info("Starting event-driven mode")
	
	// 首次执行一次全量创建
	if err := s.createAllCards(ctx); err != nil {
		s.logger.Error("Failed to create all cards on startup", zap.Error(err))
	}
	
	// 启动定时任务（每天上午9点）
	go s.startScheduledUpdate(ctx)
	
	// 启动事件消费者（阻塞）
	if s.eventConsumer != nil {
		return s.eventConsumer.Start(ctx)
	}
	
	return fmt.Errorf("event consumer not initialized")
}
```

**执行流程**:
1. ✅ **启动时立即执行**：调用 `s.createAllCards(ctx)`（第 198 行）
2. ✅ **事件驱动更新**：监听 Redis Streams 事件
3. ✅ **定时任务**：每天上午 9 点全量更新

## createAllCards 实现

**文件**: `internal/service/aggregator.go`

**函数**: `createAllCards`

**代码**:
```go
func (s *AggregatorService) createAllCards(ctx context.Context) error {
	s.logger.Info("Starting to create cards for all units")
	
	// 从配置获取 tenant_id
	tenantID := s.config.Aggregator.TenantID
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required, please set TENANT_ID environment variable")
	}
	
	// 获取所有 unit
	unitIDs, err := s.cardRepo.GetAllUnits(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get all units: %w", err)
	}
	
	s.logger.Info("Found units to process",
		zap.Int("unit_count", len(unitIDs)),
	)
	
	// 为每个 unit 创建卡片
	successCount := 0
	errorCount := 0
	
	for _, unitID := range unitIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := s.cardCreator.CreateCardsForUnit(tenantID, unitID); err != nil {
				s.logger.Error("Failed to create cards for unit",
					zap.String("unit_id", unitID),
					zap.Error(err),
				)
				errorCount++
			} else {
				successCount++
			}
		}
	}
	
	s.logger.Info("Completed creating cards",
		zap.Int("success_count", successCount),
		zap.Int("error_count", errorCount),
	)
	
	return nil
}
```

**执行步骤**:
1. ✅ 从配置获取 `TENANT_ID`
2. ✅ 调用 `GetAllUnits(tenantID)` 获取所有 unit IDs
3. ✅ 遍历每个 unit，调用 `CreateCardsForUnit(tenantID, unitID)`
4. ✅ 记录成功和失败数量

## 启动流程

### 完整启动流程

```
main.go
  └─> service.NewAggregatorService(cfg, log)
      └─> 初始化数据库、Redis、Repository、CardCreator
  └─> svc.Start(ctx)
      └─> 根据 CARD_TRIGGER_MODE 选择模式
          ├─> "polling" → startPollingMode(ctx)
          │   └─> createAllCards(ctx)  // ✅ 立即执行
          │       └─> 遍历所有 units
          │           └─> CreateCardsForUnit(tenantID, unitID)
          │
          └─> "events" → startEventDrivenMode(ctx)
              └─> createAllCards(ctx)  // ✅ 立即执行
                  └─> 遍历所有 units
                      └─> CreateCardsForUnit(tenantID, unitID)
```

## 日志输出

启动时应该看到以下日志：

```
{"level":"info","msg":"Starting card aggregator service","trigger_mode":"polling",...}
{"level":"info","msg":"Starting polling mode","interval":"1h0m0s"}
{"level":"info","msg":"Starting to create cards for all units"}
{"level":"info","msg":"Found units to process","unit_count":4}
{"level":"info","msg":"Completed creating cards","success_count":4,"error_count":0}
```

## 验证方法

### 1. 检查日志

启动服务后，应该立即看到：
- `"Starting to create cards for all units"`
- `"Found units to process"` (unit_count > 0)
- `"Completed creating cards"` (success_count > 0)

### 2. 检查数据库

启动后立即查询 cards 表：

```sql
SELECT COUNT(*) FROM cards WHERE tenant_id = 'bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c';
```

应该立即有数据（如果满足 ActiveBed 条件）。

### 3. 使用诊断脚本

```bash
cd /Users/sady3721/project/owlBack/wisefido-card-aggregator
./diagnose_card_generation.sh
```

## 常见问题

### Q: 为什么启动后没有立即创建 cards？

**可能原因**:
1. ❌ `TENANT_ID` 环境变量未设置
2. ❌ 服务启动失败（检查日志）
3. ❌ 数据库连接失败
4. ❌ 没有满足 ActiveBed 条件的 beds

**解决方法**:
1. ✅ 检查环境变量：`echo $TENANT_ID`
2. ✅ 检查服务日志：查看是否有错误
3. ✅ 运行诊断脚本：`./diagnose_card_generation.sh`
4. ✅ 检查数据库：确认有 ActiveBeds

### Q: 启动时创建失败，但定时轮询成功？

**可能原因**:
1. 启动时数据库连接未就绪
2. 启动时 Redis 连接未就绪（如果启用数据聚合）

**解决方法**:
- 检查服务依赖（PostgreSQL、Redis）是否已启动
- 检查服务启动顺序（确保数据库先启动）

## 总结

✅ **逻辑已实现**：启动时立即执行一次全量 card 创建

✅ **两种模式都支持**：
- 轮询模式：启动时 + 定时轮询
- 事件驱动模式：启动时 + 事件驱动 + 定时任务

✅ **错误处理**：如果启动时创建失败，会记录错误日志，但不会阻止服务继续运行

