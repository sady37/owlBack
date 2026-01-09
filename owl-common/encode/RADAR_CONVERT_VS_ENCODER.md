# radar_convert.go vs radar_encoder.go 的区别

## 一、文件职责对比

| 维度 | `radar_convert.go` | `radar_encoder.go` |
|------|-------------------|-------------------|
| **定位** | 底层工具库（Utility Layer） | 业务逻辑层（Business Logic Layer） |
| **职责** | 提供数据转换和映射查询的基础工具函数 | 实现完整的 Radar 数据编码业务流程 |
| **层级** | 低层（Low-level） | 高层（High-level） |
| **依赖关系** | 不依赖 `radar_encoder.go` | **依赖** `radar_convert.go` |

## 二、功能对比

### 2.1 `radar_convert.go` - 底层工具库

**核心功能**：
1. **转换表加载和管理**
   - `loadRadarConvertTable()` - 加载 `radar_convert_table.json`
   - 使用 `sync.Once` 确保只加载一次（单例模式）

2. **转换规则查询**
   - `GetFieldConversion(fieldPath)` - 根据字段路径获取转换规则
   - `GetSNOMEDMappingByFieldPath(fieldPath, rawValue)` - 获取 SNOMED 映射

3. **数值转换**
   - `ConvertFieldValue(fieldPath, rawValue)` - 转换字段值（设备 → Server）
   - `ConvertConfigValue(fieldPath, rawValue, direction)` - 转换配置参数（支持双向：read/write）
   - `convertNumericValue()` - 数值单位转换（如 dm → cm, m → cm, 10秒 → 秒）
   - `convertArrayValue()` - 数组值转换（如 rectangle, declare_area）

4. **坐标字符串处理**
   - `parseCoordinateString()` - 解析坐标字符串 `"{x1, y1; x2, y2, x3, y3, x4, y4}"`
   - `formatCoordinateString()` - 格式化坐标数组为字符串

5. **位字段解析**
   - `ParseBitField(byteValue, bitPosition)` - 从字节值中解析位字段（如 "7:6"）

6. **公式应用**
   - `applyFormula(value, formula)` - 应用转换公式（如 "value * 10", "value / 10"）

7. **辅助函数**
   - `splitFieldPath()` - 分割字段路径

**特点**：
- ✅ **通用性强**：可以用于任何字段的转换，不局限于 Radar 数据
- ✅ **可复用**：其他服务（如配置管理、设备属性读写）也可以使用
- ✅ **无状态**：所有函数都是纯函数（pure functions），没有副作用
- ✅ **支持双向转换**：支持设备 → Server（read）和 Server → 设备（write）

### 2.2 `radar_encoder.go` - 业务逻辑层

**核心功能**：
1. **完整的 Radar 数据编码流程**
   - `RadarEncode(data, topicType)` - 主入口函数，根据 topic_type 分发到不同处理函数

2. **4 类数据流的编码实现**
   - `encodeRadarMonitor()` - 实时数据（monitor）编码
   - `encodeRadarStat()` - 统计数据（stat）编码
   - `encodeRadarEvent()` - 事件数据（event）编码
   - `encodeRadarAlarm()` - 告警数据（alarm）编码

3. **复杂数据处理**
   - Base64 解码（`bh`, `sleep`, `track` 字段）
   - 位字段提取（从字节数组中提取位字段值）
   - 多字节字段解析（Big-Endian/Little-Endian）
   - 数组/对象结构的处理

4. **SNOMED 映射应用**
   - `applySNOMedMapping()` - 应用 SNOMED 映射到编码后的数据
   - 自动添加 `_snomed_code`, `_snomed_display`, `_category`, `_display_en` 字段

**特点**：
- ✅ **业务导向**：专门处理 Radar 设备的 4 类数据流
- ✅ **流程完整**：从原始数据到最终编码数据的完整流程
- ✅ **数据理解**：了解 Radar 数据的具体格式和结构
- ✅ **调用底层工具**：大量使用 `radar_convert.go` 中的工具函数

## 三、依赖关系

```
┌─────────────────────────────────┐
│   radar_encoder.go              │
│   (业务逻辑层)                   │
│                                 │
│   - RadarEncode()              │
│   - encodeRadarMonitor()       │
│   - encodeRadarStat()          │
│   - encodeRadarEvent()         │
│   - encodeRadarAlarm()         │
│   - applySNOMedMapping()       │
└──────────────┬──────────────────┘
               │ 调用
               ▼
┌─────────────────────────────────┐
│   radar_convert.go              │
│   (底层工具库)                   │
│                                 │
│   - GetFieldConversion()       │
│   - GetSNOMEDMappingByFieldPath()│
│   - ConvertFieldValue()        │
│   - ConvertConfigValue()       │
│   - ParseBitField()            │
│   - parseCoordinateString()    │
│   - applyFormula()             │
└─────────────────────────────────┘
               │ 使用
               ▼
┌─────────────────────────────────┐
│   radar_convert_table.json      │
│   (配置数据)                     │
└─────────────────────────────────┘
```

## 四、使用场景对比

### 4.1 `radar_convert.go` 的使用场景

1. **配置参数读写**（`wisefido-radar` 服务）：
   ```go
   // 读取设备属性：设备单位 → Server 单位
   height, _, err := ConvertConfigValue("config.radar_install_height", deviceValue, "read")
   // height: 28 (dm) → 280 (cm)

   // 写入设备属性：Server 单位 → 设备单位
   deviceHeight, _, err := ConvertConfigValue("config.radar_install_height", serverValue, "write")
   // serverValue: 280 (cm) → 28 (dm)
   ```

2. **坐标边界设置**：
   ```go
   // Server: 厘米 → 设备: 分米
   rectangle, _, err := ConvertConfigValue("config.rectangle", serverRect, "write")
   ```

3. **单个字段转换**：
   ```go
   // 查询字段转换规则
   conv, err := GetFieldConversion("monitor.track.position_x")
   
   // 获取 SNOMED 映射
   mapping, err := GetSNOMEDMappingByFieldPath("monitor.track.pose", 5)
   ```

### 4.2 `radar_encoder.go` 的使用场景

1. **完整数据流处理**（`wisefido-radar` MQTT Consumer）：
   ```go
   // 处理设备上报的 MQTT 消息
   encoded, err := RadarEncode(rawData, "monitor")
   // 自动完成：base64 解码、位字段提取、单位转换、SNOMED 映射
   ```

2. **统计数据解码**：
   ```go
   // 处理 sleep/track 统计数据
   encoded, err := RadarEncode(statData, "stat")
   // 自动完成：base64 解码、位字段提取、单位转换、SNOMED 映射
   ```

3. **事件/告警数据处理**：
   ```go
   // 处理事件数据
   encoded, err := RadarEncode(eventData, "event")
   
   // 处理告警数据
   encoded, err := RadarEncode(alarmData, "alarm")
   ```

## 五、代码示例对比

### 5.1 `radar_convert.go` - 底层工具函数

```go
// 单个字段转换
height, mapping, err := ConvertConfigValue("config.radar_install_height", 28, "read")
// 输入: 28 (dm)
// 输出: 280 (cm), mapping (包含 SNOMED 信息)

// 位字段解析
bitValue, err := ParseBitField(0x85, "7:6")
// 输入: 0x85 (133), "7:6" (bit 7 到 bit 6)
// 输出: "10" (二进制字符串)

// 查询转换规则
conv, err := GetFieldConversion("monitor.track.pose")
// 返回: FieldConversion 结构，包含 mappings, unit_conversion 等
```

### 5.2 `radar_encoder.go` - 业务逻辑函数

```go
// 完整的数据编码流程
rawData := map[string]interface{}{
    "device_id": "RADAR001",
    "tenant_id": "T001",
    "topic_type": "monitor",
    "bh": "AAECAwQFBgcICQoLDA0ODw==",  // base64 编码的呼吸心率数据
    "pose": 5,
    "position_x": 12,  // 分米
}

encoded, err := RadarEncode(rawData, "monitor")
// 自动完成：
// 1. Base64 解码 bh 字段
// 2. 提取 sleep_status (bit 7:6), stability (bit 1:0)
// 3. 单位转换 position_x: 12 (dm) → 120 (cm)
// 4. SNOMED 映射 pose: 5 → Fall (safety)
// 5. 添加所有 SNOMED 字段：pose_snomed_code, pose_category, pose_display_en 等

// 输出：
// {
//   "device_id": "RADAR001",
//   "pose": 5,
//   "pose_snomed_code": "161898004",
//   "pose_category": "safety",
//   "pose_display_en": "Fall",
//   "position_x": 120,  // 已转换为厘米
//   "sleep_status": "11",
//   "sleep_status_category": "activity",
//   "sleep_status_display_en": "Awake",
//   ...
// }
```

## 六、设计模式

### 6.1 `radar_convert.go` - 策略模式（Strategy Pattern）

- 转换规则定义在配置文件中（`radar_convert_table.json`）
- 通过 `fieldPath` 查找对应的转换策略
- 支持不同的转换类型（enum, bit_field, numeric, array）

### 6.2 `radar_encoder.go` - 模板方法模式（Template Method Pattern）

- `RadarEncode()` 是模板方法，定义整体编码流程
- 根据 `topicType` 调用不同的具体实现（`encodeRadarMonitor`, `encodeRadarStat` 等）
- 每个具体实现都遵循相同的模式：单位转换 → SNOMED 映射 → 字段复制

## 七、关键区别总结

| 特性 | `radar_convert.go` | `radar_encoder.go` |
|------|-------------------|-------------------|
| **抽象级别** | 低（原子操作） | 高（业务流程） |
| **数据理解** | 不了解 Radar 数据格式 | 了解 Radar 数据格式 |
| **使用粒度** | 单个字段或值 | 完整数据对象 |
| **Base64 处理** | ❌ 不处理 | ✅ 处理（解码、提取位字段） |
| **业务流程** | ❌ 无业务流程 | ✅ 完整的编码流程 |
| **可复用性** | ✅ 高（通用工具） | ❌ 低（Radar 专用） |
| **调用方** | `radar_encoder.go` 和其他服务 | MQTT Consumer、HTTP Handler 等 |
| **复杂度** | 低（单一职责） | 高（多步骤处理） |

## 八、最佳实践

1. **需要单个字段转换时**：使用 `radar_convert.go` 的函数
   ```go
   mapping, err := GetSNOMEDMappingByFieldPath("monitor.track.pose", 5)
   ```

2. **需要完整数据编码时**：使用 `radar_encoder.go` 的函数
   ```go
   encoded, err := RadarEncode(rawData, "monitor")
   ```

3. **配置参数读写时**：使用 `radar_convert.go` 的 `ConvertConfigValue`
   ```go
   deviceValue, _, err := ConvertConfigValue("config.radar_install_height", serverValue, "write")
   ```

4. **需要自定义处理流程时**：组合使用两个文件中的函数
   ```go
   // 先解码 base64
   bytes, _ := base64.StdEncoding.DecodeString(base64Str)
   // 再提取位字段
   bitValue, _ := ParseBitField(bytes[13], "7:6")
   // 最后获取 SNOMED 映射
   mapping, _ := GetSNOMEDMappingByFieldPath("stat.sleep.hr_breath_event.breath_state", bitValue)
   ```

## 九、总结

- **`radar_convert.go`**：底层工具库，提供数据转换和映射查询的基础能力，**可复用、通用性强**。
- **`radar_encoder.go`**：业务逻辑层，实现 Radar 数据编码的完整流程，**专用性强、流程完整**。

两者是**互补关系**：`radar_encoder.go` 依赖 `radar_convert.go` 提供的工具函数来完成底层的数据转换和映射查询，而 `radar_convert.go` 可以被其他服务直接使用，不依赖于 `radar_encoder.go`。

这种设计符合**分层架构**和**单一职责原则**，使得代码更加清晰、可维护、可测试。
