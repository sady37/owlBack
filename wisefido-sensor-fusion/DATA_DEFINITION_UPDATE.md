# 数据定义更新说明

## ✅ 修改完成

根据用户要求，已更新数据定义和相关代码：

### 1. 数据库定义更新 (`cards.sql`)

**修改前**：
```sql
-- 格式：[{"device_id": "...", "device_name": "...", "device_type": "...", "device_model": "...", "binding_type": "direct|indirect"}, ...]
```

**修改后**：
```sql
-- 格式：[{"device_id": "...", "device_name": "...", "device_type": "...", "device_model": "...", "bed_id": "...", "bed_name": "...", "room_id": "...", "room_name": "...", "unit_id": "..."}, ...]
-- 注意：
--   - 如果设备绑定到床：bed_id 和 bed_name 不为空，room_id 和 room_name 为空
--   - 如果设备绑定到房间：room_id 和 room_name 不为空，bed_id 和 bed_name 为空
--   - unit_id 始终存在（设备必须绑定到某个单元）
```

### 2. wisefido-card-aggregator 更新

#### 2.1 `DeviceInfo` 结构体
- ❌ 移除 `BindingType` 字段
- ✅ 添加 `BedName` 字段（床名称）
- ✅ 添加 `RoomName` 字段（房间名称）

#### 2.2 `DeviceJSON` 结构体
- ❌ 移除 `BindingType` 字段
- ✅ 添加 `BedName` 字段
- ✅ 添加 `RoomName` 字段

#### 2.3 `GetDevicesByBed` 和 `GetUnboundDevicesByUnit`
- ✅ 查询时 JOIN `beds` 表获取 `bed_name`
- ✅ 查询时 JOIN `rooms` 表获取 `room_name`
- ✅ 扫描时填充 `BedName` 和 `RoomName` 字段

#### 2.4 `ConvertDevicesToJSON`
- ✅ 将 `BedName` 和 `RoomName` 包含到 JSON 中

### 3. wisefido-sensor-fusion 更新

#### 3.1 `DeviceInfo` 结构体
- ❌ 移除 `BindingType` 字段
- ✅ 添加 `BedName` 字段
- ✅ 添加 `RoomName` 字段

#### 3.2 融合逻辑
- ❌ 不再使用 `binding_type` 来判断
- ✅ 使用 `bed_id` 来判断：
  - 对于 ActiveBed 卡片：只融合 `bed_id` 有效且相同的设备
  - 对于 Location 卡片：融合所有设备（`bed_id` 为 NULL）

### 4. 测试文件更新

- ✅ 更新 `card_creator_test.go`，移除所有 `BindingType` 字段
- ✅ 添加 `BedName` 和 `RoomName` 字段到测试数据

## 📝 数据格式

### 设备绑定到床
```json
{
  "device_id": "device-123",
  "device_name": "Radar01",
  "device_type": "Radar",
  "device_model": "Model-A",
  "bed_id": "bed-456",
  "bed_name": "BedA",
  "room_id": null,
  "room_name": null,
  "unit_id": "unit-789"
}
```

### 设备绑定到房间
```json
{
  "device_id": "device-123",
  "device_name": "Radar01",
  "device_type": "Radar",
  "device_model": "Model-A",
  "bed_id": null,
  "bed_name": null,
  "room_id": "room-456",
  "room_name": "Room1",
  "unit_id": "unit-789"
}
```

## ✅ 验证

- ✅ `wisefido-card-aggregator` 编译通过
- ✅ `wisefido-sensor-fusion` 编译通过
- ✅ 测试通过（`card_creator_test.go`）

## 🎯 关键改进

1. **更直接的数据结构**：直接存储 `bed_id`/`room_id` 及其名称，不需要通过 `binding_type` 来判断
2. **更清晰的融合逻辑**：使用 `bed_id` 来判断是否应该融合，逻辑更清晰
3. **完整的信息**：包含 `bed_name` 和 `room_name`，便于前端显示

