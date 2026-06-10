---
name: feedback-replay-alarm-sound-policy
description: Replay alarm marker spawn 时按 ttl 决定是否响 alarmSound：30s marker（推断类 gap>3s）响一下；0.2s flash（firmware gap≤3s）不响。
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 02ff7aec-d416-4282-aa86-023ffc9704bd
---

Replay 期间 alarm marker spawn 时按 ttl 区别处理 sound：
- **ttl >= 30000ms**（gap > 3s，推断类如 bedroom_person_silent / lost_track / sleepad_radar_conflict）：调一次 `alarmSound.playAlarm()`
- **ttl < 30000ms**（gap ≤ 3s，firmware Fall pose=5 直发）：不响

**Why:** 推断类 fall 是长时间无活动 / 双源矛盾推出来的严重事件，user review 时听到一声等于"这是关键事件"；firmware Fall 反正 monitor pose=5 那一秒已经在 track 上有连续姿态可视，再加声音吵。BE wisefido-sensor 自身永不接 sound 触发（sound 是 FE-only 概念）。

**How to apply:** 实现见 [WaveMonitor.vue:playNextFrame](owlFront/src/components/Radar/WaveMonitor.vue) `if (p.ttlMs >= 30000) alarmSound.playAlarm()`。未来扩展 replay 类型（vital replay 等）按相同原则：高严重度的推断事件响一下，连续可视的实时类事件不响。

**Replay 结束 marker 清理强约束**：`handleStop` → `clearAllData` 里必须同步清 `alarmMarkers = []`，**不能**让 marker 走自己的 setTimeout 自然 GC。原因：replay 3min 结束 → setPlaybackMode(false) → SSE realtime 重新接管 ringBuffer，此时残留的 alarm marker 钉在画布上会跟实时 track 视觉混叠，让 user 误判"现在还有 fall"。无论 marker ttl 多长（甚至刚 spawn 0.5 秒），replay 边界结束就立即清。配套：`setPlaybackMode(true)` 起步也清一次防上次残留。
