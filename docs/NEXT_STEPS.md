# 下一步行动计划

## 当前状态

### ✅ 已完成
1. **wisefido-card-aggregator 卡片创建功能**
   - ✅ 卡片创建规则实现（场景 A/B/C）
   - ✅ 事件驱动模式（事件消费者）
   - ✅ 定时任务（每天上午9点）

### ⚠️ 待实现
1. **wisefido-data API 服务的事件发布功能**
   - ⚠️ 设备绑定/解绑 API 需要发布事件
   - ⚠️ 住户绑定/解绑 API 需要发布事件
   - ⚠️ 单元信息更新 API 需要发布事件

## 下一步：实现 wisefido-data API 服务的事件发布功能

### 目标
在 `wisefido-data` API 服务中，当设备/住户/床位绑定关系变化时，发布事件到 Redis Streams `card:events`，触发 `wisefido-card-aggregator` 重新计算卡片。

### 需要实现的 API 端点

#### 1. 设备绑定/解绑 API
- **端点**：`PUT /admin/api/v1/device/:id` 或 `PUT /device/api/v1/device/:id`
- **触发条件**：当 `bound_bed_id`、`bound_room_id`、`unit_id`、`monitoring_enabled` 变化时
- **发布事件**：
  ```go
  event := map[string]interface{}{
      "event_type": "device.bound",  // 或 "device.unbound"
      "tenant_id":  tenantID,
      "device_id":  deviceID,
      "bed_id":     bedID,      // 如果绑定到床
      "unit_id":    unitID,     // 从 bed 或 room 查询得到
      "timestamp":  time.Now().Unix(),
      "metadata": {
          "old_bed_id": oldBedID,
          "new_bed_id": newBedID,
      },
  }
  rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
  ```

#### 2. 住户绑定/解绑 API
- **端点**：`PUT /admin/api/v1/residents/:id`
- **触发条件**：当 `bed_id`、`unit_id`、`status` 变化时
- **发布事件**：
  ```go
  event := map[string]interface{}{
      "event_type": "resident.bound",  // 或 "resident.unbound"
      "tenant_id":  tenantID,
      "resident_id": residentID,
      "bed_id":     bedID,
      "unit_id":    unitID,
      "timestamp":  time.Now().Unix(),
  }
  rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
  ```

#### 3. 单元信息更新 API
- **端点**：`PUT /admin/api/v1/addresses/:id` 或 `PUT /admin/api/v1/units/:id`
- **触发条件**：当 `unit_name`、`branch_tag`、`building` 等变化时
- **发布事件**：
  ```go
  event := map[string]interface{}{
      "event_type": "unit.info_changed",
      "tenant_id":  tenantID,
      "unit_id":    unitID,
      "timestamp":  time.Now().Unix(),
  }
  rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
  ```

### 实现步骤

#### 步骤 1：检查 wisefido-data 服务是否存在
- [ ] 检查 `wisefido-data` 服务是否已创建
- [ ] 如果不存在，需要先创建项目结构

#### 步骤 2：添加 Redis 客户端
- [ ] 在配置中添加 Redis 配置
- [ ] 初始化 Redis 客户端
- [ ] 添加事件发布工具函数

#### 步骤 3：在设备绑定 API 中添加事件发布
- [ ] 找到设备更新 API 的 handler
- [ ] 检测绑定关系变化（`bound_bed_id`、`bound_room_id`、`unit_id`）
- [ ] 发布 `device.bound` 或 `device.unbound` 事件

#### 步骤 4：在住户绑定 API 中添加事件发布
- [ ] 找到住户更新 API 的 handler
- [ ] 检测绑定关系变化（`bed_id`、`unit_id`、`status`）
- [ ] 发布 `resident.bound` 或 `resident.unbound` 事件

#### 步骤 5：在单元信息更新 API 中添加事件发布
- [ ] 找到单元更新 API 的 handler
- [ ] 检测信息变化（`unit_name`、`branch_tag` 等）
- [ ] 发布 `unit.info_changed` 事件

#### 步骤 6：测试
- [ ] 测试设备绑定 → 事件发布 → 卡片更新流程
- [ ] 测试住户绑定 → 事件发布 → 卡片更新流程
- [ ] 测试单元信息更新 → 事件发布 → 卡片更新流程
- [ ] 验证定时任务（每天上午9:00）正常工作

### 注意事项

1. **事件幂等性**：确保重复处理事件不会导致数据不一致
2. **错误处理**：事件发布失败不应该影响 API 的正常响应
3. **事务一致性**：数据库更新和事件发布应该在同一个事务中，或者使用补偿机制
4. **unit_id 查询**：如果只有 `bed_id`，需要查询对应的 `unit_id`

### 相关文档

- `wisefido-card-aggregator/docs/EVENT_TRIGGER_MECHANISM.md` - 事件触发机制说明
- `wisefido-card-aggregator/docs/EVENT_DRIVEN_IMPLEMENTATION.md` - 事件驱动实现方案
- `docs/system_architecture_complete.md` - 系统架构文档

