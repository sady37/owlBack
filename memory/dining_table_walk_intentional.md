---
name: 餐桌区保持 Walk 是设计意图
description: roomengine 中餐桌/茶几学不成 AreaDeny 是预期行为，不修复
type: feedback
originSessionId: 1e0f3f52-3831-4ea0-86e2-4b4bce2f7275
---
餐桌区域被学成 Walk（而非 AreaDeny）是预期，不要"修复"。

**Why:** 老人在餐桌区高发跌倒（被椅子绊、心梗、滑倒）。`AreaDeny` 会拒绝 fall 检测；`AreaActive`(Walk) 允许 fall 触发。保持 Walk 才能让餐桌跌倒被报。地图美观度让位于业务正确性。

**根因**（供未来排查类似"为什么 X 没学成 Deny"问题参考）：餐桌四周是椅子 → 学成 `AreaSit` → 在 [cell_learning.go:317](owlBack/wisefido-ai/internal/roomengine/cell_learning.go#L317) 是 BFS source → 桌面中心 dist=1 < 2 → Auto-Deny 永远不触发。这是机制和业务都对的巧合。

**How to apply:** 用户截图指出餐桌/茶几"描述得不好"时，先确认是 Walk 还是 Unknown：
- 是 Walk → 不动，解释为什么 Walk 更安全
- 是 Unknown 散点 → 也不动（Unknown 也允许 fall）
- 用户坚持要标 → 走 PR-18 layout 工具手标 Furniture，不要改 BFS source 规则
不要把 AreaSit 从 BFS source 移除——会连带影响床/沙发周边几何。
