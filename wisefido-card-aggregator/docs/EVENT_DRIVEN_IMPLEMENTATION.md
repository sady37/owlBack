# 事件驱动模式实现方案

## 需求

- **事件触发**：当设备/住户/床位绑定关系变化时，立即触发卡片重新计算
- **定时兜底**：每天上午9点全量更新，确保数据一致性

## 事件触发方案

### 方案选择：Redis Streams

**理由**：
1. ✅ 项目已有 Redis Streams 基础设施（`owl-common/redis/streams.go`）
2. ✅ 其他服务（`wisefido-sensor-fusion`, `wisefido-data-transformer`）都在使用
3. ✅ 支持消息持久化和消费者组
4. ✅ 解耦：API 层发布事件，卡片聚合服务消费事件

### 架构设计

```
┌─────────────────┐
│  API 服务        │
│ (wisefido-data) │
│                 │
│  设备绑定/解绑   │───发布事件───┐
│  住户绑定/解绑   │              │
│  床位状态变化    │              ▼
└─────────────────┘      ┌──────────────────┐
                         │  Redis Streams    │
                         │  card:events      │
                         └──────────────────┘
                                  │
                                  │ 消费事件
                                  ▼
                         ┌──────────────────┐
                         │ wisefido-card-   │
                         │ aggregator       │
                         │                  │
                         │  监听事件        │
                         │  重新计算卡片    │
                         └──────────────────┘
```

### 事件类型

根据 `21_cards.sql` 文档，需要监听以下事件：

1. **设备绑定事件** (`device.bound`)
   - 触发时机：`devices.bound_bed_id` 变化
   - 触发时机：`devices.bound_room_id` 变化
   - 触发时机：`devices.unit_id` 变化
   - 触发时机：`devices.monitoring_enabled` 变化

2. **住户绑定事件** (`resident.bound`)
   - 触发时机：`residents.bed_id` 变化
   - 触发时机：`residents.unit_id` 变化
   - 触发时机：`residents.status` 变化

3. **床位状态事件** (`bed.status_changed`)
   - 触发时机：`beds.bound_device_count` 变化（ActiveBed ↔ NonActiveBed）

4. **地址信息事件** (`unit.info_changed`)
   - 触发时机：`units.unit_name` 变化
   - 触发时机：`units.branch_tag` 变化
   - 触发时机：`beds.bed_name` 变化

### 事件消息格式

```json
{
  "event_type": "device.bound",
  "tenant_id": "tenant-123",
  "unit_id": "unit-456",
  "bed_id": "bed-789",
  "device_id": "device-001",
  "timestamp": 1234567890,
  "metadata": {
    "old_bed_id": null,
    "new_bed_id": "bed-789"
  }
}
```

### 实现步骤

#### 1. API 层发布事件（需要在 wisefido-data 服务中实现）

当设备/住户/床位绑定关系变化时，发布事件：

```go
// 在设备绑定 API 中
func (s *DeviceService) BindDeviceToBed(deviceID, bedID string) error {
    // 1. 更新数据库
    err := s.repo.UpdateDeviceBinding(deviceID, bedID)
    if err != nil {
        return err
    }
    
    // 2. 发布事件到 Redis Streams
    event := map[string]interface{}{
        "event_type": "device.bound",
        "tenant_id":  tenantID,
        "device_id":  deviceID,
        "bed_id":     bedID,
        "unit_id":    unitID, // 从 bed 查询得到
        "timestamp":  time.Now().Unix(),
    }
    
    _, err = rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
    return err
}
```

#### 2. wisefido-card-aggregator 监听事件

```go
// internal/consumer/event_consumer.go
type EventConsumer struct {
    redisClient *redis.Client
    cardCreator *aggregator.CardCreator
    logger      *zap.Logger
}

func (c *EventConsumer) Start(ctx context.Context) error {
    // 创建消费者组
    stream := "card:events"
    consumerGroup := "card-aggregator-group"
    consumerName := "card-aggregator-1"
    
    err := rediscommon.CreateConsumerGroup(ctx, c.redisClient, stream, consumerGroup)
    if err != nil {
        return err
    }
    
    // 消费事件
    for {
        messages, err := rediscommon.ReadFromStream(
            ctx, c.redisClient, stream,
            consumerGroup, consumerName, 10,
        )
        if err != nil {
            c.logger.Error("Failed to read events", zap.Error(err))
            continue
        }
        
        for _, msg := range messages {
            if err := c.processEvent(ctx, msg); err != nil {
                c.logger.Error("Failed to process event", zap.Error(err))
            }
        }
    }
}

func (c *EventConsumer) processEvent(ctx context.Context, msg rediscommon.StreamMessage) error {
    eventType := msg.Values["event_type"].(string)
    tenantID := msg.Values["tenant_id"].(string)
    unitID := msg.Values["unit_id"].(string)
    
    // 根据事件类型触发卡片重新计算
    switch eventType {
    case "device.bound", "device.unbound":
        return c.cardCreator.CreateCardsForUnit(tenantID, unitID)
    case "resident.bound", "resident.unbound":
        return c.cardCreator.CreateCardsForUnit(tenantID, unitID)
    case "bed.status_changed":
        bedID := msg.Values["bed_id"].(string)
        // 获取 bed 的 unit_id
        unitID := c.getUnitIDByBedID(tenantID, bedID)
        return c.cardCreator.CreateCardsForUnit(tenantID, unitID)
    case "unit.info_changed":
        return c.cardCreator.CreateCardsForUnit(tenantID, unitID)
    }
    
    return nil
}
```

#### 3. 定时任务（每天上午9点）

```go
// internal/service/aggregator.go
func (s *AggregatorService) startScheduledUpdate(ctx context.Context) error {
    // 计算到明天上午9点的时间
    now := time.Now()
    next9AM := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
    if next9AM.Before(now) {
        next9AM = next9AM.Add(24 * time.Hour)
    }
    
    duration := next9AM.Sub(now)
    timer := time.NewTimer(duration)
    defer timer.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return nil
        case <-timer.C:
            // 执行全量更新
            if err := s.createAllCards(ctx); err != nil {
                s.logger.Error("Failed to create all cards in scheduled update", zap.Error(err))
            }
            
            // 重置定时器到明天上午9点
            timer.Reset(24 * time.Hour)
        }
    }
}
```

## 实现计划

### 阶段 1：事件消费（wisefido-card-aggregator）

1. ✅ 创建事件消费者 (`internal/consumer/event_consumer.go`)
2. ✅ 监听 Redis Streams `card:events`
3. ✅ 解析事件并触发卡片重新计算
4. ✅ 实现定时任务（每天上午9点）

### 阶段 2：事件发布（wisefido-data API 服务）

1. ⚠️ 在设备绑定 API 中发布事件
2. ⚠️ 在住户绑定 API 中发布事件
3. ⚠️ 在床位状态变化时发布事件
4. ⚠️ 在地址信息变化时发布事件

### 阶段 3：增量更新优化

1. ⚠️ 实现增量更新逻辑（只更新变化的卡片）
2. ⚠️ 结合事件驱动，达到最佳性能

## 注意事项

1. **事件幂等性**：确保重复处理事件不会导致数据不一致
2. **错误处理**：事件处理失败时，需要重试机制
3. **消息确认**：处理成功后确认消息，避免重复处理
4. **并发控制**：同一 unit 的多个事件可能并发，需要处理竞态条件

