# wisefido-card-aggregator 输出格式定义

本文档定义 card-aggregator 写入 Redis 的 key、value 的 JSON 结构及 TTL。消费者（wisefido-data、wisefido-ai 等）依此解析。

**约定**：对内/前端展示用 **display 或设备原始值**（如 "Light sleep"、"lying"），否则前端不显示；**仅在与外部系统对接时**使用 SNOMED 码（如 248220003）。

---

## 1. Redis 键命名

| Key 模式 | 写入方 | TTL | 说明 |
|----------|--------|-----|------|
| `vital-focus:card:{card_id}:realtime` | iot_stream_consumer → CacheManager | 10s | 按源实时数据（radar/sleepad） |
| `vital-focus:card:{card_id}:full` | DataAggregator → CacheManager | 10s | 聚合后的完整卡片，供 API 返回 |
| `vital-focus:card:{card_id}:device:{device_id}:data` | iot_stream_consumer | 6s | 单设备原始时序，用于 fuse 出 realtime；TTL=2s×3=6s，与 HR/RR 2s 对齐，容错约 3 次漏报 |
| `vital-focus:card:{card_id}:alarms` | **wisefido-ai**（非 card-aggregator） | — | 报警列表；card-aggregator 仅**读取**，合并进 full |

---

## 2. `vital-focus:card:{card_id}:realtime`

**Go 类型**：`models.RealtimeData`（`internal/models/iot_timeseries.go`）

**JSON 结构**：

```json
{
  "radar": {
    "heart": 72,
    "breath": 16,
    "sleep_status": "Awake",
    "bed_status": "on_bed"
  },
  "sleepad": {
    "heart": 70,
    "breath": 18,
    "sleep_status": "Light sleep",
    "bed_status": "on_bed"
  },
  "person_count": 1,
  "postures": [
    {
      "tracking_id": "0",
      "posture_code": "lying",
      "posture_display": "Lying",
      "position_x": 160,
      "position_y": 340,
      "position_z": 50,
      "area_id": 1
    }
  ],
  "timestamp": 1769378367
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `radar` | object | Radar 源：heart, breath, sleep_status, bed_status |
| `sleepad` | object | Sleepad 源：同上 |
| `person_count` | int | 人数（tracking_id 数量） |
| `postures` | array | 姿态列表，元素为 `Posture` |
| `timestamp` | int64 | Unix 秒 |

`VitalSource`（radar/sleepad 元素）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `heart` | int | 心率 bpm |
| `breath` | int | 呼吸率 /min |
| `sleep_status` | string | 睡眠状态：**用 display 或原始**（如 "Awake","Light sleep","Deep sleep"）；对外对接时才用 SNOMED（如 248218005） |
| `bed_status` | string | on_bed / off_bed / ENTER_BED / LEFT_BED 等 |

`Posture`：

| 字段 | 类型 | 说明 |
|------|------|------|
| `tracking_id` | string | Radar 轨迹 ID，**用设备原始值**（如 target_id："0"、"1"） |
| `posture_code` | string | **用 display 或原始**（如 "lying","walk","sitting"）；对外对接时才用 SNOMED |
| `posture_display` | string | 显示名（如 "Lying"） |
| `position_x` | int | 轨迹字节 2，厘米 |
| `position_y` | int | 轨迹字节 3，厘米 |
| `position_z` | int | 轨迹字节 4，厘米 0-255，原样透传 |
| `area_id` | int | 区域 ID |

---

## 3. `vital-focus:card:{card_id}:full`

**Go 类型**：`models.VitalFocusCard`（`internal/models/vital_focus_card.go`）

**JSON 结构**（字段与 `VitalFocusCard` 的 `json` tag 一致）：

```json
{
  "card_id": "uuid",
  "tenant_id": "uuid",
  "card_type": "ActiveBed",
  "bed_id": "uuid",
  "location_id": null,
  "card_name": "1-101-1",
  "card_address": "1栋1单元101室1床",
  "primary_resident_id": "uuid",
  "residents": [{"resident_id":"","nickname":"","unit_id":null,"bed_id":null}],
  "devices": [{"device_id":"","device_name":"","device_type":"Radar","device_model":"","bed_id":null,"bed_name":null,"room_id":null,"room_name":null,"unit_id":""}],
  "device_count": 1,
  "resident_count": 1,
  "unhandled_alarm_0": 0,
  "unhandled_alarm_1": 0,
  "unhandled_alarm_2": 0,
  "unhandled_alarm_3": 0,
  "unhandled_alarm_4": 0,
  "total_unhandled_alarms": 0,
  "icon_alarm_level": 3,
  "pop_alarm_emerge": 0,
  "r_connection": 1,
  "s_connection": 0,
  "heart": 72,
  "breath": 16,
  "heart_source": "r",
  "breath_source": "r",
  "sleep_stage": 2,
  "sleep_state_snomed_code": null,
  "sleep_state_display": "Light sleep",
  "bed_status": 0,
  "person_count": 1,
  "postures": [1, 6],
  "bed_status_timestamp": "2026-01-25 14:59:00",
  "status_duration": "00:05:00",
  "alarms": []
}
```

- `heart`/`breath`/`heart_source`/`breath_source`/`sleep_stage`/`bed_status`/`person_count`/`postures`：由 DataAggregator 从 `:realtime` 经 `mergeRealtimeData` 计算出 display 后写入。
- `sleep_state_display`：前端展示用；`sleep_state_snomed_code` 仅对外对接时用。
- `alarms`：从 `:alarms` 读取后填入；格式见第 5 节。

---

## 4. `vital-focus:card:{card_id}:device:{device_id}:data`

**Go 类型**：`models.IoTTimeSeries`（`internal/models/iot_timeseries.go`）

**JSON 结构**：与 `IoTTimeSeries` 的 `json` tag 一致。示例：

```json
{
  "id": "uuid",
  "tenant_id": "uuid",
  "device_id": "uuid",
  "timestamp": "2026-01-25T21:59:27Z",
  "heart_rate": 72,
  "heart_rate_code": null,
  "heart_rate_display": null,
  "respiratory_rate": 16,
  "respiratory_rate_code": null,
  "respiratory_rate_display": null,
  "posture_snomed_code": null,
  "posture_display": "Lying",
  "tracking_id": "0",
  "position_x": 160,
  "position_y": 340,
  "position_z": 50,
  "area_id": 1,
  "bed_status_snomed_code": null,
  "bed_status_display": "In bed",
  "sleep_state_snomed_code": null,
  "sleep_state_display": "Light sleep",
  "device_type": "Radar"
}
```

供 `fuseCardDataFromCache` 融合为 `RealtimeData` 后写 `:realtime`。TTL 6s（2s×3）。`*_display`、`tracking_id` 用 display/设备原始值；`*_snomed_code` 仅对外对接时用。

---

## 5. `vital-focus:card:{card_id}:alarms`（card-aggregator 仅读取）

**写入方**：wisefido-ai（非 card-aggregator）。

**card-aggregator 解析类型**：`[]aggregator.AlarmEvent`（`internal/aggregator/data_aggregator.go`），再转为 `models.AlarmItem` 填入 full。

**期望的 JSON**：`AlarmEvent` 数组：

```json
[
  {
    "event_id": "uuid",
    "event_type": "Fall",
    "category": "safety",
    "alarm_level": "1",
    "alarm_status": "active",
    "triggered_at": 1769378367,
    "triggered_by": "Radar-1",
    "trigger_data": {"heart_rate":72,"respiratory_rate":16},
    "iot_timeseries_id": 12345
  }
]
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `event_id` | string | 事件 ID |
| `event_type` | string | Fall, SuspectedFall, OfflineAlarm 等 |
| `category` | string | safety, clinical, behavioral, device |
| `alarm_level` | string | '0'~'4' 或 EMERG/ALERT/CRIT/ERR/WARNING |
| `alarm_status` | string | active, acknowledged |
| `triggered_at` | int64 | Unix 秒 |
| `triggered_by` | string | 可选，设备名或 'Cloud' |
| `trigger_data` | object | 可选 |
| `iot_timeseries_id` | int64 | 可选 |

---

## 6. 参考

- `internal/models/vital_focus_card.go`：VitalFocusCard, CardResident, CardDevice, AlarmItem
- `internal/models/iot_timeseries.go`：RealtimeData, VitalSource, Posture, IoTTimeSeries
- `internal/aggregator/data_aggregator.go`：AlarmEvent（:alarms 解析）、mergeRealtimeData
- `internal/aggregator/cache_manager.go`：:realtime、:full 的写入
- `internal/consumer/iot_stream_consumer.go`：:realtime、:device:{id}:data 的写入
