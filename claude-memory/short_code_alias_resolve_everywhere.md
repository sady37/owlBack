---
name: short-code-alias-resolve-everywhere
description: 短码/别名 ID 必须在所有下游 callsite resolve；只在 permission check 处解析下游用原始字符串 → cache miss 静默退化（device 全 offline 等）
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 0ef86edd-98ed-4293-bb19-0c02b8af18a4
---

引入短码/别名 ID 层（如 `cards.dns_short_name` → `spatial_prefix`、未来可能的 device short ID）后，**每个**用 ID 当下游 key 的 callsite 都要走 resolve；只在权限校验处 resolve 而下游沿用原始字符串 = silent cache miss = 全栈降级且无 error。

**Why:** 2026-05-15 改 URL 用 dns_short_name 当 cardId param 后 detail 页所有 device 显示 offline。Root cause = `GetCardStatus` handler 权限校验调 `GetCardInfo`（已支持短码解析），但下游 `realtime.GetCardStatus(ctx, cardID)` 用了原始 cardID 字符串（短码 `poqfqu`）查 Redis，而 Redis key 是 IPv6 spatial_prefix → 全部 miss → 状态默认 false。无 error log，FE 收到合法响应但内容全 offline，靠肉眼对照才发现。同模式 bug 还存在于 `SubscribeRealtimeStream`（一并修了）。

与 [[v2-cutover-type-assert-silent-drop]] 同源：**新引入的标识层在某些 callsite 没接到，无 error 但行为退化**。

**How to apply:**
- 引入新别名层时，handler 入口 resolve 后用 resolved ID 调下游，**永远不要**让下游函数收到原始用户输入（即使权限已过）
  ```go
  card, err := h.cardService.GetCardInfo(ctx, ..., rawCardID)  // resolve + perm check
  resolvedID := card.CardID  // 必是 canonical IPv6
  data, _ := h.realtime.GetCardStatus(ctx, resolvedID)  // ← 用 resolved，不是 rawCardID
  ```
- audit 已有代码：grep handler 函数体里 `rawParam` 形参在 `GetCardInfo(...rawParam)` 之后是否还被传给其它 service。如果是，必是 bug
- 短码不可逆 + 下游 cache 通常按 canonical key — 不查 cache 就是 silent miss，没有 sql.ErrNoRows 那种醒目错误
- 如果 service 边界改不动，至少 handler 局部变量改名（`rawCardID` vs `resolvedCardID`），让 reviewer 容易看出哪个该传给谁

相关：[[v2-cutover-type-assert-silent-drop]] 同样的 silent-drop 模式
