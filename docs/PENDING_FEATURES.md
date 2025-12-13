# 待实现功能清单

本文档记录待实现的功能，方便后续开发时参考。

## 🎯 优先级 1：事件发布功能

### 功能描述
在 `wisefido-data` API 服务中实现事件发布功能，当设备/住户/床位绑定关系变化时，发布事件到 Redis Streams，触发 `wisefido-card-aggregator` 实时更新卡片。

### 当前状态
- ⏸️ **已暂停**：`wisefido-data` 服务尚未实现
- ✅ **临时方案**：使用轮询模式（每60秒全量更新卡片）
- 📝 **备注**：轮询模式已实现并可用，但延迟较高（最多60秒）

### 需要实现的内容

#### 1. 创建 wisefido-data 服务
- [ ] 创建项目结构（参考其他服务）
- [ ] 实现 HTTP API 框架（Gin 或 Echo）
- [ ] 实现认证中间件（JWT）
- [ ] 实现权限过滤

#### 2. 实现设备绑定/解绑 API
- [ ] 端点：`PUT /admin/api/v1/device/:id` 或 `PUT /device/api/v1/device/:id`
- [ ] 检测绑定关系变化（`bound_bed_id`、`bound_room_id`、`unit_id`、`monitoring_enabled`）
- [ ] 发布事件到 Redis Streams：`card:events`
  ```go
  event := map[string]interface{}{
      "event_type": "device.bound",  // 或 "device.unbound"
      "tenant_id":  tenantID,
      "device_id":  deviceID,
      "bed_id":     bedID,
      "unit_id":    unitID,
      "timestamp":  time.Now().Unix(),
  }
  rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
  ```

#### 3. 实现住户绑定/解绑 API
- [ ] 端点：`PUT /admin/api/v1/residents/:id`
- [ ] 检测绑定关系变化（`bed_id`、`unit_id`、`status`）
- [ ] 发布事件到 Redis Streams：`card:events`
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

#### 4. 实现单元信息更新 API
- [ ] 端点：`PUT /admin/api/v1/addresses/:id` 或 `PUT /admin/api/v1/units/:id`
- [ ] 检测信息变化（`unit_name`、`branch_tag`、`building` 等）
- [ ] 发布事件到 Redis Streams：`card:events`
  ```go
  event := map[string]interface{}{
      "event_type": "unit.info_changed",
      "tenant_id":  tenantID,
      "unit_id":    unitID,
      "timestamp":  time.Now().Unix(),
  }
  rediscommon.PublishToStream(ctx, redisClient, "card:events", event)
  ```

### 相关文档
- `wisefido-card-aggregator/docs/EVENT_TRIGGER_MECHANISM.md` - 事件触发机制说明
- `wisefido-card-aggregator/docs/EVENT_DRIVEN_IMPLEMENTATION.md` - 事件驱动实现方案
- `docs/EVENT_PUBLISH_EXPLANATION.md` - 事件发布机制详细说明
- `docs/system_architecture_complete.md` - 系统架构文档

### 实现后的效果
- ✅ 设备/住户绑定关系变化后，**实时**（秒级）更新卡片
- ✅ 减少数据库查询压力（只在变化时更新，而不是每60秒全量更新）
- ✅ 提高系统响应速度

---

## 🎯 优先级 2：增量更新功能

### 功能描述
优化 `wisefido-card-aggregator` 的卡片更新策略，从全量重建改为增量更新（只更新变化的卡片）。

### 当前状态
- ⏸️ **待实现**：当前为全量重建（DELETE + INSERT）
- 📝 **备注**：功能已设计，待实现

### 需要实现的内容
- [ ] 比较现有卡片和期望卡片
- [ ] 只 CREATE 新卡片
- [ ] 只 UPDATE 变化的卡片
- [ ] 只 DELETE 不再需要的卡片

### 相关文档
- `wisefido-card-aggregator/docs/CARD_UPDATE_STRATEGIES.md` - 卡片更新策略说明

---

## 🎯 优先级 3：其他功能

### wisefido-alarm 服务
- [ ] 实现报警规则评估
- [ ] 实现 AI 智能评估（可选）

### wisefido-card-aggregator 数据聚合
- [ ] 实现卡片数据聚合（从 Redis 读取实时数据和报警数据）
- [ ] 组装完整的 VitalFocusCard 对象

### wisefido-data HTTP API
- [ ] 实现 HTTP API 端点
- [ ] 实现权限过滤
- [ ] 实现 Focus 过滤

---

## 📝 备注

### 当前使用的临时方案
- **轮询模式**：`wisefido-card-aggregator` 每60秒全量更新所有卡片
- **优点**：简单可靠，不依赖外部服务
- **缺点**：延迟较高（最多60秒），资源消耗较大

### 切换到事件驱动模式
当 `wisefido-data` 服务实现后，可以通过配置切换到事件驱动模式：
```bash
export CARD_TRIGGER_MODE=events  # 从 polling 改为 events
```

### 更新日期
- 创建日期：2024-12-19
- 最后更新：2024-12-19

