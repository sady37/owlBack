---
name: alarm-status-no-acked-auto-resolved
description: alarm_events.alarm_status 只 5 个值（不加 acked_auto_resolved）；2026-05-21 砍掉 sensor_v2 决定 17 该分支
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 3f1734f3-aa75-4cd3-ab56-deae78211024
---

`alarm_events.alarm_status` CHECK 取值集**只 5 个**：`active` / `acked` / `resolved` / `auto_resolved` / `expired`。**不要**加 `acked_auto_resolved`。

**Why**：sensor_v2 决定 17（2026-05-18 拍板）原打算给 Critical "物理自动恢复 + 人工 review" 路径加独立终态 `acked_auto_resolved` 区分"现场处理"与"事后回顾"。但 (a) `operation` 字段 (verified/false_alarm/test/auto_resolved) + `handler` (user_id) 复合已能表达同一语义 — `WHERE operation='auto_resolved' AND handler IS NOT NULL` ≡ 等价集合；(b) 多一个状态值是冗余编码 + 多一条 SQL CASE 分支；(c) 5-18 migration 文件存在但**从未 apply 到 PG**，0 数据行 / 0 写入路径 — 砍掉零损失。2026-05-21 决定撤销该状态值。

**How to apply**：
- Critical alarm transition 走标准 `active → acked → resolved` 或 `active → auto_resolved → resolved`（auto_resolved 人工 review 后**直接 resolved**，不经 acked）
- KPI 查询"物理自动恢复后人工 review 过"：`WHERE operation='auto_resolved' AND handler IS NOT NULL`
- "Critical 必须强制人工 review"原则仍成立：`operation='auto_resolved' AND handler IS NULL` = 未 review
- `QueryCardAlarmState` SQL 计数（alarmBell counter）：Critical 仍 `IN ('active','acked','auto_resolved')`，`resolved` 是终态不计
- **Pending list（/alarm/records 与 by-card Detail）**：active + Critical(level<=2) auto_resolved unreviewed；**acked 不入 Pending**（acked=staff 已确认收到=task done 进 Resolved tab）。即 `(alarm_status='active' OR (operation='auto_resolved' AND handler IS NULL AND alarm_level<=2))`。与 alarmBell counter 语义不同——counter 包含 acked，list 不含
- doc/sensor_v2.md 残留 7 处决定 17 关于 acked_auto_resolved 的描述待后续清理（不影响代码/schema 一致性）

相关：[[feedback_schema_review_via_dbv2]] CREATE 是权威；[[p4_next_session_prompt]] 决定 17 P4 此分支已撤销
