# 融合逻辑修正说明

## 🚨 问题

之前的融合逻辑只使用了 `binding_type` 来判断是否应该融合，但这是不完整的。

**正确的判断方式**：
- 应该使用设备绑定的 `bed_id`, `room_id`, `unit_id` 来判断
- 如果 `bed_id` 有效，则所有 `bed_id` 相同的设备都是绑在同一床上的，应该融合
- 如果 `bed_id` 为 NULL，则不参与融合（未绑床的设备）

## ✅ 修改内容

### 1. wisefido-card-aggregator 修改

#### 1.1 更新 `DeviceInfo` 结构体

添加 `BoundRoomID` 字段：
```go
type DeviceInfo struct {
    // ... 其他字段
    BoundBedID        *string
    BoundRoomID       *string // 新增：设备绑定的房间ID
    UnitID            string
    // ...
}
```

#### 1.2 更新 `DeviceJSON` 结构体

添加 `bed_id`, `room_id`, `unit_id` 字段：
```go
type DeviceJSON struct {
    DeviceID    string  `json:"device_id"`
    DeviceName  string  `json:"device_name"`
    DeviceType  string  `json:"device_type"`
    DeviceModel string  `json:"device_model"`
    BindingType string  `json:"binding_type"`
    BedID       *string `json:"bed_id,omitempty"`       // 新增
    RoomID      *string `json:"room_id,omitempty"`      // 新增
    UnitID      string  `json:"unit_id"`                // 新增
}
```

#### 1.3 更新 `GetDevicesByBed` 和 `GetUnboundDevicesByUnit`

- 查询时包含 `bound_room_id` 字段
- 扫描时填充 `BoundRoomID` 字段

#### 1.4 更新 `ConvertDevicesToJSON`

将 `bed_id`, `room_id`, `unit_id` 字段包含到 JSON 中：
```go
deviceJSONs = append(deviceJSONs, DeviceJSON{
    // ... 其他字段
    BedID:  device.BoundBedID,
    RoomID: device.BoundRoomID,
    UnitID: device.UnitID,
})
```

### 2. wisefido-sensor-fusion 修改

#### 2.1 更新 `DeviceInfo` 结构体

添加 `bed_id`, `room_id`, `unit_id` 字段：
```go
type DeviceInfo struct {
    DeviceID    string  `json:"device_id"`
    DeviceName  string  `json:"device_name"`
    DeviceType  string  `json:"device_type"`
    DeviceModel string  `json:"device_model"`
    BindingType string  `json:"binding_type"`
    BedID       *string `json:"bed_id,omitempty"`       // 新增
    RoomID      *string `json:"room_id,omitempty"`      // 新增
    UnitID      string  `json:"unit_id"`                // 新增
}
```

#### 2.2 更新融合逻辑

**修改前**：
- 使用 `binding_type = "direct"` 来判断是否应该融合

**修改后**：
- 使用 `bed_id` 来判断是否应该融合
- 对于 ActiveBed 卡片：
  - 如果 `bed_id` 有效，则所有 `bed_id` 相同的设备都是绑在同一床上的，应该融合
  - 如果 `bed_id` 为 NULL，则不参与融合（未绑床的设备）
- 对于 Location 卡片：
  - 融合所有设备（因为它们都是未绑床的设备，`bed_id` 为 NULL）

**代码逻辑**：
```go
var bedIDForFusion *string // 用于 ActiveBed 卡片，记录第一个有效 bed_id

for _, device := range devices {
    if deviceType == "Radar" || deviceType == "Sleepace" || deviceType == "SleepPad" {
        if cardType == "ActiveBed" {
            // ActiveBed 卡片：只融合绑定到同一床上的设备
            if device.BedID != nil && *device.BedID != "" {
                // 如果这是第一个有效 bed_id，记录它
                if bedIDForFusion == nil {
                    bedIDForFusion = device.BedID
                }
                // 只融合 bed_id 相同的设备（绑定到同一床上的设备）
                if bedIDForFusion != nil && *device.BedID == *bedIDForFusion {
                    fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
                }
            }
            // bed_id 为 NULL 的设备不参与融合（未绑床的设备）
        } else {
            // Location 卡片：融合所有设备
            fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
        }
    }
}
```

## 📝 数据流

1. **wisefido-card-aggregator**：
   - 从 `devices` 表查询设备信息（包括 `bound_bed_id`, `bound_room_id`, `unit_id`）
   - 转换为 `DeviceJSON` 格式（包含 `bed_id`, `room_id`, `unit_id`）
   - 存储到 `cards.devices` JSONB 字段

2. **wisefido-sensor-fusion**：
   - 从 `cards.devices` JSONB 字段读取设备信息（包含 `bed_id`, `room_id`, `unit_id`）
   - 根据 `bed_id` 判断是否应该融合
   - 只融合 `bed_id` 相同且有效的设备

## ✅ 验证

- ✅ `wisefido-card-aggregator` 编译通过
- ✅ `wisefido-sensor-fusion` 编译通过
- ✅ 融合逻辑现在使用 `bed_id` 来判断，而不是 `binding_type`

## 🎯 关键改进

1. **更准确的判断**：使用 `bed_id` 来判断设备是否绑定到同一床，而不是依赖 `binding_type`
2. **完整的字段**：`cards.devices` JSONB 现在包含 `bed_id`, `room_id`, `unit_id` 等完整信息
3. **场景 A 正确处理**：场景 A 中，ActiveBed 卡片包含床上的设备（`bed_id` 有效）和未绑床的设备（`bed_id` 为 NULL），现在只融合床上的设备

