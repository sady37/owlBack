---
name: startup-clear-and-initial-publish
description: 2026-05-20 修订 — sensor 启动 publish 全部 ts=0（不是 now），cardagg builder 视作 no-data → FE 显 "—"；杜绝重启所有卡同步显"Active 3m ago / InBed 3m"假活跃
metadata: 
  node_type: memory
  type: project
  originSessionId: a23f1c6c-09c3-4b4c-82f1-1aaf419b5178
---

## 双侧机制

**问题**：sensor 重启后 cardagg `card:state:<addr>` hash 保留昨日 `RoomState.LastExitTime / BedState.StartTime`（如「OOR 29h ago」），因为 sensor 只在 zone transition 时 publish，启动时无 transition；Redis hash 是跨重启持久的。

**解法（双侧）**：

| 侧 | 触发 | 行为 | 文件 |
|---|---|---|---|
| **C** | wisefido-data 启动 | SCAN+DEL `card:state:*` 全清 | [cmd/wisefido-data/startup_clear.go](../../../owl/owlBack/wisefido-data/cmd/wisefido-data/startup_clear.go) |
| **A** | wisefido-sensor 启动 | publish 所有 room/bed 占位 RoomState/BedState/BedSleepStage，**所有 ts=0** | [cmd/wisefido-sensor/initial_publish.go](../../../owl/owlBack/wisefido-sensor/cmd/wisefido-sensor/initial_publish.go) |

## 关键修订（2026-05-20 ts=0 重设计）

**旧实现（已废）**：publish 时 LastExitTime=now / StartTime=now / UpdatedAt=now → 重启后所有卡同步显"Active 3m ago / InBed 3m / LeftBed 3m"假活跃。

**新实现（[initial_publish.go:19-26](../../../owl/owlBack/wisefido-sensor/cmd/wisefido-sensor/initial_publish.go#L19) 注释明示）**：
- 所有时间字段 UpdatedAt / StartTime / LastExitTime = **0**，不伪造 now anchor
- cardagg `BuildCardDisplay` 看 UpdatedAt=0 → `bedHas/roomHas=false` → 不进 display 派生
- FE 渲染：active_anchor_ms=0 → "—"；scene_state=0 default → "OOR"；bed_status undefined → "No visitor today"

**Why:** 旧 now anchor 让重启后整片卡片同步显示假活跃，nurse 误以为人都刚活动；ts=0 让 FE 老实显示 "—"，等真 transition 触发自然更新。

**How to apply:** 任何"重启清残留"的初始 publish 都按 ts=0 模式，别拿 nowMs 当 anchor；cardagg `mergeBedStateSensorOwner` 字段级 merge 保 prev 时间戳。

## 启动顺序

`start-owlback-full.sh`：data → 10s → cardagg → qinglan → sleepace → iot → sensor。

- data 先清 Redis（empty hashes）
- cardagg 起来挂上 `sensor:derived:stream` 消费者
- sensor 最后启动并 publish 初始占位（ts=0）→ cardagg 字段级覆盖 owner 字段、保留 prev ts → FE 看到"—"

## A 路径覆盖范围

| publish | 字段（当前实际）| cardagg 端 merge 行为 |
|---|---|---|
| `RoomState` (category=room.state) | TotalPeople:0（ts/RoomType 未设；RoomType 已挪 CardStatic）| 整段覆盖（v2 全 sensor owner）|
| `BedState` (category=bed.state) | BedStatus:1 占位 LeftBed（其他字段 implicit zero / ts=0）| 字段级 merge：sensor owner 字段覆盖 |
| `BedSleepStage` (category=bed.sleepstage) | SleepStage=Initial, SleepConfidence=0 | 字段级 merge：仅这两字段（避免残留 prev SleepStage）|

未涵盖的 hash 字段（依赖 organic event 重建）：
- `Target` — 等首个 sensor target.state event
- `AlarmState` — 等首个 alarm event 或 next QueryCardAlarmState 触发
- `Display` — cardagg 每次 sub-state 写时自动 rebuild

## 已知字面 log 不一致

[initial_publish.go:105](../../../owl/owlBack/wisefido-sensor/cmd/wisefido-sensor/initial_publish.go#L105) 仍 log "anchors reset to now"，但代码实际 ts=0；log 字面落后于重设计，无功能影响。

## 失败/边缘

- data clear 失败 log warning 不阻塞启动（fail-open）
- sensor publish 失败逐项 log warning 不中断（partial 也可接受）
- 单独 data 重启（sensor 不动）→ FE 卡片短暂 blank 直到 organic event 重建；trade-off 已知

## 关联

- [[device_class_alarm_auto_recover]] — 与本机制正交（per-device offline 触发 in-memory clear，不写 Redis）
- [[sensor_target_state_unwired]] — Target 字段 unwired 已修，但启动初始 publish 不含 Target（依赖事件流）
- [[card_display_projector_handoff]] — Display rebuild 路径
