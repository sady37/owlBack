---
name: debug_query_pg_first
description: "排查 sensor 事件/监控数据时**首先查 PG**(event_log/alarm_events/monitor_stream),不要只翻 Redis 流——流会滚导致漏事件"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: e8750e7d-da7e-421d-98f1-fc1e56c285ed
---

排查设备事件(InBed/LeftBed/Fall/track 等)时,**第一步查 PG 持久表,不是 Redis 流**。

**Why**:Redis 流(iot:event:stream / iot:monitor:stream / iot:alarm:stream)有 retention 窗口会滚,只翻 Redis 容易**漏掉真实存在的事件**。2026-06-05 实例:我断言"9e7 没有上床事件"(只翻了 Redis 25min 窗口),被用户打脸——PG `event_log` 里 9e7 明明有 `InBed @ 15:03:15`(+ 14:32:48),Redis 没捞到。

**How to apply**:
- 事件 → `event_log`(owl_v2,列:ts/device_addr(inet /128)/event_kind/subject_addr/payload jsonb;90d)。按 `device_addr=...::inet AND ts>now()-interval` 查,`event_kind` 含 InBed/LeftBed/EnterRoom/ExitRoom/activity/number_people/sleep-stage/analysis 等。
- 报警 → `alarm_events`(7y)。
- 监控 raw → `monitor_stream`(track XY/pose/HR/RR;90d)。
- DB=owl_v2 @ 127.0.0.1:5432 user/pass=postgres/postgres;Redis 密码 TeLunSu-36kr。
- Redis 流只用来看**最新实时态**(card:state hash / card:realtime),不要拿来做"有没有发生过某事件"的判断。

关联 [[feedback_no_unverified_claims]](不未验证先断言)、[[radar_realtime_subscribe_not_delivered]]。
