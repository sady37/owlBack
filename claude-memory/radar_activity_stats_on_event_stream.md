---
name: radar-activity-stats-on-event-stream
description: 数据流事实：radar 分钟级活动统计（walk_distance/walk_duration/stand_duration/lie_duration/multi_person_duration/pose）在 iot:event:stream category=activity，不在 monitor 流
metadata: 
  node_type: memory
  type: reference
  originSessionId: 0566c6c8-ddbd-4e2d-b8b6-79fef2a5a715
---

**Radar 分钟级活动统计字段流向**：
- Producer：qinglan `publishStatActivity`（[wisefido-qinglan/internal/consumer/mqtt_consumer.go:1117-1147](../../../owl/owlBack/wisefido-qinglan/internal/consumer/mqtt_consumer.go#L1117-L1147)）
- 流：`iot:event:stream`（**不是** monitor 流）
- topic_type：`"event"`
- Category：`observation.FieldActivity` = `"activity"`
- 字段（observation/fields.go:46-50）：
  - `walk_distance` (m, 1min)
  - `walk_duration` (s, 1min cap 900)
  - `stand_duration` (s, 1min cap 600)
  - `lie_duration` (s, 1min)
  - `multi_person_duration` (s)
  - `pose` (FieldPose, 当前姿态)
- 频率：每分钟一发（firmware 分钟级聚合）

**monitor 流（topic_type=monitor）只有 1Hz raw track**：
- 字段：XY 坐标、HR / RR / track_id / 姿态瞬时态等
- 不含 walk/stand 累加统计

**Pose 也分两种**：
- 瞬时 pose 切换 → type=2 pose event（iot:event:stream category=pose）
- 当前 pose 状态 → activity stat 里的 pose 字段（per minute）

**sensor 侧消费方式**（FU3/FU4 接通时）：
- TargetStateAggregator 需要订阅 `iot:event:stream` filter category∈{activity, pose, fall, ...}
- `iot:alarm:stream` 仅供 WeakBio score 累加（HR/RR/ApneaH/WeakBio raw alarm）
- `iot:monitor:stream` 不进 aggregator（aggregator 关心的是分钟级派生量；1Hz raw 是 zoneengine track_manager 的事）

**心跳契约就是 firmware activity stat 每分钟上报这件事**：
不用 sensor 自造心跳；只要 firmware 每分钟 push activity，sensor aggregator 就每分钟 dirty，StreamPublisher 60s tick 自然每分钟发 target.state，cardagg 2min staleness 阈值天然成立。
