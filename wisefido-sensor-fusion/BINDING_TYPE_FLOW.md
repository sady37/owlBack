# binding_type 数据流说明

## 📊 数据流概览

```
devices 表 (bound_bed_id)
    ↓
wisefido-card-aggregator (设置 binding_type)
    ↓
cards.devices JSONB 字段
    ↓
wisefido-sensor-fusion (读取 binding_type)
```

## 🔍 详细流程

### 1. 数据源：`devices` 表

`devices` 表中有 `bound_bed_id` 字段：
- 如果设备绑定到床：`bound_bed_id IS NOT NULL`
- 如果设备未绑床：`bound_bed_id IS NULL`

### 2. wisefido-card-aggregator 设置 binding_type

#### 2.1 GetDevicesByBed（获取床上的设备）

```go
// wisefido-card-aggregator/internal/repository/card.go
func (r *CardRepository) GetDevicesByBed(tenantID, bedID string) ([]DeviceInfo, error) {
    // 查询条件：bound_bed_id = bedID
    // ...
    
    if boundBedID.Valid {
        device.BoundBedID = &boundBedID.String
        device.BindingType = "direct"  // ← 绑定到床，设置为 "direct"
    } else {
        device.BindingType = "indirect" // ← 理论上不会到这里（因为查询条件已经过滤了）
    }
}
```

**说明**：
- 查询条件：`WHERE d.bound_bed_id = $2`
- 所以返回的设备都是 `bound_bed_id IS NOT NULL`
- 因此 `binding_type = "direct"`

#### 2.2 GetUnboundDevicesByUnit（获取未绑床的设备）

```go
// wisefido-card-aggregator/internal/repository/card.go
func (r *CardRepository) GetUnboundDevicesByUnit(tenantID, unitID string) ([]DeviceInfo, error) {
    // 查询条件：unit_id = unitID AND bound_bed_id IS NULL
    // ...
    
    device.BindingType = "indirect"  // ← 未绑床，设置为 "indirect"
}
```

**说明**：
- 查询条件：`WHERE d.unit_id = $2 AND d.bound_bed_id IS NULL`
- 所以返回的设备都是 `bound_bed_id IS NULL`
- 因此 `binding_type = "indirect"`

### 3. 转换为 JSON 并存储到 cards.devices

```go
// wisefido-card-aggregator/internal/repository/card.go
func ConvertDevicesToJSON(devices []DeviceInfo) ([]byte, error) {
    var deviceJSONs []DeviceJSON
    for _, device := range devices {
        deviceJSONs = append(deviceJSONs, DeviceJSON{
            DeviceID:    device.DeviceID,
            DeviceName:  device.DeviceName,
            DeviceType:  device.DeviceType,
            DeviceModel: device.DeviceModel,
            BindingType: device.BindingType,  // ← 从 DeviceInfo 复制到 DeviceJSON
        })
    }
    return json.Marshal(deviceJSONs)
}
```

**存储位置**：
- `cards.devices` JSONB 字段
- 格式：`[{"device_id": "...", "binding_type": "direct|indirect", ...}, ...]`

### 4. wisefido-sensor-fusion 读取 binding_type

```go
// wisefido-sensor-fusion/internal/repository/card.go
func (r *CardRepository) GetCardDevices(cardID string) ([]DeviceInfo, error) {
    // 从 cards.devices JSONB 字段读取
    var devicesJSON []byte
    err := r.db.QueryRow(query, cardID).Scan(&devicesJSON)
    
    // 解析 JSONB
    var devices []DeviceInfo
    err := json.Unmarshal(devicesJSON, &devices)
    // devices 中的每个设备都有 BindingType 字段
}
```

```go
// wisefido-sensor-fusion/internal/repository/card.go
type DeviceInfo struct {
    DeviceID    string `json:"device_id"`
    DeviceName  string `json:"device_name"`
    DeviceType  string `json:"device_type"`
    DeviceModel string `json:"device_model"`
    BindingType string `json:"binding_type"` // ← 从 JSONB 解析出来
}
```

### 5. wisefido-sensor-fusion 使用 binding_type

```go
// wisefido-sensor-fusion/internal/fusion/sensor_fusion.go
func (f *SensorFusion) FuseCardData(tenantID, cardID, cardType string) (*models.RealtimeData, error) {
    devices, err := f.cardRepo.GetCardDevices(cardID)
    
    for _, device := range devices {
        if deviceType == "Radar" || deviceType == "Sleepace" || deviceType == "SleepPad" {
            if cardType == "ActiveBed" {
                if device.BindingType == "direct" {  // ← 使用 binding_type 过滤
                    fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
                }
            }
        }
    }
}
```

## 📝 总结

### binding_type 的判断规则

| 设备绑定情况 | bound_bed_id | binding_type | 说明 |
|------------|-------------|--------------|------|
| 绑定到床 | `IS NOT NULL` | `"direct"` | 设备直接绑定到床 |
| 未绑床（绑定到 unit） | `IS NULL` | `"indirect"` | 设备绑定到 unit，但未绑床 |

### 数据流路径

1. **数据库**：`devices.bound_bed_id`（原始数据）
2. **wisefido-card-aggregator**：
   - `GetDevicesByBed` → `binding_type = "direct"`
   - `GetUnboundDevicesByUnit` → `binding_type = "indirect"`
   - `ConvertDevicesToJSON` → 转换为 JSON
   - `CreateCard` → 存储到 `cards.devices` JSONB
3. **wisefido-sensor-fusion**：
   - `GetCardDevices` → 从 `cards.devices` JSONB 读取并解析
   - `FuseCardData` → 使用 `binding_type` 过滤设备

### 关键点

- ✅ `binding_type` 不是数据库字段，而是**计算字段**
- ✅ 由 `wisefido-card-aggregator` 在创建卡片时计算并存储到 `cards.devices` JSONB
- ✅ `wisefido-sensor-fusion` 从 JSONB 中读取并使用
- ✅ 判断依据：`devices.bound_bed_id IS NOT NULL` → `"direct"`，否则 → `"indirect"`

