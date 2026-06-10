---
name: feedback-low-freq-invalidate-simple
description: 低频 invalidation 流不要做 incremental / event-sourced / unified-spec 抽象；频率属性应导向更简单方案不是更复杂
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 2c3246e4-2e55-4933-a33c-8f7f15eab443
---

低频变化流（card config / spatial / monitor toggle 等 10 次/天 量级）的 invalidation：
- 用"通知受影响范围 + consumer lazy reload from DB"足够，**DB 是真相，message 只是 hint**
- 不抽 `LoadAffectedSpec` 公共 SQL（看到 3 处类似别想统一，cache 形状本来就不同 = leaky abstraction）
- 不画 TDPv2 event-sourced envelope 北极星（那是为高频流设计的，低频不痛）
- 短期方案 = 长期方案

**Why**：2026-05-22 [data_v2_todo S10] 讨论 monitor on/off 信号同步时我连续 over-engineer 两次。先抽 LoadAffectedSpec 被纠正，再画 TDPv2 演进路径被纠正。用户提示"card 变更很少"才让我意识到频率属性应该直接导向更简单方案。

**How to apply**：评估任何 cache invalidation / cross-service sync 设计时，第一个问题问"这事多高频"。10 次/天 = stop thinking about abstractions，直接焊缺的线。
