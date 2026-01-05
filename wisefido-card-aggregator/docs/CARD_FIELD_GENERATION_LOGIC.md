# 卡片字段生成逻辑说明

## 1. 事件订阅方处理状态

### ✅ 已实现的事件处理

所有新发布的事件类型都已在 `event_consumer.go` 中处理：

| 事件类型 | 处理位置 | 处理逻辑 | 状态 |
|---------|---------|---------|------|
| `device.bound` | `event_consumer.go:162` | 调用 `CreateCardsForUnit` 重新计算相关 unit 的卡片 | ✅ 已实现 |
| `device.unbound` | `event_consumer.go:162` | 调用 `CreateCardsForUnit` 重新计算相关 unit 的卡片 | ✅ 已实现 |
| `device.monitoring_changed` | `event_consumer.go:162` | 调用 `CreateCardsForUnit` 重新计算相关 unit 的卡片 | ✅ 已实现 |
| `resident.caregivers_changed` | `event_consumer.go:192` | 调用 `CreateCardsForUnit` 重新计算相关 unit 的卡片 | ✅ 已实现 |
| `resident.bound` | `event_consumer.go:177` | 调用 `CreateCardsForUnit` 重新计算相关 unit 的卡片 | ✅ 已实现 |
| `resident.unbound` | `event_consumer.go:177` | 调用 `CreateCardsForUnit` 重新计算相关 unit 的卡片 | ✅ 已实现 |
| `unit.info_changed` | `event_consumer.go:203` | 调用 `CreateCardsForUnit` 重新计算相关 unit 的卡片 | ✅ 已实现 |

### ⚠️ 注意事项

1. **事件驱动模式需要配置启用**：
   - 默认模式：`polling`（轮询模式，每 60 分钟全量更新）
   - 事件驱动模式：需要设置 `CARD_TRIGGER_MODE=events` 环境变量
   - 如果未启用事件驱动模式，事件不会被实时处理，只能等待轮询更新

2. **事件处理逻辑**：
   - 所有事件都会触发 `CreateCardsForUnit`，重新计算整个 unit 的卡片
   - 如果事件中包含 `unit_id`，直接使用
   - 如果只有 `bed_id`，会先查询 `unit_id`，然后重新计算

---

## 2. 卡片字段生成逻辑

### 2.1 `resident_id` 字段（Primary resident for ActiveBed）

**位置**：`card_creator.go:calculateExpectedActiveBedCard` (第 710-728 行)

**生成逻辑**：
```go
// 1. 获取 bed 上绑定的 resident
resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)

var residentID *string
if resident != nil {
    // 如果 bed 有绑定的 resident，使用该 resident_id
    residentID = &resident.ResidentID
} else {
    // 如果 bed 没有绑定 resident，residentID = nil
    residentID = nil
}
```

**规则**：
- **ActiveBed 卡片**：`resident_id` = bed 上绑定的 resident_id（如果有），否则为 `NULL`
- **Location 卡片**：`resident_id` = `NULL`（Location 卡片没有主住户）

**数据来源**：
- `residents` 表的 `bed_id` 字段

---

### 2.2 `devices` JSONB 字段（Precomputed associations）

**位置**：`card_creator.go:calculateExpectedActiveBedCard` (第 730-734 行)

**生成逻辑**：

#### ActiveBed 卡片：
```go
// 1. 获取 bed 上绑定的设备（monitoring_enabled = TRUE）
bedDevices, err := c.repo.GetDevicesByBed(tenantID, bed.BedID)

// 2. 如果是 Scenario A（只有 1 个 ActiveBed），还要获取 unit 下未绑定的设备
if includeUnboundDevices {
    unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitInfo.UnitID)
    allDevices = append(bedDevices, unboundDevices...)
} else {
    allDevices = bedDevices
}

// 3. 转换为 JSON
devicesJSON, err := repository.ConvertDevicesToJSON(allDevices)
```

#### Location 卡片：
```go
// 1. 获取 unit 下未绑定的设备（monitoring_enabled = TRUE）
unboundDevices, err := c.repo.GetUnboundDevicesByUnit(tenantID, unitInfo.UnitID)

// 2. 转换为 JSON
devicesJSON, err := repository.ConvertDevicesToJSON(unboundDevices)
```

**JSON 格式**（`card.go:DeviceJSON`）：
```json
[
  {
    "device_id": "uuid",
    "device_name": "Radar-XXXX",
    "device_type": "Radar",
    "device_model": "BM8701-2",
    "bed_id": "uuid",      // 如果绑定到 bed
    "bed_name": "Bed 1",   // 如果绑定到 bed
    "room_id": "uuid",     // 如果绑定到 room
    "room_name": "Room 1", // 如果绑定到 room
    "unit_id": "uuid"      // 设备所在的 unit_id
  }
]
```

**规则**：
- 只包含 `monitoring_enabled = TRUE` 的设备
- ActiveBed 卡片：包含 bed 上绑定的设备 + unit 下未绑定的设备（Scenario A）
- Location 卡片：只包含 unit 下未绑定的设备

**数据来源**：
- `devices` 表的 `bound_bed_id`、`bound_room_id`、`monitoring_enabled` 字段
- `device_store` 表的 `device_type`、`device_model` 字段（通过 JOIN）

---

### 2.3 `residents` JSONB 字段（Precomputed associations）

**位置**：`card_creator.go:calculateExpectedActiveBedCard` (第 710-728 行)

**生成逻辑**：

#### ActiveBed 卡片：
```go
// 1. 获取 bed 上绑定的 resident
resident, err := c.repo.GetResidentByBed(tenantID, bed.BedID)

var residents []repository.ResidentInfo
if resident != nil {
    // 如果 bed 有绑定的 resident，只包含该 resident
    residents = []repository.ResidentInfo{*resident}
} else {
    // 如果 bed 没有绑定 resident，获取 unit 下所有 residents
    unitResidents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
    residents = unitResidents
}
```

#### Location 卡片：
```go
// 获取 unit 下所有 residents
residents, err := c.repo.GetResidentsByUnit(tenantID, unitInfo.UnitID)
```

**JSON 格式**（`card.go:ResidentJSON`）：
```json
[
  {
    "resident_id": "uuid",
    "nickname": "John Doe"
  }
]
```

**规则**：
- ActiveBed 卡片：
  - 如果 bed 有绑定的 resident → 只包含该 resident
  - 如果 bed 没有绑定 resident → 包含 unit 下所有 residents
- Location 卡片：包含 unit 下所有 residents

**数据来源**：
- `residents` 表的 `bed_id`、`unit_id`、`nickname` 字段

---

### 2.4 `unhandled_alarm_0` ~ `unhandled_alarm_4` 字段（Unhandled alarm counters）

**位置**：`card.go:CreateCard` (第 660-661 行)

**生成逻辑**：
```sql
-- 在 CreateCard 的 INSERT 语句中，这些字段不插入，使用数据库默认值
INSERT INTO cards (
    tenant_id,
    card_type,
    bed_id,
    unit_id,
    card_name,
    card_address,
    resident_id,
    devices,
    residents
    -- 注意：unhandled_alarm_0 ~ unhandled_alarm_4 不插入，使用默认值 0
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
```

**规则**：
- **创建时**：所有字段都使用数据库默认值 `0`
- **更新时**：这些字段不会被更新（`UpdateCard` 方法不更新这些字段）
- **实际更新**：这些字段由其他服务（如告警处理服务）更新，不是由卡片聚合服务维护

**数据来源**：
- 数据库默认值：`DEFAULT 0`
- 后续由告警处理服务更新（不在卡片生成逻辑中）

---

## 3. 字段更新时机

| 字段 | 创建时 | 更新时 | 更新来源 |
|------|--------|--------|---------|
| `resident_id` | ✅ 生成 | ✅ 更新 | 卡片聚合服务（当 resident 绑定/解绑时） |
| `devices` JSONB | ✅ 生成 | ✅ 更新 | 卡片聚合服务（当设备绑定/解绑/监控状态变化时） |
| `residents` JSONB | ✅ 生成 | ✅ 更新 | 卡片聚合服务（当 resident 绑定/解绑时） |
| `unhandled_alarm_0~4` | ✅ 默认值 0 | ❌ 不更新 | 告警处理服务（不在卡片聚合服务中） |

---

## 4. 总结

### 事件订阅方
- ✅ 所有新事件类型都已处理
- ⚠️ 需要启用事件驱动模式（`CARD_TRIGGER_MODE=events`）才能实时响应

### 卡片字段生成
- ✅ `resident_id`：从 bed 或 unit 获取
- ✅ `devices` JSONB：从 bed 或 unit 获取设备列表（只包含 `monitoring_enabled = TRUE`）
- ✅ `residents` JSONB：从 bed 或 unit 获取住户列表
- ✅ `unhandled_alarm_0~4`：使用数据库默认值 0，由其他服务更新

