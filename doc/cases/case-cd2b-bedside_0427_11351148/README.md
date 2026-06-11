# Lost Fall 测试 — Radar_CD2B + Sleepad_1641 11:35-11:48 MDT

设备：
- Radar `9D8A32A1CD2B` (Radar_CD2B) — `device_id=ac790ba1-ac17-47d1-9cd1-c65848bd0027`
- Sleepad `BM87224601641` (Sleepad_1641) — `device_id=ad16fdff-383b-444a-b7c0-d2c021f97f7b`，绑定本房间床位

房间 `b368ee77-0a13-4584-9ce1-bbcff6205e48` (Bedroom)，单元 `3db05d02-1257-407d-88bb-cbe9a7498b27`，
tenant `43b8fbf7-b55f-4b48-bd8b-27bb14f48870`。

## 文件清单

| 文件 | 内容 |
|---|---|
| `room_layout.json` | rooms.layout_config raw（walls / Bed / Enter / radar mount） |
| `2026-04-27_11-35_to_11-48_MDT.json` | Radar 559 条（534 monitor / 24 event / 1 alarm）|
| `sleepad_BM1641_2026-04-27_11-35_to_11-48_MDT.json` | Sleepad 同窗口全部记录 |

时间窗：
- **MDT 11:35:00 - 11:48:00**（本案所有时间 = MDT，下同）
- 服务器存的是 epoch ms `bigint`：1777311300000-1777312080000
- 等价 UTC 17:35-17:48（postgres server timezone = UTC，查询时需 `AT TIME ZONE 'America/Denver'` 才显示 MDT）

## 现场场景

人**进入卧室 → 上床 → 离床（疑似跌倒）→ 雷达看不到 5.5 分钟 → 重新出现**。

由于床高遮挡 + 人离床后落地角度，雷达完全失锁；firmware 持续广播最后一次"真实"track 5.5 分钟后才放弃。
全程 **alarm_events 表无 Fall 报警**（漏报场景）。

## 双源 Timeline（关键事件，全部 MDT）

| 时间 (MDT) | 来源 | 事件 / 状态 |
|---|---|---|
| 11:35:00 | radar | 帧流开始（人尚未进房） |
| **11:37:03** | **radar** | **EnterRoom start** — 人走进卧室 |
| 11:37:03-11:37:35 | radar | 走到床边、躺下，pose 4→1→6 转换正常 |
| 11:37:40 | sleepad | **InBed start + instant** — 床压感感测到人 |
| 11:38:20-11:39:04 | radar | 床上躺姿 frozen 帧（45s 完全相同 `(10,200,0) pose=6 tc=60`）|
| 11:39:05-11:39:35 | radar | 床上微动 → 起身 → 站立 → 走出床外 |
| **11:39:35** | **sleepad + radar** | **同秒：sleepad LeftBed start + radar 最后一次"真实"track** `(-90, 320, 0) pose=4` |
| **11:39:35-11:45:06** | **radar** | **🔴 5.5 分钟 frozen：334 帧完全相同 `(x=-90, y=320, z=0, pose=4, tc=60)`** |
| 11:40:25 | sleepad | OfflineRecover end |
| 11:43:55 | radar | OfflineRecover / SignalPoorRecover / AngleExceptionRecover (firmware 状态机 end，**非真 alarm**)|
| 11:45:07 | radar | 切到 `tid=88`（heartbeat 无人帧）—— **firmware 放弃 tracking** |
| 11:47:08-11:47:22 | radar | 人重新出现，从 (-240, 220) 移动到 (-410, 0) |
| **11:47:25** | **radar** | **ExitRoom start** —— 人离开房间 |

## Frozen 帧的物理含义

雷达 firmware 在 11:39:35 看到人最后位置是 `(-90, 320, 0) pose=4` —— 这个位置在**床尾下方/床外**（床中心约 y≈200）。
之后 firmware 失锁：
- **不愿意立即丢弃 track**（怕 1-2 帧抖动），所以**保持发布最后位置**
- **持续 5.5 分钟**，每秒发 1 帧完全一致的内容（含 tc=60 不变）
- 5.5 分钟超时后才切到 tid=88（无人）

人在这 5.5 分钟里：
- 没有触发 ExitRoom（11:39:35 - 11:47:25 之间无 ExitRoom 事件）→ 人**没出房间**
- sleepad LeftBed → 人**确实离床**
- 11:47:08 重新出现在 (-240, 220) 走动 → 人是从某个雷达盲区走回来的（卫生间？床的另一侧？）

**两种可能**：
1. 人离床 → 倒地（盲区）→ 自己爬起来 → 走出盲区
2. 人离床 → 进卫生间（盲区）→ 出来

但本案 alarm_events 无 Fall，handle_type 也无 false_alarm 标记，需要用户判定到底发生了什么。

## 设计含义（讨论中）

### 1. 雷达 frozen-frame 是强 lost-fall 信号

**特征**：连续 N 帧（N≥30）`(x, y, z, pose, tc)` 完全一致 + 无任何 sleepad/door 事件支持「人没动」。

与「真静止」区别：真静止时 engine 会看到 cell.ActiveType[Stand/Lie] 累计 + 偶尔的 quality 抖动；frozen 帧 quality 完全不变（tc=60 一动不动）。

### 2. Lost-fall 等待时间应 > radar 自身 frozen 超时

当前设计：lost fall walkway 5min。但 radar firmware 的 frozen → 放弃超时本身就是 ~5.5min。
所以我们的 5min 阈值在**radar 还在「装作能看到」**的时候就触发了，可能误报；建议：

- **从 frozen-stop（tid=88）那一刻**起算等待，而不是从 LeftBed/最后真实 track 起算 → 此时已经过了 5.5min，再等 2-3min 即可（总 7-8min）
- 或：直接消费 frozen-pattern 作为「lost 起点」，从 frozen 出现连续 30s 之时起算（~5.5min 后超时报警），保守等到 8min

### 3. 「人重新出现」要取消 pending

本案 11:47:08 人重新出现 → 应取消任何 pending lost-fall。
触发：track 在 lost-fall pending 期内有新 track 出现（不限位置，因为人可能从盲区走回）。

### 4. ExitRoom 事件做反证

本案没有 ExitRoom 事件支持「人走出去了」，所以 lost 假设成立。
如果 frozen 期间有 ExitRoom，应取消 pending（人正常离开）。

## 用途

回归 fixture，用于：
- Lost fall 规则触发条件实测验证
- Frozen-frame 检测算法的 ground truth
- 双源融合（radar frozen + sleepad LeftBed）作为 lost-fall 触发组合
- Cell history integral：本段 InBed 期间无 alarm 应累计 ToleratedStillCount

## JSON Schema

```
[
  {
    "id": <bigint>,
    "device_uid": <string>,
    "device_id": <uuid>,
    "timestamp": <epoch ms>,
    "topic_type": "monitor" | "event" | "alarm",
    "category": <string>,
    "data_value": <jsonb array of frame objects>
  },
  ...
]
```

monitor frame 字段（`event_name="track"`）：track_id / pose / position_x / position_y / position_z / track_confidence / bed_status / area_id / remaining_time / dataCategory
