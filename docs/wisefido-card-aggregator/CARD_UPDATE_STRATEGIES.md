# 卡片更新策略说明

## 当前实现 vs 待实现功能

### 1. 事件驱动模式（当前仅轮询模式）

#### 当前实现：轮询模式
```go
// 每 60 秒（可配置）定时轮询，全量重新创建所有卡片
for {
    select {
    case <-ticker.C:
        // 为所有 unit 创建卡片
        s.createAllCards(ctx)
    }
}
```

**特点**：
- ✅ 简单可靠，不依赖外部事件
- ❌ 延迟：最多等待一个轮询周期（如60秒）才能反映变化
- ❌ 资源消耗：即使没有变化，也会全量重新计算

**触发时机**：
- 定时触发（如每60秒）
- 无论数据是否变化，都会执行

#### 待实现：事件驱动模式
```go
// 监听设备/住户/床位绑定关系变化事件
// 当变化发生时，立即触发卡片重新计算
eventStream.Subscribe("device.bound", func(event Event) {
    // 立即重新计算相关卡片
    cardCreator.CreateCardsForUnit(tenantID, unitID)
})
```

**特点**：
- ✅ 实时响应：数据变化后立即更新
- ✅ 资源高效：只在有变化时才执行
- ❌ 复杂度：需要事件监听机制（Redis Streams、PostgreSQL NOTIFY/LISTEN 等）

**触发时机**：
- 设备绑定/解绑床位时
- 住户绑定/解绑床位时
- 床位状态变化时（ActiveBed ↔ NonActiveBed）
- 单元信息变化时（地址、名称等）

**触发场景**（根据 `21_cards.sql` 文档）：
1. 床位绑定关系变化：`residents.bed_id` 变化、`devices.bound_bed_id` 变化
2. 门牌号下住户变化：`residents.unit_id` 变化、`residents.status` 变化
3. 设备绑定关系变化：`devices.unit_id`、`devices.bound_room_id`、`devices.bound_bed_id` 变化
4. 地址信息变化：`units`、`rooms`、`beds` 名称变化

---

### 2. 增量更新（当前为全量重建）

#### 当前实现：全量重建
```go
func (c *CardCreator) CreateCardsForUnit(tenantID, unitID string) error {
    // 1. 删除该 unit 下的所有旧卡片
    c.repo.DeleteCardsByUnit(tenantID, unitID)
    
    // 2. 重新创建所有卡片
    // 场景 A/B/C 的创建逻辑...
}
```

**特点**：
- ✅ 简单：不需要比较新旧数据
- ✅ 保证一致性：删除后重建，确保没有残留数据
- ❌ 性能：即使只有1个卡片变化，也会删除并重建所有卡片
- ❌ 数据丢失风险：删除和创建之间有时间窗口

**更新策略**：
- 每次都是 `DELETE` 所有旧卡片
- 然后 `INSERT` 新卡片
- 即使卡片内容没有变化，也会删除重建

#### 待实现：增量更新
```go
func (c *CardCreator) UpdateCardsForUnit(tenantID, unitID string) error {
    // 1. 获取当前数据库中的卡片
    existingCards := c.repo.GetCardsByUnit(tenantID, unitID)
    
    // 2. 计算应该存在的卡片
    expectedCards := c.calculateExpectedCards(tenantID, unitID)
    
    // 3. 比较差异
    toCreate, toUpdate, toDelete := c.diffCards(existingCards, expectedCards)
    
    // 4. 只更新变化的卡片
    for _, card := range toCreate {
        c.repo.CreateCard(...)
    }
    for _, card := range toUpdate {
        c.repo.UpdateCard(...)  // 使用 UPDATE 而不是 DELETE+INSERT
    }
    for _, card := range toDelete {
        c.repo.DeleteCard(card.CardID)
    }
}
```

**特点**：
- ✅ 性能：只更新变化的卡片
- ✅ 数据安全：保留未变化的卡片，减少数据丢失风险
- ✅ 可追踪：可以记录哪些卡片被更新了
- ❌ 复杂度：需要比较逻辑、处理并发更新等

**更新策略**：
- 比较现有卡片和期望卡片
- 只 `CREATE` 新卡片
- 只 `UPDATE` 变化的卡片（更新 `devices`、`residents`、`card_name`、`card_address` 等字段）
- 只 `DELETE` 不再需要的卡片

---

## 总结

### 是的，这两个功能都是指卡片的新建或更新

1. **事件驱动模式**：
   - **触发时机**：何时触发卡片更新
   - 当前：定时触发（轮询）
   - 目标：事件触发（实时）

2. **增量更新**：
   - **更新策略**：如何更新卡片
   - 当前：全量删除重建
   - 目标：只更新变化的卡片

### 组合效果

| 模式 | 触发时机 | 更新策略 | 效果 |
|------|---------|---------|------|
| **当前实现** | 轮询（每60秒） | 全量重建 | 延迟高，资源消耗大 |
| **目标1** | 事件驱动 | 全量重建 | 实时响应，但资源消耗仍大 |
| **目标2** | 轮询 | 增量更新 | 延迟高，但资源消耗小 |
| **最终目标** | 事件驱动 | 增量更新 | 实时响应，资源高效 ⭐ |

### 建议实现顺序

1. **先实现增量更新**（在轮询模式下）
   - 可以立即提升性能
   - 为事件驱动模式打好基础

2. **再实现事件驱动模式**
   - 结合增量更新，达到最佳效果

