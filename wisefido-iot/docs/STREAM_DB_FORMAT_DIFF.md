# Redis Stream 格式与 iot_timeseries 数据库表结构差异分析

## 一、Redis Stream 格式（wisefido-radar 发布）

### 1.1 标准格式结构

wisefido-radar 发布到 Redis Stream 的数据格式（参考 `RADAR_REDIS_STREAM_FORMAT_STANDARD.md`）：

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "uuid-yyy",
  "timestamp": 1234567890,
  "topic_type": "monitor",
  "data_value": {
    "category": "track",
    "target_id": 1,
    "position_x": 150,
    "position_y": 200,
    "position_z": 50,
    "pose": "Walking",
    "event": 0,
    "area_id": 1,
    "raw_original": "base64_string"
  },
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- **顶层元数据字段**（按标准顺序）：
  - `device_id`: 设备 UUID
  - `device_type`: 设备类型（"Radar"）
  - `tenant_id`: 租户 ID
  - `timestamp`: 时间戳（Unix 秒）
  - `topic_type`: 主题类型（"monitor", "statistics", "event", "alarm"）
  - `data_value`: JSON 对象或数组，包含按 category 分组的数据（`category` 字段在 `data_value` 内部）
  - `branch_id`, `building_id`, `unit_id`, `room_id`, `bed_id`: 可选的位置信息

**字段顺序说明**：
字段按照以下顺序排列：`device_id` → `device_type` → `tenant_id` → `timestamp` → `topic_type` → `data_value` → 位置信息。此顺序符合访问频率和逻辑分组，便于查询和维护。

**关于 `category` 字段的说明**：
- **`category` 字段位置**：`category` 字段保留在 `data_value` 内部，不提取到顶层，避免数据冗余
- **查询方式**：查询 `category` 时需要通过 `data_values->'data_value'->>'category'` 访问（如果 `data_value` 是对象）或 `data_values->'data_value'->0->>'category'`（如果 `data_value` 是数组）
- **设计原则**：遵循数据不冗余原则，`category` 作为数据内容的一部分，保留在 `data_value` 内部，保持数据结构的完整性和一致性

### 1.2 实际发布代码（wisefido-radar/internal/consumer/mqtt_consumer.go）

```go
// 构建完整的输出对象（字段顺序：device_id → device_type → tenant_id → timestamp → topic_type → data_value → 位置信息）
// 注意：category 字段保留在 data_value 内部，不提取到顶层，避免冗余
encodedData := map[string]interface{}{
    "device_id":   getStringOrNull(device.DeviceID),
    "device_type": "Radar",
    "tenant_id":   getStringOrNull(device.TenantID),
    "timestamp":   time.Now().Unix(),
    "topic_type":  finalTopicType,
    "data_value":  dataValue,   // RadarDecoder 返回的数据值（包含 category 字段）
    "branch_id":   nil,
    "building_id": nil,
    "unit_id":     nil,
    "room_id":     getStringOrNullPtr(device.BoundRoomID),
    "bed_id":      getStringOrNullPtr(device.BoundBedID),
}
```

## 二、iot_timeseries 数据库表结构

### 2.1 表结构定义（18_iot_timeseries.sql）

```sql
CREATE TABLE iot_timeseries (
    id          BIGSERIAL PRIMARY KEY,
    uid         VARCHAR(50) NOT NULL,                  -- 设备唯一标识（从 device_store 获取）
    timestamp   TIMESTAMPTZ NOT NULL,                  -- 数据时间戳
    data_type   VARCHAR(20) NOT NULL DEFAULT 'monitor' CHECK (data_type IN ('monitor', 'statistics', 'event', 'alarm')),
    data_values JSONB NOT NULL                         -- 所有数据值存储在 JSONB 中
);
```

**字段说明**：
- `id`: 主键（自增）
- `uid`: 设备硬件 UID（从 `device_store` 表获取，通过 `device_id` 关联）
- `timestamp`: 数据时间戳（TIMESTAMPTZ）
- `data_type`: 数据类型（'monitor', 'statistics', 'event', 'alarm'）
- `data_values`: JSONB 字段，存储所有 encode 后的标准数据

### 2.2 data_values JSONB 字段期望格式（根据 SQL 注释）

```json
{
  "topic_type": "monitor",
  "tracking_id": 1,
  "position_x": 100,
  "position_y": 200,
  "position_z": 0,
  "pose": 1,
  "pose_snomed_code": "129006008",
  "pose_snomed_display": "Walking",
  "pose_category": "activity",
  "pose_display_en": "Walking",
  "event": 0,
  "event_snomed_code": null,
  "heart_rate": 75,
  "breath_rate": 16,
  "sleep_status": 1,
  "sleep_status_snomed_code": "248220002",
  "sleep_status_snomed_display": "Awake",
  "sleep_status_category": "activity",
  "sleep_status_display_en": "Awake",
  "stability": 0,
  "stability_snomed_code": null,
  "stability_category": "vital-signs",
  ...
}
```

**注意**：SQL 注释中期望的是**展开后的标准字段**，而不是包含 `data_value` 的嵌套结构。

## 三、wisefido-iot-timeseries 实际处理逻辑

### 3.1 数据插入代码（internal/repository/iot_timeseries_repo.go）

```go
func (r *IoTTimeSeriesRepository) Insert(data map[string]interface{}) (int64, error) {
    // 1. 提取 device_id
    deviceID, _ := data["device_id"].(string)
    
    // 2. 获取设备 UID（从 device_store 获取）
    _, _, _, uid, err := r.GetDeviceHardwareInfo(deviceID)
    
    // 3. 提取 timestamp
    var timestamp time.Time
    // ... 转换逻辑
    
    // 4. 确定 data_type（从 topic_type 映射）
    dataType := "monitor"
    if topicType, ok := data["topic_type"].(string); ok {
        switch topicType {
        case "monitor":
            dataType = "monitor"
        case "stat":
            dataType = "statistics"
        case "event":
            dataType = "event"
        case "alarm":
            dataType = "alarm"
        }
    }
    
    // 5. 将所有 encode 后的数据存储在 data_values JSONB 中
    // 直接使用传入的 data map，它已经包含了所有 encode 后的标准字段
    dataValuesJSON, err := json.Marshal(data)
    
    // 6. 插入数据库
    INSERT INTO iot_timeseries (uid, timestamp, data_type, data_values)
    VALUES ($1, $2, $3, $4)
}
```

### 3.2 实际存储的 data_values 格式

由于代码直接将整个 `data` map 存储到 `data_values`，实际存储的格式是（字段顺序：device_id → device_type → tenant_id → timestamp → topic_type → data_value → 位置信息）：

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "uuid-yyy",
  "timestamp": 1234567890,
  "topic_type": "monitor",
  "data_value": {
    "category": "track",
    "target_id": 1,
    "position_x": 150,
    "position_y": 200,
    "position_z": 50,
    "pose": "Walking",
    "event": 0,
    "area_id": 1
  },
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

## 四、差异分析

### 4.1 主要差异

| 项目 | Redis Stream 格式 | 数据库期望格式 | 实际存储格式 |
|------|-----------------|--------------|------------|
| **结构** | 顶层元数据 + `data_value` 嵌套对象 | 展开后的标准字段（扁平化） | 顶层元数据 + `data_value` 嵌套对象（与 Stream 一致） |
| **device_id** | 顶层字段 | 不在 data_values 中（已提取为 `uid`） | 仍在 data_values 中 |
| **tenant_id** | 顶层字段 | 不在 data_values 中 | 仍在 data_values 中 |
| **topic_type** | 顶层字段 | 在 data_values 中 | 在 data_values 中 |
| **timestamp** | 顶层字段 | 不在 data_values 中（已提取为 `timestamp` 列） | 仍在 data_values 中 |
| **data_value** | 嵌套对象/数组 | 不存在（应展开） | 存在（嵌套结构） |
| **位置信息** | 顶层字段（branch_id, unit_id 等） | 不在 data_values 中 | 仍在 data_values 中 |

### 4.2 问题点

1. **冗余字段**：
   - `device_id`, `tenant_id`, `timestamp` 在 `data_values` 中冗余存储（这些信息已通过 `uid` 和 `timestamp` 列存储）
   - 位置信息（`branch_id`, `building_id`, `unit_id`, `room_id`, `bed_id`）在 `data_values` 中存储，但这些信息可以通过 `devices` 表关联获取

2. **嵌套结构**：
   - Stream 格式使用 `data_value` 嵌套对象/数组来组织 category 数据
   - 数据库期望的是展开后的扁平化字段（如 `pose_snomed_code`, `heart_rate` 等）
   - 实际存储保留了嵌套结构，导致查询时需要访问 `data_values->'data_value'->>'position_x'` 而不是 `data_values->>'position_x'`

3. **字段命名不一致**：
   - Stream 格式：`data_value` 中包含 `category: "track"` 的对象
   - 数据库期望：直接包含 `position_x`, `pose_snomed_code` 等字段
   - 实际存储：保留了 `data_value` 嵌套结构

### 4.3 影响

1. **查询复杂度**：
   - 需要多一层 JSONB 路径访问：`data_values->'data_value'->>'position_x'`
   - 而不是：`data_values->>'position_x'`

2. **数据冗余**：
   - `device_id`, `tenant_id`, `timestamp` 等信息在 JSONB 中重复存储
   - 位置信息可以通过关联表获取，不需要存储在 JSONB 中

3. **兼容性**：
   - 如果后续需要按照 SQL 注释中的期望格式（扁平化）查询，需要修改存储逻辑
   - 或者需要修改查询逻辑以适应嵌套结构

## 五、建议

### 5.1 选项 1：保持当前格式（推荐用于快速实现）

**优点**：
- 实现简单，直接存储 Stream 数据
- 保留完整的原始数据信息
- 便于调试和问题排查

**缺点**：
- 数据冗余
- 查询路径复杂

**适用场景**：
- 当前阶段，快速实现功能
- 数据量不大，冗余可接受

### 5.2 选项 2：展开 data_value 到顶层（符合 SQL 注释期望）

**实现方式**：
```go
// 在 Insert 方法中，展开 data_value
if dataValue, ok := data["data_value"]; ok {
    // 如果 data_value 是对象，展开到顶层
    if dataValueMap, ok := dataValue.(map[string]interface{}); ok {
        for k, v := range dataValueMap {
            data[k] = v
        }
    }
    // 如果 data_value 是数组，需要特殊处理（可能需要拆分为多条记录）
    // 删除 data_value 字段
    delete(data, "data_value")
}

// 删除冗余的元数据字段
delete(data, "device_id")    // 已通过 uid 存储
delete(data, "tenant_id")    // 可通过 uid 关联获取
delete(data, "timestamp")    // 已通过 timestamp 列存储
delete(data, "branch_id")    // 可通过关联表获取
delete(data, "building_id")   // 可通过关联表获取
delete(data, "unit_id")      // 可通过关联表获取
delete(data, "room_id")      // 可通过关联表获取
delete(data, "bed_id")       // 可通过关联表获取

// 保留 device_type 和 topic_type（用于查询）
```

**优点**：
- 符合 SQL 注释中的期望格式
- 查询路径简单：`data_values->>'position_x'`
- 减少冗余数据

**缺点**：
- 实现复杂，需要处理数组类型的 `data_value`（可能需要拆分为多条记录）
- 丢失了原始 Stream 格式信息

**适用场景**：
- 需要按照 SQL 注释中的格式查询
- 数据量大，需要优化存储空间

### 5.3 选项 3：混合方案

**实现方式**：
- 保留 `data_value` 嵌套结构（用于兼容）
- 同时展开常用字段到顶层（用于查询优化）

```go
// 展开常用字段到顶层，同时保留 data_value
if dataValue, ok := data["data_value"]; ok {
    if dataValueMap, ok := dataValue.(map[string]interface{}); ok {
        // 展开常用字段
        if positionX, ok := dataValueMap["position_x"]; ok {
            data["position_x"] = positionX
        }
        if pose, ok := dataValueMap["pose"]; ok {
            data["pose"] = pose
        }
        // ... 其他常用字段
    }
    // 保留 data_value 用于完整数据访问
}
```

**优点**：
- 兼顾查询效率和数据完整性
- 向后兼容

**缺点**：
- 数据冗余（常用字段在顶层和 data_value 中都有）

## 六、当前实现状态

**当前实现**：选项 1（保持 Stream 格式）

**存储的 data_values 结构**（新格式，已统一）：
```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "uuid-yyy",
  "timestamp": 1234567890,
  "topic_type": "monitor",
  "data_value": {
    "category": "track",
    "target_id": 1,
    "position_x": 150,
    "position_y": 200,
    "position_z": 50,
    "pose": "Walking",
    "pose_snomed_code": "129006008",
    "pose_snomed_display": "Walking",
    "event": 0,
    "area_id": 1
  },
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**关键变化**：
1. ✅ `category` 保留在 `data_value` 内部，避免数据冗余
2. ✅ 字段顺序已标准化：`device_id` → `device_type` → `tenant_id` → `timestamp` → `topic_type` → `data_value` → 位置信息
3. ✅ Stream 格式与 DB 存储格式已统一

**查询示例**：
```sql
-- 查询 category（在 data_value 内部）
SELECT data_values->'data_value'->>'category' 
FROM iot_timeseries 
WHERE uid = 'device-uid' AND data_values->'data_value'->>'category' = 'track';

-- 如果 data_value 是数组，需要访问第一个元素
SELECT data_values->'data_value'->0->>'category' 
FROM iot_timeseries 
WHERE uid = 'device-uid' AND data_values->'data_value'->0->>'category' = 'track';

-- 查询 topic_type（在顶层）
SELECT data_values->>'topic_type' 
FROM iot_timeseries 
WHERE uid = 'device-uid';

-- 查询 position_x（需要访问嵌套的 data_value）
SELECT data_values->'data_value'->>'position_x' 
FROM iot_timeseries 
WHERE uid = 'device-uid';

-- 查询 timestamp（在顶层，便于时间范围查询）
SELECT data_values->>'timestamp' 
FROM iot_timeseries 
WHERE uid = 'device-uid' 
  AND (data_values->>'timestamp')::bigint >= 1234567890;
```

## 七、当前实现状态（已更新）

**当前实现**：方案3（Stream 和 DB 格式已统一）

**已完成**：
1. ✅ `category` 保留在 `data_value` 内部，避免数据冗余
2. ✅ 字段顺序已调整：`device_id` → `device_type` → `tenant_id` → `timestamp` → `topic_type` → `data_value` → 位置信息
3. ✅ Stream 格式与 DB 存储格式已统一
4. ✅ wisefido-radar 和 wisefido-sleepace 都已更新

**优势**：
- 数据不冗余，`category` 作为数据内容的一部分保留在 `data_value` 内部
- 字段顺序符合访问频率和逻辑分组
- Stream 和 DB 格式一致，便于调试和维护

## 八、后续建议

1. **已完成**：Stream 和 DB 格式已统一，`category` 保留在 `data_value` 内部，避免冗余
2. **监控**：观察查询性能，如果 `data_value` 嵌套访问成为瓶颈，可考虑展开常用字段
3. **优化**：根据实际查询模式，考虑在 `data_values` JSONB 上创建更精确的索引

## 九、字段顺序调整说明

### 9.1 调整原因

文档中所有字段顺序已统一调整为：`device_id` → `device_type` → `tenant_id` → `timestamp` → `topic_type` → `data_value` → 位置信息（`category` 保留在 `data_value` 内部）。

**调整原因**：

1. **符合实际代码实现**：
   - wisefido-radar 的 `mqtt_consumer.go` 中已按照此顺序构建 `encodedData`
   - 确保文档与实际代码实现保持一致，避免混淆

2. **逻辑分组清晰**：
   - **设备标识**：`device_id`, `device_type`（设备基本信息）
   - **租户信息**：`tenant_id`（多租户隔离）
   - **时间信息**：`timestamp`（数据时间戳）
   - **数据类型**：`topic_type`（`category` 在 `data_value` 内部）
   - **数据内容**：`data_value`（实际数据）
   - **位置信息**：`branch_id`, `building_id`, `unit_id`, `room_id`, `bed_id`（可选位置信息）

3. **查询优化**：
   - 将最常用的查询字段（`device_id`, `tenant_id`, `timestamp`, `topic_type`）放在前面
   - `category` 保留在 `data_value` 内部，作为数据内容的一部分
   - 位置信息放在最后，因为通常通过关联表查询，较少直接使用

4. **维护一致性**：
   - 与 `RADAR_REDIS_STREAM_FORMAT_STANDARD.md` 保持一致
   - 与 `18_iot_timeseries.sql` 中的注释保持一致
   - 确保整个系统的文档和代码都遵循相同的字段顺序规范

### 9.2 调整内容

本次调整更新了以下部分：
- ✅ 1.1 标准格式结构：更新字段顺序说明
- ✅ 1.2 实际发布代码：更新代码示例，去掉 category 提取逻辑，category 保留在 data_value 内部
- ✅ 3.2 实际存储格式：更新 JSON 示例中的字段顺序
- ✅ 六、当前实现状态：更新字段顺序说明
- ✅ 七、当前实现状态（已更新）：确保字段顺序描述一致

### 9.3 影响范围

- **代码层面**：无需修改，代码已按新顺序实现
- **数据库层面**：无需修改，JSONB 字段顺序不影响查询
- **文档层面**：已统一更新，确保文档准确性
- **查询层面**：字段顺序不影响 JSONB 查询性能，但遵循统一顺序有助于代码可读性
