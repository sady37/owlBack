# Radar Redis Stream 标准数据格式（审查）

## 一、Redis Stream 名称

| topic_type | Redis Stream 名称 |
|-----------|------------------|
| `monitor` | `iot:monitor:stream` |
| `stat` | `iot:stat:stream` |
| `event` | `iot:event:stream` |
| `alarm` | `iot:alarm:stream` |
| `auth` | ` iot:auth:stream` |

## 二、标准格式结构

  所有数据都遵循以下标准格式（顶层字段顺序：`device_id` → `device_type` → `card_id` → `tenant_id` → `timestamp` → `topic_type` → `category` → `data_value`）。**位置信息（addressInfo）不再出现在流中**，需要时由消费方按 `card_id` 查询 Card 获取。

  ```json
  {
    "device_id": "uuid-ddd",
    "device_type": "Radar",
    "card_id": "uuid-ccc",
    "tenant_id": "uuid-yyy",
    "timestamp": 1234567890,
    "topic_type": "monitor",
    "category": "track2.vital1",
    "data_value": [
      {
        "category": "track",
        "track_id": 1,
        "position_x": 150,
        "position_y": 200,
        "position_z": 50,
        "remaining_time": 30,
        "pose": "Walking",
        "event": 0,
        "area_id": 1,
        "raw_original": "base64_string"
      },
      {
        "category": "track",
        "track_id": 2,
        "position_x": 100,
        "position_y": 180,
        "position_z": 50,
        "remaining_time": 0,
        "pose": "Lying position",
        "event": 0,
        "area_id": 1
      },
      {
        "category": "vital",
        "vital_flag": 0,
        "respiratory_rate": 14,
        "heart_rate": 72,
        "sleep_status": "Sleep state undefined",
        "stability": "11",
        "raw_original": "base64_string"
      }
    ]
  }
  ```

**顶层字段说明**（按标准顺序）：
- `device_id`: 设备 ID（UUID 格式），系统内部唯一标识符
- `device_type`: 设备类型（字符串），如 "Radar"、"Sleepace"
- `card_id`: 卡片 ID（字符串），未绑卡可空或省略
- `tenant_id`: 租户 ID（字符串），服务端添加
- `timestamp`: 时间戳（整数），服务端添加
- `topic_type`: 主题类型（字符串），值为 `monitor`、`stat`、`event`、`alarm`、`auth`
- `category`: 数据类别（字符串），多类用 `.` 拼接，如 "track2.vital1" 表示 2 条 track、1 条 vital
- `data_value`: **JSON 数组**，每项为一条数据，项内含 `category` 字段；单条时也为单元素数组

**关于 `data_value` 与 `category`**：
- `data_value` 统一为**数组**，每项为一条观测/事件，项内必含 `category` 字段，与顶层 `category` 的 N 对应（如 track2 即数组中有 2 个 category 为 "track" 的项）。
- 顶层 `category` 便于订阅者判读与过滤，无需解析数组即可知本条消息包含的类别及条数。

## 三、Monitor 数据格式（实时数据）

### 3.1 标准格式（data_value 为数组）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "monitor",
  "category": "track2.vital1",
  "data_value": [
    {
      "category": "track",
      "track_id": 1,
      "position_x": 150,
      "position_y": 200,
      "position_z": 50,
      "remaining_time": 30,
      "pose": "Walking",
      "event": 0,
      "area_id": 1,
      "raw_original": "base64_string"
    },
    {
      "category": "track",
      "track_id": 2,
      "position_x": 150,
      "position_y": 200,
      "position_z": 50,
      "remaining_time": 30,
      "pose": "Walking",
      "event": 0,
      "area_id": 1,
      "raw_original": "base64_string"
    },
    {
      "category": "vital",
      "vital_flag": 0,
      "respiratory_rate": 14,
      "heart_rate": 72,
      "sleep_status": "Sleep state undefined",
      "stability": "11",
      "raw_original": "base64_string"
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
| `pose` | string | 姿态（display_en 值）。映射：0→"Initialization"，1→"Walking"，2→"SuspectedFall"，3→"Sitting position"，4→"Standing position"，5→"Fall"，6→"Lying position"，7→"SuspectedSittingOnGround"，8→"SittingOnGround"，9→"Sitting up in bed"，10→"SuspectedBedSitUp"，11→"Sitting up in bed" | track 字节 13，SNOMED 映射 |
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
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "stat",
  "category": "sleep",
  "data_value": [
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
      "raw_original": "base64_string"
    }
  ]
}
```

### 4.2 包含 track 统计

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "stat",
  "category": "track",
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
      "raw_original": "base64_string"
    }
  ]
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

### 4.4 其他字段（data_value 为数组，原始 base64 放在项内 raw_original）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "stat",
  "category": "sleep.track",
  "data_value": [
    { "category": "sleep", "raw_original": "base64_string" },
    { "category": "track", "raw_original": "base64_string" }
  ]
}
```

## 五、Event 数据格式（事件数据 - Redis Stream 标准格式）

**说明**：Event 数据使用 `cmd: "event"` 格式，包含4种事件类型（type=1/2/3/5/7/8/9）。对接平台仅接收 event 的事件，对于 log 的事件可选择忽略。

### 5.1 标准格式结构

所有 Event 数据都遵循以下标准格式（顶层无 addressInfo，位置信息需时按 card_id 查 Card）：

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "enter2out",
  "data_value": [
    {
      "category": "enter2out",
      "track_id": 1,
      "event": "Enter room",
      "area_type": "Bed"
    }
  ]
}
```

**字段说明**：
- **顶层字段**：`device_id`, `device_type`, `card_id`, `tenant_id`, `timestamp`, `topic_type`, `category`, `data_value`（见二、标准格式结构）
- `data_value`: 数组，每项包含 `category` 及该类别对应字段

### 5.2 type=1（进出事件）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "enter2out",
  "data_value": [
    {
      "category": "enter2out",
      "track_id": 1,
      "event": "Enter room",
      "area_type": "Bed"
    }
  ]
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
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "pose",
  "data_value": [
    {
      "category": "pose",
      "track_id": 1,
      "pose": "Fall"
    }
  ]
}
```

**字段说明**：
- `data_value`: 数组，包含姿态变化事件信息
- `category`: 固定为 "pose"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `track_id`: 人员轨迹 ID（数值）
- `pose`: 姿态类型（字符串，display_en 值）
  - "Walking" (SNOMED code: 129006008, SNOMED display: "Walking", display_en: "Walking") - 对应原始值 1（行走）
  - "SuspectedFall" (SNOMED code: 129839007, SNOMED display: SuspectedFall", display_en: "SuspectedFall") - 对应原始值 2（疑似跌倒）
  - "Fall" (SNOMED code: 161898004, SNOMED display: "Fall", display_en: "Fall") - 对应原始值 5（确认跌倒）
  - "SuspectedSittingOnGround" (SNOMED code: 129839007, SNOMED display: "SuspectedSittingOnGround", display_en: "SuspectedSittingOnGround") - 对应原始值 7（疑似坐地）
  - "SittingOnGround" (SNOMED code: 161898004, SNOMED display: "SittingOnGround", display_en: "SittingOnGround") - 对应原始值 8（确认坐地）
  - "SuspectedBedSitUp" (SNOMED code: 225698008, SNOMED display: "SuspectedBedSitUp", display_en: "SuspectedBedSitUp") - 对应原始值 10（疑似床上坐起）
  - "BedSitUp" (SNOMED code: 225698008, SNOMED display: "SuspectedBedSitUp", display_en: "BedSitUp") - 对应原始值 11（确认床上坐起）
  - 其他 = 其他姿态 (根据具体值映射，若有 SNOMED 映射则使用 SNOMED display，否则使用 display_en)

### 5.4 type=3（人数变化事件）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "number-people",
  "data_value": [
    {
      "category": "number-people",
      "number_people": 1
    }
  ]
}
```

**字段说明**：
- `data_value`: 数组，包含人数变化事件信息
- `category`: 固定为 "number-people"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `number_people`: 人数（数值）

### 5.5 type=5（设备在线状态）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "deviceStatus",   //StatusFieldOnline 。。。预定义的const
  "data_value": [
    {
      "category": "online",    // //StatusFieldtype 。。。预定义的const
      "StatusFieldValue": "1",       //StatusFieldValue   string ,防止有一些其它字符串
      "device_uid": "E598A2ACD523"
    }
  ]
}
```


```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "deviceStatus",
  "data_value": [
    {
      "category": "deviceStatus",
      "online": 1,
      "device_uid": "E598A2ACD523"
    }
  ]
}
```

**字段说明**：
- `data_value`: 数组，包含设备在线状态事件信息
- `category`: 固定为 "deviceStatus"
- `device_status`: 设备在线状态（字符串），"online 1" / "offline 0"
- `device_uid`: 需要时放在 data_value 项内（字符串）

### 5.6 type=7（信号差事件）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "deviceStatus",
  "data_value": [
    {
      "category": "deviceStatus",
      "signal_pool": "0",
      "device_uid": "E598A2ACD523"
    }
  ]
}
```

**字段说明**：
- `data_value`: 数组，包含信号差事件信息
- `category`: 固定为 "signal_poor"（event_type: 1=Enter2out, 2=pose, 3=number-people, 5=isOnline, 7=signal_poor, 8=angle_abnormal, 9=other）
- `recovery`: 恢复状态（字符串）
  - "signal_poor" = 信号差（对应原始值 0，触发）
  - "signal_recovery" = 信号恢复（对应原始值 1，恢复） 这是清L 硬件发出的
- `device_uid`: 需要时放在 data_value 项内

**信号差说明**：
- 雷达网络信号超过正常阈值范围时产生
- **阈值**：`<=-88` 即偏弱（触发 signal_poor）
- **正常范围**：
  - Wifi: -88~-20
  - 4G: 11~31
- **固件版本**：2024年3月份版本后

### 5.7 type=8（倾角异常事件）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "angle_abnormal",
  "data_value": [
    {
      "category": "angle_abnormal",
      "recovery": "angle_recovery",
      "device_uid": "E598A2ACD523"
    }
  ]
}
```

**字段说明**：
- `data_value`: 数组，包含倾角异常事件信息
- `category`: 固定为 "angle_abnormal"
- `recovery`: 恢复状态（字符串），"angle_abnormal" / "angle_recovery"
- `device_uid`: 需要时放在 data_value 项内

**倾角异常说明**：
- 雷达倾角低于阈值时产生
- **雷达倾角格式**：`X:Y:Z:V` 四个值
  - `V`：雷达角度是否已出厂校准
    - `0` = 未校准
    - `1` = 已校准
  - `XYZ`：雷达与空间XYZ三个轴的夹角
- **安装要求**：
  - **顶装**：X和Y应在±10°以内
  - **侧装**：X应在±10°以内，Y应在-60°到-90°之间（墙面夹角30°以内）
- **注意**：只有HC2系列支持倾角测量，TK2系列不支持
- **固件版本**：2024年3月份版本后


### 5.8 type=9（其他告警）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "other",
  "data_value": [
    {
      "category": "other",
      "alarmType": "2",
      "device_uid": "E598A2ACD523"
    }
  ]
}
```

**字段说明**：
- `data_value`: 数组，包含其他告警事件信息
- `category`: 固定为 "other"
- `alarmType`: 告警类型（字符串），"1"=离床未归, "2"=滞留, "3"=长时间无人活动, "4"=坐地告警, "5"=床上坐起告警 等
- `device_uid`: 需要时放在 data_value 项内

### 5.9 事件类型汇总

| type | 事件类型 | category | data_value 格式 | 说明 |
|------|---------|----------|----------------|------|
| 1 | 进出事件 | "enter2out" | 数组 | 包含 `track_id`, `event`, `area_type` |
| 2 | 姿态变化事件 | "pose" | 数组 | 包含 `track_id`, `pose` |
| 3 | 人数变化事件 | "number-people" | 数组 | 包含 `number_people` |
| 5 | 设备在线状态 | "isOnline" | 数组 | 包含 `device_status`, `device_uid` |
| 7 | 信号差事件 | "signal_poor" | 数组 | 包含 `recovery`, `device_uid` |
| 8 | 倾角异常事件 | "angle_abnormal" | 数组 | 包含 `recovery`, `device_uid` |
| 9 | 其他告警 | "other" | 数组 | 包含 `alarmType`, `device_uid` |

### 5.10 BedState 事件

由 Wisefido-AI 触发，card_agg 汇总为卡片动态数据。

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "BedState",
  "data_value": [
    {
      "category": "BedState",
      "bed_id": "uuid-bbb",
      "CurrentState": "in_bed",
      "StateSince": 1234567890
    }
  ]
}
```


### 5.11 RoomState 事件

由 Wisefido-AI 触发，card_agg 汇总为卡片动态数据。

```json
{
  "device_id": "uuid-ddd",
  "device_type": "AI",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "event",
  "category": "RoomState",
  "data_value": [
    {
      "category": "RoomState",
      "room_id": "uuid-rrr",
      "room_name": "101",
      "PeopleCount": 1,
      "StayTime": 5,
      "StateSince": 1234567890
    }
  ]
}
```



## 六、Alarm 数据格式（告警数据）

**说明**：告警数据从事件（Event）和统计（Stat）数据中抽取，当满足告警使能条件时生成告警事件。将topic改为 alarm
top    "category": "alarmLevel.alarmType"

以下个为例，alarm_device中，fall 报警设置为启用

### 6.1 跌倒告警（从 type=2 姿态变化事件中抽取）

当 `pose` 为 "Fall" 或 "SuspectedFall" 时，且跌倒告警 enable。

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "alarm",
  "category": "alarmLevel.pose",   // alarmLevel.alarmType
  "data_value": [
    {
      "category": "pose",
      "track_id": 1,
      "pose": "Fall"
    }
  ]
}
```

**字段说明**：
- `data_value`: 数组，包含跌倒告警信息
--`category`: alarmLevel.alarmCode
- `category`: 固定为 "pose"
- `pose`: 姿态类型（字符串），"Fall" / "SuspectedFall"
- `track_id`: 人员轨迹 ID（数值）

### 6.3 从统计 Stat（统计数据）

```json
{
  "device_id": "uuid-ddd",
  "device_type": "Radar",
  "card_id": "uuid-ccc",
  "tenant_id": "TENANT001",
  "timestamp": 1234567890,
  "topic_type": "alarm",
  "category": "sleep",
  "data_value": [
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
      "raw_original": "base64_string"
    }
  ]
}
```

## 七、Auth 数据格式（设备认证数据）

### 7.1 认证请求格式（设备端发起）

```json
{
  "device_id": "uuid-ddd",
  "device_uid": "E598A2ACD523",    
  "device_type": "Radar",
  "tenant_id": "'00000000-0000-0000-0000-000000000000'",
  "timestamp": 1234567890,
  "topic_type": "auth",
  "category": "auth_request",      //auth_request  ,auth_response  
  "data_value": {
    "category": "auth_request",      //auth_request  ,auth_response
    "device_uid": "E598A2ACD523",
    "remote_addr": "10.0.0.187:57087"
    "log":{                           //author 产生的log,格式上auth决定
      "device_type": 1,
      "mcu_hw": "2.0",
      "mcu_sw": "Dec 17 2025 10:22:19",
      "radar_hw": "2.3",
      "radar_sw": "Jun 25 2025 11:33:44"            
  }
}
```

对于radar, log可能如下，但其它厂家设备可能没有这些version信息

### 7.2 认证响应格式（服务器端返回）

**响应码说明**（设备响应中的 `code` 字段）：
- `200` - 成功
- `500` - 失败
- `777` - 设备离线
- `778` - 该设备不适用该模式


```json
{
  "device_id": "uuid-ddd",
   "device_uid": "E598A2ACD523",    
   "device_type": "Radar",
  "tenant_id": "bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c",
  "timestamp": 1234567890,
  "topic_type": "auth",
  "category": "auth_response",  
  "data_value": {
    "category": "auth_response",
    "auth_status": "success",           //success, deny, login , password_der
    "device_uid": "E598A2ACD523",
    "log":                            //author 产生的信息或log
    }
  }
}
```

### 7.3 字段说明

| 字段名 | 类型 | 说明 |
|-------|------|------|
| `category` | string | 认证阶段标识：`"auth_request"`（设备请求）或 `"auth_response"`（服务器响应） |
| `auth_method` | string | 认证方法，当前固定为 `"device_uid"` |
| `device_uid` | string | 设备唯一标识符（MAC地址或硬件ID），如 `"E598A2ACD523"` |
| `device_type` | int | 设备类型代码（1=Radar） |
| `auth_status` | string | **（仅在auth_response中出现）** 认证结果，值为 `"success"` 或 `"failure"` |
| `mqtt_server` | string | **（仅在auth_response中出现）** MQTT服务器地址 |
| `mqtt_port` | int | **（仅在auth_response中出现）** MQTT服务器端口 |
| `firmware_version` | object | 固件版本信息，包含四个字段 |
| `firmware_version.mcu_hw` | string | MCU硬件版本 |
| `firmware_version.mcu_sw` | string | MCU软件版本（编译日期和时间） |
| `firmware_version.radar_hw` | string | Radar硬件版本 |
| `firmware_version.radar_sw` | string | Radar软件版本（编译日期和时间） |
| `remote_addr` | string | **（仅在auth_request中出现）** 设备远程地址（IP地址:端口） |

### 7.4 设计说明

- **`category` 字段**：用于区分认证的不同阶段
  - `"auth_request"`：设备发起认证请求时使用，此时 `tenant_id` 为默认值
  - `"auth_response"`：服务器返回认证结果时使用，此时 `tenant_id` 为实际分配的租户ID
  
- **`tenant_id` 说明**：
  - 设备请求时不包含此信息，故使用默认值 `'00000000-0000-0000-0000-000000000000'`
  - 服务器响应时填充实际的租户ID
  - 此字段保证后端数据库字段的完整性

- **`firmware_version` 嵌套格式**：
  - 与Device_store格式保持一致
  - 支持统一所有厂家的固件版本格式

## 八、字段映射规则

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

## 九、实现说明

### 8.1 topic_type 映射
- `topic_type = "monitor"` → `topic_type = "monitor"`
- `topic_type = "stat"` → `topic_type = "stat"`（统一使用 "stat"）
- `topic_type = "event"` → `topic_type = "event"`
- `topic_type = "alarm"` → `topic_type = "alarm"`

### 8.2 data_value 结构
- **统一为数组**：所有 topic_type（monitor、stat、event、alarm）的 `data_value` 均为 **JSON 数组**，每项为一条数据，项内必含 `category` 字段；单条时也为单元素数组。
- **Monitor**：数组项为 track 或 vital 等，每项含 `category` 及对应字段。
- **Stat**：数组项为 sleep、track 等统计对象，每项含 `category` 及对应字段。
- **Event/Alarm**：数组项为事件/告警对象，每项含 `category` 及对应字段。

### 8.3 category 分组
- **"track"**: 轨迹相关数据（position_x, position_y, position_z, target_id, area_id, remaining_time, pose, event）
- **"vital"**: 生命体征相关数据（respiratory_rate, heart_rate, sleep_status, stability）
- **"sleep"**: 睡眠统计相关数据（sleep_flag, respiratory_rate, heart_rate, avg_respiratory_rate, avg_heart_rate, breath_state, heart_state, vital_signs_state, sleep_state）
- **"enter2out"**: 进出事件（track_id, event, area_type）
- **"pose"**: 姿态变化事件（track_id, pose）
- **"number-people"**: 人数变化事件（number_people）
- **"isOnline"**: 设备在线状态（device_status, device_uid）
- **"signal_poor"**: 信号差事件（recovery, device_uid）
- **"angle_abnormal"**: 倾角异常事件（recovery, device_uid）
- **"other"**: 其他告警（alarmType, device_uid）
- **告警 category**: "fall_alarm", "device_offline", "signal_poor_alarm", "angle_abnormal_alarm", "vital_signs_weak" 等

## 十、字段顺序与约定说明

### 10.1 顶层字段顺序

文档中 iot:\*:stream 顶层字段顺序统一为：`device_id` → `device_type` → `card_id` → `tenant_id` → `timestamp` → `topic_type` → `category` → `data_value`。**不包含** `device_uid`、位置信息（addressInfo）；需要时由消费方按 `card_id` 查询 Card 获取。

**逻辑分组**：
- **设备与卡**：`device_id`, `device_type`, `card_id`（未绑卡可空）
- **租户与时间**：`tenant_id`, `timestamp`
- **类型与内容**：`topic_type`, `category`, `data_value`（数组，项内含 `category`）

### 10.2 data_value 与 addressInfo

- **data_value**：统一为 **JSON 数组**，每项含 `category` 及该类别对应字段；单条时为单元素数组。
- **位置信息**：流中不再携带 `branch_id`, `building_id`, `unit_id`, `room_id`, `bed_id`；需用时按 `card_id` 向 wisefido-data 查询 Card 静态结构。

### 10.3 本次调整内容

- ✅ 二、标准格式结构：顶层仅保留 device_id, device_type, card_id, tenant_id, timestamp, topic_type, category, data_value；data_value 为数组；去掉 addressInfo
- ✅ 三、Monitor：data_value 改为数组，顶层去掉 device_uid 与 addressInfo
- ✅ 四、Stat：data_value 改为数组，顶层去掉 addressInfo
- ✅ 五、Event / 六、Alarm：顶层去掉 device_uid 与 addressInfo，data_value 保持数组
- ✅ 八、实现说明：data_value 统一为数组；category 分组说明
