---
name: feedback_analyze_options_not_just_answer
description: 用户/项目组给选项或想法时，必须先给必要分析(利弊/影响/风险)再实施或回答，不简单照做
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 55e53c08-0788-49f8-a133-bf627cbc2a3f
---

2026-06-07 用户明确要求:**对于项目组(或用户本人)给出的选项/想法/方案,不能简单回答或直接照做,要先给出必要的分析**——利弊、影响面、风险、可能的副作用、更优替代,然后再实施或给建议。

**Why:** 用户是系统架构师,给的"选项"往往是抛出的想法,期待我用工程判断帮他权衡,而不是当执行器盲目落地;简单照做会错过设计缺陷(如交互盲点、功能回归)。

**How to apply:** 收到选项/需求时,先点出关键 tradeoff 和潜在问题(尤其交互/数据/兼容性回归),给出推荐并说明理由,必要时反推/提替代;确认方向后再写代码。与 [[feedback_no_unverified_claims]]、[[feedback_read_design_before_modify]] 一脉相承。
