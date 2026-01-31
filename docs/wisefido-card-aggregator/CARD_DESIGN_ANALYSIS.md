# 卡片设计分析文档

## 1. 事件驱动模式配置状态

### 当前配置

**启动脚本**：`owlBack/start_all_services.sh` (第 78 行)
```bash
export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
```

**默认值**：`polling`（轮询模式）

### 问题分析

#### ❌ 当前状态：未启用事件驱动模式

- **默认模式**：`polling`（每 60 分钟轮询一次）
- **事件发布**：✅ 已实现（`device.bound/unbound/monitoring_changed`, `resident.caregivers_changed`）
- **事件消费**：✅ 已实现（`event_consumer.go`）
- **实际运行**：❌ 默认使用轮询模式，事件不会被实时处理

#### ⚠️ 影响

1. **延迟问题**：
   - 设备绑定/解绑后，最多延迟 60 分钟才更新卡片
   - Resident 绑定/解绑后，最多延迟 60 分钟才更新卡片
   - `resident_caregivers` 更新后，最多延迟 60 分钟才更新卡片

2. **实时性**：
   - 如果目标是实时响应数据变化，当前配置无法满足
   - 如果目标是最终一致性（1 小时内），当前配置可以接受

### 解决方案

#### 方案 A：修改启动脚本默认值（推荐）

```bash
# owlBack/start_all_services.sh 第 78 行
export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-events}"  # 改为 events
```

**优点**：
- 默认启用事件驱动，实时响应
- 保留轮询作为兜底（每天上午 9 点全量更新）

**缺点**：
- 需要确保 Redis Streams 正常工作
- 如果 Redis 不可用，服务可能无法启动

#### 方案 B：保持默认 polling，但提供明确说明

在启动脚本中添加提示：
```bash
if [ "$CARD_TRIGGER_MODE" = "polling" ]; then
    echo -e "${YELLOW}Note: Using polling mode (60 min interval).${NC}"
    echo -e "${YELLOW}      To enable real-time updates, set: CARD_TRIGGER_MODE=events${NC}"
fi
```

**优点**：
- 更安全（不依赖 Redis）
- 用户可以根据需要选择

**缺点**：
- 默认情况下延迟较高

### 建议

**从目标考虑**：
- **如果目标是实时响应**：应该修改默认值为 `events`
- **如果目标是最终一致性**：保持 `polling` 模式，但应该缩短轮询间隔（如 5-10 分钟）

---

## 2. `resident_id` 字段命名分析

### 当前设计

**字段名**：`resident_id`
**含义**：Primary resident for ActiveBed (Location may be NULL)

### 使用场景分析

#### 场景 1：ActiveBed 卡片
- **当前逻辑**：`resident_id` = bed 上绑定的 resident_id
- **用途**：
  - 权限过滤（`cards_repository.go:120`）：`c.card_type = 'ActiveBed' AND c.resident_id = rc.resident_id`
  - 快速查询主住户信息

#### 场景 2：Location 卡片
- **当前逻辑**：`resident_id` = `NULL`
- **用途**：无（Location 卡片不使用此字段）

### 命名建议分析

#### 选项 A：保持 `resident_id`
**优点**：
- 简洁明了
- 已广泛使用（权限过滤、查询等）
- 符合常见数据库设计习惯

**缺点**：
- 对于 Location 卡片，含义不够明确（为什么是 NULL？）

#### 选项 B：改为 `bed_resident_id`
**优点**：
- 更明确表示这是 bed 上的 resident
- 对于 Location 卡片，NULL 的含义更清晰（Location 没有 bed）

**缺点**：
- 需要修改所有使用此字段的代码
- 数据库迁移成本
- 如果未来 Location 卡片也需要主住户，命名会变得不准确

### 建议

**保持 `resident_id`**，原因：
1. 当前命名已经足够清晰（注释说明了用途）
2. 修改成本高（涉及权限过滤、查询等多个地方）
3. 如果改为 `bed_resident_id`，未来 Location 卡片需要主住户时会变得不准确

---

## 3. Location 卡片的 `resident_id` 逻辑分析

### 当前逻辑

**Location 卡片**：`resident_id` = `NULL`

**代码位置**：`card_creator.go:calculateExpectedUnitCard` (第 801 行)
```go
return &ExpectedCard{
    CardType:      "Location",
    BedID:         nil,
    UnitID:        unitInfo.UnitID,
    CardName:      cardName,
    CardAddress:   cardAddress,
    ResidentID:    nil,  // ← 当前总是 NULL
    DevicesJSON:   devicesJSON,
    ResidentsJSON: residentsJSON,
}
```

### 建议逻辑分析

**建议**：
- 如果不是 `sharedUnit` → `resident_id` = residents 中的第 1 个
- 如果是 `sharedUnit` → `resident_id` = `NULL`

### 可行性分析

#### ✅ 技术可行性：可行

可以修改 `calculateExpectedUnitCard` 方法：
```go
var residentID *string
if !unitInfo.IsSharedUnit {
    // 非 SharedUnit：使用第一个 resident
    if len(residents) > 0 {
        residentID = &residents[0].ResidentID
    }
} else {
    // SharedUnit：保持 NULL
    residentID = nil
}
```

#### ⚠️ 业务逻辑考虑

**问题 1：Location 卡片的主住户概念**
- Location 卡片代表整个 unit（可能有多个 residents）
- 选择"第一个 resident"作为主住户是否合理？
- 如果 residents 列表顺序变化，`resident_id` 也会变化

**问题 2：权限过滤影响**
- 当前权限过滤逻辑（`cards_repository.go:199-220`）：
  - ActiveBed：直接匹配 `c.resident_id`
  - Location：从 `c.residents` JSONB 数组中提取 `resident_id` 进行匹配
- 如果 Location 卡片也有 `resident_id`，权限过滤逻辑需要调整

**问题 3：一致性**
- ActiveBed：`resident_id` = bed 上绑定的 resident（明确、稳定）
- Location：`resident_id` = residents 列表的第一个（可能变化）

### 建议

**暂不建议修改**，原因：
1. **业务语义不清晰**：Location 卡片代表整个 unit，不应该有"主住户"概念
2. **数据稳定性**：residents 列表顺序可能变化，导致 `resident_id` 不稳定
3. **权限过滤复杂化**：当前权限过滤逻辑已经能正确处理 Location 卡片（从 JSONB 数组提取）

**如果确实需要**：
- 可以考虑添加新字段 `primary_resident_id`（用于显示），而不是修改 `resident_id` 的语义
- 或者使用 `card_name` 来显示主住户信息（当前已经这样做了）

---

## 4. JSONB 字段设计分析（devices / residents）

### 当前设计

#### `devices` JSONB 格式（`card.go:DeviceJSON`）
```json
[
  {
    "device_id": "uuid",        // ✅ ID
    "device_name": "Radar-XXXX", // ✅ Name
    "device_type": "Radar",
    "device_model": "BM8701-2",
    "bed_id": "uuid",           // ✅ ID
    "bed_name": "Bed 1",        // ✅ Name
    "room_id": "uuid",          // ✅ ID
    "room_name": "Room 1",      // ✅ Name
    "unit_id": "uuid"           // ✅ ID
  }
]
```

#### `residents` JSONB 格式（`card.go:ResidentJSON`）
```json
[
  {
    "resident_id": "uuid",      // ✅ ID
    "nickname": "John Doe"      // ✅ Name
  }
]
```

### 分析结果

#### ✅ 当前设计已经是 `id:name` 模式

**优点**：
1. **程序传递用 ID**：所有关联关系都使用 ID（`device_id`, `resident_id`, `bed_id`, `room_id`, `unit_id`）
2. **显示用 Name**：同时提供 name 字段（`device_name`, `nickname`, `bed_name`, `room_name`）
3. **数据一致性**：即使 name 变化，ID 保持不变，关联关系不会断裂

#### ⚠️ 潜在问题

**问题 1：Name 可能过时**
- 如果 `device_name` 或 `nickname` 在数据库中更新，JSONB 中的 name 不会自动更新
- 需要重新生成卡片才能更新 JSONB 中的 name

**问题 2：冗余存储**
- JSONB 中同时存储 ID 和 Name，占用更多空间
- 但考虑到查询性能（避免 JOIN），这是合理的权衡

### 使用场景分析

#### 场景 1：权限过滤（`cards_repository.go`）
```sql
-- 使用 ID 进行匹配
(c.residents->0->>'resident_id')::text = $resident_id
```
✅ **正确**：使用 ID，不依赖 Name

#### 场景 2：前端显示
- 前端可以直接使用 JSONB 中的 `device_name`、`nickname` 显示
- 不需要额外查询数据库获取 name

#### 场景 3：数据更新
- 如果 name 变化，需要重新生成卡片才能更新 JSONB
- 但 ID 不变，关联关系保持稳定

### 建议

#### ✅ 保持当前设计（id:name 模式）

**理由**：
1. **已经实现了 id:name 模式**：程序使用 ID，显示使用 Name
2. **性能优化**：避免频繁 JOIN 查询 name
3. **数据稳定性**：ID 不变，关联关系稳定

#### ⚠️ 注意事项

1. **Name 更新策略**：
   - 当前：通过事件驱动或轮询重新生成卡片
   - 如果 name 变化频繁，可以考虑缩短轮询间隔或确保事件及时触发

2. **数据一致性**：
   - JSONB 中的 name 可能与数据库中的 name 不同步
   - 这是可接受的权衡（最终一致性）

3. **如果确实需要实时 name**：
   - 可以考虑在查询时 JOIN 数据库获取最新 name
   - 但会增加查询复杂度

---

## 5. 总结与建议

### 1. 事件驱动模式

**当前状态**：❌ 未启用（默认 polling）

**建议**：
- **如果目标是实时响应**：修改 `start_all_services.sh` 默认值为 `events`
- **如果目标是最终一致性**：保持 `polling`，但考虑缩短轮询间隔

### 2. `resident_id` 字段命名

**当前状态**：✅ 命名合理

**建议**：
- **保持 `resident_id`**，不需要改为 `bed_resident_id`
- 原因：已广泛使用，修改成本高，未来扩展性更好

### 3. Location 卡片的 `resident_id`

**当前状态**：`NULL`

**建议**：
- **暂不建议修改**
- 原因：Location 卡片不应该有"主住户"概念，当前权限过滤逻辑已能正确处理

### 4. JSONB 字段设计

**当前状态**：✅ 已经是 id:name 模式

**建议**：
- **保持当前设计**
- 原因：程序使用 ID，显示使用 Name，性能和数据稳定性都得到保障

---

## 6. 需要决策的问题

1. **事件驱动模式默认值**：是否改为 `events`？
2. **Location 卡片的 `resident_id`**：是否需要支持非 SharedUnit 的主住户？
3. **Name 更新策略**：当前通过重新生成卡片更新，是否可接受？

