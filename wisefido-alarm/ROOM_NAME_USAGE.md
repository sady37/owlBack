# room_name 使用说明

## 📊 用途

`room_name` 主要是在 `wisefido-alarm` 中，用于判断 room 是不是 bathroom（卫生间）。

## 🔍 使用场景

### 事件3：Bathroom可疑跌倒检测

根据 `alarm_rule.md`：
- **触发条件**：在bathroom（卫生间）房间内
- **房间识别**：通过 `room_name` 或 `unit_name` 中是否包含以下词（不区分大小写）：
  - bathroom
  - restroom
  - toilet

## 📝 数据来源

### 方案：从 `cards.devices` JSONB 读取（当前实现）✅

**优点**：
- ✅ 预计算数据，性能好
- ✅ 不需要额外查询 `rooms` 表
- ✅ 数据已经在 `cards.devices` JSONB 中

**实现**：
```go
// wisefido-alarm 从 cards.devices JSONB 读取设备信息
devices, err := cardRepo.GetCardDevices(cardID)

// 检查设备是否在 bathroom 中
for _, device := range devices {
    if device.RoomName != nil {
        // 判断是否是 bathroom
        roomNameLower := strings.ToLower(*device.RoomName)
        isBathroom := strings.Contains(roomNameLower, "bathroom") ||
                     strings.Contains(roomNameLower, "restroom") ||
                     strings.Contains(roomNameLower, "toilet")
        
        if isBathroom {
            // 执行事件3：Bathroom可疑跌倒检测
        }
    }
}
```

## ✅ 当前实现

### 1. `cards.devices` JSONB 包含 `room_name`

```json
{
  "device_id": "device-123",
  "device_name": "Radar01",
  "device_type": "Radar",
  "device_model": "Model-A",
  "bed_id": null,
  "bed_name": null,
  "room_id": "room-456",
  "room_name": "Bathroom",  // ← 用于 alarm 判断是否是 bathroom
  "unit_id": "unit-789"
}
```

### 2. `wisefido-alarm` 的 `DeviceInfo` 包含 `RoomName`

```go
type DeviceInfo struct {
    DeviceID    string
    DeviceName  string
    DeviceType  string
    DeviceModel string
    BedID       *string
    BedName     *string
    RoomID      *string
    RoomName    *string  // ← 用于判断是否是 bathroom
    UnitID      string
}
```

### 3. 使用方式（在 Evaluator 层实现）

在 `wisefido-alarm` 的 Evaluator 层（事件3）中：
```go
// 从 cards.devices JSONB 读取设备信息
devices, err := cardRepo.GetCardDevices(cardID)

// 检查设备是否在 bathroom 中
for _, device := range devices {
    if device.RoomName != nil {
        roomNameLower := strings.ToLower(*device.RoomName)
        isBathroom := strings.Contains(roomNameLower, "bathroom") ||
                     strings.Contains(roomNameLower, "restroom") ||
                     strings.Contains(roomNameLower, "toilet")
        
        if isBathroom {
            // 执行事件3：Bathroom可疑跌倒检测
            // 条件检查：
            // 1. 在bathroom房间内 ✅
            // 2. 雷达检测范围内仅1人
            // 3. 1个人处于站立状态（不是坐着）
            // 4. 位置未有变化（位置变化小于10cm，超过10分钟）
            // 5. 房间内仅有1个track_id
        }
    }
}
```

## 🎯 结论

✅ **当前实现是正确的**：
- `cards.devices` JSONB 中包含 `room_name`，方便 `wisefido-alarm` 直接使用
- 避免 alarm 需要额外查询 `rooms` 表
- 提供预计算数据，提高性能

✅ **`room_name` 的主要用途**：
- 在 `wisefido-alarm` 中判断房间是否是 bathroom（事件3）
- 也可以用于前端显示（如果需要）

## 📝 注意事项

- `room_name` 主要用于 alarm 判断是否是 bathroom
- 如果设备绑定到床，`room_name` 可能为 NULL（需要通过 `bed_id` 查询 `rooms` 表获取）
- 如果设备绑定到房间，`room_name` 不为 NULL，可以直接使用
