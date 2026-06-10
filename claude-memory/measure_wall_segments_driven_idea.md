---
name: measure-wall-segments-driven-idea
description: 2026-05-28 furniture rect 绿方案 (segments-driven) 实测胜出后下一步思路 — 套用到 Wall 设计 + Enter 改后置叠加 + 整体后移 backend
metadata: 
  node_type: memory
  type: project
  originSessionId: 57ba5657-d24d-4eef-8454-6da260ea880d
---

把 furniture rect 的"segments-driven"（蓝拟合线驱动 rect，DEV 绿框）思路套用到 Wall fit 设计；Enter 处理改为"最后叠加"而不是 wall pipeline 内部 split；Wall 验证 OK 后整体迁移到 backend 实现。

**Why**:
- 2026-05-28 furniture rect 实测：绿（segments-driven）相比红（raw-dots TBLR LS）在 4/5 fixture 上更准（修了 bed_b17f tail 拉伸 -40cm、other_b17f -y 偏 -38cm 等）。蓝拟合线已经过 snap+absorb+merge 清洗，比 raw dots 抗 tail outlier。同样的 logic 在 Wall（更长更复杂 polygon）上理论上收益更大。
- 当前 Wall recipe (`makeWallRecipe`) 用 `detectAnomalousGaps` 在 wall fit pipeline 内部 split N → (gapSegments for Enter, M for wall)，Enter 处理跟 wall 拟合耦合。User 想拆开：先纯按 wall 拟合，Enter 作后置 overlay step，独立判定。
- 整套 measure 算法目前 100% 客户端跑（owlFront/src/utils/radar/measureSketch.ts），算法成熟后挪到 backend：客户端只送 raw dots，server 出 layout objects。理由：跨设备一致性、可批量 reprocess（fixture 数据 server 持有，算法迭代不需重测）。

**How to apply**:
- **Step 1** (近期): Wall recipe 重构 — `makeWallRecipe` 用 segments-driven 思路：
  - 复用 `classifyPathShape` 路 H/V 段（wall 也是 axial Manhattan world）
  - polygon 顶点 = axial segments 的端点交点（H/V 段相交点）
  - 不在 wall pipeline 里 split Enter
- **Step 2**: Enter 后置 — 拿到 wall polygon 后再跑独立 Enter 检测（在 wall path 上找 ED 候选 + merge），输出 gapSegments 叠加渲染
- **Step 3** (验证后): Backend 化 —
  - 客户端只送 Dot[] 到 server
  - server 实现同套 measureSketch.ts 逻辑（Go 移植）
  - server 返回 BaseObject[] (wall polygon + Enter rects + furniture rects)
  - 算法迭代时 server 端 reprocess 所有历史 fixture batch 验证

**关联**:
- [[measure_time_continuity_projection]] — dot 投影到拟合线后按时间相邻差区分同圈连续 vs 跨圈共线，本质同源（segments 上做后处理）
- [[radar_measure_flow_design]] — 4 intent 平等 + cell 着色人画原则，wall + enter 拆开符合该原则
- [[server_internal_utc_only]] — 若 backend 化，所有 time 字段（Dot.time）按 UTC 处理
- 设计 snapshot doc: `owlFront/docs/measure_rect_red_vs_blue.md` (commit `a781c99`)
- 实现入口: `owlFront/src/utils/radar/measureSketch.ts` `makeWallRecipe` (line ~1295), `inscribedRectFromSegments` (line ~614)
