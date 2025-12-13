# 数据定义分析

## 🤔 问题

用户指出：我没有想清楚为什么要这么改，数据定义对不对。

## 📊 当前情况

### 1. 数据库定义 (`cards.sql`)

```sql
-- 设备列表：存储该卡片绑定的所有设备信息（预计算结果）
-- 格式：[{"device_id": "...", "device_name": "...", "device_type": "...", "device_model": "...", "binding_type": "direct|indirect"}, ...]
devices JSONB DEFAULT '[]'::jsonb,
```

**问题**：
- 数据库定义中**没有** `bed_id`, `room_id`, `unit_id` 字段
- 只有 `binding_type` 字段

### 2. 融合逻辑的需求

**场景 A**（门牌下只有 1 个 ActiveBed）：
- ActiveBed 卡片包含：
  - 床上的设备（`bound_bed_id = bed_id`，`binding_type = "direct"`）
  - 未绑床的设备（`bound_bed_id IS NULL`，`binding_type = "indirect"`）

**融合要求**：
- 只融合床上的设备（`bed_id` 有效且相同）
- 不融合未绑床的设备（`bed_id` 为 NULL）

### 3. 我的修改

我在 `DeviceJSON` 中添加了 `bed_id`, `room_id`, `unit_id` 字段，但这与数据库定义不一致。

## ❓ 需要澄清的问题

1. **数据库定义是否需要更新？**
   - `cards.sql` 中的 `devices` JSONB 格式是否需要包含 `bed_id`, `room_id`, `unit_id`？
   - 还是应该保持原样，只包含 `binding_type`？

2. **融合逻辑的正确实现方式？**
   - 方案1：使用 `cards.bed_id`（卡片级别的 bed_id）
     - 对于 ActiveBed 卡片，`cards.bed_id` 已经存在
     - 融合时，只融合 `binding_type = "direct"` 的设备（因为 `binding_type = "direct"` 表示绑定到床）
     - **问题**：但用户说 `binding_type` 不完整
   - 方案2：在 `cards.devices` JSONB 中添加 `bed_id` 字段
     - 需要更新数据库定义
     - 融合时，比较每个设备的 `bed_id` 与卡片的 `bed_id`
     - **问题**：数据冗余（卡片的 `bed_id` 和设备 JSONB 中的 `bed_id`）

3. **用户的要求**
   - 用户说：应该用设备绑定的 `unit_id`, `room_id`, `room_name`, `bed_id` 来判断
   - 如果 `bed_id` 有效，则所有 `bed_id` 的 device 都是绑在同一床上的
   - **这意味着**：`cards.devices` JSONB 应该包含这些字段

## 💡 可能的解决方案

### 方案 A：更新数据库定义（推荐）

更新 `cards.sql`，在 `devices` JSONB 中添加 `bed_id`, `room_id`, `unit_id` 字段：

```sql
-- 格式：[{"device_id": "...", "device_name": "...", "device_type": "...", "device_model": "...", "binding_type": "direct|indirect", "bed_id": "...", "room_id": "...", "unit_id": "..."}, ...]
```

**优点**：
- 融合逻辑可以直接从 JSONB 中获取 `bed_id`，不需要额外查询
- 数据完整，符合用户要求

**缺点**：
- 需要更新数据库定义
- 数据冗余（卡片的 `bed_id` 和设备 JSONB 中的 `bed_id`）

### 方案 B：使用卡片级别的 bed_id

对于 ActiveBed 卡片，使用 `cards.bed_id` 来判断：
- 只融合 `binding_type = "direct"` 的设备（因为 `binding_type = "direct"` 表示绑定到床，而床的 `bed_id` 就是卡片的 `bed_id`）

**优点**：
- 不需要更新数据库定义
- 不需要在 JSONB 中存储冗余数据

**缺点**：
- 用户说 `binding_type` 不完整
- 如果场景 A 中有多个床（虽然规则说只有 1 个），可能会有问题

### 方案 C：查询 devices 表

融合时，不依赖 `cards.devices` JSONB，而是查询 `devices` 表获取每个设备的 `bound_bed_id`。

**优点**：
- 数据准确，直接从源表获取

**缺点**：
- 需要额外的数据库查询
- 性能较差

## 🎯 建议

**我建议采用方案 A**：
1. 更新 `cards.sql` 的注释，说明 `devices` JSONB 应该包含 `bed_id`, `room_id`, `unit_id` 字段
2. 更新 `wisefido-card-aggregator` 的 `DeviceJSON`，包含这些字段
3. 更新 `wisefido-sensor-fusion` 的融合逻辑，使用 `bed_id` 来判断

**但需要确认**：
- 数据库定义是否需要更新？
- 还是应该保持原样，使用其他方式实现？

