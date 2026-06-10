---
name: cardagg-v1-to-v2-migration-audit
description: 2026-05-19 audit + 实施完成：cardagg v1 EventHandler 功能 v2 落点全表；S1-S6 全部已落地（commits f910ac3/12b72d8/74178af/04191ad/41d6146/9b19a84/1c107c9）；BedState 4 字段全归 sensor owner；v1 sleepad track override v2 自然不需要（S5c audit-only）；S6 设备类 alarm gate 防 SensorDetached 等污染 FSM
metadata: 
  node_type: memory
  type: project
  originSessionId: 0566c6c8-ddbd-4e2d-b8b6-79fef2a5a715
---

## v1 → v2 功能映射

| v1 cardagg EventHandler 路由 | v2 落点 | 状态 |
|---|---|---|
| EnterRoom / ExitRoom / NumberPeople → RoomState 派生 | sensor zoneengine adapter_radar.go + translator.TranslateRoomState | ✅ 已实施 |
| InBed / LeftBed → BedState/bed FSM 协调 | sensor zoneengine bed FSM + sensor:derived bed.state | ✅ 已实施 |
| LeftBed → pending alarm（45min 阈值 + rest_window + sleepad 跳过） | cardagg AlarmRouter | ✅ 已实施（旧路径仍在 cardagg） |
| Activity → RoomState walk/stand/multi_person | v2 RoomState 已删 walk/stand 字段；改为 TargetState.LastActiveTs/StandingContinuousMin per device，由 sensor aggregator（S2）写 | ✅ 设计变更，无需 cardagg expand |
| EnterRoom/ExitRoom → clear AIVerdict cache | cardagg ai_verdict_handler 缺（见下） | ❌ |
| EnterRoom/ExitRoom → ClearBedStateSleepStage | 见 SleepStage 融合行 | ❌ |
| SleepStage / SleepConfidence（**不是加权融合，是 confidence ladder 覆盖**）+ OOB drop + device_failure 报告 + LeftBed/ExitRoom/EnterRoom 触发 Clear | **归 sensor**：sensor 订阅 sleepace+radar SleepStage event；**直接取 event 的 SleepStage 值**（不算术加权）；SleepConfidence=60(Radar)/90(Sleepad)/100(双设备同步)；仅 `confidence ≥ current` 才覆盖。OOB 时（自家 zoneengine bed FSM = leaving/vacant）整段 drop + emit device_failure。Clear 路径由 sensor 自家 RoomState/BedState FSM transition 触发。写自家物理 BedState（sensor:derived:stream subject_entity=/96 bed CIDR）。**projector.go SleepStage/SleepConfidence 改 sensor owner**。S4 wire 即覆盖 | 🟡 等 S4 |
| track_verdict ghost cache → aiOverrides | sensor roomengine emit 到 iot:event:stream（已实施）；cardagg `ai_verdict_handler` consumer 缺 | ❌ 缺 cardagg consumer |
| Stay alarm（bathroom 武装窗 + activity 累计） | sensor zonealarm Supervisor | ✅ 已实施 |
| sleepad track_id override（InBed 时换 track） | 待审：v2 仍需要？v1 `sleepadTrackOverrideForInBed` 在 cardagg event_handler.go | ⚠️ |

## v2 BedState 字段 ownership 实际状况

**核心订正（用户 2026-05-19）**：SleepStage/SleepConfidence 在 v2 归 sensor 写自家物理 BedState；sensor_state_projector.go 注释里"SleepStage 非 sensor owner"是 v1 时代旧标记，**S4 wire 时一并订正**。

| 字段 | sensor 写? | cardagg 写? | 当前 writer | 状态 |
|---|---|---|---|---|
| UpdatedAt / BedStatus / BedEvent / StartTime / DurationSec | ✅ | - | sensor zoneengine translator | OK |
| **TrackNumber**（=床上人数 0..2，**不是** track_id；v1 InBed +1 / LeftBed -1，Sleepad 用 sleepadInBedTrackCount 直接覆盖；v2 中应归 sensor 自家 zoneengine bed FSM 写——bed FSM 知道 occupied count）| ? | - | **零 writer** | 🟡 S5b 决策：sensor bed FSM 接管写 or 删字段（用户 2026-05-19 指出 v1 还有 "15s 同类事件 dedup" 但代码搜不到——可能更早设计；S5b 实施时确认是否加入 dedup） |
| **BedConfidence**（v1 InBed 时 = Sleepad 90 / Radar 60；意义 = 数据可信度）| ? | - | **零 writer** | 🟡 S5b 决策：同 TrackNumber（sensor bed FSM 写 or 删） |
| **SleepStage** | ✅ 应该 | - | **零 writer** | 🟡 S4 wire 后归 sensor，projector merge owner 标记一并订正 |
| **SleepConfidence** | ✅ 应该 | - | **零 writer** | 🟡 同 SleepStage |

## v2 RoomState 字段 ownership

所有字段 sensor 全 owner，无 cardagg 补写需要。v2 已删 walk/stand/multi_person 字段（这些走 TargetState per-device）。**RoomState 不存在"完整版待补"问题**——用户 Q2 原话"RoomState 完整版融合 sleepstate track_id"实际指 BedState。

## v2 TargetState 字段 ownership

| 字段 | writer | 状态 |
|---|---|---|
| TrackID / LogicID | sensor per-device | ✅（v3 logicID 真融合再说）|
| LastActiveTs | sensor aggregator（S2 完工后）| 🟡 待 wire |
| StandingContinuousMin | sensor aggregator（S2 完工后）| 🟡 待 wire |
| WeakBiometricSignal | sensor aggregator（S1 完工后）| 🟡 待 wire |
| VisitorStartTs / TodayMaxVisitorMin / HasVisitorToday | cardagg VisitorDeriver | ✅ 已实施 |
| UpdatedAt | sensor + cardagg max-merge | ✅ |

## 必须新增的 S4 / S5 节点（v2 残留收尾）

**S4 重写（sensor 端，sensor 端，**不是** cardagg）— 沿用 v1 逻辑不再算术加权**：
- sensor 订阅 iot:event:stream category=alarm.SleepStage（sleepace + radar 双源 publisher）
- **直接取 event 中的 SleepStage 值**（不做 Sleepad=8/Radar=4 算术加权——那是 sensor weights.go 别处的 vital 信任度语义，与 SleepStage 无关）
- **SleepConfidence = 60(Radar) / 90(Sleepad) / 100(双设备同步态)**；仅 `confidence ≥ current` 才覆盖，否则 drop（v1 [cardaggv1/internal/service/state_service.go:572 PublishBedStateSleepStage](../../../owl/owlBack/wisefido-cardaggv1/internal/service/state_service.go#L572)）
- OOB 判定走 sensor 自家 zoneengine bed FSM（不读 card:state；FSM 知道 leaving/vacant），异常 → emit device_failure alarm
- LeftBed/ExitRoom/EnterRoom 触发清零：由 sensor 自家 zoneengine transition 自然触发 BedState rewrite（SleepStage=0, SleepConfidence=0）
- 写自家物理 BedState（sensor:derived:stream bed.state，subject_entity=/96 bed CIDR）
- sensor_state_projector.go 同 PR 订正：SleepStage / SleepConfidence 从"非 sensor owner"→"sensor owner"；mergeBedStateSensorOwner 加这两字段
- 入口 Producer "sensor.*" 前缀 skip 防 loop

**S5 cardagg 端剩余收尾**：

| # | 项 | 内容 |
|---|---|---|
| ~~S5a~~ | **已完成 2026-05-19 commit `41d6146`** | cardagg AIVerdictHandler port from v1；订阅 ai:track:verdict:stream；MonitorHandler.Apply 合并到 track fields；4 个清理触发（tid=88 / EnterRoom / ExitRoom / TTL 60s GC）；sandbox/release 双模式 deploy 按域名区分（`8da3ad4`） |
| ~~S5b~~ | **已完成 2026-05-19 commit `9b19a84`** | TrackNumber = `ZoneState.Count`（v2 bed FSM 0/1；多人共床 cap 0..2 留 v3）；BedConfidence = LastSource 映射 sleepace=90/radar=60/else=0；10s 同类事件 dedup helper（`BedEventDedup`）在 radar/sleepace adapter 入口；cardagg `mergeBedStateSensorOwner` 加 TrackNumber/BedConfidence owner |
| ~~S5c~~ | **已确认 audit-only 完成 2026-05-19** | v1 sleepad track override 链路（`sleepadTrackOverrideForInBed` / `sleepadTrackOverride` / `BedEventCoordinator.sleepadInBedTrackCount`）全部留在 `wisefido-cardaggv1/`，v2 一开始就没 port，无需删除。S5b 落地的 TrackNumber 取 `ZoneState.Count`（bed FSM 0/1），不需 monitor buffer 数 track；多人共床 cap 0..2（v1 行为）v2 不建模，**留待 v3**。grep `wisefido-cardagg/` 已验证零残留 |
| ~~S6~~ | **设备类 alarm gate sensor 端（2026-05-19 用户拍板补）commit `1c107c9`** | 4 类设备 alarm (Offline/SensorDetached/AngleException/SignalPoor + Recover) → DeviceFitnessTracker per-device 位标 → radar/sleepace adapter handleMsg 入口 IsFit gate；防 SensorDetached 后 sleepace 乱发数据污染 bed FSM (P0 风险)；27 Tier1 case |

## 关联

- [[sensor_stream_subscriptions]] — sensor 订阅口径（producer 4 wire）
- [[p4_next_session_prompt]] — 把 S5 加进第 1 段后置节点
- [[radar_activity_stats_on_event_stream]] — activity 字段在 event 流不是 monitor 流
- [[feedback_producer_first]] — sensor producer 4 wire 全绿才能动 cardagg S5（S5 依赖 sensor target/bed state publish 正常）
