# Bed Repository 分析

## 🤔 问题

为什么创建了 `card.go`、`device.go`、`room.go`，但没有 `bed.go`？

## 📊 分析

### 1. wisefido-sensor-fusion 的工作方式

**wisefido-sensor-fusion** 通过卡片获取设备：
- ✅ 使用 `GetCardDevices(cardID)` 从 `cards.devices` JSONB 字段读取设备列表
- ✅ 对于 **ActiveBed 卡片**，`cards.devices` 已经包含了该床上的所有设备（由 wisefido-card-aggregator 预计算）
- ✅ 所以 wisefido-sensor-fusion **不需要直接查询床上的设备**

**代码示例**：
```go
// wisefido-sensor-fusion/internal/fusion/sensor_fusion.go
func (f *SensorFusion) FuseCardData(tenantID, cardID, cardType string) (*models.RealtimeData, error) {
    // 1. 获取卡片关联的所有设备（从 cards.devices JSONB）
    devices, err := f.cardRepo.GetCardDevices(cardID)
    // ...
}
```

### 2. wisefido-alarm 的工作方式

**wisefido-alarm** 的数据流：
- ✅ 读取融合后的实时数据：`vital-focus:card:{card_id}:realtime`（已经是卡片级别的数据）
- ✅ 如果需要查询设备信息，可以通过：
  - `card.go` 的 `GetCardDevices(cardID)` - 从卡片获取设备列表（推荐）
  - `device.go` 的 `GetDevicesByBed(tenantID, bedID)` - 直接查询床上的设备（备用）

### 3. alarm_rule.md 的需求分析

#### 事件1：防止雷达漏报 - 床上跌落检测
- **触发**：sleepad检测到离床事件
- **需要**：查询床上的 sleepad 和 radar 设备
- **方案**：
  - ✅ 通过 `card.go` 的 `GetCardDevices(cardID)` 获取（ActiveBed 卡片已包含床上的所有设备）
  - ✅ 或通过 `device.go` 的 `GetDevicesByBed(tenantID, bedID)` 直接查询

#### 事件2：Sleepad可靠性判断
- **分支A**：床上绑radar
  - **条件**：绑到床上的雷达，未检测到区域ID-床有人存在
  - **需要**：查询床上的 radar 设备
  - **方案**：
    - ✅ 通过 `card.go` 的 `GetCardDevices(cardID)` 获取
    - ✅ 或通过 `device.go` 的 `GetDevicesByBed(tenantID, bedID)` 直接查询

### 4. 当前实现情况

**已实现的功能**：
- ✅ `card.go` - `GetCardDevices(cardID)` - 从卡片获取设备列表
- ✅ `device.go` - `GetDevicesByBed(tenantID, bedID)` - 直接查询床上的设备

**是否需要 bed.go**：
- ❌ **不需要单独的 bed.go**
- ✅ 原因：
  1. **查询床上的设备**：可以通过 `card.go` 的 `GetCardDevices`（推荐，因为卡片已预计算）
  2. **直接查询床上的设备**：可以通过 `device.go` 的 `GetDevicesByBed`（备用方案）
  3. **查询床信息**（bed_id, bed_name 等）：可以通过 `card.go` 的 `GetCardByID` 获取 `BedID`，然后查询 `beds` 表（如果需要）

### 5. 是否需要 bed.go 的场景

**可能需要 bed.go 的场景**：
- ❌ 查询床信息（bed_id, bed_name, room_id 等）- 可以通过 `card.go` 获取 `BedID`，然后直接查询 `beds` 表
- ❌ 查询床上的设备 - 已有 `card.go` 和 `device.go` 提供
- ❌ 查询床的住户 - 可以通过 `card.go` 获取 `BedID`，然后查询 `residents` 表

**结论**：
- ✅ **不需要 bed.go**
- ✅ 现有的 `card.go` 和 `device.go` 已经足够

## 📝 建议

### 方案1：优先使用卡片（推荐）✅

```go
// 1. 通过卡片获取设备列表（推荐）
card, err := cardRepo.GetCardByID(tenantID, cardID)
if err != nil {
    return err
}

devices, err := cardRepo.GetCardDevices(cardID)
if err != nil {
    return err
}

// 对于 ActiveBed 卡片，devices 已经包含床上的所有设备
```

**优点**：
- ✅ 使用预计算的数据（cards.devices），性能好
- ✅ 与 wisefido-sensor-fusion 保持一致
- ✅ 数据已经由 wisefido-card-aggregator 维护

### 方案2：直接查询床上的设备（备用）

```go
// 2. 直接查询床上的设备（备用方案）
if card.BedID != nil {
    devices, err := deviceRepo.GetDevicesByBed(tenantID, *card.BedID)
    if err != nil {
        return err
    }
}
```

**优点**：
- ✅ 不依赖卡片数据
- ✅ 可以获取最新的设备绑定关系

**缺点**：
- ⚠️ 需要额外的数据库查询
- ⚠️ 与 wisefido-sensor-fusion 的工作方式不一致

## ⚠️ 重要发现

**用户指出的问题**：即使是对于 ActiveBed 卡片，也并不是所有 device 都属于该 bed！

**卡片创建规则分析**：

### 场景 A：门牌下只有 1 个 ActiveBed
- **ActiveBed 卡片绑定的设备**：
  1. ✅ 该 bed 绑定的设备：`devices.bound_bed_id = bed_id` 且 `monitoring_enabled = TRUE`
  2. ⚠️ **该 unit 下未绑床的设备**：`devices.unit_id = unit_id` 且 `devices.bound_bed_id IS NULL` 且 `monitoring_enabled = TRUE`

**结论**：场景 A 中，ActiveBed 卡片的 `devices` 字段包含了**未绑床的设备**，这些设备不属于该 bed！

### 场景 B：门牌下有多个 ActiveBed（≥2）
- **ActiveBed 卡片绑定的设备**：
  - ✅ 只包含该 bed 绑定的设备：`devices.bound_bed_id = bed_id` 且 `monitoring_enabled = TRUE`

**结论**：场景 B 中，ActiveBed 卡片的 `devices` 字段只包含该 bed 绑定的设备。

## ✅ 修正后的结论

**不需要 bed.go**，但需要**区分使用场景**：

### 1. 查询"卡片上的所有设备"（用于融合）
- ✅ 使用 `card.go` 的 `GetCardDevices(cardID)`
- ✅ 适用于 wisefido-sensor-fusion（需要融合卡片上的所有设备）

### 2. 查询"床上的设备"（用于报警评估）
- ⚠️ **不能直接使用** `card.go` 的 `GetCardDevices`（场景 A 会包含未绑床的设备）
- ✅ **必须使用** `device.go` 的 `GetDevicesByBed(tenantID, bedID)`
- ✅ 适用于 wisefido-alarm 的事件评估（如事件2的分支A：床上绑radar）

**推荐做法**：
- **wisefido-sensor-fusion**：使用 `card.go` 的 `GetCardDevices`（需要融合卡片上的所有设备）
- **wisefido-alarm**：如果需要查询"床上的设备"，使用 `device.go` 的 `GetDevicesByBed`（确保只获取该 bed 绑定的设备）
