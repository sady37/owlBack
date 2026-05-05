# 4D8710D5CABB 24h Fall + Ghost 综合分析（2026-05-03 ~ 05-04）

设备 `4D8710D5CABB` (Radar_CABB)，bound_room `0af07f65-b4f3-465b-bbcf-21d2abc7f7a6`，绑卡 `9ad06295-9a2c-49de-8917-ed37bf77684a` (ActiveBedCard "Hun.Z – BedA" / `Denver-LakeW-C2601`)，与 Sleepad `BM87225200672` 同卡。

- firmware: `2.3-Jun 25 2025 11:33:44`
- mcu: `2.0-Apr 29 2026 17:11:23`
- **不是 D523 firmware** → `number_people=0 ExitRoom 兜底`规则不适用本设备

24h 内出现 3 起 Fall alarm + 多段 Ghost 嫌疑 track。replay 重现不出原因，本文件汇总 forensics + 6 份 case fixture 的导出索引。

---

## Part 1: 三起 Fall alarm 概况

| 标识 | triggered (MDT) | source | score | cell_area | spatial_jump | 处置 | 触发模式 |
|---|---|---|---|---|---|---|---|
| **A** | 2026-05-04 00:16:51 | sensor.caregiver01 | 96 | 4 (sensing) | false | verified | frozen-track lost-fall (真人 → z 失真 → 卡死) |
| **B** | 2026-05-03 19:47:01 | AI.Caregiver01 | 98 | 0 (none)    | false | verified | health-tick 后 ~4min commit pending |
| **C** | 2026-05-03 18:03:19 | AI.Caregiver01 | 98 | 4 (sensing) | **true**  | false_alarm | health-tick 后 5s commit + spatial_jump |

3 起共性：`reason=lost_track`，`evidence.context=track_lost_no_exit_room_no_recovery`，`last_verdict=1`。

### A 事件 — frozen-track lost-fall（最有价值的 ghost-fall 因果案例）

`frozen_start_ms = 1777874887556 = 2026-05-04 00:08:07.556 MDT`。track 数据时序：

```
00:07:33-00:07:45  pose=4 (stand)  位置 (-10, 0, 0~107)            人在站
00:07:46-00:08:05  pose=1 (lying)  位置 (0~50, -50~50, 60~125)     人躺下，z 还正常
00:08:06           pose=1          位置 (-30, -50, 0)              z 首次跌 0
00:08:08-00:08:14  pose=1          位置 (~0, -50~-70, 0)           z 持续 0
00:08:15-∞         pose=4 (stand)  位置 (0, -60, 0)                z=0 + 位置完全冻结
                                                                    pose 翻成站立但 z=0 矛盾
```

**真因**：人在 00:07:46 躺下 (pose 4→1, z=124→63)，雷达 z 通道质量劣化 → 00:08:06 z 失真为 0 → 00:08:15 firmware 把这条 frozen track 的 pose 改回 4 但 (x,y,z)=(0,-60,0) 卡死不动 → AI engine 看作"站立 + 不动 + lost-track" → 86s wait_ms 走完 → fire lostfall。

`number_people=0` 兜底事件在 00:12:59 才到达（frozen_start 后 4:52），但本设备非 D523 firmware，兜底规则不适用，没拦住。最终 00:16:51 fall fire（相对 frozen_start +8.7 min）。

trigger_data.position 是 (0, 100, 0) — 与冻结位置 (0, -60, 0) 不一致 → engine verdict 内部还有另一份 reference position。

### B 事件 — health-tick 后 4 min commit

`triggered=19:47:01.889 MDT`, `frozen_start=19:40:18 MDT`, source=`AI.Caregiver01`, cell_area=0 (no cell training)。

时序对照：
```
19:43:14   alarm topic 三连发：OfflineRecover + SignalPoorRecover + AngleException event_status=start
19:47:01   Fall fire
```

19:43:14 那组三连发是 **qinglan 10-min health_check ticker** 的产物（[wisefido-qinglan/internal/subscriber/health_check.go:40](../../wisefido-qinglan/internal/subscriber/health_check.go#L40)），**不是真实的 offline/recover** —— 但 AI engine 看到 OfflineRecover 后按 reconnect 处理，触发 reconcile，4 分钟内把之前 frozen 的 lost-fall pending commit。

### C 事件 — health-tick 后 5 秒 + spatial_jump=true

`triggered=18:03:19.901 MDT`, `frozen_start=17:58:24 MDT`, position=(-230, 490, 0)，cell_area=4。

```
18:03:14   alarm topic 三连发（10-min health tick）
18:03:19   Fall fire           （仅 5 秒后）
```

position_y=490cm = 4.9m，远超雷达正常工作范围。`spatial_jump=true` 是引擎自标记的"位置突变不可信"。被人工判 false_alarm 是正确的。

---

## Part 2: 雷达基线状况

### AngleException 长期未 recover

alarm topic 显示 17:53 ~ 22:33 期间每 10 分钟稳定出现一组：
- `OfflineRecover`（健康"未离线"）
- `SignalPoorRecover`（健康"未弱信号"）
- `AngleException event_status="start" event_value=1 angle_abnormal=1`（**始终在报角度异常**）

按 [publishHealthIfChanged](../../wisefido-qinglan/internal/subscriber/health_check.go#L191) 的 transition-only 去重，AngleException onset 只应在 0→1 跳变时发一次。**重复发 = `prevHealth` 内存被反复清空 = qinglan 频繁重启**（user 已确认本周期重启是主动重启）。

后果：
1. AI engine 把每次 OfflineRecover 当 reconnect 信号 → reconcile 触发 → 把 frozen pending commit 成 alarm
2. AngleException 始终 onset 不 recover，cell-level 可信度阈值受累积影响
3. alarm_events 表灌入大量 noise

### Ghost track 24h 统计

#### Track-ID 共存（硬证据）

5 段 `track_0 + track_1` 同时存在，全部短瞬态，总计 218 秒：

| 时间 (MDT) | 时长 | 模式 | Ghost 嫌疑 | fixture |
|---|---|---|---|---|
| 05-03 13:07:26 | 57s | 末段双方都 frozen z=0 (dist=168cm) | **高** (frozen ghost pair) | `cabb-ghost-pair-1307/` |
| 05-03 21:17:17 | 29s | track_0 frozen pose=3 z=0 30 帧 | **高** | `cabb-ghost-frozen-sit-2117/` |
| 05-03 22:30:24 | 57s | 双方都有运动 | 中 (可能真双人/分裂) | — |
| 05-04 04:15:10 | 41s | track_0 frozen pose=3 z=0 43 帧 | **高** | `cabb-ghost-frozen-sit-0415/` |
| 05-04 08:44:43 | 34s | 双方都有运动 + z>0 | 中 (可能真双人) | — |

#### Track-ID 总分布

| tid | 频次 | 解读 |
|---|---|---|
| 0 | 7,784 | 主 track |
| 1 | 1,399 | 第二 track（含 ghost）|
| 88 | 2,581 | firmware sentinel "no person"，**非 ghost** |

#### 单 track frozen-at-z=0（软证据）

A 事件本身就是这种模式：单 track 卡死在 (0,-60,0) z=0 数分钟。23:07:10-23:07:39 也观察到 `track_0 stuck at (-10,0,0) z=0` 30+ 秒。这是这台雷达 24h 内反复出现的 ghost 表现。

---

## Part 3: 因果链 + 为什么 replay 漏报

### 因果链

```
雷达 angle long-time abnormal (硬件/安装因素)
  │
  ▼
z 通道质量差，真人垂直方向定位失真
  │
  ▼
真 track z 跌 0 后被 firmware 冻结 (frozen ghost)
  │
  ▼
AI engine: 不动 + pose 突变 + track 不更新 → 标 lost-track
  │                            ┌──────────────┐
  ▼                            │ 同时 qinglan  │
积累 wait_ms 计数              │ 10-min health │
  │                            │ tick 发"假"   │
  │                            │ OfflineRecover│
  │                            └──────┬────────┘
  │                                   │
  └────── 引擎把 OfflineRecover 当 reconnect ────┐
                                                  ▼
                                           commit pending
                                                  │
                                                  ▼
                                              fire Fall
```

### Replay 漏报原因（5 项）

| 因素 | 说明 |
|---|---|
| ① replay 通常只读 `monitor` topic | 但 B/C 的真实 trigger 信号是 `alarm` topic 的 OfflineRecover 三连发，replay 看不到 |
| ② qinglan 重启状态不复现 | health_check `prevHealth` 被重启清空才会反复发 AngleException onset；replay 没"qinglan 进程重启"事件 |
| ③ AI engine in-memory frozen-track state | `frozen_start_ms` 是引擎内部 wall-clock 状态机，replay 重置后从干净状态开始，timing 对不上 |
| ④ angle_abnormal 长期=1 的累积影响 | 长期角度异常对 cell-level 阈值有自适应影响，replay 不复现累积态 |
| ⑤ A 的 86s wait_ms 内部计时 | wait_ms 不是 wall delta（A 实际 frozen_start→trigger = 8.7 min），是引擎内部 verdict 累积器，replay 状态机不对齐 |

---

## Part 4: 已导出的 6 份 case fixture

全部在 `doc/cases/`，全程 `--tz America/Denver`：

| 目录 | 时间窗 | 行数 | 用途 |
|---|---|---|---|
| `cabb-ghost-pair-1307/` | 2026-05-03 13:05–13:10 MDT | 323 | 双 track frozen-pair，47s 共存 |
| `cabb-ghost-frozen-sit-2117/` | 2026-05-03 21:15–21:20 MDT | 346 | track_0 frozen pose=3 z=0 30 帧 |
| `cabb-ghost-frozen-sit-0415/` | 2026-05-04 04:13–04:18 MDT | 309 | track_0 frozen pose=3 z=0 43 帧（最长） |
| `cabb-fall-A-frozen-0016/` | 2026-05-04 00:05–00:20 MDT | 495 | A 事件：真人躺下 → frozen ghost → lost-fall（**最重要**） |
| `cabb-fall-B-healthtick-1947/` | 2026-05-03 19:42–19:52 MDT | 43 | B 事件：health-tick 后 4min commit |
| `cabb-fall-C-spatialjump-1803/` | 2026-05-03 17:58–18:08 MDT | 98 | C 事件：health-tick 后 5s commit + spatial_jump=true |

每份目录含 `room_layout.json` + iot_timeseries 时间窗 JSON，结构与 `d5f7-ghost/` 一致。

---

## Part 5: 检测建议（按命中率优先级）

按 memory `ghost_detection_todos`：

1. **frozen-at-z=0 + position 不变 ≥ N 秒** → 标 ghost，prune track。**A 误报根因，命中率最高**
2. **track-ID 共存 + 任一 z=0 + 距离 > 100cm** → 标 ghost pair（13:07 / 21:17 / 04:15 三例命中）
3. **AngleException long-time onset 期间** → 全局降权该雷达 fall verdict 可信度（24h 都在报 angle_abnormal=1，引擎应当感知）
4. **OfflineRecover 后 N 秒内禁止 commit pending verdict** → 直接堵 B/C 这种 health-tick 假触发模式

附加：
- 已规划的 **birth-position 检测**对本台命中率有限（CABB 的 ghost 主要是 frozen 残留型，不是新生型）
- qinglan **prevHealth 持久化**（Redis or disk）能消除 AngleException 重发，独立于本 ghost 修法

---

## Part 6: 同会话发现 + 已落地的修复

### Bug A（已修）：qinglan 把 Sleepad 错打成 Radar

[wisefido-qinglan/internal/repository/postgres_device.go:880](../../wisefido-qinglan/internal/repository/postgres_device.go#L880) 的 `GetAllAccessibleDevices` SQL 缺 `device_type='Radar'` 过滤，导致 8 台 Sleepad 进入 qinglan subscription，每次重启 `publishOnlineForConnectedAfterStartup` 给它们发 OfflineRecover 并硬编码 `DeviceTypeRadar`，污染 iot_timeseries。BM87225200672 6 周累计 211 条 Radar 错标。

### Bug B（已修）：sleepace probe 风暴

启动时 config:card:stream backlog 28+ 条，每条 deviceID="" → 旧逻辑 fallback `scanAll` → 同一 UID 1s 内被探 17 次。
- [health_check.go:202-218](../../wisefido-sleepace/internal/subscriber/health_check.go#L202)：probeOne 加 per-UID 30s 闸门
- [health_check.go:170-200](../../wisefido-sleepace/internal/subscriber/health_check.go#L170)：scanAll/scanAllNow 拆分 + 30s 去重
- [health_check.go:126](../../wisefido-sleepace/internal/subscriber/health_check.go#L126) + [config_subscriber.go:90](../../wisefido-sleepace/internal/subscriber/config_subscriber.go#L90)：`ProbeAfterCardChange` 接受 `affectedUIDs` 优先精准探测

### 同卡 Sleepad BM87225200672 状况

- 自 2026-03-19 入库以来从未真正联网（`sleepace_report=0`、`monitor=0`、cardagg `last_seen_ms=0`）
- 6 周仅有 connectionStatus 翻转事件
- 4-22 ~ 4-26 触发 5 条 BedNightAbsence 假阳性（设备未联网造成）
- user 确认：**预配置设备，没有接入网络**

---

*导出时间: 2026-05-04*
*来源会话: 4D8710D5CABB 24h forensics + Bug A/B fix*
