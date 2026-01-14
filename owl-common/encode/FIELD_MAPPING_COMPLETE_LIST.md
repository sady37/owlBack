# 完整字段映射清单

## 字段命名标准规范

### 标准字段名（各厂家统一）
- **心率**：`heart_rate`（Radar: `heart_rate`, Sleepace: `heart` → `heart_rate`）
- **呼吸率**：`respiratory_rate`（Radar: `breath_rate` → `respiratory_rate`, Sleepace: `breath` → `respiratory_rate`）

### SNOMED 映射字段命名规范
对于需要 SNOMED 映射的字段，统一使用以下字段命名：
```
{fieldName}                    // 原始值（保留）
{fieldName}_snomed_code       // SNOMED 编码
{fieldName}_snomed_display    // SNOMED 显示（标准显示字段，各厂家统一使用）
{fieldName}_category          // FHIR Category
{fieldName}_display_en        // 英文显示名称
```

### 字段使用规范
- **数值字段**（不需要 SNOMED 映射）：统一使用标准字段名 `heart_rate`, `respiratory_rate`
- **枚举/状态字段**（需要 SNOMED 映射）：使用 `{fieldName}_snomed_display` 作为标准显示字段
- **查询规范**：查询数值字段使用标准字段名，查询显示字段使用 `{fieldName}_snomed_display`

**注意**：不再保留原始字段名（`breath_rate`, `heart`, `breath`），统一使用标准字段名。

---

## 一、元数据字段（各厂家统一）

| 标准字段名 | 说明 | 来源 | Radar | Sleepace |
|----------|------|------|-------|----------|
| `device_id` | 设备 ID | 原始数据 | ✅ | ✅ |
| `tenant_id` | 租户 ID | 原始数据 | ✅ | ✅ |
| `device_type` | 设备类型 | 原始数据 | ✅ | ✅ |
| `topic_type` / `data_key` | 数据类型 | 原始数据 | `topic_type` | `data_key` |
| `timestamp` | 时间戳 | 原始数据 | ✅ | ✅ |

## 二、生命体征字段（各厂家统一标准字段名）

| 标准字段名 | 说明 | 单位 | Radar 原始字段 | Sleepace 原始字段 | SNOMED 映射 |
|----------|------|------|--------------|-----------------|------------|
| `heart_rate` | 心率 | bpm | `heart_rate` (bh 字节 2) | `heart` | ❌ 不需要 |
| `respiratory_rate` | 呼吸率 | 次/分钟 | `breath_rate` (bh 字节 1) | `breath` | ❌ 不需要 |
| `realtime_heart_rate` | 实时心率（统计数据） | bpm | `heart_rate` (sleep 字节 2) | - | ❌ 不需要 |
| `realtime_respiratory_rate` | 实时呼吸率（统计数据） | 次/分钟 | `breath_rate` (sleep 字节 1) | - | ❌ 不需要 |
| `avg_heart_rate` | 平均心率（统计数据） | bpm | `heart_rate` (sleep 字节 6) | - | ❌ 不需要 |
| `avg_respiratory_rate` | 平均呼吸率（统计数据） | 次/分钟 | `breath_rate` (sleep 字节 5) | - | ❌ 不需要 |

**注意**：所有厂家统一使用 `heart_rate` 和 `respiratory_rate` 作为标准字段名。

## 三、位置和轨迹字段（Radar）

| 标准字段名 | 说明 | 单位 | 来源 | SNOMED 映射 |
|----------|------|------|------|------------|
| `position_x` | X 坐标 | 厘米 | track 解码 (dm → cm) | ❌ 不需要 |
| `position_y` | Y 坐标 | 厘米 | track 解码 (dm → cm) | ❌ 不需要 |
| `position_z` | Z 坐标（高度） | 厘米 | track 解码 | ❌ 不需要 |
| `tracking_id` / `target_id` | 目标 ID | - | track 解码 | ❌ 不需要 |
| `area_id` | 区域 ID | - | track 解码 | ❌ 不需要 |
| `remaining_time` | 剩余时间 | 秒 | track 解码 | ❌ 不需要 |
| `duration` | 持续时间 | 秒 | 原始数据 | ❌ 不需要 |

## 四、统计数据字段（Radar stat）

| 标准字段名 | 说明 | 单位 | 来源 | SNOMED 映射 |
|----------|------|------|------|------------|
| `walk_duration` | 行走时长 | 秒 | 原始数据 | ❌ 不需要 |
| `lie_duration` | 躺卧时长 | 秒 | 原始数据 | ❌ 不需要 |
| `stand_duration` | 站立时长 | 秒 | 原始数据 | ❌ 不需要 |
| `multi_person_duration` | 多人时长 | 秒 | 原始数据 | ❌ 不需要 |
| `walk_distance` | 行走距离 | 厘米 | 原始数据 (m → cm) | ❌ 不需要 |
| `sit_duration` | 坐位时长 | 秒 | track 解码 | ❌ 不需要 |

## 五、需要 SNOMED 映射的字段（各厂家统一显示字段：{field}_snomed_display）

### 5.1 Radar - Monitor (实时数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `pose` | `pose` | `pose_snomed_display` | 姿态 | track 解码 | `monitor.track.pose` |
| `event` | `event` | `event_snomed_display` | 事件 | track 解码 | `monitor.track.event` |
| `sleep_status` | `sleep_status` | `sleep_status_snomed_display` | 睡眠状态 | bh 解码 (bit 7:6) | `monitor.bh.sleep_status` |
| `stability` | `stability` | `stability_snomed_display` | 稳定度 | bh 解码 (bit 1:0) | `monitor.bh.stability` |

### 5.2 Radar - Stat (统计数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `breath_state` | `breath_state` | `breath_state_snomed_display` | 呼吸状态 | sleep 解码 (bit 1:0) | `stat.sleep.hr_breath_event.breath_state` |
| `heart_state` | `heart_state` | `heart_state_snomed_display` | 心率状态 | sleep 解码 (bit 3:2) | `stat.sleep.hr_breath_event.heart_state` |
| `vital_signs_state` | `vital_signs_state` | `vital_signs_state_snomed_display` | 生命体征状态 | sleep 解码 (bit 5:4) | `stat.sleep.hr_breath_event.vital_signs_state` |
| `sleep_state` | `sleep_state` | `sleep_state_snomed_display` | 睡眠状态 | sleep 解码 (bit 7:6) | `stat.sleep.hr_breath_event.sleep_state` |

### 5.3 Radar - Event (事件数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `event` | `event` | `event_snomed_display` | 进出事件 | 原始数据 / data 数组 | `event.enter_leave.event` |
| `pose` | `pose` | `pose_snomed_display` | 姿态变化事件 | 原始数据 / data 数组 | `event.pose_change.pose` |

### 5.4 Radar - Alarm (告警数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `event` | `event` | `event_snomed_display` | 进出事件告警 | data 数组 | `event.enter_leave.event` |
| `pose` | `pose` | `pose_snomed_display` | 姿态变化告警 | data 数组 | `event.pose_change.pose` |

### 5.5 Sleepace - Realtime (实时数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `bedStatus` | `bedStatus` | `bedStatus_snomed_display` | 床状态（0=在床, 1=离床） | 原始数据 | `realtime.bedStatus` |
| `sitUp` | `sitUp` | `sitUp_snomed_display` | 床上坐起（sitUp > 0） | 原始数据 | `realtime.sitUp` |

### 5.6 Sleepace - SleepStage (睡眠阶段数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `sleepStage` | `sleepStage` | `sleepStage_snomed_display` | 睡眠阶段（0=清醒, 1=浅睡眠, 2=深睡眠, 3=REM） | 原始数据 | `sleepStage.sleepStage` |

### 5.7 Sleepace - ConnectionStatus (连接状态数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `connectionStatus` | `connectionStatus` | `connectionStatus_snomed_display` | 连接状态（0=离线, 1=在线） | 原始数据 | `connectionStatus.connectionStatus` |

### 5.8 Sleepace - AlarmNotify (告警通知数据)

| 原始字段名 | 标准字段名 | SNOMED 显示字段 | 说明 | 来源 | 映射路径 |
|----------|----------|---------------|------|------|---------|
| `alarmType` / `type` | `alarmType` | `alarmType_snomed_display` | 告警类型 | 原始数据 | `alarmNotify.type` |

## 六、其他字段（不进行 SNOMED 映射）

### 6.1 Radar

| 字段名 | 说明 | 来源 |
|-------|------|------|
| `track` | 原始 track base64 字符串 | 原始数据 |
| `track_list` | 多人 track 数据（数组） | track 解码 |
| `version` | 版本标识符 | track 解码 |
| `people_count` | 人数 | track 解码 |
| `number-people` / `number_people` | 人数（事件数据） | 原始数据 |
| `type` | 事件类型（1=进出, 2=姿态变化, 3=人数变化） | 原始数据 |
| `track-id` / `track_id` | 轨迹 ID | 原始数据 |
| `cmd` | 命令字段 | 原始数据 |

### 6.2 Sleepace

| 字段名 | 说明 | 来源 |
|-------|------|------|
| `turnOver` | 翻身次数 | 原始数据 |
| `bodyMove` | 体动次数 | 原始数据 |
| `leftRight` | 左侧/右侧 | 原始数据 |
| `initStatus` | 初始化状态 | 原始数据 |
| `signalQuality` | 信号质量 | 原始数据 |
| `status` | 告警状态 | 原始数据 |
| `userId` | 用户 ID | 原始数据 |
| `relieveReason` | 解除原因 | 原始数据 |
| `relieveTime` | 解除时间 | 原始数据 |

## 七、SNOMED 映射字段命名规范

对于需要 SNOMED 映射的字段，统一使用以下字段命名：

| 字段名 | 说明 | 示例 |
|-------|------|------|
| `{fieldName}` | 原始值 | `pose: 1` |
| `{fieldName}_snomed_code` | SNOMED 编码 | `pose_snomed_code: "129006008"` |
| `{fieldName}_snomed_display` | **SNOMED 显示（标准显示字段，各厂家统一使用）** | `pose_snomed_display: "Walking"` |
| `{fieldName}_category` | FHIR Category | `pose_category: "activity"` |
| `{fieldName}_display_en` | 英文显示名称 | `pose_display_en: "Walking"` |

**重要**：各厂家统一使用 `{fieldName}_snomed_display` 作为标准显示字段进行查询和显示。

## 八、字段映射完整示例

### 8.1 Radar Monitor 数据示例

```json
{
  "device_id": "RADAR001",
  "tenant_id": "TENANT001",
  "device_type": "Radar",
  "topic_type": "monitor",
  "timestamp": 1234567890,
  
  // 生命体征（标准字段名）
  "heart_rate": 75,
  "respiratory_rate": 16,
  
  // 位置（标准字段名）
  "position_x": 120,
  "position_y": 200,
  "position_z": 0,
  "tracking_id": 1,
  "area_id": 1,
  
  // 姿态（SNOMED 映射）
  "pose": 1,
  "pose_snomed_code": "129006008",
  "pose_snomed_display": "Walking",
  "pose_category": "activity",
  "pose_display_en": "Walking",
  
  // 事件（SNOMED 映射）
  "event": 0,
  "event_snomed_code": null,
  "event_snomed_display": null,
  
  // 睡眠状态（SNOMED 映射）
  "sleep_status": 0,
  "sleep_status_snomed_code": null,
  "sleep_status_snomed_display": "Sleep state undefined",
  "sleep_status_category": "activity",
  "sleep_status_display_en": "Sleep state undefined",
  
  // 稳定度（SNOMED 映射）
  "stability": 0,
  "stability_snomed_code": null,
  "stability_snomed_display": "No interference",
  "stability_category": "vital-signs",
  "stability_display_en": "No interference"
}
```

### 8.2 Sleepace Realtime 数据示例

```json
{
  "device_id": "SLEEPACE001",
  "tenant_id": "TENANT001",
  "device_type": "Sleepace",
  "data_key": "realtime",
  "timestamp": 1234567890,
  
  // 生命体征（标准字段名）
  "heart_rate": 72,
  "respiratory_rate": 14,
  
  // 床状态（SNOMED 映射）
  "bedStatus": 0,
  "bedStatus_snomed_code": "248569007",
  "bedStatus_snomed_display": "In bed",
  "bedStatus_category": "activity",
  "bedStatus_display_en": "In bed",
  
  // 坐起（SNOMED 映射，仅当 sitUp > 0 时）
  "sitUp": 0,
  
  // 其他字段
  "turnOver": 2,
  "bodyMove": 5
}
```

## 九、SNOMED 映射字段列表

所有需要 SNOMED 映射的字段，统一使用 `{fieldName}_snomed_display` 作为标准显示字段：

- **Radar**: `pose_snomed_display`, `event_snomed_display`, `sleep_status_snomed_display`, `stability_snomed_display`, `breath_state_snomed_display`, `heart_state_snomed_display`, `vital_signs_state_snomed_display`, `sleep_state_snomed_display`
- **Sleepace**: `bedStatus_snomed_display`, `sitUp_snomed_display`, `sleepStage_snomed_display`, `connectionStatus_snomed_display`, `alarmType_snomed_display`
