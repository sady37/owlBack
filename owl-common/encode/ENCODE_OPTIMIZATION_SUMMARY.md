# Encode 公共库优化总结

## 一、优化目标

完成并优化两个厂家（Radar 和 Sleepace）的数据格式转换公共函数，消除重复代码，提高代码可维护性。

## 二、优化内容

### 2.1 统一 SNOMED 映射函数

**问题**：
- `radar_encoder.go` 中有 `applySNOMedMapping()` 函数
- `sleepace_encoder.go` 中有 `applySleepaceSNOMedMapping()` 函数
- 两个函数功能完全相同，只是调用的映射函数不同

**解决方案**：
- 创建 `snomed_mapping.go` 文件，统一 SNOMED 映射逻辑
- 实现通用的 `applySNOMedMappingInternal()` 函数（接受映射函数作为参数）
- 提供便捷函数：
  - `applyRadarSNOMedMapping()` - Radar 专用
  - `applySleepaceSNOMedMapping()` - Sleepace 专用
  - `applySNOMedMapping()` - Radar 编码器使用的向后兼容函数

**文件结构**：
```
owl-common/encode/
├── snomed_mapping.go          # 新增：统一的 SNOMED 映射函数
├── radar_encoder.go           # 使用 applySNOMedMapping()（向后兼容）
└── sleepace_encoder.go        # 使用 applySleepaceSNOMedMapping()
```

### 2.2 代码复用

**优化前**：
- Radar 和 Sleepace 各自实现 SNOMED 映射逻辑
- 代码重复，维护成本高

**优化后**：
- 统一的 SNOMED 映射实现
- 通过函数参数区分不同的映射查询函数
- 代码复用率提高，维护成本降低

## 三、文件清单

### 3.1 核心文件

| 文件 | 功能 | 状态 |
|------|------|------|
| `common.go` | 公共辅助函数（copyOtherFields, parseInt, parseFloat） | ✅ 已完成 |
| `radar_convert.go` | Radar 转换表加载和查询工具 | ✅ 已完成 |
| `sleepace_convert.go` | Sleepace 转换表加载和查询工具 | ✅ 已完成 |
| `radar_encoder.go` | Radar 数据编码函数（单位转换、SNOMED 映射） | ✅ 已完成 |
| `sleepace_encoder.go` | Sleepace 数据编码函数（字段重命名、SNOMED 映射） | ✅ 已完成 |
| `snomed_mapping.go` | **新增**：统一的 SNOMED 映射函数 | ✅ 已优化 |

### 3.2 配置文件

| 文件 | 功能 | 状态 |
|------|------|------|
| `config/radar_convert_table.json` | Radar 字段转换表（单位转换、SNOMED 映射） | ✅ 已完成 |
| `config/sleepace_convert_table.json` | Sleepace 字段转换表（SNOMED 映射） | ✅ 已完成 |
| `config/Radar_HTTPS_MQTT_Protocol_Formatted.md` | Radar 协议文档 | ✅ 已完成 |
| `config/06_FHIR_Simple_Conversion_Guide.md` | FHIR/SNOMED 标准参考 | ✅ 已完成 |

## 四、函数接口

### 4.1 Radar 编码函数

```go
// RadarEncode 编码 Radar 数据
// data: 包含 device_id, tenant_id, device_type, topic_type, timestamp 和原始数据字段
// topicType: 主题类型 ("monitor", "stat", "event", "alarm")
// 返回: 编码后的数据（单位转换、格式统一、SNOMED 映射）
func RadarEncode(data map[string]interface{}, topicType string) (map[string]interface{}, error)
```

**支持的 topicType**：
- `"monitor"` - 实时数据（轨迹、呼吸心率）
- `"stat"` - 统计数据（轨迹统计、睡眠统计）
- `"event"` - 事件数据（进出事件、姿态变化事件）
- `"alarm"` - 告警数据（跌倒、异常等）

### 4.2 Sleepace 编码函数

```go
// SleepaceEncode 编码 Sleepace 数据
// data: 包含 device_id, tenant_id, device_type, data_key, timestamp 和原始数据字段
// dataKey: 数据类型 ("realtime", "sleepStage", "connectionStatus", "alarmNotify")
// 返回: 编码后的数据（格式统一、字段重命名、SNOMED 映射）
func SleepaceEncode(data map[string]interface{}, dataKey string) (map[string]interface{}, error)
```

**支持的 dataKey**：
- `"realtime"` - 实时数据（呼吸心率、床状态、坐起等）
- `"sleepStage"` - 睡眠阶段数据
- `"connectionStatus"` - 连接状态数据
- `"alarmNotify"` - 告警通知数据

### 4.3 公共辅助函数

```go
// copyOtherFields 复制其他字段（排除指定字段）
func copyOtherFields(source, target map[string]interface{}, exclude []string)

// parseInt 将值转换为整数
func parseInt(value interface{}) (int, error)

// parseFloat 将值转换为浮点数
func parseFloat(value interface{}) (float64, error)
```

### 4.4 SNOMED 映射函数（统一后）

```go
// applySNOMedMappingInternal 应用 SNOMED 映射（内部通用函数）
func applySNOMedMappingInternal(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{}, getMappingFunc func(string, interface{}) (*SNOMEDMapping, error))

// applyRadarSNOMedMapping 应用 Radar SNOMED 映射（便捷函数）
func applyRadarSNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{})

// applySleepaceSNOMedMapping 应用 Sleepace SNOMED 映射（便捷函数）
func applySleepaceSNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{})

// applySNOMedMapping 应用 Radar SNOMED 映射（向后兼容函数）
func applySNOMedMapping(encoded map[string]interface{}, fieldName, fieldPath string, rawValue interface{})
```

## 五、功能特性

### 5.1 Radar 编码功能

✅ **单位转换**：
- 位置坐标：dm → cm（`position_x`, `position_y`）
- 行走距离：m → cm（`walk_distance`）
- 时间单位：10秒 → 秒（`fall_alarm_duration`, `sitting_alarm_duration`）
- 配置参数：双向转换（`radar_install_height`, `rectangle`, `declare_area`）

✅ **Base64 解码**：
- `bh` 字段：解码并提取位字段（`sleep_status`, `stability`）
- `sleep` 字段：解码并提取位字段（`breath_state`, `heart_state`, `vital_signs_state`, `sleep_state`）
- `track` 字段：解码并提取多字节字段（`walk_distance`）

✅ **SNOMED 映射**：
- 姿态（`pose`）：行走、站立、坐位、卧位、跌倒等
- 事件（`event`）：进入房间、离开房间、进入区域、离开区域等
- 睡眠状态（`sleep_state`）：清醒、浅睡眠、深睡眠
- 呼吸心率状态（`breath_state`, `heart_state`）：正常、过低、过高、暂停
- 生命体征情况（`vital_signs_state`）：正常、弱

### 5.2 Sleepace 编码功能

✅ **字段重命名**：
- `breath` → `respiratory_rate`（保留原始字段名）
- `heart` → `heart_rate`（保留原始字段名）

✅ **SNOMED 映射**：
- 床状态（`bedStatus`）：在床、离床
- 睡眠阶段（`sleepStage`）：清醒、浅睡眠、深睡眠、REM睡眠
- 坐起（`sitUp`）：坐起事件
- 连接状态（`connectionStatus`）：离线、在线
- 告警类型（`alarmType`）：离床、坐起、呼吸暂停、心率异常、设备离线等

## 六、代码质量

### 6.1 代码复用

- ✅ 统一的 SNOMED 映射实现（`snomed_mapping.go`）
- ✅ 公共辅助函数（`common.go`）
- ✅ 转换表加载模式统一（`radar_convert.go`, `sleepace_convert.go`）

### 6.2 可维护性

- ✅ 清晰的函数命名和注释
- ✅ 统一的错误处理
- ✅ 向后兼容性（保留原有函数接口）

### 6.3 可扩展性

- ✅ 支持新增设备厂家（只需添加转换表和编码函数）
- ✅ 支持新增字段类型（通过转换表配置）
- ✅ 支持新增 SNOMED 映射（通过转换表配置）

## 七、使用示例

### 7.1 Radar 数据编码

```go
// 实时数据（monitor）
data := map[string]interface{}{
    "device_id": "RADAR001",
    "tenant_id": "TENANT001",
    "device_type": "Radar",
    "topic_type": "monitor",
    "timestamp": 1234567890,
    "pose": 5,
    "position_x": 12,  // 分米
    "position_y": 8,   // 分米
    "position_z": 150, // 厘米
}

encoded, err := encode.RadarEncode(data, "monitor")
// 输出：
// {
//   "device_id": "RADAR001",
//   "pose": 5,
//   "pose_snomed_code": "161898004",
//   "pose_category": "safety",
//   "pose_display_en": "Fall",
//   "position_x": 120,  // 已转换为厘米
//   "position_y": 80,   // 已转换为厘米
//   "position_z": 150,  // 保持厘米
//   ...
// }
```

### 7.2 Sleepace 数据编码

```go
// 实时数据（realtime）
data := map[string]interface{}{
    "device_id": "SLEEPACE001",
    "tenant_id": "TENANT001",
    "device_type": "Sleepace",
    "data_key": "realtime",
    "timestamp": 1234567890,
    "breath": 20,
    "heart": 75,
    "bedStatus": 0,
    "sitUp": 1,
}

encoded, err := encode.SleepaceEncode(data, "realtime")
// 输出：
// {
//   "device_id": "SLEEPACE001",
//   "breath": 20,                    // 保留原始字段
//   "respiratory_rate": 20,          // 字段重命名
//   "heart": 75,                     // 保留原始字段
//   "heart_rate": 75,                // 字段重命名
//   "bedStatus": 0,
//   "bedStatus_snomed_code": "248569007",
//   "bedStatus_category": "activity",
//   "bedStatus_display_en": "In bed",
//   "sitUp": 1,
//   "sitUp_snomed_code": "422256002",
//   "sitUp_category": "activity",
//   "sitUp_display_en": "Sitting up in bed",
//   ...
// }
```

## 八、优化成果

### 8.1 代码减少

- **SNOMED 映射函数**：从 2 个重复实现 → 1 个统一实现
- **代码行数**：减少约 30 行重复代码

### 8.2 维护性提升

- ✅ 统一的 SNOMED 映射逻辑，修改一处即可
- ✅ 清晰的函数职责划分
- ✅ 向后兼容，不影响现有代码

### 8.3 可扩展性提升

- ✅ 新增设备厂家时，只需实现编码函数和转换表
- ✅ 统一的 SNOMED 映射接口，便于扩展

## 九、后续优化建议

### 9.1 转换表加载器统一

**当前状态**：
- `radar_convert.go` 和 `sleepace_convert.go` 有相似的加载逻辑
- 可以进一步抽象为通用的转换表加载器

**建议**：
- 创建通用的 `ConvertTableLoader` 接口
- 实现 `RadarConvertTableLoader` 和 `SleepaceConvertTableLoader`
- 进一步减少重复代码

### 9.2 单元测试

**建议**：
- 为 `snomed_mapping.go` 添加单元测试
- 为 `radar_encoder.go` 和 `sleepace_encoder.go` 添加集成测试
- 确保代码质量和功能正确性

### 9.3 文档完善

**建议**：
- 添加 API 文档（godoc）
- 添加使用示例和最佳实践
- 添加性能优化建议

## 十、总结

✅ **已完成**：
1. 统一 SNOMED 映射函数（`snomed_mapping.go`）
2. 消除重复代码
3. 保持向后兼容性
4. 提高代码可维护性

✅ **功能完整**：
1. Radar 数据编码（单位转换、Base64 解码、SNOMED 映射）
2. Sleepace 数据编码（字段重命名、SNOMED 映射）
3. 公共辅助函数（字段复制、类型转换）

✅ **代码质量**：
1. 清晰的函数命名和注释
2. 统一的错误处理
3. 良好的可扩展性

**两个厂家的数据格式转换公共函数已完成并优化！** 🎉
