# Sleepace Encoder SNOMED 映射实现总结

## 一、实现概述

参考 `radar_encoder.go` 的实现方式，为 `sleepace_encoder.go` 实现了完整的 SNOMED 映射功能。

## 二、新增文件

### 2.1 `config/sleepace_convert_table.json`

Sleepace 设备字段到 SNOMED 标准格式的转换表，包含：

- **realtime** 数据类型：
  - `bedStatus` (0=在床, 1=离床) → SNOMED 映射（category: activity）
  - `sleepStage` (0=清醒, 1=浅睡眠, 2=深睡眠, 3=REM睡眠) → SNOMED 映射（category: activity）
  - `sitUp` (0=不坐起, 1=坐起) → SNOMED 映射（category: activity）
  - `breath` / `heart` (生命体征) → category: vital-signs（数值本身不需要 SNOMED）

- **sleepStage** 数据类型：
  - `sleepStage` → SNOMED 映射（category: activity）

- **connectionStatus** 数据类型：
  - `connectionStatus` (0=离线, 1=在线) → SNOMED 映射（category: device）

- **alarmNotify** 数据类型：
  - `type` (报警类型) → SNOMED 映射（category: safety/clinical/behavioral/device）

### 2.2 `sleepace_convert.go`

Sleepace 专用的转换表加载和查询工具，提供：

- `loadSleepaceConvertTable()` - 加载 `sleepace_convert_table.json`
- `GetSleepaceFieldConversion(fieldPath)` - 获取字段转换规则
- `GetSleepaceSNOMEDMappingByFieldPath(fieldPath, rawValue)` - 获取 SNOMED 映射

## 三、更新文件

### 3.1 `sleepace_encoder.go`

实现了完整的 SNOMED 映射功能：

#### 3.1.1 `encodeSleepaceRealtime()` - 实时数据编码

- **bedStatus**: 应用 SNOMED 映射（`realtime.bedStatus`）
  - 0 → `248569007` (In bed, activity)
  - 1 → `248570008` (Not in bed / Left bed, activity)

- **sleepStage**: 应用 SNOMED 映射（`realtime.sleepStage`）
  - 0 → `248220002` (Awake, activity)
  - 1 → `248232005` (Light sleep, activity)
  - 2 → `248233000` (Deep sleep, activity)
  - 3 → `248234006` (REM sleep, activity)

- **sitUp**: 应用 SNOMED 映射（`realtime.sitUp`）
  - `sitUp > 0` → 映射为 `1`，获取 SNOMED 映射（`422256002` - Sitting up in bed, activity）
  - `sitUp = 0` → 不映射，只保留原始值

- **breath** / **heart**: 字段重命名（`respiratory_rate`, `heart_rate`），保留原始字段名
  - 数值本身不需要 SNOMED 映射，但数据流的 category 应为 `vital-signs`

- **turnOver** / **bodyMove**: 直接保留，不进行 SNOMED 映射

#### 3.1.2 `encodeSleepaceSleepStage()` - 睡眠阶段数据编码

- **sleepStage**: 应用 SNOMED 映射（`sleepStage.sleepStage`）
  - 映射规则与 `realtime.sleepStage` 相同

#### 3.1.3 `encodeSleepaceConnectionStatus()` - 连接状态数据编码

- **connectionStatus**: 应用 SNOMED 映射（`connectionStatus.connectionStatus`）
  - 0 → `null` (Offline, device)
  - 1 → `706689003` (Online, device)

#### 3.1.4 `encodeSleepaceAlarmNotify()` - 告警通知数据编码

- **alarmType**: 应用 SNOMED 映射（`alarmNotify.type`）
  - `LeftBed` → `248570008` (Not in bed, behavioral)
  - `SitUp` → `422256002` (Sitting up in bed, behavioral)
  - `ApneaHypopnea` → `67905006` (Apnea, clinical)
  - `AbnormalHeartRate` / `AbnormalRespiratoryRate` → (clinical)
  - `OfflineAlarm` → `397942008` (Device offline, device)
  - `LowBattery` → `703507001` (Low battery, device)
  - `DeviceFailure` → (device)

#### 3.1.5 `applySleepaceSNOMedMapping()` - SNOMED 映射应用函数

参考 `radar_encoder.go` 中的 `applySNOMedMapping()` 函数，为编码后的数据添加：
- `{field_name}_snomed_code` - SNOMED 编码（如果有）
- `{field_name}_snomed_display` - SNOMED 显示名称
- `{field_name}_category` - Category 分类
- `{field_name}_display_en` - 英文显示名称

## 四、SNOMED 映射字段

### 4.1 realtime 数据类型

| 字段名 | Category | SNOMED 映射字段 |
|-------|---------|----------------|
| `bedStatus` | `activity` | `bedStatus_snomed_code`, `bedStatus_snomed_display`, `bedStatus_category`, `bedStatus_display_en` |
| `sleepStage` | `activity` | `sleepStage_snomed_code`, `sleepStage_snomed_display`, `sleepStage_category`, `sleepStage_display_en` |
| `sitUp` | `activity` | `sitUp_snomed_code`, `sitUp_snomed_display`, `sitUp_category`, `sitUp_display_en` |
| `breath` / `heart` | `vital-signs` | 无 SNOMED 映射字段（数值本身不需要映射） |

### 4.2 sleepStage 数据类型

| 字段名 | Category | SNOMED 映射字段 |
|-------|---------|----------------|
| `sleepStage` | `activity` | `sleepStage_snomed_code`, `sleepStage_snomed_display`, `sleepStage_category`, `sleepStage_display_en` |

### 4.3 connectionStatus 数据类型

| 字段名 | Category | SNOMED 映射字段 |
|-------|---------|----------------|
| `connectionStatus` | `device` | `connectionStatus_snomed_code`, `connectionStatus_snomed_display`, `connectionStatus_category`, `connectionStatus_display_en` |

### 4.4 alarmNotify 数据类型

| 字段名 | Category | SNOMED 映射字段 | 说明 |
|-------|---------|----------------|------|
| `alarmType` | `behavioral` / `clinical` / `device` | `alarmType_snomed_code`, `alarmType_snomed_display`, `alarmType_category`, `alarmType_display_en` | 根据具体报警类型确定 category |

## 五、Category 分类规则

### 5.1 Observation 类型（realtime, sleepStage, connectionStatus）

1. **`activity`**：用于行为观测数据
   - 床状态（bedStatus）：在床/离床
   - 睡眠阶段（sleepStage）：清醒/浅睡眠/深睡眠/REM睡眠
   - 行为事件（sitUp）：床上坐起

2. **`vital-signs`**：用于生命体征基础测量值
   - 呼吸率（breath）：次/分钟
   - 心率（heart）：次/分钟

3. **`device`**：用于设备技术状态
   - 连接状态（connectionStatus）：在线/离线

### 5.2 Flag 类型（alarmNotify）

1. **`behavioral`**：行为健康告警
   - `LeftBed` - 离床告警
   - `SitUp` - 床上坐起告警

2. **`clinical`**：生命体征异常告警
   - `ApneaHypopnea` - 呼吸暂停告警
   - `AbnormalHeartRate` - 心率异常告警
   - `AbnormalRespiratoryRate` - 呼吸频率异常告警

3. **`device`**：设备技术告警
   - `OfflineAlarm` - 设备离线告警
   - `LowBattery` - 低电量告警
   - `DeviceFailure` - 设备故障告警

## 六、关键实现细节

### 6.1 sitUp 字段处理

- `sitUp = 0`：不映射，只保留原始值
- `sitUp > 0`：映射为 `1`，然后获取 SNOMED 映射（`422256002` - Sitting up in bed）

### 6.2 breath/heart 字段处理

- 字段重命名：`breath` → `respiratory_rate`，`heart` → `heart_rate`
- 保留原始字段名：`breath` 和 `heart` 也保留在编码后的数据中
- 不进行 SNOMED 映射：数值本身不需要映射，但数据流的 category 应为 `vital-signs`

### 6.3 alarmType 字段处理

- 优先使用 `alarmType` 字段（从 `handleAlarmNotify` 中设置）
- 如果只有 `type` 字段，也进行映射（兼容性处理）
- 根据不同的报警类型，category 可能为 `behavioral`、`clinical` 或 `device`

## 七、与 Radar 实现的对比

| 特性 | Radar | Sleepace |
|------|-------|----------|
| 转换表文件 | `radar_convert_table.json` | `sleepace_convert_table.json` |
| 转换表加载器 | `radar_convert.go` | `sleepace_convert.go` |
| 数据编码器 | `radar_encoder.go` | `sleepace_encoder.go` |
| Base64 解码 | ✅ 需要（`bh`, `sleep`, `track`） | ❌ 不需要 |
| 位字段解析 | ✅ 需要（bit-fields） | ❌ 不需要 |
| 单位转换 | ✅ 需要（dm→cm, m→cm, 10s→s） | ❌ 不需要 |
| SNOMED 映射 | ✅ 完整实现 | ✅ 完整实现 |
| Category 分类 | ✅ 完整实现 | ✅ 完整实现 |

## 八、测试建议

1. **单元测试**：
   - 测试 `encodeSleepaceRealtime()` 的所有字段映射
   - 测试 `encodeSleepaceSleepStage()` 的 sleepStage 映射
   - 测试 `encodeSleepaceConnectionStatus()` 的 connectionStatus 映射
   - 测试 `encodeSleepaceAlarmNotify()` 的不同报警类型映射

2. **集成测试**：
   - 测试完整的编码流程（从原始数据到编码后数据）
   - 验证 SNOMED 映射字段是否正确添加
   - 验证 category 是否正确设置

3. **边界情况测试**：
   - `sitUp = 0` vs `sitUp > 0`
   - `bedStatus` 的 0 和 1 值
   - `sleepStage` 的所有可能值（0-3）
   - `alarmType` 的各种报警类型

## 九、使用示例

```go
// 实时数据编码
rawData := map[string]interface{}{
    "device_id": "SLEEPACE001",
    "tenant_id": "T001",
    "data_key": "realtime",
    "bedStatus": 0,
    "sleepStage": 2,
    "sitUp": 1,
    "breath": 20,
    "heart": 75,
}

encoded, err := SleepaceEncode(rawData, "realtime")
// 输出：
// {
//   "device_id": "SLEEPACE001",
//   "bedStatus": 0,
//   "bedStatus_snomed_code": "248569007",
//   "bedStatus_category": "activity",
//   "bedStatus_display_en": "In bed",
//   "sleepStage": 2,
//   "sleepStage_snomed_code": "248233000",
//   "sleepStage_category": "activity",
//   "sleepStage_display_en": "Deep sleep",
//   "sitUp": 1,
//   "sitUp_snomed_code": "422256002",
//   "sitUp_category": "activity",
//   "sitUp_display_en": "Sitting up in bed",
//   "breath": 20,
//   "respiratory_rate": 20,
//   "heart": 75,
//   "heart_rate": 75,
//   ...
// }
```

## 十、总结

✅ **已完成**：
- 创建了 `sleepace_convert_table.json` 转换表
- 创建了 `sleepace_convert.go` 转换表加载器
- 更新了 `sleepace_encoder.go`，实现了完整的 SNOMED 映射功能
- 所有 4 类数据流（realtime, sleepStage, connectionStatus, alarmNotify）都已实现 SNOMED 映射
- Category 分类规则已实现

✅ **与 Radar 实现一致**：
- 使用相同的 SNOMED 映射字段命名规范（`{field}_snomed_code`, `{field}_category`, `{field}_display_en`）
- 使用相同的 `applySNOMedMapping()` 函数模式
- 转换表结构一致

✅ **代码质量**：
- 编译通过
- 无 linter 错误
- 遵循 Go 编码规范
