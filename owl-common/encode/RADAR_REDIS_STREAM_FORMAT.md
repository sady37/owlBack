# Radar Redis Stream 数据格式

## 一、Redis Stream 名称

| topic_type | Redis Stream 名称 |
|-----------|------------------|
| `monitor` | `iot:monitor:stream` |
| `stat` | `iot:stat:stream` |
| `event` | `iot:event:stream` |
| `alarm` | `iot:alarm:stream` |

## 二、Monitor 数据格式（实时数据）

### 2.1 元数据字段（所有类型都有）
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `device_id` | string | 设备 ID |
| `tenant_id` | string | 租户 ID |
| `device_type` | string | 设备类型（固定为 "Radar"） |
| `topic_type` | string | 主题类型（固定为 "monitor"） |
| `timestamp` | int64 | 时间戳（Unix 时间戳） |

### 2.2 从 track base64 解码的字段
| 字段名 | 类型 | 说明 | 来源 |
|-------|------|------|------|
| `position_x` | int | X 坐标（厘米） | track 字节 1，dm→cm 转换 |
| `position_y` | int | Y 坐标（厘米） | track 字节 2，dm→cm 转换 |
| `position_z` | int | Z 坐标（厘米） | track 字节 3 |
| `target_id` | int | 目标 ID | track 字节 0 |
| `area_id` | int | 区域 ID | track 字节 15 |
| `remaining_time` | int | 剩余时间（秒，0-60，仅在自动测量边界时使用，正常不用） | track 字节 12 |
| `pose` | string | 姿态（display_en 值） | track 字节 13，SNOMED 映射 |
| `event` | int/string | 事件（原始值或 display_en 值） | track 字节 14，如果有映射则使用 display_en，否则保留原始值 |
| `track` | string | 原始 track base64 字符串（保留） | 原始数据，16 字节 * N（N 为人数） |

### 2.3 从 bh base64 解码的字段
| 字段名 | 类型 | 说明 | 来源 |
|-------|------|------|------|
| `respiratory_rate` | int | 呼吸率（次/分钟） | bh 字节 1 |
| `heart_rate` | int | 心率（bpm） | bh 字节 2 |
| `sleep_status` | string | 睡眠状态（display_en 值） | bh 字节 13 bit 7:6，SNOMED 映射 |
| `stability` | string | 稳定度（原始 bit 值或 display_en 值） | bh 字节 14 bit 1:0，如果 display_en 为空则保留原始 bit 值（如 "11"） |
| `bh` | string | 原始 bh base64 字符串（保留） | 原始数据 |

### 2.4 其他字段（原始数据中的字段，直接保留）
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `duration` | int | 持续时间（秒，原始数据中的字段，不是从 track 解码的） |
| `track_list` | array | 多人数据（如果 track 包含多人，每个元素包含 target_id, position_x, position_y, position_z, remaining_time, pose, event, area_id） |

### 2.5 完整示例
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "monitor",
  "timestamp": 1234567890,
  "position_x": 150,
  "position_y": 200,
  "position_z": 50,
  "target_id": 1,
  "area_id": 1,
  "remaining_time": 30,
  "pose": "Walking",
  "event": 0,
  "respiratory_rate": 14,
  "heart_rate": 72,
  "sleep_status": "Sleep state undefined",
  "stability": "11",
  "duration": 60,
  "track": "base64_string",
  "bh": "base64_string"
}
```

## 三、Stat 数据格式（统计数据）

### 3.1 元数据字段（同 Monitor）

### 3.2 轨迹统计字段
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `walk_duration` | int | 行走时长（秒） |
| `lie_duration` | int | 躺卧时长（秒） |
| `stand_duration` | int | 站立时长（秒） |
| `multi_person_duration` | int | 多人时长（秒） |
| `walk_distance` | int | 行走距离（厘米，m→cm 转换） |
| `sit_duration` | int | 坐位时长（秒） |

### 3.3 从 sleep base64 解码的字段
| 字段名 | 类型 | 说明 | 来源 |
|-------|------|------|------|
| `realtime_respiratory_rate` | int | 实时呼吸率（次/分钟） | sleep 字节 1 |
| `realtime_heart_rate` | int | 实时心率（bpm） | sleep 字节 2 |
| `avg_respiratory_rate` | int | 平均呼吸率（次/分钟） | sleep 字节 5 |
| `avg_heart_rate` | int | 平均心率（bpm） | sleep 字节 6 |
| `breath_state` | string | 呼吸状态（display_en 值） | sleep 字节 13 bit 1:0，SNOMED 映射 |
| `heart_state` | string | 心率状态（display_en 值） | sleep 字节 13 bit 3:2，SNOMED 映射 |
| `vital_signs_state` | string | 生命体征状态（display_en 值） | sleep 字节 13 bit 5:4，SNOMED 映射 |
| `sleep_state` | string | 睡眠状态（display_en 值） | sleep 字节 13 bit 7:6，SNOMED 映射 |
| `sleep` | string | 原始 sleep base64 字符串（保留） | 原始数据 |

### 3.4 其他字段
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `version` | int | 版本号 |
| `people_count` | int | 人数 |
| `track` | string | 原始 track base64 字符串（如果有） |

### 3.5 完整示例
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "stat",
  "timestamp": 1234567890,
  "walk_duration": 300,
  "lie_duration": 1800,
  "stand_duration": 120,
  "multi_person_duration": 0,
  "walk_distance": 5000,
  "sit_duration": 60,
  "realtime_respiratory_rate": 14,
  "realtime_heart_rate": 72,
  "avg_respiratory_rate": 15,
  "avg_heart_rate": 70,
  "breath_state": "Normal",
  "heart_state": "Normal",
  "vital_signs_state": "Normal",
  "sleep_state": "Light sleep",
  "version": 1,
  "people_count": 1,
  "sleep": "base64_string"
}
```

## 四、Event 数据格式（事件数据）

### 4.1 元数据字段（同 Monitor）

### 4.2 type 字段
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `type` | int | 事件类型：1=进出事件, 2=姿态变化, 3=人数变化 |

### 4.3 data 字段（数组或对象）

**type=1（进出事件）** - data 是数组：
```json
{
  "data": [
    {
      "event": "Enter room",   // display_en 值（如果有映射）
      "area_type": 1,
      "track-id": 1
    }
  ]
}
```

**type=2（姿态变化事件）** - data 是数组：
```json
{
  "data": [
    {
      "pose": "Sitting up in bed",  // display_en 值（如果有映射）
      "track-id": 1
    }
  ]
}
```

**type=3（人数变化事件）** - data 是对象：
```json
{
  "data": {
    "number-people": 2
  }
}
```

### 4.4 兼容处理字段（如果 event/pose 直接在顶层）
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `event` | string | display_en 值（如果有映射，type=1 时） |
| `pose` | string | display_en 值（如果有映射，type=2 时） |
| `area_type` | int | 区域类型（type=1 时） |
| `track-id` | int | 轨迹 ID（type=1 或 2 时） |
| `number-people` | int | 人数（type=3 时） |

### 4.5 其他字段
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `cmd` | string | 命令字符串（如果有） |

### 4.6 完整示例

**type=1（进出事件）**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 1,
  "data": [
    {
      "event": "Enter room",
      "area_type": 1,
      "track-id": 1
    }
  ]
}
```

**type=2（姿态变化事件）**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 2,
  "data": [
    {
      "pose": "Sitting up in bed",
      "track-id": 1
    }
  ]
}
```

**type=3（人数变化事件）**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "event",
  "timestamp": 1234567890,
  "type": 3,
  "data": {
    "number-people": 2
  }
}
```

## 五、Alarm 数据格式（告警数据）

### 5.1 元数据字段（同 Monitor）

### 5.2 type 字段
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `type` | int | 告警类型：1=进出事件, 2=姿态变化, 3=人数变化 |

### 5.3 data 字段（数组或对象）

**type=1（进出事件告警）** - data 是数组：
```json
{
  "data": [
    {
      "event": "Left bed",     // display_en 值（如果有映射）
      "track-id": 1
    }
  ]
}
```

**type=2（姿态变化告警）** - data 是数组：
```json
{
  "data": [
    {
      "pose": "Fall",          // display_en 值（如果有映射）
      "track-id": 1
    }
  ]
}
```

**type=3（人数变化告警）** - data 是对象：
```json
{
  "data": {
    "number-people": 2
  }
}
```

### 5.4 兼容处理字段（如果 event/pose 直接在顶层）
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `event` | string | display_en 值（如果有映射，type=1 时） |
| `pose` | string | display_en 值（如果有映射，type=2 时） |
| `track-id` | int | 轨迹 ID（type=1 或 2 时） |
| `number-people` | int | 人数（type=3 时） |

### 5.5 其他字段
| 字段名 | 类型 | 说明 |
|-------|------|------|
| `cmd` | string | 命令字符串（如果有） |

### 5.6 完整示例

**type=1（进出事件告警）**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "alarm",
  "timestamp": 1234567890,
  "type": 1,
  "data": [
    {
      "event": "Left bed",
      "track-id": 1
    }
  ]
}
```

**type=2（姿态变化告警）**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "alarm",
  "timestamp": 1234567890,
  "type": 2,
  "data": [
    {
      "pose": "Fall",
      "track-id": 1
    }
  ]
}
```

**type=3（人数变化告警）**：
```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "alarm",
  "timestamp": 1234567890,
  "type": 3,
  "data": {
    "number-people": 2
  }
}
```

## 六、字段映射规则

### 6.1 SNOMED 映射字段
对于需要 SNOMED 映射的字段（pose, event, sleep_status, stability, breath_state, heart_state, vital_signs_state, sleep_state）：
- **字段名**：保持不变
- **字段值**：使用 `display_en`（如果有映射且不为空），否则保留原始值
- **不添加**：`_snomed_code`, `_snomed_display`, `_category`, `_display_en` 等额外字段

### 6.2 数值字段
对于数值字段（respiratory_rate, heart_rate 等）：
- 使用标准字段名
- 直接使用数值，不添加 SNOMED 相关字段

### 6.3 原始 base64 字符串
- `track`, `bh`, `sleep` 等原始 base64 字符串会保留在编码后的数据中
