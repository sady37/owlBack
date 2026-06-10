---
name: producer-consumer
description: "流名带 realtime/projection/state 类的\"加工产品\"流，producer 负责 TTL/dedup/分组/snapshot，consumer 只读不重做维护"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 102f29bf-0317-4351-b343-2f5e6fe0d658
---

card:realtime:stream / card:status:<addr> / 各种 projection 流，**producer 是 maintainer**：负责 TTL prune、per-key 分组、snapshot 节流。**consumer 只读不重做这些维护工作**。

**Why**：让 consumer 自己重做维护工作 = 同一份逻辑双写在两边，TTL 不一致 / dedup 窗口不同步就 silent bug。如果允许"data 直订 iot:monitor:stream 自己 fan-out"，data 端要复刻 cardagg 的 12s TTL + 1s snapshot + per-card 分组，cardagg 改 TTL 时 data 不跟改 = drift。

**How to apply**：
- 流名 `iot:*:stream` = 原始总线，consumer 自己过滤是常态
- 流名 `<domain>:realtime:*` / `<domain>:projection:*` / `<domain>:state:*` = 已加工产品，consumer 只读
- 设计新流时问：producer 有没有"维护态"（TTL/累积/分组）？有 = maintainer pattern，consumer 必单一读路径

实例：
- ✅ cardagg 订 iot:monitor:stream（raw）维护 card:realtime:stream（加工）→ data 订 card:realtime:stream → 前端 SSE
- ❌ data 直订 iot:monitor:stream 自己 fan-out（曾经的提议，因为 data 不是 maintainer）

详 [owlBack/CLAUDE.md 规则 #2.4](../../../owl/owlBack/CLAUDE.md)。
