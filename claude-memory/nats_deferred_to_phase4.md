---
name: nats-migration-phase-4-v2-stable-1-2
description: 当前打地鼠根因是 v1→v2 cutover 代码债，不是 Redis Streams；切 NATS 不解决问题，反增三线作战
metadata: 
  node_type: memory
  type: project
  originSessionId: 102f29bf-0317-4351-b343-2f5e6fe0d658
---

2026-05-16 讨论 — 是否把 Redis Streams 切 NATS。结论：**推迟到 Phase 4**。

**Why**：
- 当前打地鼠根因是 v1→v2 cutover 留的代码债（cardagg pending stub / iot_timeseries_id 散落 / short_code alias 没全 resolve / sensor adapter_redis 直写违反 writer 边界）— NATS 解决不了任何一条
- Redis 在 owl 还要留（card:status hash / cache / KMS / IPAM 都 Redis）→ 切 NATS = 双 broker 运维
- v2 cutover 没完成（owlFront 迁移 / iot_timeseries 退役 / Phase 2-5 backlog 进行中）— 叠加 broker 迁移 = 三线作战，错误难定位
- NATS 真优势（subject 通配 / native cluster / 10k subjects 千分之一成本）owl 当前规模根本用不到

**How to apply**：
- 现阶段任何 messaging 设计走 Redis Streams 不要假设 NATS
- 不要在新代码里加"NATS-ready"抽象层（YAGNI + 加摩擦）
- Phase 4 启动条件：v2 cutover 全部 done + owlFront 上线 + 运营稳定 1-2 个月，那时才评估 NATS 是否解决具体新需求（多租户隔离 / 多 region federation 等）

**当下正事**：cardagg 重写 pilot 验证"重写 vs 手术"实证收益，顺手清 Redis stream wiring 上的债（命名 / consumer group / payload schema），数据攒齐再决 NATS。
