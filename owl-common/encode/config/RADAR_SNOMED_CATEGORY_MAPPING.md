# Radar 数据流 SNOMED 格式统一规范

## 一、4类数据流的 SNOMED 映射

根据 FHIR/SNOMED 标准，Radar 设备的 4 类数据流（实时数据、轨迹/睡眠统计数据、事件/日志、告警）已统一按照 SNOMED 格式进行映射。

| 数据流类型 | MQTT Topic | 数据类型 | 说明 |
|----------|-----------|---------|------|
| **实时数据** | `monitor` | Observation | 实时轨迹、呼吸心率数据 |
| **统计数据** | `stat` | Observation | 分钟级轨迹统计、睡眠统计 |
| **事件数据** | `event` | Observation | 进出事件、姿态变化事件 |
| **告警数据** | `alarm` | Flag | 告警事件（跌倒、异常等） |

## 二、Category 分类体系

### 2.1 FHIR Flag（报警标志）的 4 大类

根据 FHIR 标准，**Flag（报警标志）分为 4 大类**，对应不同的 `flag-category`：

| 领域 | FHIR flag-category | 说明 | 用于资源类型 | 示例告警 |
|------|-------------------|------|------------|---------|
| **临床安全** | `safety` | 跌倒、滞留、危险姿态等直接威胁生命安全的事件 | Flag | Fall, SuspectedFall, Stay, NoActivity24h |
| **生命体征** | `clinical` | 心率/呼吸率异常等生命体征告警（FHIR推荐使用 `clinical`） | Flag | ApneaHypopnea, AbnormalHeartRate, AbnormalRespiratoryRate, VitalsWeak |
| **行为健康** | `behavioral` | 长时间无活动、异常行为模式等行为健康告警 | Flag | LeftBed, SitUp, AbnormalBodyMovement（统一，不分设备） |
| **设备技术** | `device` | 设备故障、网络异常等技术告警 | Flag | OfflineAlarm, LowBattery, DeviceFailure, AngleException |

### 2.2 Observation（观测数据）的 Category 分类

对于 Observation 类型的数据（实时数据、统计数据、事件数据），使用以下 category：

| FHIR Category | 用于数据流 | 说明 | 示例字段 |
|--------------|-----------|------|---------|
| `activity` | monitor, stat, event | 活动/行为观测（姿态、事件、睡眠状态） | pose, event, sleep_state |
| `vital-signs` | monitor, stat | 生命体征基础测量值（心率、呼吸率原始数值和状态） | breath_state, heart_state, stability |

### 2.3 数据流与 Category 的对应关系

| 数据流类型 | 资源类型 | Category 值 | 示例 |
|----------|---------|------------|------|
| **实时数据 (monitor)** | Observation | `activity`, `vital-signs` | pose (activity), breath_rate (vital-signs) |
| **统计数据 (stat)** | Observation | `activity`, `vital-signs`, `clinical` | sleep_state (activity), breath_state (vital-signs), apnea (clinical) |
| **事件数据 (event)** | Observation | `activity`, `safety` | enter_leave (activity), pose_change (safety) |
| **告警数据 (alarm)** | Flag | `safety`, `clinical`, `behavioral`, `device` | Fall (safety), Apnea (clinical) |

## 三、SNOMED 映射字段

每个需要 SNOMED 映射的字段，在编码后会添加以下字段：

- `{field_name}` - 原始值
- `{field_name}_snomed_code` - SNOMED 编码（如果有）
- `{field_name}_snomed_display` - SNOMED 显示名称
- `{field_name}_category` - Category 分类（`activity`, `vital-signs`, `safety`, `clinical`, `behavioral`, `device`）
- `{field_name}_display_en` - 英文显示名称

## 四、各数据流的 SNOMED 映射字段

### 4.1 实时数据 (monitor)

| 字段名 | Category | SNOMED 映射字段 |
|-------|---------|----------------|
| `pose` | `activity` | `pose_snomed_code`, `pose_snomed_display`, `pose_category`, `pose_display_en` |
| `event` | `activity` | `event_snomed_code`, `event_snomed_display`, `event_category`, `event_display_en` |
| `sleep_status` | `activity` | `sleep_status_snomed_code`, `sleep_status_snomed_display`, `sleep_status_category`, `sleep_status_display_en` |
| `stability` | `vital-signs` | `stability_snomed_code`, `stability_snomed_display`, `stability_category`, `stability_display_en` |

### 4.2 统计数据 (stat)

| 字段名 | Category | SNOMED 映射字段 |
|-------|---------|----------------|
| `breath_state` | `vital-signs` | `breath_state_snomed_code`, `breath_state_snomed_display`, `breath_state_category`, `breath_state_display_en` |
| `heart_state` | `vital-signs` | `heart_state_snomed_code`, `heart_state_snomed_display`, `heart_state_category`, `heart_state_display_en` |
| `vital_signs_state` | `vital-signs` | `vital_signs_state_snomed_code`, `vital_signs_state_snomed_display`, `vital_signs_state_category`, `vital_signs_state_display_en` |
| `sleep_state` | `activity` | `sleep_state_snomed_code`, `sleep_state_snomed_display`, `sleep_state_category`, `sleep_state_display_en` |

**注意**：`breath_state` 为 `"11"` (呼吸暂停) 时，category 为 `clinical`。

### 4.3 事件数据 (event)

| 字段名 | Category | SNOMED 映射字段 |
|-------|---------|----------------|
| `event` (type=1) | `activity` | `event_snomed_code`, `event_snomed_display`, `event_category`, `event_display_en` |
| `pose` (type=2) | `safety` | `pose_snomed_code`, `pose_snomed_display`, `pose_category`, `pose_display_en` |

### 4.4 告警数据 (alarm)

告警数据（alarm）作为 **Flag 类型**，按照 FHIR Flag 的 **4 大类**进行分类：

| 字段名 | Category | SNOMED 映射字段 | 说明 |
|-------|---------|----------------|------|
| `event` (type=1) | `activity` / `behavioral` | `event_snomed_code`, `event_snomed_display`, `event_category`, `event_display_en` | 进出事件告警（根据具体事件类型：进入床/离开床为 `behavioral`，其他为 `activity`） |
| `pose` (type=2) | `safety` / `activity` | `pose_snomed_code`, `pose_snomed_display`, `pose_category`, `pose_display_en` | 姿态变化告警（跌倒相关：pose=2, 5, 7, 8 为 `safety`；床上坐起：pose=9, 10, 11 为 `activity`） |

**重要说明**：告警数据的 category 根据具体告警类型确定：

1. **`safety`** - 临床安全告警：
   - 跌倒相关告警（pose=5, 8）：`Fall`
   - 疑似跌倒告警（pose=2, 7）：`SuspectedFall` / `At risk for falls`
   - 滞留告警：`Stay`
   - 24小时无人告警：`NoActivity24h`

2. **`clinical`** - 生命体征异常告警：
   - 呼吸暂停告警：`ApneaHypopnea`（统一，不分设备，从 breath_state="11" 触发）
   - 心率异常告警：`AbnormalHeartRate`（统一，不分设备）
   - 呼吸频率异常告警：`AbnormalRespiratoryRate`（统一，不分设备）
   - 生命体征微弱告警：`VitalsWeak`（统一，不分设备，从 vital_signs_state="11" 触发）

3. **`behavioral`** - 行为健康告警：
   - 离床告警：`LeftBed`（统一，不分设备，从 event=6 或 bedStatus=1 触发）
   - 床上坐起告警：`SitUp` / `BED_SIT_UP` / `pose=9,10,11`（统一 SNOMED: 225698008）
   - 异常体动告警：`AbnormalBodyMovement` / `SleepPad_AbnormalBodyMovement`

4. **`device`** - 设备技术告警：
   - 设备离线告警：`OfflineAlarm`
   - 低电量告警：`LowBattery`
   - 设备故障告警：`DeviceFailure`
   - 角度异常告警：`AngleException`

**注意**：
- Radar 设备直接上报的告警（通过 `/prefix/alarm/productId/UID/post`）目前只包含基础的 pose/event 数据
- 高级告警（如呼吸暂停、心率异常、离床等）通常由云端服务（如 `wisefido-alarm`、`wisefido-card-aggregator`）基于统计数据计算得出
- `radar_encoder.go` 中的 `encodeRadarAlarm()` 处理的是设备直接上报的告警，category 从 `radar_convert_table.json` 中获取
- 对于云端计算的告警，category 应在生成告警的服务中确定（如 `wisefido-alarm` 中的 `GetAlarmCategory()`）

## 五、Category 值的确定规则

### 5.1 Observation 类型（monitor, stat, event）

1. **`activity`**：用于行为观测数据
   - 姿态（pose）：行走、站立、坐位、卧位、床上坐起（pose=9,10,11）等
   - 事件（event）：进入房间、离开房间、进入区域、离开区域、进入床、离开床
   - 睡眠状态（sleep_state）：清醒、浅睡眠、深睡眠

2. **`vital-signs`**：用于生命体征基础测量值
   - 呼吸状态（breath_state）：正常、过低、过高（非暂停）
   - 心率状态（heart_state）：正常、过低、过高
   - 生命体征情况（vital_signs_state）：正常、弱
   - 稳定度（stability）：无干扰、小动作、大动作

3. **`clinical`**：用于生命体征异常告警（仅在 stat 中使用）
   - 呼吸暂停（breath_state="11"）：`67905006` (Apnea)

### 5.2 Flag 类型（alarm）

1. **`safety`**：临床安全告警
   - 跌倒（pose=5, 8, 11）
   - 疑似跌倒（pose=2, 7, 10）
   - 滞留告警
   - 24小时无人告警

2. **`clinical`**：生命体征异常告警（统一，不分设备）
   - 呼吸暂停告警（`ApneaHypopnea`）
   - 心率异常告警（`AbnormalHeartRate`）
   - 呼吸频率异常告警（`AbnormalRespiratoryRate`）
   - 生命体征微弱告警（`VitalsWeak`）

3. **`behavioral`**：行为健康告警（统一，不分设备）
   - 离床告警（`LeftBed`，统一 SNOMED: 248570008）
   - 床上坐起告警（`SitUp` / `BED_SIT_UP`，统一 SNOMED: 422256002）
   - 异常体动告警（`AbnormalBodyMovement`）

4. **`device`**：设备技术告警
   - 设备离线告警（OfflineAlarm）
   - 低电量告警（LowBattery）
   - 设备故障告警（DeviceFailure）
   - 角度异常告警（AngleException）

## 六、实现说明

### 6.1 4 类数据流的 SNOMED 映射实现

所有 4 类数据流的 SNOMED 映射已统一在 `radar_encoder.go` 中实现：

- `encodeRadarMonitor()` - 处理实时数据（monitor）的 SNOMED 映射
- `encodeRadarStat()` - 处理统计数据（stat）的 SNOMED 映射
- `encodeRadarEvent()` - 处理事件数据（event）的 SNOMED 映射
- `encodeRadarAlarm()` - 处理告警数据（alarm）的 SNOMED 映射

每个函数都会调用 `applySNOMedMapping()` 来添加 SNOMED 映射字段，包括：
- `{field_name}_snomed_code` - SNOMED 编码（如果有）
- `{field_name}_snomed_display` - SNOMED 显示名称（中文）
- `{field_name}_category` - Category 分类（`activity`, `vital-signs`, `safety`, `clinical`, `behavioral`, `device`）
- `{field_name}_display_en` - 英文显示名称

所有映射规则定义在 `radar_convert_table.json` 中，通过 `GetSNOMEDMappingByFieldPath()` 函数查询。

### 6.2 告警数据的 Category 确定流程

告警数据的 category 确定分为两个层次：

#### 6.2.1 设备直接上报的告警（`radar_encoder.go` 处理）

设备通过 `/prefix/alarm/productId/UID/post` 主题直接上报的告警：
- **处理位置**：`radar_encoder.go` 中的 `encodeRadarAlarm()`
- **Category 来源**：从 `radar_convert_table.json` 中获取（通过 `applySNOMedMapping()`）
- **示例**：
  - `pose=5, 8` (跌倒) → `safety`
  - `pose=9, 10, 11` (床上坐起) → `activity`
  - `pose=2, 7` (疑似跌倒) → `safety`
  - `pose=9, 10, 11` (床上坐起) → `activity`
  - `event=6` (离床) → `activity`（注：离床告警的 category 在云端计算时可能需要调整为 `behavioral`）

#### 6.2.2 云端计算的告警（`wisefido-alarm` 或 `wisefido-card-aggregator` 处理）

云端服务基于统计数据计算得出的告警：
- **处理位置**：`wisefido-alarm` 或 `wisefido-card-aggregator` 中的告警评估逻辑
- **Category 确定**：使用 `GetAlarmCategory(eventType)` 函数（定义在 `wisefido-card-aggregator/internal/alarm/alarm_handler.go`）
- **4 大类分类规则**：
  - **`safety`**：Fall, SuspectedFall, Stay, NoActivity24h
  - **`clinical`**：ApneaHypopnea, AbnormalHeartRate, AbnormalRespiratoryRate, VitalsWeak（统一，不分设备）
  - **`behavioral`**：LeftBed（统一 SNOMED: 248570008）, SitUp/BED_SIT_UP（统一 SNOMED: 225698008）, AbnormalBodyMovement, NoTurning.2H, NoBodyMovement.2H（统一，不分设备）
  - **`device`**：OfflineAlarm, LowBattery, DeviceFailure, AngleException

### 6.3 数据流与 Category 的对应关系总结

| 数据流类型 | 资源类型 | Category 值 | Category 来源 | 说明 |
|----------|---------|------------|--------------|------|
| **实时数据 (monitor)** | Observation | `activity`, `vital-signs` | `radar_convert_table.json` | 姿态、事件、睡眠状态（activity），呼吸心率状态（vital-signs） |
| **统计数据 (stat)** | Observation | `activity`, `vital-signs`, `clinical` | `radar_convert_table.json` | 睡眠状态（activity），呼吸心率状态（vital-signs），呼吸暂停（clinical） |
| **事件数据 (event)** | Observation | `activity`, `safety` | `radar_convert_table.json` | 进出事件（activity），姿态变化事件（safety） |
| **告警数据 (alarm)** | Flag | `safety`, `clinical`, `behavioral`, `device` | `radar_convert_table.json`（设备上报）或 `GetAlarmCategory()`（云端计算） | 设备直接上报的告警（safety），云端计算的告警（4 大类） |

### 6.4 统一 SNOMED 格式的好处

1. **不同厂家设备数据统一**：所有厂家的设备数据（Radar、Sleepace 等）都转换为统一的 SNOMED 格式
2. **字段命名统一**：使用 `display_en` 字段统一英文显示名称，避免大小写或命名不一致
3. **Category 分类统一**：按照 FHIR 标准统一分类，便于后续查询和分析
4. **便于扩展**：新增设备厂家时，只需添加对应的转换规则到 `radar_convert_table.json`
