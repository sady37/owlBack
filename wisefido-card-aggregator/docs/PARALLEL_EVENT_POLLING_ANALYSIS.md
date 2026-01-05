# 事件驱动 + 轮询并行模式分析

## 1. 当前实现分析

### 当前设计：互斥模式（Mutually Exclusive）

**代码位置**：`aggregator.go:Start` (第 102-114 行)

```go
// 根据触发模式启动不同的处理逻辑
if s.config.Aggregator.TriggerMode == "polling" {
    return s.startPollingMode(ctx)  // 阻塞，在主 goroutine 中运行
} else if s.config.Aggregator.TriggerMode == "events" {
    return s.startEventDrivenMode(ctx)  // 阻塞，在主 goroutine 中运行
}
```

**特点**：
- ❌ 只能选择一种模式
- ❌ `startPollingMode` 是阻塞的（`for` 循环在主 goroutine）
- ❌ `startEventDrivenMode` 是阻塞的（`eventConsumer.Start` 在主 goroutine）

---

## 2. 为什么不能并行？技术原因分析

### 原因 1：阻塞设计

**`startPollingMode`**：
```go
func (s *AggregatorService) startPollingMode(ctx context.Context) error {
    // ...
    for {
        select {
        case <-ctx.Done():
            return nil
        case <-ticker.C:
            s.createAllCards(ctx)  // 阻塞调用
        }
    }
    // 永远不会返回（除非 ctx.Done）
}
```

**`startEventDrivenMode`**：
```go
func (s *AggregatorService) startEventDrivenMode(ctx context.Context) error {
    // ...
    return s.eventConsumer.Start(ctx)  // 阻塞调用，永远不会返回
}
```

**问题**：
- 两个方法都是阻塞的，如果同时调用，第二个永远不会执行
- 当前使用 `if-else` 确保只执行一个

---

## 3. 是否可以并行？技术可行性分析

### ✅ 技术上完全可行

#### 理由 1：`CreateCardsForUnit` 有比较逻辑保护

**代码位置**：`card_creator.go:CreateCardsForUnit` (第 114-123 行)

```go
// 5. Compare existing cards with expected cards
if c.compareCards(existingCards, expectedCards) {
    // Cards are unchanged, skip update to preserve card_id
    stats.UnchangedCount = len(existingCards)
    c.logger.Debug("Cards unchanged, skipping update",
        zap.String("unit_id", unitID),
        zap.Int("card_count", len(expectedCards)),
    )
    return stats, nil  // ← 如果卡片没有变化，直接返回，不执行更新
}
```

**保护机制**：
- 如果卡片没有变化，直接返回，不执行任何数据库操作
- 即使两个模式同时调用 `CreateCardsForUnit`，如果卡片没有变化，都不会执行更新
- 如果卡片有变化，会执行更新，但因为有比较逻辑，不会重复创建

#### 理由 2：数据库事务保护

**代码位置**：`card.go:CreateCard`, `UpdateCard`, `DeleteCard`

- 所有操作都在数据库事务中执行
- PostgreSQL 的 ACID 特性保证并发安全
- 即使两个 goroutine 同时更新同一个 unit，数据库会处理并发冲突

#### 理由 3：事件驱动模式已经有定时任务

**代码位置**：`aggregator.go:startEventDrivenMode` (第 269-270 行)

```go
// 启动定时任务（每天上午9点）
go s.startScheduledUpdate(ctx)
```

**说明**：
- 事件驱动模式已经有一个定时任务（每天上午 9 点全量更新）
- 这说明设计上已经考虑了"事件驱动 + 定时兜底"的组合

---

## 4. 并行模式的实现方案

### 方案 A：简单并行（推荐）

**修改 `Start` 方法**：

```go
func (s *AggregatorService) Start(ctx context.Context) error {
    s.logger.Info("Starting card aggregator service",
        zap.String("trigger_mode", s.config.Aggregator.TriggerMode),
        zap.Bool("aggregation_enabled", s.config.Aggregator.Aggregation.Enabled),
    )
    
    // 启动数据聚合任务（如果启用）
    if s.config.Aggregator.Aggregation.Enabled {
        go s.startDataAggregation(ctx)
    }
    
    // ✅ 修改：支持并行模式
    enablePolling := s.config.Aggregator.TriggerMode == "polling" || 
                     s.config.Aggregator.TriggerMode == "both"
    enableEvents := s.config.Aggregator.TriggerMode == "events" || 
                    s.config.Aggregator.TriggerMode == "both"
    
    // 首次执行一次全量创建
    if err := s.createAllCards(ctx); err != nil {
        s.logger.Error("Failed to create all cards on startup", zap.Error(err))
    }
    
    // 启动轮询模式（如果启用）
    if enablePolling {
        go func() {
            if err := s.startPollingMode(ctx); err != nil {
                s.logger.Error("Polling mode failed", zap.Error(err))
            }
        }()
    }
    
    // 启动事件驱动模式（如果启用）
    if enableEvents {
        if s.eventConsumer == nil {
            return fmt.Errorf("event consumer not initialized (required for events mode)")
        }
        // 启动定时任务（每天上午9点）
        go s.startScheduledUpdate(ctx)
        // 启动事件消费者（阻塞）
        return s.eventConsumer.Start(ctx)
    }
    
    // 如果只启用轮询模式，需要阻塞主 goroutine
    if enablePolling {
        select {
        case <-ctx.Done():
            return nil
        }
    }
    
    return fmt.Errorf("no trigger mode enabled")
}
```

**配置支持**：
```go
// config.go
TriggerMode string // "polling" 或 "events" 或 "both"
```

**优点**：
- ✅ 事件驱动：实时响应数据变化
- ✅ 轮询兜底：确保即使事件丢失，也能最终一致性
- ✅ 资源高效：轮询有比较逻辑，不会重复更新

**缺点**：
- ⚠️ 需要修改配置逻辑（支持 "both" 模式）
- ⚠️ 两个模式可能同时更新同一个 unit（但数据库事务保护）

---

### 方案 B：事件驱动 + 长间隔轮询（当前事件驱动模式）

**当前实现**：`startEventDrivenMode` 已经包含定时任务

```go
// 启动定时任务（每天上午9点）
go s.startScheduledUpdate(ctx)
```

**特点**：
- ✅ 事件驱动：实时响应
- ✅ 定时兜底：每天上午 9 点全量更新
- ✅ 不需要修改代码

**问题**：
- ⚠️ 定时任务间隔太长（24 小时）
- ⚠️ 如果事件丢失，需要等待 24 小时

**改进建议**：
- 可以将定时任务间隔缩短（如 1 小时或 6 小时）
- 或者添加一个可配置的轮询间隔（即使事件驱动模式也启用轮询）

---

## 5. 并发安全性分析

### 潜在问题

#### 问题 1：两个模式同时更新同一个 unit

**场景**：
- 事件驱动：收到 `device.bound` 事件，调用 `CreateCardsForUnit(unitA)`
- 轮询：定时触发，调用 `createAllCards()`，也会更新 `unitA`

**影响**：
- ✅ **安全**：`CreateCardsForUnit` 有比较逻辑，如果卡片没有变化，不会执行更新
- ✅ **安全**：如果卡片有变化，会执行更新，但数据库事务保证一致性
- ⚠️ **性能**：可能造成不必要的数据库查询（但不会重复更新）

#### 问题 2：资源消耗

**场景**：
- 事件驱动：实时响应，只更新变化的 unit
- 轮询：定时全量更新所有 unit

**影响**：
- ⚠️ **资源消耗**：轮询会查询所有 unit，即使没有变化
- ✅ **优化**：`CreateCardsForUnit` 有比较逻辑，如果卡片没有变化，不会执行数据库更新
- ✅ **可接受**：查询成本较低，更新成本被比较逻辑保护

---

## 6. 建议方案

### 推荐：方案 B（改进版）- 事件驱动 + 可配置轮询兜底

**实现**：
1. **默认模式**：`events`（事件驱动）
2. **自动启用轮询兜底**：即使事件驱动模式，也启用一个长间隔的轮询（如 1-6 小时）
3. **配置选项**：`CARD_POLLING_FALLBACK_INTERVAL`（轮询兜底间隔，默认 3600 秒）

**代码修改**：

```go
func (s *AggregatorService) startEventDrivenMode(ctx context.Context) error {
    s.logger.Info("Starting event-driven mode")
    
    // 首次执行一次全量创建
    if err := s.createAllCards(ctx); err != nil {
        s.logger.Error("Failed to create all cards on startup", zap.Error(err))
    }
    
    // 启动定时任务（每天上午9点全量更新）
    go s.startScheduledUpdate(ctx)
    
    // ✅ 新增：启动轮询兜底（可配置间隔，默认 1 小时）
    fallbackInterval := s.config.Aggregator.Polling.FallbackInterval
    if fallbackInterval > 0 {
        go s.startPollingFallback(ctx, fallbackInterval)
    }
    
    // 启动事件消费者（阻塞）
    if s.eventConsumer != nil {
        return s.eventConsumer.Start(ctx)
    }
    
    return fmt.Errorf("event consumer not initialized")
}

// startPollingFallback 启动轮询兜底（长间隔，用于确保最终一致性）
func (s *AggregatorService) startPollingFallback(ctx context.Context, interval int) {
    ticker := time.NewTicker(time.Duration(interval) * time.Second)
    defer ticker.Stop()
    
    s.logger.Info("Starting polling fallback",
        zap.Int("interval_seconds", interval),
    )
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            s.logger.Debug("Running polling fallback check")
            if err := s.createAllCards(ctx); err != nil {
                s.logger.Error("Failed to create cards in polling fallback", zap.Error(err))
            }
        }
    }
}
```

**配置**：
```go
// config.go
Polling struct {
    Interval        int // 轮询模式间隔（秒），默认 3600
    FallbackInterval int // 事件驱动模式下的轮询兜底间隔（秒），默认 3600，0 表示禁用
}
```

**优点**：
- ✅ 事件驱动：实时响应（主要更新方式）
- ✅ 轮询兜底：确保最终一致性（即使事件丢失）
- ✅ 资源高效：轮询有比较逻辑，不会重复更新
- ✅ 配置灵活：可以禁用轮询兜底（设置为 0）

---

## 7. 总结

### 当前设计限制

**为什么不能并行**：
1. **阻塞设计**：两个模式都是阻塞的，不能同时运行
2. **互斥逻辑**：使用 `if-else` 确保只执行一个

### 是否可以并行？

**答案**：✅ **技术上完全可行**

**理由**：
1. ✅ `CreateCardsForUnit` 有比较逻辑保护（不会重复更新）
2. ✅ 数据库事务保护（并发安全）
3. ✅ 事件驱动模式已经有定时任务（说明设计上支持组合）

### 建议

**推荐方案**：事件驱动 + 可配置轮询兜底
- 主要更新方式：事件驱动（实时响应）
- 兜底机制：长间隔轮询（1-6 小时，确保最终一致性）
- 配置灵活：可以禁用轮询兜底

**优点**：
- ✅ 实时响应 + 最终一致性保障
- ✅ 资源高效（比较逻辑保护）
- ✅ 不需要修改太多代码（在 `startEventDrivenMode` 中添加轮询兜底）

