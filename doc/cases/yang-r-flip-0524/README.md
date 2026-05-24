# yang-r-flip-0524 — Yang.R bed FSM 5-7s 周期反转

## 现象

2026-05-24 00:36-00:42 UTC，5oi0n5 卡 (Yang.R unit `fd00:0:3:112:3::/80`) 的 FE
显示 bed_status 反复在 "not in bed" 与 "awake (in bed)" 之间翻转。

时间点：**正是 Tier 2.A.8 + 2.C commit `3d823e3` 部署 (sensor restart 00:36:36) 之后**。
此前 Tier 2.A `e8a1297` 部署后 (00:21:55) 仅有 1 次真实 InBed (00:22:01) 后稳定。
新 commit 部署后立刻开始 5-7s 周期翻转。

## 设备拓扑

| Device | Addr (/128) | DeviceType | 设备身份 |
|---|---|---|---|
| sleepad | `fd00:0:3:112:3:101:2460:1641` | Sleepad | BM87224601641 |
| radar (bedroom) | `fd00:0:3:112:3:100:32a1:cd2b` | Radar | 9D8A32A1CD2B |
| radar (bathroom) | `fd00:0:3:112:3:200:59b8:333b` | Radar | 25A859B8333B |

派生 spatial:
- 床 /96: `fd00:0:3:112:3:101::/96` (only sleepad)
- 卧室 /88: `fd00:0:3:112:3:100::/88` (radar bedroom)
- 卫生间 /88: `fd00:0:3:112:3:200::/88` (radar bathroom)
- 单元 /80: `fd00:0:3:112:3::/80` (Yang.R card aka 5oi0n5)

## 关键观察 (sensor_derived.jsonl)

sensor:derived bed.state subject = `fd00:0:3:112:3:101::/96` 时间序列（部分）：

```
00:36:25.606  bed.state    bs=1 be=0 tn=0 bc=0   ← initial publish (sensor restart 1)
00:36:29.601  bed.state    bs=0 be=0 tn=1 bc=0   ← FLIP Occupied (无 sleepace event!)
00:36:32.407  bed.state    bs=1 be=0 tn=0 bc=0   ← restart 2 initial publish
00:36:36.354  bed.state    bs=1 be=0 tn=0 bc=0   ← restart 3 initial publish
00:36:40.349  bed.state    bs=0 be=0 tn=1 bc=0
00:36:48.349  bed.state    bs=1 be=1 tn=0 bc=0   ← LeftBed transition
00:36:53.348  bed.state    bs=0 be=0 tn=1 bc=0   ← InBed
00:36:58.349  bed.state    bs=1 be=1 tn=0 bc=0   ← LeftBed
00:37:05.348  bed.state    bs=0 be=0 tn=1 bc=0
00:37:09.348  bed.state    bs=1 be=1 tn=0 bc=0
...（每 5-7s 一次翻转，持续）
```

**`bc=0` (bed_confidence=0)** 反常：translator.go `bedConfidenceForSource` 仅
sleepace=90/radar=60；为 0 表示 LastSource ∈ {"vital", "invariant_repair",
"subset_invariant", "decay"} 等非 sleepace/radar 源。

## 输入侧（iot:event:stream）

5min 窗口内 Yang.R unit 全部事件 (python 精解析):

| device | category | count |
|---|---|---|
| sleepad 1641 | InBed | **1** (00:22:01.992，14 min 前) |
| sleepad 1641 | sleep-stage | 2 (00:22:01.994 + 1 后续) |
| radar cd2b | activity | 59 |
| radar cd2b | number_people | 12 |
| radar cd2b | EnterRoom | 4 |
| radar cd2b | ExitRoom | 4 |
| radar 333b | activity | 59 |
| radar 333b | number_people | 11 |

**关键：radar 完全 0 个 InBed/LeftBed** —— Yang.R 卧室 layout `beds: 0`，firmware
没法判 bed-area，不发 bed-event。

## 物理输入侧 (iot:monitor:stream → monitor_stream.jsonl)

1641 sleepad.track 5min 窗口 163 条，HR 58-70 (有效)：

```
00:22:01  HR=60 RR=?  (与 InBed event 同 ts)
00:22:05  HR=60 turn_over=1
00:22:15  HR=58
00:22:25  HR=68 turn_over=1
00:22:35  (body_move only, 无 HR)
00:22:37  HR=70
...
```

VitalSource.hrRRPresent 需 HR>0 **AND** RR>0；payload 截断未确认 RR 是否同时
present。**关键开放问题**：sustain 路径是否真正激活。

## 输出侧 (sensor:derived → cardagg → Redis)

Redis `card:state:fd00:0:3:112:3::/80`.bed_state 当前快照 (见 redis_snapshot.txt):
```json
{
  "bed_status": 0,
  "bed_status_ts": 1779583233348,    // 00:40:33 UTC (实时刷新)
  "bed_confidence": 0,                // ⚠️ 不是 90 也不是 60
  "sleep_stage": 1,                   // 00:22:01 时的 stale 值
  "sleep_stage_ts": 1779582121994,
  "sleep_confidence": 90              // 00:22:01 时的 stale 值
}
```

bed_status_ts 持续刷新 (cardagg state-change-anchored merge 看到值变化)，
sleep_stage 卡在 14 min 前的 ts (无新 sleep-stage event 进 ladder)。

## 候选根因 (待验)

1. **vital 持续 sustain ↔ Tick decay 的振荡**：
   - vital adapter 4s 周期；HR 有效 → emit sustain
   - 5-7s 周期匹配 vital 失效 (HR/RR 不同时出现 → hrRRPresent false → sustain 缺席) 与 score decay 越过 exit_threshold
   - 验证：看 vital_source 是否对单独 HR (无 RR) 的 track 跳过

2. **repairSubsetInvariant 触发**：
   - Tick 内的 repairSubsetInvariant：bed.IsPresent + room.Vacant + bed stale → force bed Vacant
   - Yang.R bedroom radar cd2b 发 ExitRoom 时 /88 room 转 Vacant → bed FSM 被 force Vacant
   - 但下一个 sleep-stage event 来又 re-flip Occupied
   - 验证：看每次 flip 前后是否伴随 room transition

3. **scorer enter/leave 异步 latch 冲突**：
   - 单源 sustain 不足以维持，但又不算 leave evidence
   - 中间态在 entry/exit threshold 之间循环越界

## 与 Tier 2 refactor 关系

Tier 2.A 拆 CardID 后 zone FSM 从 (CardID, ZoneType, ZoneID) 三元组变 (ZoneType, ZoneID)
二元组。预期改进是"同 /96 床多源 sleepad+radar 共享 FSM"。

但 Yang.R 现状是 radar 不在 bed-/96（radar 在 room-/88-/96=`112:3:100::/96`），
sleepad 单独在 bed-/96=`112:3:101::/96`。**没有多源 fusion**，所以 FSM 收敛不来。
单源 (sleepad) 在 sleep-stage 稀疏 + vital 间歇时不稳定。

之前 Tier 2.A 部署的窗口 (00:21:55-00:36:25) 也是单源，为什么没翻转？
- 00:22:01 sleepad InBed (硬 evidence Source=sleepace 强 score)
- 后续 14 min 每分钟 60s tick 复发 cached state (不重新走 Apply)
- 没新事件 → 没新 Apply → 不会 re-evaluate → cached 持续 publish bs=0 bc=90 ✓

**00:36:25 sensor restart 后**：
- FSM 状态丢失，初始化为 Vacant
- 1641 已经 14 min 没发新 InBed（已在床但 sleepace 不发周期 InBed）
- vital sustain 4s 周期勉强维持 Occupied，但偶发缺失 → 卡在阈值边缘震荡

## 数据 fixtures

| 文件 | 说明 | 大小 |
|---|---|---|
| `event_log.jsonl` | event_log 表 (0 rows — wisefido-iot 在 v2 cutover 后未生效?) | 0 |
| `monitor_stream.jsonl` | monitor_stream 表，包含 1641 HR/RR | 760KB / 2348 rows |
| `alarm_events.jsonl` | alarm_events 表 (4 rows) | 191B |
| `sensor_derived.jsonl` | Redis sensor:derived:stream Yang.R /96 bed.* (python 解析) | 67KB / 112 entries |
| `redis_snapshot.txt` | Redis card:state hash + device:status snapshot | 1.9KB |
| `sensor.log` | sensor.log slice 00:20-00:45 含 Yang.R | 21KB |
| `sleepace.log` | sleepace.log slice 同窗口 | 815B (仅 connectionStatus) |
| `cardagg.log` | cardagg.log slice 含 Yang.R | 304B |

## 重现步骤（理论）

1. `bedside_grace_sec/lost_*_silent_sec` 参数 + Tick decay rate 已知
2. monitor_stream.jsonl 时间戳重放给 sensor + sleepace InBed @ ts=0 only
3. 看 sensor 是否在 ~14 min 后开始 5-7s 周期翻转

## 后续工作

- [ ] 检查 vital_source HR/RR present 判定：单独 HR (无 RR) 算不算？
- [ ] 看 zone_rules.yaml entry/exit_threshold + decay rate
- [ ] 加 sensor 调试 log：FSM transition 时打 Source/Score/Decision
- [ ] 单元测试：覆盖 sustain-only 路径的稳定性
