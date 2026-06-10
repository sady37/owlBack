---
name: alarm-anchored-to-card
description: alarm 锚定 card_id snapshot 不跟 device 走；card 删除时同 card_id 非终态 alarm 自动 expired
metadata: 
  node_type: memory
  type: project
  originSessionId: 2bffdee5-d5ae-4711-879c-844ea5b2555b
---

**事件锚定卡**：`alarm_events.card_id` 是 trigger 瞬间 cards GiST LPM 锁定的 snapshot，**永不更新**。device 迁去别的 card → alarm 留在原 card 作历史，**不跟 device 走**。

**Why**：alarm 是发生在某个**空间**的事件（"卧室检测到 Fall"），不是设备属性。device 是数据采集源不是事件主体。空间换手后，原空间的历史不应被新主人继承，也不应消失。HIPAA 审计要求事件发生地不可篡改。

**How to apply**：
- 改 alarm 数据时禁止 UPDATE card_id（不论 device 迁移、设备解绑、卡 rename）
- 原 card-G Detail 仍能查到 alarm（Resolved tab，operation/handler 历史完整）
- 新 card-H Detail 看不到这条 alarm（card_id 不匹配）
- 数据 layout：alarm_events.card_id INET 无 FK 约束，故意 — 见 [owlRD/dbv2/60_alarm_events.sql L80](../../owlBack/wisefido-data/../../owlRD/dbv2/60_alarm_events.sql)

**Card 真正删除时的清理规则**（2026-05-22 拍板）：
- 单卡删除 / unit 删除 / reconcile delete → 同 card_id 上所有 `active/acked/auto_resolved` 强制 UPDATE → `expired`
- 实现：owl-common `ExpireAlarmsByCardID` (单卡 `card_id = $1`) + `ExpireAlarmsByCardPrefix` (`card_id <<= $1` 批量)；接到 `postgres_card.DeleteCard` / `DeleteCardsByUnit` / `card_reconcile.applyDiffs` 三处删除路径
- preserve `operation` / `handler`（不覆盖原 staff 决策历史）；只动 `alarm_status`/`hand_time`/`handler_notes`
- 旧 `ExpireAlarmsByDeviceAddrs` 已删（按 device_addr 过严，device 迁走时漏掉孤儿 alarm）

**未行动的边界**：
- device 仅解绑（unit/card 未删）→ alarm 不动（仍是历史），故意保留
- card_id 列无 FK 约束 → 删 card 时 alarm 行 card_id 变成"指向不存在卡"的悬挂值；查询时仍正常返回（因为 card_id 是值不是引用），Resolved tab 仍可查；alarmBell counter / Pending list 走 `card_id = $1` 谓词，找不到对应 card 自然 0 — 设计预期

相关：[[alarm_status_no_acked_auto_resolved]] alarm 状态机；[[v2_cutover_rules]] R-002 禁止硬删（HIPAA）；alarm 走 expired 而不是 DELETE
