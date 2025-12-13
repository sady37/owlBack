# wisefido-sensor-fusion 融合逻辑问题分析

## 🚨 问题描述

根据卡片创建规则，**场景 A**（门牌下只有 1 个 ActiveBed）中：
- ActiveBed 卡片的 `devices` 字段包含了：
  1. **该 bed 绑定的设备**（`binding_type = "direct"`）
  2. **该 unit 下未绑床的设备**（`binding_type = "indirect"`）

**融合逻辑要求**：
- 应该只融合**同一床上的 Radar 和 Sleepace 设备**
- 不应该把床上的 Sleepace 和未绑床的 Radar 进行融合

## ⚠️ 当前实现的问题

### 当前代码（`sensor_fusion.go`）

```go
// 1. 获取卡片关联的所有设备
devices, err := f.cardRepo.GetCardDevices(cardID)

// 2. 过滤设备类型：只查询 Radar 和 Sleepace 设备
var fusionDeviceIDs []string
for _, device := range devices {
    deviceType := device.DeviceType
    if deviceType == "Radar" || deviceType == "Sleepace" || deviceType == "SleepPad" {
        fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
    }
}

// 3. 判断是否需要融合
needFusion := len(sleepaceData) > 0 && len(radarData) > 0
```

**问题**：
- ❌ 没有检查 `binding_type`
- ❌ 对于 ActiveBed 卡片，可能会把床上的 Sleepace 和未绑床的 Radar 进行融合
- ❌ 这会导致错误的融合结果

### 示例场景

```
门牌号：201（unit_id: unit-201）
ActiveBed：BedA（bed_id: bed-a）
设备：
  - SleepPad01（绑 BedA，binding_type="direct"）
  - Radar01（绑 BedA，binding_type="direct"）
  - Radar02（绑门牌号，未绑床，binding_type="indirect"）

ActiveBed 卡片的 devices：
  - SleepPad01（direct）
  - Radar01（direct）
  - Radar02（indirect）← 不应该参与融合！
```

**当前行为**：
- ❌ 会融合 SleepPad01 + Radar01 + Radar02（错误！）
- ❌ Radar02 不应该参与融合，因为它不是床上的设备

**期望行为**：
- ✅ 只融合 SleepPad01 + Radar01（床上的设备）
- ✅ Radar02 不参与融合

## ✅ 解决方案

### 方案 1：根据 `binding_type` 过滤（推荐）

对于 **ActiveBed 卡片**，只融合 `binding_type = "direct"` 的设备（即绑定到床上的设备）。

对于 **Location 卡片**，可以融合所有设备（因为它们都是未绑床的设备）。

```go
// 修改 FuseCardData 函数
func (f *SensorFusion) FuseCardData(tenantID, cardID, cardType string) (*models.RealtimeData, error) {
    // 1. 获取卡片关联的所有设备
    devices, err := f.cardRepo.GetCardDevices(cardID)
    if err != nil {
        return nil, fmt.Errorf("failed to get card devices: %w", err)
    }
    
    // 2. 过滤设备类型和绑定类型
    var fusionDeviceIDs []string
    for _, device := range devices {
        deviceType := device.DeviceType
        if deviceType == "Radar" || deviceType == "Sleepace" || deviceType == "SleepPad" {
            // 对于 ActiveBed 卡片，只融合绑定到床上的设备（binding_type = "direct"）
            if cardType == "ActiveBed" {
                if device.BindingType == "direct" {
                    fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
                }
            } else {
                // Location 卡片：融合所有设备（因为它们都是未绑床的设备）
                fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
            }
        }
    }
    
    // ... 后续逻辑不变
}
```

### 方案 2：根据 `bed_id` 查询床上的设备（备用）

如果 `DeviceInfo` 结构体不包含 `binding_type`，可以通过查询 `bed_id` 来过滤设备。

但这种方式需要额外的数据库查询，不如方案 1 高效。

## 📝 需要修改的文件

1. **`wisefido-sensor-fusion/internal/fusion/sensor_fusion.go`**
   - 修改 `FuseCardData` 函数，添加 `binding_type` 过滤逻辑

2. **`wisefido-sensor-fusion/internal/repository/card.go`**
   - 确认 `DeviceInfo` 结构体包含 `BindingType` 字段（已包含 ✅）

## ✅ 验证

修改后，需要验证：

1. **场景 A**（门牌下只有 1 个 ActiveBed）：
   - ✅ 只融合床上的 Radar 和 Sleepace 设备
   - ✅ 未绑床的设备不参与融合

2. **场景 B**（门牌下有多个 ActiveBed）：
   - ✅ 每个 ActiveBed 卡片只融合该床上的设备（场景 B 中，ActiveBed 卡片只包含床上的设备）

3. **Location 卡片**：
   - ✅ 融合所有设备（因为它们都是未绑床的设备）

