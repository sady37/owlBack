# Lost Fall 测试 — Radar_CD2B 11:35-11:48 MDT

设备 `9D8A32A1CD2B` (Radar_CD2B)，房间 `b368ee77-0a13-4584-9ce1-bbcff6205e48` (Bedroom)，
单元 `3db05d02-1257-407d-88bb-cbe9a7498b27`，tenant `43b8fbf7-b55f-4b48-bd8b-27bb14f48870`。

## 文件清单

| 文件 | 内容 |
|---|---|
| `room_layout.json` | rooms.layout_config raw（walls / Enter / Bed / radar mount 等） |
| `2026-04-27_11-35_to_11-48_MDT.json` | 全量 iot_timeseries 559 条（monitor / event / alarm 三类） |

时间窗（MDT = UTC-6）：
- MDT 11:35:00 – 11:48:00 = UTC 17:35:00 – 17:48:00
- bigint 时间戳：1777311300000 – 1777312080000（ms）

## 数据统计

| topic_type | 条数 |
|---|---|
| monitor | 534 |
| event | 24 |
| alarm | 1 |

## Event/Alarm Timeline (UTC)

| 时间 | topic | event_name | status | track_id | 备注 |
|---|---|---|---|---|---|
| 17:35:05 | event | activity | instant | 9 | 进入房间前 activity |
| 17:35:36 | event | activity | instant | 9 | |
| 17:36:40 | event | number_people | start | 10 | |
| 17:36:40 | event | activity | instant | 9 | |
| 17:37:03 | event | number_people | start | 10 | |
| **17:37:03** | **event** | **EnterRoom** | **start** | | **进房** |
| 17:37:33 | event | activity | instant | 9 | |
| 17:38:33 | event | activity | instant | 9 | |
| **17:39:28** | **event** | **InBed** | **start** | | **躺床上** |
| 17:39:32 | event | activity | instant | 9 | |
| 17:40:32 | event | activity | instant | 9 | |
| 17:41:31 | event | activity | instant | 9 | |
| 17:42:32 | event | activity | instant | 9 | |
| 17:43:31 | event | activity | instant | 9 | |
| 17:43:55 | event | SignalPoorRecover | end | 11 | firmware 信号差恢复 |
| 17:43:55 | alarm | OfflineRecover | end | 11 | firmware 离线恢复（非真 alarm） |
| 17:43:55 | event | AngleExceptionRecover | end | 11 | firmware 角度异常恢复 |
| 17:44:30 | event | activity | instant | 9 | |
| 17:45:07 | event | number_people | start | 10 | |
| 17:45:39 | event | activity | instant | 9 | |
| 17:46:43 | event | activity | instant | 9 | |
| 17:47:08 | event | number_people | start | 10 | |
| **17:47:25** | **event** | **ExitRoom** | **start** | | **离房** |
| 17:47:26 | event | number_people | start | 10 | |
| 17:47:46 | event | activity | instant | 9 | |

## 关键观察

1. **alarm_events 表此窗口无 Fall 报警** — 漏报场景（用户测试 lost fall 用）
2. **monitor 流连续**：17:37:03（EnterRoom）→ 17:45:39 之间无大于 5s 的间隔（最长 gap 是进房前/离房前的 ~30s 静默）
3. **17:43:55 三个 Recover 事件**：firmware 状态机内部信号差/离线/角度异常的恢复信号；event_since 与 timestamp 同值意味着此时刻报告"刚恢复"，但其 start 事件不在本窗口（推测早于 11:35）
4. **InBed 持续 ~8 分钟**（17:39:28 ~ 17:47:25 ExitRoom）

## 用途

回归 fixture，用于：
- Lost fall 规则验证：track 在 Bed area 长时间静止（AreaBed/AreaSit cell timeout = 60min）应不触发，而 ExitRoom 后才正常退出
- Cell history integral 学习入口：本段 InBed 期间长时间无 alarm，对应 cell 应累计 ToleratedStillCount
- 验证 firmware Recover 事件不被误当作真 alarm 灌入告警表（topic_type=alarm 但 event_name=OfflineRecover）

## JSON Schema

```
[
  {
    "id": <bigint>,
    "device_uid": "9D8A32A1CD2B",
    "device_id": "ac790ba1-ac17-47d1-9cd1-c65848bd0027",
    "timestamp": <epoch ms>,
    "topic_type": "monitor" | "event" | "alarm",
    "category": <string>,
    "data_value": <jsonb array of frame objects>
  },
  ...
]
```

monitor frame `data_value[i]` 字段：track_id / pose / x / y / z / heart_rate / resp_rate / sleep_state / ...
