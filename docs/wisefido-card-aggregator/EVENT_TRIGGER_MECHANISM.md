# 事件触发机制说明

## 实现方案

### 事件触发：Redis Streams

**架构**：
```
API 服务 (wisefido-data)
    ↓ 发布事件
Redis Streams (card:events)
    ↓ 消费事件
wisefido-card-aggregator
    ↓ 重新计算
Cards 表更新
```

### 事件类型

根据 `21_cards.sql` 文档，需要监听以下事件：

1. **设备绑定事件**
   - `device.bound` - 设备绑定到床位/房间
   - `device.unbound` - 设备解绑
   - `device.monitoring_changed` - 设备监护状态变化

2. **住户绑定事件**
   - `resident.bound` - 住户绑定到床位/单元
   - `resident.unbound` - 住户解绑
   - `resident.status_changed` - 住户状态变化

3. **床位状态事件**
   - `bed.status_changed` - 床位状态变化（ActiveBed ↔ NonActiveBed）
   - `bed.device_count_changed` - 床位设备数量变化

4. **单元信息事件**
   - `unit.info_changed` - 单元信息变化（地址、名称等）

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

## 实现状态

### ✅ 已完成（wisefido-card-aggregator）

1. **事件消费者** (`internal/consumer/event_consumer.go`)
   - ✅ 监听 Redis Streams `card:events`
   - ✅ 解析事件并触发卡片重新计算
   - ✅ 支持多种事件类型
   - ✅ 错误处理和重试机制

2. **定时任务** (`internal/service/aggregator.go`)
   - ✅ 每天上午9点全量更新
   - ✅ 与事件驱动模式并行运行

3. **配置支持**
   - ✅ `CARD_TRIGGER_MODE=events` 启用事件驱动模式
   - ✅ `CARD_EVENT_STREAM` 配置事件流名称
   - ✅ `CARD_CONSUMER_GROUP` 配置消费者组
   - ✅ `CARD_CONSUMER_NAME` 配置消费者名称

### ⚠️ 待实现（wisefido-data API 服务）⏸️ **已暂停，使用轮询模式**

**📝 当前状态**：
- `wisefido-data` 服务尚未实现
- 暂时使用轮询模式（每60秒全量更新）
- 待实现文档：`../../docs/PENDING_FEATURES.md`

**需要在 API 层发布事件**（待 wisefido-data 服务实现后）：

1. **设备绑定 API** (`/api/devices/:id/bind`)
   ```go
   // 当设备绑定/解绑时
   event := map[string]interface{}{
       "event_type": "device.bound",
       "tenant_id":  tenantID,
       "device_id":  deviceID,
       "bed_id":     bedID,
       "unit_id":    unitID,
       "timestamp":  time.Now().Unix(),
   }
   rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
   ```

2. **住户绑定 API** (`/api/residents/:id/bind`)
   ```go
   // 当住户绑定/解绑时
   event := map[string]interface{}{
       "event_type": "resident.bound",
       "tenant_id":  tenantID,
       "resident_id": residentID,
       "bed_id":     bedID,
       "unit_id":    unitID,
       "timestamp":  time.Now().Unix(),
   }
   rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
   ```

3. **床位状态变化**
   - 当 `beds.bound_device_count` 变化时（通过数据库触发器或应用层）
   - 发布 `bed.status_changed` 事件

4. **单元信息变化**
   - 当 `units.unit_name`、`units.branch_tag` 等变化时
   - 发布 `unit.info_changed` 事件

## 使用方式

### 环境变量配置

```bash
# 启用事件驱动模式
export CARD_TRIGGER_MODE=events

# Redis Streams 配置
export CARD_EVENT_STREAM=card:events
export CARD_CONSUMER_GROUP=card-aggregator-group
export CARD_CONSUMER_NAME=card-aggregator-1

# 租户ID（必需）
export TENANT_ID=your-tenant-id

# Redis 配置
export REDIS_ADDR=localhost:6379
export REDIS_PASSWORD=
```

### 运行服务

```bash
# 事件驱动模式
CARD_TRIGGER_MODE=events ./wisefido-card-aggregator

# 轮询模式（备用）
CARD_TRIGGER_MODE=polling ./wisefido-card-aggregator
```

## 工作流程

### 事件驱动模式

1. **API 层**：设备/住户绑定关系变化
2. **发布事件**：发送到 Redis Streams `card:events`
3. **事件消费**：`wisefido-card-aggregator` 监听并消费事件
4. **触发计算**：根据事件类型，重新计算相关 unit 的卡片
5. **定时兜底**：每天上午9点全量更新，确保数据一致性

### 定时任务

- **触发时间**：每天上午9点
- **执行内容**：全量重新创建所有卡片
- **作用**：兜底机制，确保数据最终一致性（避免凌晨2点可能出现的意外情况）

## 注意事项

1. **事件幂等性**：确保重复处理事件不会导致数据不一致
2. **错误处理**：事件处理失败时，消息会保留在 Stream 中，可以重试
3. **消息确认**：处理成功后确认消息（ACK），避免重复处理
4. **并发控制**：同一 unit 的多个事件可能并发，当前实现会顺序处理
5. **事件发布**：需要在 API 层（wisefido-data）实现事件发布逻辑

## 下一步

1. ✅ **wisefido-card-aggregator**：事件消费和定时任务已实现
2. ⚠️ **wisefido-data API**：需要在设备/住户绑定 API 中发布事件
3. ⚠️ **测试**：需要测试事件驱动的完整流程

