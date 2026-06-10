---
name: firmware-fall-qualification
description: Radar firmware 自带 30~90s 时间阈值；pose=2 (SuspectedFall) 持续超时才升级 pose=5 (Fall confirmed)
metadata: 
  node_type: memory
  type: project
  originSessionId: 102f29bf-0317-4351-b343-2f5e6fe0d658
---

**Why**：radar firmware（Qinglan 协议）在源头分两段判定 fall-class 事件，server 不再重复判定：

| pose | EnumPose | firmware 当前行为（实测 D523 2026-05-17 16:24） | server 处理（Option C） |
|---|---|---|---|
| 2 | suspected_fall | **会发 /event/ type=2**（实测确认） | event_log SuspectedFall row，不入 alarm_events |
| 5 | fallen | 持续 fall_alarm_duration（30~90s）confirmed 后发 /event/ | event_log + alarm_events (CRITICAL) |
| 7 | suspected_sitting_ground | 推测对称（未实测） | event_log SuspectedSittingOnGround row，不入 alarm_events |
| 8 | sitting_ground | 持续 confirmed 后发 /event/ | event_log + alarm_events (CRITICAL) |

**实测 D523 2026-05-17 16:24-26**（PoseMap 修复后）：
- 16:24:51 firmware 发 /event/ type=2 pose=2 → event_log `SuspectedFall pose=2` row
- 16:26:00 firmware 发 /event/ type=2 pose=0 → event_log `Initialization pose=0` row
- alarm_events: 0 行（Option C sensor 早返）

**修正历史认知**：早先以为"firmware 30s 内不发 /event/"是错的。原因是 PoseMap 在 decoder 层把 pose=2 collapse 成 "Fall"，所以以前测试看到的"Fall row"其实是 SuspectedFall 被磨平。修 PoseMap 后才暴露 firmware 真实会发 SuspectedFall。

阈值时长来自 fall_param 配置（observation.FieldFallAlarmDuration / EncodeFallParam，单位秒，BLE 配置时下发 firmware）。

**Option C 设计原则（2026-05-17 拍板）**：
1. `alarm_events` 表语义 = "需要人响应的事件"（HIPAA actionable trail），Suspected* 不构成响应必要
2. firmware 30~90s qualification = 内置去抖，server 加一层会双重过滤 → 漏报
3. 事件链不丢：event_log 全留（SuspectedFall + Fall + Initialization 完整序列），事后调查能完整回放
4. 避免护士 alarm fatigue：30s 内自动升级到 Fall CRITICAL 已足够及时，WARNING 中间态打扰意义低

**How to apply**：
- qinglan handleEventMessage case eventType=2 保留 4 个 envelope.Category 各自独立（不再合并）
- sensor engine.go handleEventMessage：Suspected* 早返不调 ParseRadarFallAlarm + RecordRadarAlarm
- ParseRadarFallAlarm 白名单仅 Fall + SittingOnGround
- RadarFallAlarm.Category 字段透传，修复了之前 SittingOnGround 被误转发为 Fall 的 hardcoded bug
- DefaultAlarmSetting.Radar：SuspectedFall = IsEnabledOff + DisplayNone（保留 schema 占位，前端切换可重启）
- 不要在 server 端再叠一层"持续期"判定 — firmware 已经做过

**实测 2026-05-17 15:36 D523**：
- firmware 上报 pose=1 (Walking) → pose=2 (SuspectedFall) → 9s 内回 pose=0 (Init)
- 当时 server 还是 Option A 合并实现，event_log 落 Fall row + alarm_events ALERT row
- Option C 之后再测 30s+ 摔倒：event_log 4 行（Walking → SuspectedFall → Fall → Init），alarm_events 1 行 Fall CRITICAL

[[qinglan_log_to_zap_done]] [[event_name_literal_audit_done]]
