---
name: bathroom-state-merged-into-room-state
description: 2026-05-18 BathRoomState/UnitState/BedNightAbsence 全删，统一到 RoomState + Kind + LastExitToOutside；unit 卡 FE 找 active child room 渲染
metadata: 
  node_type: memory
  type: project
  originSessionId: 26834bd4-1e50-431d-b029-8287b0926981
---

## 数据模型最终态

- `card.RoomState`（/88 hash 唯一 state 块）：`Kind` 区分 bedroom/bathroom/living/kitchen/dining/public；`StaySec` / `LastExitToOutside` 字段共用
- `card.BedState`（/96 hash）：同前
- 删除 `BathRoomState` / `UnitState` —— /80 unit 没有自己的 state hash
- FE unit 卡渲染：按 /80 spatial_prefix 列子 /88 room card，挑最近 `room_state.updated_at` 的那个显示
- OOR vs OOU：`room_state.last_exit_to_outside == true` 时 FE 标签从 OOR 切到 OOU（EnterArea==outside 触发；sensor 上游信号源待接入）

## 报警重排

- 删 `BedNightAbsence`：床上无人不等于风险，可由 LeftBed 长时长（FE 显 OOB.Xh）替代
- 保留 `NightAbsence`：room→vacant 持续 N min（21-7 时段），真风险
- sensor 内部 unitPersonCount 留作 NightAbsence of Unit 未来扩展前置（当前未实现，无 caller）

**Why:** v1 BathRoomState/UnitState 字段独立设计在 v2 cards-as-INET-prefix 模型下成了冗余；多 radar 同绑 /80 时 room_state vs bathroom_state 在同 hash 冲突；NightAbsence of Bed 语义模糊（沙发睡觉算 absence）。

**How to apply:**
- 新加 room-级字段直接加在 `RoomState`，按 `Kind` 在 sensor/FE 分支处理 risk 阈值，不要再造独立 state 块
- unit-level 派生：sensor 内部算（如 cross-room risk），不写 card:state hash
- FE unit 卡 = "active room 视图"，不是 "rooms 聚合视图"
- sensor `stream_publisher.OnZoneEvent` 写 cardID 用 `e.ZoneID`（/88 或 /96），不用 `e.CardID`（= radar 绑卡，可能是 /80）—— 避免多 radar 写同 hash 互覆盖
- `TranslateRoomState` count_change 分支处理 0↔N 跨越的 LastEnter/LastExit 更新（否则 FE OOR 时钟永远停在最初 Vacant 时刻）

涉及落地 commit/未提交 PR：
- BE: owl-common/card + owl-common/alarm + wisefido-sensor (zoneengine + zonealarm) + wisefido-cardagg (projector) + wisefido-data + scripts
- FE: monitorModel.ts + monitor.ts + Overview.vue + AlarmCloud.vue + sleepace-monitor-settings.vue
