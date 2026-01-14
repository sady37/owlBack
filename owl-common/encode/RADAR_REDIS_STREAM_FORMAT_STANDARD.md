# Radar Redis Stream 标准数据格式（待审查）

## 一、Redis Stream 名称

| topic_type | Redis Stream 名称 |
|-----------|------------------|
| `monitor` | `iot:monitor:stream` |
| `stat` | `iot:stat:stream` |
| `event` | `iot:event:stream` |
| `alarm` | `iot:alarm:stream` |

## 二、标准格式结构

所有数据都遵循以下标准格式（字段顺序：`device_id` → `device_type` → `tenant_id` → `timestamp` → `topic_type` → `data_value` → 位置信息）：

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "uuid-yyy",
  "timestamp": 1234567890,
  "topic_type": "monitor",
  "data_value": {
    "category": "track",
    // 按 category 分组的数据
  },
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- **必需字段**（按标准顺序）：
  - `device_id`: 系统内部 UUID（字符串），统一设备标识
  - `device_type`: 设备类型（字符串），如 "Radar"
  - `tenant_id`: 租户 ID（字符串），服务端添加
  - `timestamp`: 时间戳（整数），服务端添加
  - `topic_type`: 主题类型（字符串），值为 `'monitor'`, `'statistics'`, `'event'`, `'alarm'`
  - `data_value`: JSON 对象或数组，包含按 category 分组的数据（`category` 字段在 `data_value` 内部）
- **可选字段**（设备绑定信息，未绑定时为空或不存在，放在最后）：
  - `branch_id`: 分支 ID（字符串），通过 unit 关联
  - `building_id`: 建筑 ID（字符串），通过 unit 关联
  - `unit_id`: 单元 ID（字符串），设备绑定到 unit 时
  - `room_id`: 房间 ID（字符串），设备绑定到 room 时
  - `bed_id`: 床位 ID（字符串），设备绑定到 bed 时

**字段顺序说明**：
字段按照以下顺序排列：`device_id` → `device_type` → `tenant_id` → `timestamp` → `topic_type` → `data_value` → 位置信息。此顺序符合访问频率和逻辑分组，便于查询和维护。

**关于 `category` 字段的说明**：
- **`category` 字段位置**：`category` 字段保留在 `data_value` 内部，不提取到顶层，避免数据冗余
- **设计原则**：遵循数据不冗余原则，`category` 作为数据内容的一部分，保留在 `data_value` 内部，保持数据结构的完整性和一致性

## 三、Monitor 数据格式（实时数据）

### 3.1 标准格式

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "monitor",
  "data_value": {
    "category": "track",
    "target_id": 1,
    "position_x": 150,
    "position_y": 200,
    "position_z": 50,
    "remaining_time": 30,
    "pose": "Walking",
    "event": 0,
    "area_id": 1,
    "raw_original": "base64_string"  // 可选，原始 track base64 字符串（16 字节 * N，N 为人数）
  },
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

### 3.2 多人情况（多个 track 对象）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "monitor",
  "category": "track",
  "data_value": [
    {
      "category": "track",
      "target_id": 1,
      "position_x": 150,
      "position_y": 200,
      "position_z": 50,
      "remaining_time": 30,
      "pose": "Walking",
      "event": 0,
      "area_id": 1
    },
    {
      "category": "track",
      "target_id": 2,
      "position_x": 200,
      "position_y": 250,
      "position_z": 60,
      "remaining_time": 0,
      "pose": "Standing position",
      "event": 0,
      "area_id": 1,
      "raw_original": "base64_string"  // 可选，原始 track base64 字符串（16 字节 * N，N 为人数）
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

### 3.3 包含 vital 数据（呼吸心率）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "branch_id": "uuid-bbb",
  "building_id": "uuid-aaa",
  "unit_id": "uuid-zzz",
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb",
  "topic_type": "monitor",
  "timestamp": 1234567890,
  "data_value": [
    {
      "category": "track",
      "target_id": 1,
      "position_x": 150,
      "position_y": 200,
      "position_z": 50,
      "remaining_time": 30,
      "pose": "Walking",
      "event": 0,
      "area_id": 1,
      "raw_original": "base64_string"  // 可选，原始 track base64 字符串（16 字节 * N，N 为人数）
    },
    {
      "category": "vital",
      “vital_flag”: 0,
      "respiratory_rate": 14,
      "heart_rate": 72,
      "sleep_status": "Sleep state undefined",
      "stability": "11",
      "raw_original": "base64_string"  // 可选，原始 bh base64 字符串
    }
  ]
}
```

### 3.4 字段说明

#### track category 字段
| 字段名 | 类型 | 说明 | 来源 |
|-------|------|------|------|
| `category` | string | 固定为 "track" | - |
| `target_id` | int | 目标 ID：有人状态时取值 0-7，代表人员编号；无人状态时固定为 88，表示无人 | track 字节 0 |
| `position_x` | int | X 坐标（厘米），原始单位为分米，取值范围-127~127，已转换为厘米 | track 字节 1，dm→cm 转换 |
| `position_y` | int | Y 坐标（厘米），原始单位为分米，取值范围-127~127，已转换为厘米 | track 字节 2，dm→cm 转换 |
| `position_z` | int | Z 坐标（厘米），原始单位就是厘米，取值范围 0-255 | track 字节 3 |
| `remaining_time` | int | 剩余时间（秒，0-60，仅在自动测量边界时使用） | track 字节 12 |
| `pose` | string | 姿态（display_en 值）。映射：0→"Initialization"，1→"Walking"，2→"FallSuspected"，3→"Sitting position"，4→"Standing position"，5→"Fall"，6→"Lying position"，7→"SitGroundSuspected"，8→"SitGround"，9→"Sitting up in bed"，10→"Sitting up in bed:Suspected"，11→"Sitting up in bed" | track 字节 13，SNOMED 映射 |
| `event` | int/string | 事件（display_en 值）。映射：0=无事件="No event"，1=进入房间="Enter room"，2=离开房间="Leave room"，3=进入区域="Enter area"，4=离开区域="Leave area" | track 字节 14 |
| `area_id` | int | 区域 ID。在字节 14（event）内容为 3 或 4 时，此字段标识人员进出的区域 ID | track 字节 15 |
| `raw_original` | string | 可选，原始 track base64 字符串（16 字节 * N，N 为人数） | 原始数据 |

#### vital category 字段
| 字段名 | 类型 | 说明 | 来源 |
|-------|------|------|------|
| `category` | string | 固定为 "vital" | - |
| `vital_flag` | int | 标识符，固定为 0，表示实时呼吸心率 | bh 字节 0 |
| `respiratory_rate` | int | 呼吸率（次/分钟） | bh 字节 1 |
| `heart_rate` | int | 心率（次/分钟bpm） | bh 字节 2 |
| `sleep_status` | string | 睡眠状态（display_en 值），bit 7 & bit 6 代表过去 1 分钟的睡眠状态。映射：00=未定义="Sleep state undefined"，01=浅睡="Light sleep"，10=深睡="Deep sleep"，11=清醒="Awake" | bh 字节 13 bit 7:6，SNOMED 映射 |
| `stability` | string | 稳定度（display_en 值），bit 1 & bit 0 表示过去 10 秒钟是否有动作干扰测量。映射：00=未定义="Stability undefined"，01=有较大动作="Large movement"，10=有较小动作="Small movement"，11=无干扰="No interference" | bh 字节 14 bit 1:0 |
| `raw_original` | string | 可选，原始 bh base64 字符串 | 原始数据 |



## 四、Stat 数据格式（统计数据）

### 4.1 标准格式

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "statistics",
  "data_value": {
    "category": "sleep",
    "sleep_flag": 255,
    "respiratory_rate": 14,
    "heart_rate": 72,
    "avg_respiratory_rate": 15,
    "avg_heart_rate": 70,
    "breath_state": "Breath rate normal",
    "heart_state": "Heart rate normal",
    "vital_signs_state": "Vital signs normal",
    "sleep_state": "Light sleep",
    "raw_original": "base64_string"  // 可选，原始 sleep base64 字符串
  },
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

### 4.2 包含 track 统计

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "statistics",
  "data_value": [
    {
      "category": "track",
      "version": 2,
      "people_count": 1,
      "walk_distance": 50,
      "walk_duration": 300,
      "lie_duration": 1800,
      "stand_duration": 120,
      "multi_person_duration": 0,
      "raw_original": "base64_string"  // 可选，原始 track base64 字符串（统计数据的 track 部分）
    },
    {
      "category": "sleep",
      "sleep_flag": 255,
      "respiratory_rate": 14,
      "heart_rate": 72,
      "avg_respiratory_rate": 15,
      "avg_heart_rate": 70,
      "breath_state": "Breath rate normal",
      "heart_state": "Heart rate normal",
      "vital_signs_state": "Vital signs normal",
      "sleep_state": "Light sleep",
      "raw_original": "base64_string"  // 可选，原始 sleep base64 字符串
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

### 4.3 字段说明

#### track category 字段（统计）
| 字段名 | 类型 | 说明 | 来源 |
|-------|------|------|------|
| `category` | string | 固定为 "track" | - |
| `version` | int | 版本标识符，当前取值 1 或 2 | track 字节 0 |
| `people_count` | int | 人数，当前上报时刻雷达检测到的人数，取值 0-8 | track 字节 1 |
| `walk_distance` | int | 行走距离（米），过去一分钟统计的行走距离，原始单位为米，取值范围 0-32767，采用 Big-Endian 字节序，不转换为厘米。注意：在多人情况下不做统计 | track 字节 2-3，m不转换 |
| `walk_duration` | int | 行走时长（秒），过去一分钟统计的行走时间。注意：在多人情况下不做统计 | track 字节 4 |
| `lie_duration` | int | 躺卧时长（秒），过去一分钟统计的躺卧时长 | track 字节 6 |
| `stand_duration` | int | 站立时长（秒），过去一分钟统计的站立时长。注意：在多人情况下不做统计 | track 字节 7 |
| `multi_person_duration` | int | 多人时长（秒），过去一分钟统计的有多人存在的时间。这段时间将不做行走距离、行走时长、静止时长等统计 | track 字节 8 |
| - | - | **备注**：字节 5（静坐时长）、9-15（rsv）未开放使用，**不解码** | track 字节 5、9-15 |
| `raw_original` | string | 可选，原始 track base64 字符串（统计数据的 track 部分） | 原始数据 |

#### sleep category 字段
| 字段名 | 类型 | 说明 | 来源 |
|-------|------|------|------|
| `category` | string | 固定为 "sleep" | - |
| `sleep_flag` | int | 标识符，固定为 255 (0xff)，表示为睡眠统计 | sleep 字节 0 |
| `respiratory_rate` | int | 实时呼吸率（rpm） | sleep 字节 1 |
| `heart_rate` | int | 实时心率（bpm） | sleep 字节 2 |
| `avg_respiratory_rate` | int | 分钟级平均呼吸率（rpm） | sleep 字节 5 |
| `avg_heart_rate` | int | 分钟级平均心率（bpm） | sleep 字节 6 |
| `breath_state` | string | 呼吸状态（display_en 值）。映射：00=正常="Breath rate normal"，01=偏低="Breath rate low"，10=偏高="Breath rate high"，11=呼吸暂停="Apnea" | sleep 字节 13 bit 1:0，SNOMED 映射 |
| `heart_state` | string | 心率状态（display_en 值）。映射：00=正常="Heart rate normal"，01=偏低="Heart rate low"，10=偏高="Heart rate high"，11=未定义="Heart rate undefined" | sleep 字节 13 bit 3:2，SNOMED 映射 |
| `vital_signs_state` | string | 生命体征状态（display_en 值）。映射：00=正常="Vital signs normal"，01=未定义="Vital signs undefined"，10=未定义="Vital signs undefined"，11=微弱="Vital signs weak" | sleep 字节 13 bit 5:4，SNOMED 映射 |
| `sleep_state` | string | 睡眠状态（display_en 值）。映射：00=未定义="Sleep state undefined"，01=浅睡="Light sleep"，10=深睡="Deep sleep"，11=清醒="Awake" | sleep 字节 13 bit 7:6，SNOMED 映射 |
| - | - | **备注**：14-15 为睡眠统计信息，睡眠算法未公开，本字段信息不公开，**不解码** ||
| `raw_original` | string | 可选，原始 sleep base64 字符串 | 原始数据 |

### 4.4 其他字段（保留在 data_value 顶层或单独对象）

```json
{
  "data_value": {
    "sleep": "base64_string",  // 原始 sleep base64 字符串（保留）
    "track": "base64_string"   // 原始 track base64 字符串（如果有）
  }
}
```

## 五、Event 数据格式（事件数据 - Redis Stream 标准格式）

**说明**：Event 数据使用 `cmd: "event"` 格式，包含4种事件类型（type=1/2/3/4）。对接平台仅接收 event 的事件，对于 log 的事件可选择忽略。

### 5.1 标准格式结构

所有 Event 数据都遵循以下标准格式：

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "data_value": [
    {
      "category": "enter2out",  // event_type: 1=Enter2out, 2=pose, 3=number-people
      "track_id": 1,
      "event": "Enter room",
      "area_type": "Bed"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- **顶层元数据字段**（按标准顺序）：
  - `device_id`, `device_type`, `tenant_id`, `timestamp`, `topic_type`: 元数据字段
  - `data_value`: 事件内容（包含 `category` 字段，根据 category 不同而不同）
- **位置信息字段**（可选，放在最后）：
  - `branch_id`, `building_id`, `unit_id`, `room_id`, `bed_id`: 设备绑定信息

### 5.2 type=1（进出事件）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "data_value": [
    {
      "category": "enter2out",  // event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other
      "track_id": 1,
      "event": "Enter room",    // display_en 值（如果有 SNOMED 映射则使用 SNOMED display，否则使用 display_en）
      "area_type": "Bed"        // display_en 值（area_type 为标识符，无 SNOMED 映射）
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含进出事件信息
- `category`: 固定为 "enter2out"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `track_id`: 人员轨迹 ID（数值）
- `event`: 事件类型（字符串，display_en 值）
  - "Enter room" (SNOMED display: "Enter room", display_en: "Enter room") - 对应原始值 1
  - "Leave room" (SNOMED display: "Leave room", display_en: "Leave room") - 对应原始值 2
  - "Enter area" (SNOMED display: "Enter area", display_en: "Enter area") - 对应原始值 3
  - "Leave area" (SNOMED display: "Leave area", display_en: "Leave area") - 对应原始值 4
  - "Enter bed" (SNOMED code: 248569007, SNOMED display: "In bed", display_en: "Enter bed") - 对应原始值 5
  - "Left bed" (SNOMED code: 248570008, SNOMED display: "Not in bed", display_en: "Left bed") - 对应原始值 6
- `area_type`: 区域类型（字符串，display_en 值）
  - "Bed" (display_en: "Bed") - 对应原始值 2，进入离开此区域会产生进床离床事件
  - "Monitoring bed" (display_en: "Monitoring bed") - 对应原始值 5
  - "Sensor area" (display_en: "Sensor area") - 对应原始值 6

### 5.3 type=2（姿态变化事件）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "pose",
  "data_value": [
    {
      "category": "pose",  // event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other
      "track_id": 1,
      "pose": "Fall"       // display_en 值（如果有 SNOMED 映射则使用 SNOMED display，否则使用 display_en）
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含姿态变化事件信息
- `category`: 固定为 "pose"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `track_id`: 人员轨迹 ID（数值）
- `pose`: 姿态类型（字符串，display_en 值）
  - "Walking" (SNOMED code: 129006008, SNOMED display: "Walking", display_en: "Walking") - 对应原始值 1（行走）
  - "FallSuspected" (SNOMED code: 129839007, SNOMED display: "At risk for falls", display_en: "FallSuspected") - 对应原始值 2（疑似跌倒）
  - "Fall" (SNOMED code: 161898004, SNOMED display: "Fall", display_en: "Fall") - 对应原始值 5（确认跌倒）
  - "SitGroundSuspected" (SNOMED code: 129839007, SNOMED display: "At risk for falls", display_en: "At risk for falls") - 对应原始值 7（疑似坐地）
  - "SitGround" (SNOMED code: 161898004, SNOMED display: "Fall", display_en: "Fall") - 对应原始值 8（确认坐地）
  - "Sitting up in bed:Suspected" (SNOMED code: 225698008, SNOMED display: "Sitting up in bed", display_en: "Sitting up in bed:Suspected") - 对应原始值 10（疑似床上坐起）
  - "Sitting up in bed" (SNOMED code: 225698008, SNOMED display: "Sitting up in bed", display_en: "Sitting up in bed") - 对应原始值 11（确认床上坐起）
  - 其他 = 其他姿态 (根据具体值映射，若有 SNOMED 映射则使用 SNOMED display，否则使用 display_en)


### 5.4 type=3（人数变化事件）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "data_value": [
    {
      "category": "number-people",  // event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other
      "number_people": 1            // 人数（数值）
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含人数变化事件信息
- `category`: 固定为 "number-people"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `number_people`: 人数（数值）


### 5.5 type=5（设备在线状态）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "data_value": [
    {
      "category": "isOnline",  // event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other
      "device_status": "offline",  // "0=online"  "1=offline"
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含设备在线状态事件信息
- `category`: 固定为 "isOnline"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `device_status`: 设备在线状态（字符串）
  - "online" = 在线（对应原始值 0）
  - "offline" = 离线（对应原始值 1）
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 5.6 type=7（信号差事件）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "data_value": [
    {
      "category": "signal_poor",  // event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other
      "recovery": "signal_recovery",  // 0="signal_poor" 1="signal_recovery"
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含信号差事件信息
- `category`: 固定为 "signal_poor"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `recovery`: 恢复状态（字符串）
  - "signal_poor" = 信号差（对应原始值 0，触发）
  - "signal_recovery" = 信号恢复（对应原始值 1，恢复）
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 5.7 type=8（倾角异常事件）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "data_value": [
    {
      "category": "angle_abnormal",  // event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other
      "recovery": "angle_recovery",  // 0="angle_abnormal"=倾角异常 "1=angle_recovery"
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含倾角异常事件信息
- `category`: 固定为 "angle_abnormal"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `recovery`: 恢复状态（字符串）
  - "angle_abnormal" = 倾角异常（对应原始值 0，触发）
  - "angle_recovery" = 倾角恢复（对应原始值 1，恢复）
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 5.8 type=9（其他告警）

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "data_value": [
    {
      "category": "other",  // event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other
      "alarmType": "2",     // 告警类型（字符串）
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含其他告警事件信息
- `category`: 固定为 "other"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `alarmType`: 告警类型（字符串）
  - "1" = 离床未归（display_en: "Left bed not returned"）
  - "2" = 滞留（display_en: "Stay"）
  - "3" = 长时间无人活动（display_en: "No activity for long time"）
  - "4" = 坐地告警（display_en: "Sitting on ground alarm"）
  - "5" = 床上坐起告警（display_en: "Sitting up in bed alarm"）
  - "6" = 空（未开放使用）
  - "7" = 空（未开放使用）
  - "8" = 按钮SOS告警（display_en: "Button SOS alarm"，未开放使用）
  - "9" = 拉绳SOS告警（display_en: "Rope SOS alarm"，未开放使用）
  - "10" = 呼救SOS告警（display_en: "Call for help SOS alarm"，未开放使用）
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 5.9 事件类型汇总

| type | 事件类型 | category | data_value 格式 | 说明 |
|------|---------|----------|----------------|------|
| 1 | 进出事件 | "enter2out" | 数组 | 包含 `track_id`, `event`, `area_type` |
| 2 | 姿态变化事件 | "pose" | 数组 | 包含 `track_id`, `pose` |
| 3 | 人数变化事件 | "number-people" | 数组 | 包含 `number_people` |
| 5 | 设备在线状态 | "isOnline" | 数组 | 包含 `device_status`, `device_id` |
| 7 | 信号差事件 | "signal_poor" | 数组 | 包含 `recovery`, `device_id` |
| 8 | 倾角异常事件 | "angle_abnormal" | 数组 | 包含 `recovery`, `device_id` |
| 9 | 其他告警 | "other" | 数组 | 包含 `alarmType`, `device_id` |

## 六、Alarm 数据格式（告警数据）

**说明**：告警数据从事件（Event）和统计（Stat）数据中抽取，当满足告警条件时生成告警事件。

### 6.1 跌倒告警（从 type=2 姿态变化事件中抽取）

当 `pose` 为 "Fall" 或 "FallSuspected" 时，生成跌倒告警。

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "alarm",
  "data_value": [
    {
      "category": "fall_alarm",
      "pose": "Fall",  // "Fall"=确认跌倒 "FallSuspected"=疑似跌倒
      "track_id": 1,
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含跌倒告警信息
- `category`: 固定为 "fall_alarm"
- `pose`: 姿态类型（字符串）
  - "Fall" = 确认跌倒（对应原始值 5）
  - "FallSuspected" = 疑似跌倒（对应原始值 2）
- `track_id`: 人员轨迹 ID（数值）
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 6.2 设备掉线告警（从 type=5 设备在线状态中抽取）

当 `device_status` 为 "offline" 时，生成设备掉线告警。

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "alarm",
  "data_value": [
    {
      "category": "device_offline",
      "device_status": "offline",  // "offline"=离线
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含设备掉线告警信息
- `category`: 固定为 "device_offline"
- `device_status`: 设备状态（字符串），固定为 "offline"
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 6.3 信号差告警（从 type=7 信号差事件中抽取）

当 `recovery` 为 "signal_poor" 时，生成信号差告警。

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "alarm",
  "data_value": [
    {
      "category": "signal_poor_alarm",
      "recovery": "signal_poor",  // "signal_poor"=信号差
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含信号差告警信息
- `category`: 固定为 "signal_poor_alarm"
- `recovery`: 恢复状态（字符串），"signal_poor" 表示信号差（对应原始值 0，触发）
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 6.4 倾角异常告警（从 type=8 倾角异常事件中抽取）

当 `recovery` 为 "angle_abnormal" 时，生成倾角异常告警。

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "alarm",
  "data_value": [
    {
      "category": "angle_abnormal_alarm",
      "recovery": "angle_abnormal",  // "angle_abnormal"=倾角异常
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含倾角异常告警信息
- `category`: 固定为 "angle_abnormal_alarm"
- `recovery`: 恢复状态（字符串），"angle_abnormal" 表示倾角异常（对应原始值 0，触发）
- `device_id`: 设备 ID（字符串，系统内部 UUID）

### 6.5 生命体征微弱告警（从 Stat 统计数据中抽取）

当 `vital_signs_state` 为 "Vital signs weak" 时，生成生命体征微弱告警。

```json
{
  "device_id": "uuid-xxx",
  "device_type": "Radar",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "alarm",
  "data_value": [
    {
      "category": "vital_signs_weak",
      "vital_signs_state": "Vital signs weak",  // 对应原始值 11（bit 5:4 = 11）
      "device_id": "uuid-xxx"
    }
  ],
  "branch_id": null,
  "building_id": null,
  "unit_id": null,
  "room_id": "uuid-rrr",
  "bed_id": "uuid-bbb"
}
```

**字段说明**：
- `data_value`: 数组，包含生命体征微弱告警信息
- `category`: 固定为 "vital_signs_weak"
- `vital_signs_state`: 生命体征状态（字符串），固定为 "Vital signs weak"（对应原始值 11，bit 5:4 = 11）
- `device_id`: 设备 ID（字符串，系统内部 UUID）


## 七、字段映射规则

### 7.1 SNOMED 映射字段
对于需要 SNOMED 映射的字段（pose, event, sleep_status, stability, breath_state, heart_state, vital_signs_state, sleep_state）：
- **字段名**：保持不变
- **字段值**：使用 `display_en`（如果有映射且不为空），否则保留原始值
- **不添加**：`_snomed_code`, `_snomed_display`, `_category`, `_display_en` 等额外字段

### 7.2 数值字段
对于数值字段（respiratory_rate, heart_rate 等）：
- 使用标准字段名
- 直接使用数值，不添加 SNOMED 相关字段

### 7.3 原始 base64 字符串
- `track`, `bh`, `sleep` 等原始 base64 字符串会保留在 `data_value` 中

## 八、实现说明

### 8.1 topic_type 映射
- `topic_type = "monitor"` → `topic_type = "monitor"`
- `topic_type = "stat"` → `topic_type = "statistics"`
- `topic_type = "event"` → `topic_type = "event"`
- `topic_type = "alarm"` → `topic_type = "alarm"`

### 8.2 data_value 结构
- **Monitor**: 
  - 单人：`data_value` 是对象，包含 track 和 vital 数据（如果都有，则用数组）
  - 多人：`data_value` 是数组，每个元素是一个 track 对象，vital 单独一个对象
- **Stat**: 
  - `data_value` 是对象或数组（如果同时有 track 和 sleep 统计，则用数组）
- **Event/Alarm**: 
  - `data_value` 保持原始结构（包含 type 和 data 字段）

### 8.3 category 分组
- **"track"**: 轨迹相关数据（position_x, position_y, position_z, target_id, area_id, remaining_time, pose, event）
- **"vital"**: 生命体征相关数据（respiratory_rate, heart_rate, sleep_status, stability）
- **"sleep"**: 睡眠统计相关数据（sleep_flag, respiratory_rate, heart_rate, avg_respiratory_rate, avg_heart_rate, breath_state, heart_state, vital_signs_state, sleep_state）
- **"enter2out"**: 进出事件（track_id, event, area_type）
- **"pose"**: 姿态变化事件（track_id, pose）
- **"number-people"**: 人数变化事件（number_people）
- **"isOnline"**: 设备在线状态（device_status, device_id）
- **"signal_poor"**: 信号差事件（recovery, device_id）
- **"angle_abnormal"**: 倾角异常事件（recovery, device_id）
- **"other"**: 其他告警（alarmType, device_id）
- **告警 category**: "fall_alarm", "device_offline", "signal_poor_alarm", "angle_abnormal_alarm", "vital_signs_weak" 等

## 九、字段顺序调整说明

### 9.1 调整原因

文档中所有 JSON 示例的字段顺序已统一调整为：`device_id` → `device_type` → `tenant_id` → `timestamp` → `topic_type` → `data_value` → 位置信息。

**调整原因**：

1. **符合实际代码实现**：
   - wisefido-radar 的 `mqtt_consumer.go` 中已按照此顺序构建 `encodedData`
   - `category` 字段已从 `data_value` 中提取到顶层，便于查询和过滤
   - 确保文档与实际代码实现保持一致，避免混淆

2. **逻辑分组清晰**：
   - **设备标识**：`device_id`, `device_type`（设备基本信息）
   - **租户信息**：`tenant_id`（多租户隔离）
   - **时间信息**：`timestamp`（数据时间戳）
   - **数据类型**：`topic_type`（数据分类）
   - **数据内容**：`data_value`（实际数据，包含 `category` 字段）
   - **位置信息**：`branch_id`, `building_id`, `unit_id`, `room_id`, `bed_id`（可选位置信息）

3. **查询优化**：
   - 将最常用的查询字段（`device_id`, `tenant_id`, `timestamp`, `topic_type`, `category`）放在前面
   - `category` 提取到顶层后，放在 `topic_type` 之后，便于按类别过滤
   - 位置信息放在最后，因为通常通过关联表查询，较少直接使用

4. **维护一致性**：
   - 与 `STREAM_DB_FORMAT_DIFF.md` 保持一致
   - 与 `18_iot_timeseries.sql` 中的注释保持一致
   - 确保整个系统的文档和代码都遵循相同的字段顺序规范

### 9.2 调整内容

本次调整更新了以下部分：
- ✅ 二、标准格式结构：更新字段顺序说明，`category` 保留在 `data_value` 内部
- ✅ 三、Monitor 数据格式：更新所有示例的字段顺序，去掉顶层 `category`（3.1, 3.2, 3.3）
- ✅ 四、Stat 数据格式：更新所有示例的字段顺序，去掉顶层 `category`（4.1, 4.2）
- ✅ 五、Event 数据格式：更新所有示例的字段顺序，去掉顶层 `category`（5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8）
- ✅ 六、Alarm 数据格式：更新所有示例的字段顺序，去掉顶层 `category`（6.1, 6.2, 6.3, 6.4, 6.5）
- ✅ 八、实现说明：更新 category 分组说明

### 9.3 影响范围

- **代码层面**：无需修改，代码已按新顺序实现
- **数据库层面**：无需修改，JSONB 字段顺序不影响查询
- **文档层面**：已统一更新，确保文档准确性
- **查询层面**：字段顺序不影响 JSONB 查询性能，但遵循统一顺序有助于代码可读性和维护性
