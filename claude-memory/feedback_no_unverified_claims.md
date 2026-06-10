---
name: feedback-no-unverified-claims
description: "禁止\"已清理/已修复/已收口/已完工\"类断言不先 grep 核查；30 秒 grep 能验证的事必做"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: c9715c78-74f8-4b78-864e-2103493a53a0
---

**规则**：任何"已 done / 已清理 / 已收口 / 已收敛 / 已迁移完"类完成态断言，先 grep 核查再说。

**禁忌句**（说出口前必 grep）：
- "X 已清理掉"
- "Y 已迁移完"
- "Z 包依赖已砍"
- "FieldA 全部 callsite 删了"
- "这是历史，没人在用"
- "已经收口了"

**2026-05-25 实测教训**：
说 "risk_evaluator.go 的 card 依赖是历史" — 用户 grep 一查，sensor 27 文件 / 177 处 card 引用，包括 risk_evaluator 17 处。完全没清。

**Why**：完成态断言 = 用户停止追查的信号。说错 = 用户基于错误前提做决策，下游全坏。

**How to apply**：
- 写"已 X" / "X 已经" / "Z 不再 Y" 前必跑 `grep -rn '<symbol>'` 核查
- 不确定写法："当前路径上没看到 X，但没全 grep；要不要先扫一遍？"
- 用户问"你之前说 X，是真的吗？" 直接重新 grep，不要凭印象答辩
- 类似毒句："这个早就不用了 / 这是老代码 / 这是兼容遗留" — 一律 grep verify 再说

**与其它 feedback 关系**：
- [[feedback_read_design_before_modify]] — 改代码前读 design：本条同一根（"先验证再行动"），但侧重断言而非动手
- [[feedback_store_raw_not_derived]] — 也是基础错误的复发模式（同一周内 3 次同类错）

**触发器**：自我审视有"我记得"/"应该是"/"早就"等不确定措辞 → 立刻 grep。
