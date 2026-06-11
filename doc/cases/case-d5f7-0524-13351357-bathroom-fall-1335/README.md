# d5f7-bathroom-fall-1335 — John.Y bathroom 误报 Fall（suite 缺 bathroom card）

## 现象

2026-05-24 13:35-13:57 (America/Denver)，John.Y unit (`fd00:0:3:111:3::/80`)
bathroom radar `E598A2ACD5F7` 完整记录一次入厕。FE 看到 Fall 提示，
但同一时刻 radar pose 仍是 sit (3)，已经是误报。

## 设备拓扑

| Device | Addr (/128) | DeviceType | UID |
|---|---|---|---|
| bathroom radar | `fd00:0:3:111:3:300:a2ac:d5f7` | Radar | E598A2ACD5F7 |
| bedroom radar  | `fd00:0:3:111:3:100:a2ac:d523` | Radar | (D523) |
| sleepad        | `fd00:0:3:111:3:101:3263:9e7`  | Sleepad | (9E7)  |

派生 spatial:
- 卧室 /88: `fd00:0:3:111:3:100::/88`  ✅ 有 card  (`aolr6c`, has_bed)
- 床    /96: `fd00:0:3:111:3:101::/96`
- ?     /88: `fd00:0:3:111:3:200::/88`  ✅ 有 card  (`f61e8o`, has_bed)
- **卫生间 /88: `fd00:0:3:111:3:300::/88`  ❌ cards 表无此 row**
- 单元 /80: `fd00:0:3:111:3::/80`  ✅ John.Y card (`j0yvp3`, has_bathroom=true)

## Root cause 候选

bathroom_gate 在 entry 时拒发 EnterRoom，原因 `suite_missing_or_public`：

```
sensor.log:3 (19:36:07.074Z = 13:36:07 MDT)
  msg: bathroom_gate_entry_noop
  suite_id: fd00:0:3:111:3:300::/80
  room_id:  fd00:0:3:111:3:300::/88
  reason:   suite_missing_or_public
```

bathroom 在 cards 表无对应 /88 row（只在 unit /80 上挂了 has_bathroom=true 标志）。
gate 找不到 suite 上下文 → 整路 entry 链断：
- event_log 无 EnterRoom（`event_log.jsonl` = 0 行）
- alarm_events 无任何 Fall/Stay/LeftBed/NightAbsence (`alarm_events.jsonl` = 0 行）
- cardagg `room_state.total_people = 0`，`bed_state.bed_status = 1`（NotInBed）

但 radar 本身仍在 1Hz 发 `radar.track`（`monitor_stream.jsonl` 1288 行），
FE 直接读 raw track 看到的 pose 翻转 → 显示 Fall。

## 时间线 + 5/8/15/22min 检查点

逐分钟 pose 直方图（`pose_timeline_per_minute.tsv`）：

```
13:30-13:35  pose=NULL          ← 无人
13:36        pose=1×11 + 3×41 + 4×2   ← 入厕（stand 短暂 → 落座）
13:37-13:55  pose=3 ≈60/min     ← 持续 sit（坐马桶 ~19min）
13:56        pose=1×2 + 3×57 + 4×2    ← 起身
13:57        pose=1×1 + 4×32 + NULL×4 ← 离开 + 末段 pose=4
13:58-13:59  pose=NULL          ← 离开后
```

`checkpoints_5_8_15_22min.tsv` 摘录每个检查点头 5 帧：

| +Δt   | local        | track_id | pose | (x,y,z)        | conf |
|-------|--------------|----------|------|----------------|------|
| +5m   | 13:40:00.834 | 0        | 3    | (0, 50, 46)    | 80   |
| +8m   | 13:43:00.308 | 0        | 3    | (50, 40, 0)    | 80   |
| +15m  | 13:50:00.956 | 0        | 3    | (20, 10, 0)    | 80   |
| +22m  | 13:57:00.605 | 0        | **1→4** | (-80, 180, 0) | 80   |

> 注：area_id=255（declared area 之外，正常），track_id 全程 = 0 没切换，
> 位置在很小范围内抖动（马桶半径 ~50cm）符合 sit 模式。

## Sensor verdict = ghost ×2

整段窗口内 sensor track_manager 两次给 d5f7 出 ghost 判决：

```
sensor.log:6,7 (19:48:43.628Z = 13:48:43 MDT)
  track_verdict_ghost
  device_uid: fd00:0:3:111:3:300:a2ac:d5f7  track_id: 0
  verdict: ghost  score: 19  birth_score: 50
  reason: ghost_history_zone   (x=-20, y=0)
  ai_emit → ai:track:verdict:stream  track_confidence=90

sensor.log:8,9 (19:54:08.026Z = 13:54:08 MDT)
  track_verdict_ghost
  device_uid: fd00:0:3:111:3:300:a2ac:d5f7  track_id: 0
  verdict: ghost  score: 17  birth_score: 50
  reason: low_score   (x=0, y=10)
  ai_emit → ai:track:verdict:stream  track_confidence=100
```

— 即使 gate 走通，sensor 也会把这段判为 ghost；说明该 cell 历史样本里
"长时间近原点静坐" 已被学成 ghost 区。

## Cardagg 视图（13:54:47 redis snapshot）

```
display
  section1_down_left=Bathroom  section2_left_icon=0  scene_state=0
bed_state
  bed_status=1                  ← NotInBed (memory: bed_status 默认 NotInBed)
room_state
  room_id=fd00:0:3:111:3:300::/88  total_people=0   ts=1779652653 (13:57:33)
```

cardagg 整路认为无人；FE 显示 Fall 不是来自 cardagg 的 alarm 流，
而是 FE 直读 `monitor_stream` raw pose（pose=4 == lying）的派生显示。

## 文件清单

| 文件 | 含义 |
|---|---|
| monitor_stream.jsonl | 1288 行 radar.track + radar.heart，d5f7 /128，13:35-13:57 全窗口 |
| event_log.jsonl | 直 device_addr 查 = 0 行（确认 sensor 未发任何 owl-common event） |
| event_log_bathroom.jsonl | /88 bathroom scope = 0 行 |
| event_log_unit80.jsonl | /80 unit scope = 0 行（除 deviceStatus 心跳，不在窗内） |
| alarm_events.jsonl | /80 unit scope = 0 行（confirm 无任何 alarm 触发） |
| sensor.log | 9 行：bathroom_gate_entry_noop + 2× track_verdict_ghost + 4 旁路噪声 |
| cardagg.log | 6 行：display_rebuilder 周期 tick，无业务行为 |
| pose_timeline_per_minute.tsv | 38 行：每分钟 pose 直方图 |
| checkpoints_5_8_15_22min.tsv | 22 行：4 个检查点各取头 5 帧 |
| rooms.jsonl | 3 行：unit80 下 rooms 行 |
| cards.jsonl | 3 行：cards 行，**注意缺 `fd00:0:3:111:3:300::/88`** |
| redis_card_state_unit80.txt | card:state hash snapshot |
| redis_card_realtime_unit80.txt | card:realtime hash snapshot |

## 待查 / 下一步

1. **cards 表为何没建 bathroom /88 row** — 是 import 时漏建，还是 has_bathroom flag
   设计本就只在 unit 层挂？查 card 创建链路 + bathroom card_id_kind 规则
2. **bathroom_gate "suite_missing_or_public" 应否兜底** — 若 cards 缺行是预期形态
   （unit-level has_bathroom），gate 该用 unit /80 当 suite_id，不该 noop drop
3. **FE 是否在读 raw pose 派生 Fall** — 若是，与 cardagg/sensor 决策脱钩 → 砍
   或改读 cardagg verdict；这是 [[feedback_no_dynamic_threshold_modulation]] 类问题
