---
name: sensor-target-state-wired
description: sensor target.state 2026-05-21 验证已 wire 完整；publisher/aggregator/consumer 三段连通；当前 demo 不发是因为没人走动（walk_distance<2m && walk_duration<6s 全部不达阈值）
metadata: 
  node_type: memory
  type: project
  originSessionId: 6596aa1c-6420-4198-9805-2c0620c5d97c
---

**2026-05-21 verify**：之前的"零 wire"诊断已过期。Wire 链完整，5 处源代码位置已确认：

| 段 | 实现位置 |
|---|---|
| `iot:event:stream category=activity` → `ActivityConsumer.handleMsg` | `wisefido-sensor/internal/consumer/activity_consumer_sensor.go:117` |
| `PushEventFields` → `handleEventFrame` 更新 lastActiveTs + dirty | `internal/service/target_state_aggregator.go:365-395` |
| 60s tick: `streamPublisher.tickPullAndPublish` → `aggregator.GetSnapshot` → `PublishTargetState` | `internal/zoneengine/stream_publisher.go:166-189` |
| wiring 注入 `SetAggregator` + 启动 goroutine | `internal/zoneengine/wiring/setup.go:117 / :264` |
| cardagg projector 路由 target.state → `card:state.target` | `wisefido-cardagg/internal/consumer/sensor_state_projector.go` |

触发条件: `WalkDistance≥2m OR WalkDuration≥6s` (60s 节流防同分钟重写)。

**当前 demo radar 实测**：189 activity events 全 walk_distance=0/walk_duration=0 — 雷达前没有人在走动。系统行为"不发"是正确的，不是 bug。

## 当前缓解（owlFront）保留作物理 offline / 无人场景兜底

`Overview.vue::getSection3Upper` `STALE_LAST_ACTIVE_MIN = 24*60` 仍然有效：
- `minAgo > 24h` → 显示 "—"（视作无数据）
- 真实有人走动时 publisher 会自然发包，minAgo 缩到 < 24h，"—" 自然失效，**无需回滚** fallback

## 历史背景（保留）

**2026-05-17** v2 cutover 联调时 UI "Active Xh ago" 显示最旧 35h+ → 当时诊断 PublishTargetState 接口零 caller。**实际** wire 是在该日期后补完的；记忆本身没跟着更新。本次（2026-05-21）verify 时验证 5 段链路与触发条件全在位。

## 相关字段（不要混淆）

| 字段 | 位置 | 维护状态 |
|---|---|---|
| `SuitePerson.LastActiveMs` | sensor in-memory + Redis `sensor:suite_census:*` | ✅ PR-2 维护中 |
| `card.TargetState.LastActiveTs` | cardagg `card:state:<cid>` hash → owlFront UI | ✅ wire 完整；触发条件未达即不更新 |

**Why**：避免下次会话又踩"以为零 wire 要补"的坑 — 真相是阈值 + 物理无人活动；不要重新设计 trigger 路径。

**How to apply**：
- 看到 UI "Active —" 不要立刻怀疑 wire 漏；先看 iot:event:stream activity event 是否 walk_distance/walk_duration 达阈
- 真要测试 wire：注入合成 activity event with walk_distance=5,walk_duration=10 看 `card:state:<cid>.target` 是否更新
- 阈值争议：当前 ≥2m / ≥6s 来自 v1 weights.go::WalkDistanceMetersThreshold 注释；想调阈值改 `internal/service/target_state_aggregator.go:374` 那行
