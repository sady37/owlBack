---
name: device-ipv6-single-trip
description: device_ipv6 化迁移走单程票模式（不双发/不 fallback/不观察期），同 PR 同部署窗口；checklist 在 owlBack/doc/device_ipv6_migration_checklist.md
metadata: 
  node_type: memory
  type: feedback
  originSessionId: b9c43444-5a87-4a7c-9a57-e98422659a2d
---

2026-05-13 用户校准：device_ipv6 化（cardagg/qinglan/sleepace/sensor/wisefido-data 全部从 UUID device_id 切到 INET）走单程票模式。

**Why:** cards_v2 的渐进式（双发 + fallback + 观察期）工作量翻倍且交付期长。device 这条路径用户已锁单 tenant Demo 测试 + 全栈可 1 分钟 stop/start，可承受零容错部署窗口换 ~30% 工作量节省。

**How to apply:**
- ❌ 不做 producer "双发期"（同时带 UUID + INET 字段）
- ❌ 不做 consumer "优先 INET，fallback UUID" 分支
- ❌ 不留观察期 + 二次清理
- ✅ Phase A 一次性删 owl-common UUID 入口（`LookupCardIDByDeviceID` / `BuildDeviceProducer(string)` 等）
- ✅ Phase A merge 后下游 service 立即编译失败 — 设计意图，强制全栈同步切
- ✅ Phase B-E 五处同 PR 同时改完才 push；7 service 全栈停+全栈起
- ✅ 部署前 [R-010] redis stream drain（`XLEN config:*:stream` ≤ 1000）

**Cards_v2 的渐进式不复用本模式 — cards 那条线 producer/consumer 涉及外部前端 + DDNS，不能零容错。**

详 checklist + 7 phase 拆解：[`owlBack/doc/device_ipv6_migration_checklist.md`](../../../owl/owlBack/doc/device_ipv6_migration_checklist.md)

关联：
- [[envelope_protocol_evolution]] — Datagram v1 device-as-host 派生
- [[unbound_device_card_id_fallback]] — UUID 充 card_id 的 hack 在本 PR R-009 中一次清理
- [[v2_cutover_lessons]] — 一次性 v2 重写原则（本 PR 在每个 service 内仍遵守）
- [[v2_cutover_rules]] — R-001..R-008 通则（本 PR R-001..R-010 是其特化）
