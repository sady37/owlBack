---
name: Engine event/alarm 订阅现状
description: engine 已订阅消费 EnterRoom/ExitRoom/InBed/LeftBed + iot:alarm:stream radar Fall；剩余 gap 是 firmware 门区检测覆盖不全 + R4 床边跌倒规则待加
type: project
originSessionId: 28e63f8a-abf3-408d-b6c4-76fe69672bff
---
详细设计见 [owlBack/doc/AI_event_alarm_subscription.md](../../../owl/owlBack/doc/AI_event_alarm_subscription.md)。

## 当前状态（2026-05-02 验证）

✅ **iot:alarm:stream 已订阅**：[engine.go:989 runAlarmLoop](../../../owl/owlBack/wisefido-ai/internal/roomengine/engine.go#L989) 消费 radar Fall，路由到 `RecordRadarAlarm`，未来段 7 做 narrative。

✅ **iot:event:stream 已消费 EnterRoom/ExitRoom/InBed/LeftBed/NumberPeople**：
- `RadarTrackEvent` 解析（[alarm_event.go:98 ParseRadarTrackEvents](../../../owl/owlBack/wisefido-ai/internal/roomengine/alarm_event.go#L98)）
- `RecordRadarEvent`（[track_manager.go:642](../../../owl/owlBack/wisefido-ai/internal/roomengine/track_manager.go#L642)）：
  - ExitRoom → 取消所有 pendingLostFalls（`lost_fall_cancelled_by_exit_room`）
  - InBed → cell 锁 AreaBed + 双源一致性 BedSession
  - LeftBed → 更新 `lastLeftBedAt`（R4 用）
  - **NumberPeople=0 → ExitRoom 兜底**：firmware 部分场景不发 ExitRoom（如 D523 FOV 边角离场），但发 number_people=0（实测早 track_id=88 36-44ms，三例样本 100% 命中）。track 失锁入池前若 ±5s 内有 number_people=0 → 跳过 pending；已存在 pending 同样取消。新参数 `FallRulesParam.Lost.NumberPeopleZeroFallbackMs`。
- EnterRoom 用于 birth-score 配对（`nearestEnterRoomMs` ±3s + `hasRecentEnterRoom`），加 +20 分；BirthFinalGraceMs 重算窗口

## 剩余 gap

1. **R4 床边跌倒未实现**：风险时段 LeftBed 后人未走开，在床边 100cm 内静止 >15min → 报警
   - 与 R2 (bed-fall) 区别：R2 是"radar 误以为人在床 + sleepad 离床"的物理矛盾；R4 是"人离床后到不了远方"
   - 不立即做，先等数据积累

2. **firmware 门区检测覆盖不全**（D523 实测）：
   - 同一台 D523 在 00:22 和 00:35 两次离场都 **没发 ExitRoom**（仅发 number_people=0 + track_id=88）
   - 对比同时段 Kitchen / LivingRoom 频发 EnterRoom/ExitRoom（人在两屋间走动），iot_timeseries 完整可见
   - 推断：D523 firmware 内部 enter_zone 矩形未覆盖到 (-310, 590) 和 (-510, 130) 这两个边角离场点
   - PR-C 的 number_people=0 兜底已工程上规避，但根因建议查 BLE 配置中 D523 的 enter_zone 几何

**Why**：BM87224700978 04-26 01:58 case 揭示——sleepad/radar 都报了 LeftBed 但无 Fall 检测，
当前 engine 检测体系在"安全离床后床边晕倒"场景下完全无防御。

**How to apply**：未来加 R4 时基于 `lastLeftBedAt` 已就位，配合 cell 床边邻域距离即可。
注意：双源 LeftBed 时间容差 8-15s 正常（传感器响应延迟）。
