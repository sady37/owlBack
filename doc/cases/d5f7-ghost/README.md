# D5F7 浴室 Ghost 案例

设备 `E598A2ACD5F7` (Radar_D5F7)，房间 `265665a6-1c48-4d55-9ced-665bec73e9a2` (Denver-LakeW-101 bathroom)，300×150cm 小浴室。

雷达天花板装在 (0, 240, 270cm)，FOV 边界 leftH=190 rightH=120 frontV=70 rearV=90，墙壁 h=-120~180 v=-80~70。室内 furniture：shower 区、mirror、curtain、bath 矩形。

## 文件清单

| 文件 | 内容 |
|---|---|
| `room_layout.json` | rooms.layout_config raw（含 walls / Enter / Furniture / Curtain / Interfere / radar mount） |
| `2026-04-25_17-20_to_17-35_MDT.json` | Case A 全量 iot_timeseries (monitor + event + alarm) |
| `2026-04-26_12-35_to_12-55_MDT.json` | Case B 全量 iot_timeseries |

JSON 结构：`[{id, device_uid, device_id, timestamp(ms), topic_type, category, data_value(jsonb)}, ...]`

## Case A: 04-25 17:26-17:28 — firmware track-ID 切换 ghost (~2min)

人进浴室 → 走动 → 雷达内部 tracking 失锁 → 同一真人被 split 成 2 个 track ID。

### Timeline

| 时间 (MDT) | 事件 | 位置 (h, v) | pose | z | 备注 |
|---|---|---|---|---|---|
| 17:26:17 | track 0 出生 | (-50, 80) | 4 | 99 | 在 Enter 区 (-120~-40, 60~80) 进房 |
| 17:26:17-37 | track 0 走动 | → (160, -20) | 4→1→3 | 0-83 | 真实运动，pose/z/位置都在变 |
| 17:26:39-41 | track 0 到右侧 | (170, -30)→(160, 0) | 4 | 0-123 | 接近右墙 h=180 |
| **17:26:42** | **track 0 冻结** | (190, -50) | 4 | 0 | **位置完全不变 90+ 秒**，且 h=190 已超出右墙 |
| 17:26:42 | **track 1 出生** | (40, 0) | 4 | 0 | 房间中央凭空出现（非 Enter 区） |
| 17:26:48-59 | track 1 走动 | → (0, 50) | 1 (Walk) | 33-88 | **连贯真人轨迹** |
| 17:27:00-08 | track 1 收尾 | (-50, 60) | 4→3 | 0 | Stand 后 Sit |
| 17:27:08+ | track 1 消失 | / | / | / | 出门 |
| 17:27:08-17:28:05+ | track 0 仍冻结 | (190, -50) | 4 | 0 | 假死另一分钟 |

### Ghost signal

1. **Track 0 在 17:26:42 那一刻位置突然冻结**（前 25 秒位置变化 200+cm，后 90 秒位置变化 0cm）
2. **Track 1 在 17:26:42 同一刻凭空出生**（房间中央，非 Enter 区）
3. **Track 0 冻结位置 (190, -50) 在右墙之外**（墙壁 h=180）
4. 两 track 时间反相关：track 0 停 = track 1 始

### 关键 monitor 帧（提取自 fixture）

```
17:26:39  tid=0 (170,-30) pose=4 z=0
17:26:40  tid=0 (170,-20) pose=4 z=123
17:26:41  tid=0 (160, 0)  pose=4 z=0
17:26:42  tid=0 (160,-10) pose=4 z=66    ← 最后一次正常更新
17:26:42  tid=1 ( 40,  0) pose=4 z=0     ← track 1 出生
17:26:44  tid=0 (160, 10) pose=4 z=0     ← 开始 jitter 范围扩大
17:26:48  tid=0 (180,-20) pose=4 z=0
17:26:51  tid=0 (190,-50) pose=4 z=0     ← 冻结开始
17:26:53→ tid=0 (190,-50) pose=4 z=0     ← 持续冻结直到 17:28:05+
```

## Case B: 04-26 12:41:17-12:47:51 — 6.5min 持续双 ghost

人进浴室静止做事（如刷牙/淋浴），雷达 firmware 长时间无法淘汰错误 track，6+ 分钟"两人"同时存在。

### Event-stream timeline

| 时间 (MDT) | event_name | data |
|---|---|---|
| 12:37:05 | EnterRoom + np=1 | 1 人进入 |
| 12:37:12-41:10 | activity tc=1 | 1 人活动，walk_distance/duration > 0（真实） |
| **12:41:17** | **np: 1 → 2** | 第二个 track 出现 |
| 12:42-47 | activity tc=2 | **walk_duration / stand_duration / walk_distance 全 0**（6 分钟两人都不动） |
| 12:47:51 | ExitRoom + np: 2 → 1 | 一个 track 消失 |
| 12:48:07 | activity tc=1 stand=17 | 剩下的 track 真实活动 |
| 12:48:27 | np: 1 → 0 | 全部离开 |

### Ghost signal

1. **6 分钟内 walk_duration / stand_duration / multi_person_duration = 0**：真双人不可能 6 分钟两人都没站立没走路时间累计
2. **EnterRoom 只触发了 1 次**（12:37:05），但出现了 2 个 track —— 第二个不是从门口进的
3. 离开顺序：`np 2→1` 但不带 ExitRoom（12:47:51 那次有 ExitRoom，但应该是 1→0 才对，2→1 没 Exit 表明是 ghost 自然消失）

### 用法

```bash
# 加载 fixture 跑回放（未来）
go test ./internal/roomengine -run TestD5F7Ghost \
  -fixture doc/cases/d5f7-ghost/2026-04-26_12-35_to_12-55_MDT.json
```

或 HTTP 快速看 verdict：
```
http://127.0.0.1:7788/api/playback?uid=E598A2ACD5F7&start=2026-04-25T17:20:00-06:00&end=2026-04-25T17:35:00-06:00&format=json&log_verdicts=1
```

## 检测算法待开发要点（基于这两个 case）

A 的破绽：**位置完全冻结 + 房间几何越界**  
B 的破绽：**多 track 共存但 walk/stand 时间全 0**  

共性：firmware 给的 track_count > 1 但**第二个 track 缺乏运动学证据**。Engine 应该独立维护 track 评分，不能信任 firmware 的 track ID 划分。
