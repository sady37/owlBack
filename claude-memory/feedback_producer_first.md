---
name: feedback-producer-first
description: 跨服务多 PR 排序铁律：producer 先彻底完工，consumer/FE 才能做；否则下游对着死字段空跑
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 0566c6c8-ddbd-4e2d-b8b6-79fef2a5a715
---

**规则**：跨服务功能拆 PR 时，producer 端必须先彻底完工（订阅链路 + 数据写入 + publish wire），consumer/downstream/FE 再动。**不要因为 downstream 设计先想清楚就先做 downstream**。

**Why**：用户多次纠正后强调——sensor 是 cardagg 的 producer，cardagg 是 FE 的 producer。producer 没完工时下游做的所有合并/派生/渲染都对着死字段（0 writer 永远 nil/0），既无法实测验证也无法 catch 边界 bug；等回头 wire 上游才暴露问题就是双工返修。memory [[p4_next_session_prompt]] 2026-05-19 一次就把 WeakBio Step2 FE 横条排在 FU1/FU3 sensor wiring 之前，导致用户必须出手纠偏。

**How to apply**：
- 看 followup backlog 时按"数据流上游 → 下游"重排，不按"FE 友好度"或"用户可见度"
- sensor 端 4 类 wiring 必须先全绿才能动 cardagg 派生：(1) 订阅 iot:event/alarm:stream wire 上（数据进来）（2) handleMonitorFrame 真填字段 + 心跳契约（数据攒起来）（3) StreamPublisher tick 真发 target.state（数据出去）（4) BedState/RoomState 非死字段（cardagg 不在合并空 map）
- cardagg 同理对 FE：merger + writer wire 全 → 再做 FE 横条 / dumb render
- 死字段判定：`grep -r "FieldX" --include=\"*.go\" | grep -v _test.go` 看是否有 writer，没有就是死字段
- 拍板下一个 PR 前先问"producer 这边的 4 条 wire 都通了吗"，否则切回上游
