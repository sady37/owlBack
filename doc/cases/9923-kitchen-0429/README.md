# 9923 Kitchen Case — 04-29 07:10–07:30 MDT (multi-person + np/track 不一致)

设备 `9923003AB17F` (Radar 9923, 天花板)，房间 `fae23c9e-b9b6-4408-b1d2-5d7531beb805`
(Kitchen)，tenant `43b8fbf7-b55f-4b48-bd8b-27bb14f48870` (demo)，branch
**Denver**（**MDT**）。

## 文件清单

| 文件 | 内容 |
|---|---|
| `room_layout.json` | rooms 行 (room_id / room_name / layout_config) |
| `2026-04-29_07-10_to_07-30_MDT.json` | iot_timeseries 1788 行 |

格式与 `d5f7-ghost/` 一致。

## 时间窗对齐

- 本地：2026-04-29 07:10:00 → 07:30:00 **MDT (UTC-6)**
- epoch-ms：1777468200000 → 1777469400000

`iot_timeseries.timestamp` 是 UTC epoch ms（设备 firmware `event_since`，
`wisefido-iot` 不翻译）。**导出/查询时必须用 Denver 时区**，否则会偏 1 小时
（PDT 错查会落到 08:10–08:30 MDT，这是另一段几乎无人活动的窗口，~110 行）。

## Category 分布 (1788 总行)

| category | rows |
|---|---|
| track | 1701 |
| number_people | 21 |
| activity | 20 |
| heart | 20 |
| ExitRoom | 10 |
| EnterRoom | 8 |
| AngleExceptionRecover | 2 |
| OfflineRecover | 2 |
| SignalPoorRecover | 2 |
| Walking | 1 |
| Initialization | 1 |

## Ground truth & ghost 嫌疑

- 4 个 distinct `track_id`（人员进出循环复用）
- **`max_tracks_per_frame = 1`** —— firmware 单帧从不报多 track
- 但 `number_people` 一度报到 **3** —— **np 计数器与单帧 track 数不一致**
- 这正是 ghost 嫌疑信号：firmware 内部觉得房间里有 2-3 个人，但同一帧只画 1 个，
  说明它内部维护的 track 列表里有"幽灵"被 np 计入但没出现在当前帧上

### 关键 timeline（节选）

| 时间 (MDT) | 事件 |
|---|---|
| 07:10:15 | ExitRoom（接续上一窗口） |
| 07:10:21 | EnterRoom，**np 1→2** |
| 07:11:27 | Walking + Initialization |
| 07:15:03 | EnterRoom，**np 2→3** ★ |
| 07:15:22-23 | ExitRoom，np 3→2 |
| 07:15:53-54 | ExitRoom，np 2→1 |
| 07:15:59 | EnterRoom，np 1→2 |
| 07:16:09 | EnterRoom，**np 2→3** ★ |
| 07:16:26 | ExitRoom，np 3→2 |
| 07:17:16 | np 2→1 |
| 07:18:33-34 | EnterRoom，np 1→2 |
| 07:19:19/26 | 连续 ExitRoom，np 2→1→0 |
| 07:19:35 | np 0→1（无 EnterRoom 配对，疑 ghost 出生） |
| 07:22:04 | EnterRoom，np 1→2 |
| 07:23:54-55 | ExitRoom，np 2→1 |

## 期望 verdict（待 engine 跑出）

回放后应关注：
1. np=3 那两段（07:15:03、07:16:09 起）是真三人还是 ghost 计数？
2. 07:19:35 的 np 0→1 没有 EnterRoom 事件 —— 是真人未触发 Enter 还是 ghost 出生？
3. 多次 EnterRoom/ExitRoom 频繁切换是否对应同一个真人在边界反复进出（典型 ghost
   信号）？

## 重新导出

```bash
./scripts/export_case.sh 9923003AB17F \
  "2026-04-29 07:10:00" "2026-04-29 07:30:00" \
  --tz America/Denver 9923-kitchen-0429
```

**`--tz` 必填** —— 服务器 $TZ 是 PDT，设备在 Denver MDT，差 1 小时。
