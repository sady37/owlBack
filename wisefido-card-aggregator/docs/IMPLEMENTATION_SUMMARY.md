# wisefido-card-aggregator 实现总结

## ✅ 已完成功能

### 1. 核心功能
- ✅ ActiveBed 判断逻辑
- ✅ 卡片创建场景 A/B/C
- ✅ 卡片名称和地址计算
- ✅ 设备绑定规则
- ✅ 报警路由配置转换（从 units.groupList/userList）

### 2. 事件驱动模式
- ✅ Redis Streams 事件消费者
- ✅ 支持多种事件类型（device/resident/bed/unit）
- ✅ 错误处理和重试机制
- ✅ 消息确认机制

### 3. 定时任务
- ✅ 每天上午9点全量更新
- ✅ 与事件驱动模式并行运行

### 4. 测试
- ✅ 配置模块测试（100% 覆盖率）
- ✅ 路由转换函数测试
- ✅ Repository 层测试（44.7% 覆盖率）
- ✅ Aggregator 层测试（69.0% 覆盖率）
- ✅ 总体覆盖率：~53%

## ⚠️ 待实现功能

### 1. API 层事件发布（wisefido-data 服务）⏸️ **已暂停，使用轮询模式**
- 📝 **当前状态**：`wisefido-data` 服务尚未实现，暂时使用轮询模式（每60秒全量更新）
- 📝 **待实现文档**：`docs/PENDING_FEATURES.md`
- 需要在以下 API 中发布事件：
  - 设备绑定/解绑 API
  - 住户绑定/解绑 API
  - 床位状态变化处理
  - 单元信息变化处理

### 2. 增量更新
- ⚠️ 当前为全量重建（DELETE + INSERT）
- ⚠️ 目标：只更新变化的卡片（CREATE/UPDATE/DELETE）

## 事件触发机制

### 如何触发？

**方案：Redis Streams 事件总线**

1. **API 层发布事件**（需要在 wisefido-data 服务中实现）
   ```go
   // 当设备绑定变化时
   event := map[string]interface{}{
       "event_type": "device.bound",
       "tenant_id":  tenantID,
       "unit_id":    unitID,
       "bed_id":     bedID,
       "device_id":  deviceID,
   }
   rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
   ```

2. **wisefido-card-aggregator 消费事件**
   - 监听 Redis Streams `card:events`
   - 解析事件并触发卡片重新计算
   - 已实现 ✅

3. **定时兜底**
   - 每天上午9点全量更新
   - 已实现 ✅

### 事件发布位置

需要在以下位置发布事件：

1. **设备绑定 API** (`wisefido-data` 服务)
   - `PUT /api/devices/:id` - 更新设备绑定关系
   - 当 `bound_bed_id`、`bound_room_id`、`unit_id` 变化时

2. **住户绑定 API** (`wisefido-data` 服务)
   - `PUT /api/residents/:id` - 更新住户绑定关系
   - 当 `bed_id`、`unit_id`、`status` 变化时

3. **床位状态变化**（可通过数据库触发器或应用层）
   - 当 `beds.bound_device_count` 变化时
   - 触发 `bed.status_changed` 事件

4. **单元信息变化** (`wisefido-data` 服务)
   - `PUT /api/units/:id` - 更新单元信息
   - 当 `unit_name`、`branch_tag` 等变化时

## 配置说明

### 事件驱动模式

```bash
# 启用事件驱动模式
export CARD_TRIGGER_MODE=events

# Redis Streams 配置
export CARD_EVENT_STREAM=card:events
export CARD_CONSUMER_GROUP=card-aggregator-group
export CARD_CONSUMER_NAME=card-aggregator-1

# 租户ID
export TENANT_ID=your-tenant-id

# Redis 配置
export REDIS_ADDR=localhost:6379
```

### 轮询模式（备用）

```bash
# 使用轮询模式
export CARD_TRIGGER_MODE=polling
export CARD_POLLING_INTERVAL=60  # 秒
```

## 总结

✅ **wisefido-card-aggregator 的事件消费和定时任务已实现**

⚠️ **需要在 wisefido-data API 服务中实现事件发布**

事件触发流程：
1. API 层检测到绑定关系变化 → 发布事件到 Redis Streams
2. wisefido-card-aggregator 监听事件 → 触发卡片重新计算
3. 每天上午9点定时全量更新 → 确保数据一致性（避免凌晨意外）

