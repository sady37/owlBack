# 事件发布机制说明

## 什么是事件发布？

**事件发布**是一种**解耦**的通信方式：当一个服务完成某个操作后，它**发布一个事件**到消息队列（Redis Streams），其他服务可以**监听并消费**这个事件，然后执行相应的操作。

## 为什么需要事件发布？

### 问题场景

假设没有事件发布机制：

```
用户在前端操作：绑定设备到床位
    ↓
wisefido-data API 服务
    ├─ 更新数据库：devices.bound_bed_id = 'bed-123'
    └─ ❌ 问题：如何通知 wisefido-card-aggregator 重新计算卡片？
```

**问题**：
- `wisefido-data` 服务不知道 `wisefido-card-aggregator` 的存在
- `wisefido-card-aggregator` 不知道什么时候需要重新计算卡片
- 如果直接调用，两个服务会**紧耦合**，难以维护

### 解决方案：事件发布

```
用户在前端操作：绑定设备到床位
    ↓
wisefido-data API 服务
    ├─ 更新数据库：devices.bound_bed_id = 'bed-123'
    └─ 发布事件到 Redis Streams：card:events
        {
            "event_type": "device.bound",
            "tenant_id": "tenant-123",
            "device_id": "device-456",
            "bed_id": "bed-123",
            "unit_id": "unit-789"
        }
    ↓
Redis Streams: card:events
    ↓
wisefido-card-aggregator（监听事件）
    ├─ 收到事件：device.bound
    └─ 重新计算卡片：CreateCardsForUnit(tenantID, unitID)
```

**优势**：
- ✅ **解耦**：`wisefido-data` 不需要知道 `wisefido-card-aggregator` 的存在
- ✅ **异步**：事件发布是异步的，不会阻塞 API 响应
- ✅ **可扩展**：未来可以添加更多监听者（如日志服务、通知服务等）

## 事件发布流程

### 1. 事件发布者（wisefido-data API 服务）

当设备绑定关系变化时：

```go
// 在设备更新 API handler 中
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
    // 1. 更新数据库
    err := h.repo.UpdateDevice(deviceID, updates)
    if err != nil {
        // 处理错误
        return
    }
    
    // 2. 检测绑定关系是否变化
    if oldBedID != newBedID {
        // 3. 发布事件到 Redis Streams
        event := map[string]interface{}{
            "event_type": "device.bound",  // 或 "device.unbound"
            "tenant_id":  tenantID,
            "device_id":  deviceID,
            "bed_id":     newBedID,
            "unit_id":    unitID,  // 从 bed 查询得到
            "timestamp":  time.Now().Unix(),
        }
        
        // 发布到 Redis Streams
        _, err = rediscommon.PublishToStream(
            ctx,
            redisClient,
            "card:events",  // Stream 名称
            event,
        )
        if err != nil {
            // 记录日志，但不影响 API 响应
            log.Error("Failed to publish event", err)
        }
    }
    
    // 4. 返回 API 响应
    c.JSON(200, response)
}
```

### 2. 事件消费者（wisefido-card-aggregator）

监听 Redis Streams 并处理事件：

```go
// 在 wisefido-card-aggregator 服务中
func (c *EventConsumer) Start(ctx context.Context) error {
    // 1. 监听 Redis Streams: card:events
    for {
        messages, err := rediscommon.ReadFromStream(
            ctx, redisClient, "card:events",
            "card-aggregator-group", "card-aggregator-1", 10,
        )
        
        // 2. 处理每条消息
        for _, msg := range messages {
            eventType := msg.Values["event_type"].(string)
            tenantID := msg.Values["tenant_id"].(string)
            unitID := msg.Values["unit_id"].(string)
            
            // 3. 根据事件类型触发卡片重新计算
            switch eventType {
            case "device.bound", "device.unbound":
                // 重新计算该 unit 的卡片
                c.cardCreator.CreateCardsForUnit(tenantID, unitID)
            case "resident.bound", "resident.unbound":
                // 重新计算该 unit 的卡片
                c.cardCreator.CreateCardsForUnit(tenantID, unitID)
            }
        }
    }
}
```

## 完整流程示例

### 场景：用户在前端绑定设备到床位

```
1. 用户操作
   前端：点击"绑定设备到床位"
   ↓
2. API 请求
   PUT /admin/api/v1/device/device-123
   {
       "bound_bed_id": "bed-456"
   }
   ↓
3. wisefido-data API 服务
   ├─ 更新数据库：devices.bound_bed_id = 'bed-456'
   ├─ 查询 bed-456 的 unit_id = 'unit-789'
   └─ 发布事件到 Redis Streams
       {
           "event_type": "device.bound",
           "tenant_id": "tenant-123",
           "device_id": "device-123",
           "bed_id": "bed-456",
           "unit_id": "unit-789",
           "timestamp": 1234567890
       }
   ↓
4. Redis Streams: card:events
   （消息队列，持久化存储）
   ↓
5. wisefido-card-aggregator（监听）
   ├─ 收到事件：device.bound
   ├─ 解析事件：unit_id = 'unit-789'
   └─ 重新计算卡片：CreateCardsForUnit('tenant-123', 'unit-789')
       ├─ 查询该 unit 下的所有设备、床位、住户
       ├─ 根据规则创建/更新 cards 表
       └─ 完成
   ↓
6. 结果
   ✅ cards 表已更新，包含新绑定的设备
   ✅ wisefido-sensor-fusion 可以查询到正确的卡片
```

## 事件类型

根据 `wisefido-card-aggregator` 的实现，需要发布以下事件：

### 1. 设备绑定事件
- **事件类型**：`device.bound`、`device.unbound`、`device.monitoring_changed`
- **触发时机**：
  - 设备绑定到床位（`bound_bed_id` 变化）
  - 设备绑定到房间（`bound_room_id` 变化）
  - 设备监护状态变化（`monitoring_enabled` 变化）

### 2. 住户绑定事件
- **事件类型**：`resident.bound`、`resident.unbound`、`resident.status_changed`
- **触发时机**：
  - 住户绑定到床位（`bed_id` 变化）
  - 住户绑定到单元（`unit_id` 变化）
  - 住户状态变化（`status` 变化）

### 3. 床位状态事件
- **事件类型**：`bed.status_changed`、`bed.device_count_changed`
- **触发时机**：
  - 床位设备数量变化（`beds.bound_device_count` 变化）
  - 床位激活状态变化（`beds.is_active` 变化）

### 4. 单元信息事件
- **事件类型**：`unit.info_changed`
- **触发时机**：
  - 单元名称变化（`units.unit_name` 变化）
  - 位置标签变化（`units.branch_tag` 变化）

## 当前状态

### ✅ 已实现
- **wisefido-card-aggregator** 的事件消费者（监听 Redis Streams）
- **事件处理逻辑**（根据事件类型重新计算卡片）

### ❌ 未实现
- **wisefido-data API 服务**（根本不存在）
- **事件发布功能**（需要在 API 服务中实现）

## 总结

**事件发布**就是：
1. **当某个操作发生时**（如设备绑定），**发布一个消息**到 Redis Streams
2. **其他服务监听这个消息**（如 wisefido-card-aggregator），**执行相应的操作**（如重新计算卡片）

**好处**：
- 服务之间**解耦**，不需要直接调用
- **异步处理**，不阻塞 API 响应
- **可扩展**，可以添加更多监听者

**当前问题**：
- `wisefido-data` 服务**根本不存在**，所以无法发布事件
- 需要先**创建 `wisefido-data` 服务**，然后**实现事件发布功能**

