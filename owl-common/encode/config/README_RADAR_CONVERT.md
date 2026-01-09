# Radar 转换表使用说明

## 概述

`radar_convert_table.json` 是 Radar 设备字段到 SNOMED 标准格式的转换表。它基于两个文档：
- `06_FHIR_Simple_Conversion_Guide.md` - FHIR/SNOMED 标准
- `Radar_HTTPS_MQTT_Protocol_Formatted.md` - Radar 厂家数据格式

## 文件结构

```
encode/config/
├── 06_FHIR_Simple_Conversion_Guide.md      # FHIR/SNOMED 标准文档
├── Radar_HTTPS_MQTT_Protocol_Formatted.md  # Radar 厂家协议文档
├── radar_convert_table.json                 # 转换表（核心）
└── snomed_mapping.json                      # SNOMED 映射（已废弃，使用转换表）
```

## 转换表结构

```json
{
  "version": "1.0.0",
  "last_updated": "2026-01-05T00:00:00Z",
  "description": "Radar 设备字段到 SNOMED 标准格式的转换表",
  "conversions": {
    "monitor": { ... },  // 实时数据
    "stat": { ... },     // 统计数据
    "event": { ... }     // 事件数据
  }
}
```

## 字段类型

### 1. enum（枚举类型）

用于姿态、事件等枚举值：

```json
{
  "field_path": "monitor.track.pose",
  "field_type": "enum",
  "mappings": {
    "0": {
      "snomed_code": null,
      "snomed_display": "Initialization",
      "category": "activity",
      "display_en": "Initialization"
    },
    "1": {
      "snomed_code": "129006008",
      "snomed_display": "Walking",
      "category": "activity",
      "display_en": "Walking"
    }
  }
}
```

### 2. bit_field（位字段）

用于位字段值（如睡眠状态、呼吸状态）：

```json
{
  "field_path": "stat.sleep.hr_breath_event.breath_state",
  "field_type": "bit_field",
  "byte_position": 13,
  "bit_position": "1:0",
  "mappings": {
    "00": {
      "snomed_code": null,
      "snomed_display": "Breath rate normal",
      "category": "vital-signs",
      "display_en": "Breath rate normal"
    }
  }
}
```

### 3. numeric（数值类型）

用于需要单位转换的数值字段：

**单向转换**（数据上报）：
```json
{
  "field_path": "monitor.track.position_x",
  "field_type": "numeric",
  "byte_position": 1,
  "unit_conversion": {
    "formula": "value * 10",
    "from_unit": "dm",
    "to_unit": "cm"
  }
}
```

**双向转换**（配置参数，读取和设置）：
```json
{
  "field_path": "config.fall_param.fall_alarm_duration",
  "field_type": "numeric",
  "byte_position": 3,
  "unit_conversion": {
    "read_formula": "value * 10",
    "write_formula": "value / 10",
    "from_unit": "10秒单位",
    "to_unit": "秒",
    "direction": "bidirectional"
  }
}
```

### 4. base64_array（Base64 字节数组）

用于 base64 编码的字节数组（如 sleep、track 字段）：

```json
{
  "field_path": "stat.sleep.base64_structure",
  "field_type": "base64_array",
  "array_length": 16,
  "description": "base64 编码的 16 字节数组",
  "byte_definitions": [
    {
      "byte": 0,
      "name": "identifier",
      "type": "integer",
      "description": "固定为 0xff，表示为睡眠统计",
      "unit_conversion": null
    },
    {
      "byte": 13,
      "name": "hr_breath_event",
      "type": "bit_field",
      "description": "呼吸心率事件（位字段）",
      "bit_fields": [
        {
          "bits": [0, 1],
          "name": "breath_state",
          "field_path": "stat.sleep.hr_breath_event.breath_state",
          "mappings": { ... }
        }
      ]
    }
  ]
}
```

## 使用方法

### 1. 获取字段转换规则

```go
import "owl-common/encode"

conv, err := encode.GetFieldConversion("monitor.track.pose")
if err == nil {
    // 使用转换规则
    fmt.Printf("Field type: %s\n", conv.FieldType)
}
```

### 2. 转换字段值并获取 SNOMED 映射

```go
// 转换姿态值
value, mapping, err := encode.ConvertFieldValue("monitor.track.pose", 5)
if err == nil && mapping != nil {
    fmt.Printf("SNOMED Display: %s\n", mapping.SNOMEDDisplay)
    fmt.Printf("Category: %s\n", mapping.Category)
}
```

### 3. 直接获取 SNOMED 映射

```go
mapping, err := encode.GetSNOMEDMappingByFieldPath("monitor.track.pose", 5)
if err == nil {
    fmt.Printf("SNOMED Code: %v\n", mapping.SNOMEDCode)
    fmt.Printf("Display: %s\n", mapping.SNOMEDDisplay)
}
```

## 字段路径规范

字段路径采用点分隔的层级结构：

### 实时数据 (monitor)
- `monitor.track.pose` - 实时数据中的姿态
- `monitor.track.event` - 实时数据中的事件
- `monitor.track.position_x` - 实时数据中的 x 坐标（分米 → 厘米）
- `monitor.track.position_y` - 实时数据中的 y 坐标（分米 → 厘米）
- `monitor.track.position_z` - 实时数据中的 z 坐标（已是厘米）
- `monitor.bh.breath_rate` - 实时呼吸率
- `monitor.bh.heart_rate` - 实时心率
- `monitor.bh.sleep_status` - 睡眠状态（位字段）
- `monitor.bh.stability` - 稳定度（位字段）

### 统计数据 (stat)
- `stat.sleep.realtime_breath_rate` - 实时呼吸值
- `stat.sleep.realtime_heart_rate` - 实时心率值
- `stat.sleep.avg_breath_rate` - 分钟级平均呼吸
- `stat.sleep.avg_heart_rate` - 分钟级平均心率
- `stat.sleep.hr_breath_event.breath_state` - 呼吸状态（位字段）
- `stat.sleep.hr_breath_event.heart_state` - 心率状态（位字段）
- `stat.sleep.hr_breath_event.vital_signs_state` - 生命体征状态（位字段）
- `stat.sleep.hr_breath_event.sleep_state` - 睡眠状态（位字段）
- `stat.sleep.base64_structure` - sleep 字段的 base64 结构定义
- `stat.track.version` - 版本标识符
- `stat.track.people_count` - 人数
- `stat.track.walk_distance` - 行走距离（米 → 厘米）
- `stat.track.walk_duration` - 行走时长（秒）
- `stat.track.lie_duration` - 躺卧时长（秒）
- `stat.track.stand_duration` - 站立时长（秒）
- `stat.track.multi_person_duration` - 多人时长（秒）
- `stat.track.base64_structure` - track 字段的 base64 结构定义

### 事件数据 (event)
- `event.enter_leave.event` - 进出事件
- `event.pose_change.pose` - 姿态变化事件

### 配置参数 (config)
- `config.fall_param.fall_alarm_duration` - 跌倒告警时间（10秒单位 ↔ 秒，双向转换）
- `config.fall_param.sitting_alarm_duration` - 坐地告警时间（10秒单位 ↔ 秒，双向转换）
- `config.fall_param.base64_structure` - fall_param 字段的 base64 结构定义

## 优势

1. **统一管理**：所有转换规则集中在一个 JSON 文件中
2. **易于维护**：修改转换规则只需更新 JSON 文件
3. **类型安全**：明确的字段类型和转换规则
4. **标准化**：直接映射到 SNOMED 标准格式
5. **无需数据库**：配置文件直接嵌入到二进制文件中

## 更新流程

1. 更新 `radar_convert_table.json`
2. 重新编译服务
3. 部署新版本

## 单位转换规则

### 单向转换（数据上报）
- **分米 → 厘米**：`position_x`, `position_y` (`value * 10`)
- **米 → 厘米**：`walk_distance` (`value * 100`)

### 双向转换（配置参数）
- **10秒单位 ↔ 秒**：
  - 读取设备属性：`value * 10`（10秒单位 → 秒）
  - 设置设备属性：`value / 10`（秒 → 10秒单位）
  - 字段：`fall_alarm_duration`, `sitting_alarm_duration`

### 无需转换
- `position_z`（已是厘米）
- 所有时长字段（已是秒）
- 所有呼吸/心率字段（已是次/分钟）
- 所有枚举/状态字段

## Base64 字节数组结构

转换表包含完整的 base64 字节数组结构定义：

1. **sleep 字段**（16 字节）：
   - 字节 0: 标识符（0xff）
   - 字节 1-2: 实时呼吸/心率值
   - 字节 5-6: 分钟级平均呼吸/心率
   - 字节 13: 呼吸心率事件（位字段，包含 4 个子字段）

2. **track 字段**（16 字节）：
   - 字节 0: 版本标识符
   - 字节 1: 人数
   - 字节 2-3: 行走距离（Big-Endian，米 → 厘米）
   - 字节 4-8: 各种时长（秒）

3. **fall_param 字段**（16 字节）：
   - 字节 3: 跌倒告警时间（10秒单位 ↔ 秒）
   - 字节 4: 功能开关（位字段）
   - 字节 5: 坐地告警时间（10秒单位 ↔ 秒）

## 注意事项

1. **字段路径必须准确**：路径错误会导致转换失败
2. **原始值格式**：枚举值使用字符串格式（如 "0", "1", "00", "01"）
3. **单位转换**：数值类型的单位转换在 `radar_encoder.go` 中实现
4. **位字段提取**：位字段值需要先从字节中提取，再使用转换表
5. **Base64 解码**：base64 编码的字段需要先解码为字节数组，再根据 `byte_definitions` 解析
6. **字节序**：多字节字段（如 `walk_distance`）使用 Big-Endian 字节序
7. **双向转换**：配置参数的转换是双向的，读取和设置使用不同的公式
