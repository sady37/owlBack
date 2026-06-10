---
name: v2 cutover 教训：短路 vs 重写
description: 当 v1 service/repo 有大量 schema-coupled SQL 时，逐函数 short-circuit 是错的方法 — 直接重写整个 v2 service 反而快
type: feedback
originSessionId: 6a5f2016-6082-4856-b91a-b95fb56f8a66
---
## 问题症状

v1 service 函数（300+ 行）里所有 SQL 都引用 v1 schema：
- residents.unit_id / room_id / bed_id / branch_id / tenant_id (UUID)
- 跨表 JOIN buildings/branches/units/rooms 都用 UUID
- 业务规则（BranchOnly/AssignedOnly 权限）夹在 SQL 里

每个 endpoint 路径独立，碎片化分散。

## 错误的做法 — 逐函数 short-circuit

我做的：在每个 service 函数顶部加 `if v2 短路 else 走 v1`。

**结果**：
- 每发现一处报错才知道又有一条 v1 path 在跑（resident reset_password / available_caregivers / unit display / ...）
- 短路碎片越加越多，永远做不完
- 5 次会话还在 resident 模块 — 用户问 "在有 IPv6 情况下，为什么这么慢"

## 正确的做法 — 一次性 v2 重写

**Why**：IPv6 设计本身极大简化了空间逻辑（位掩码 + 单一 hoa key），整个 v2 service 应该是 ~500 行干净 SQL，远比 3000 行 v1 + 短路碎片简洁。

**How to apply**（v2 cutover 里再遇到任何 v1 service 大量 schema 引用时）：

1. 不要 short-circuit — 一旦发现 v1 SQL 引用 3 个以上已删/重命名的列，**立即转重写**
2. 新建 `xxx_repo_v2.go` + `xxx_service_v2.go` — 完整 v2 实现
3. handler 只调 V2，删 v1 path（保留代码作 ref 即可）
4. 一次 build / 一次部署 / 所有 endpoint 同时正常

**判定标准**：v1 函数里 v1 schema 引用 ≥ 3 处 → 整段重写；< 3 处 → 单点 fix。

## 历史触发

2026-05-10 Resident 模块改造 5 次会话还没完成；用户喊停转重写策略。

## 2026-05-16 复发 — cardagg 战役 5.1 留尾巴

平台-agent cutover (2026-05-15) 把 producer / pending-alarm / NightAbsence 计时全移到 sensor，但 cardagg 留了一堆 no-op stub：
- 5 个 `func ...Pending...() (...) { return nil }` 空壳 + 15 处 `_ = h.alarms.RemovePendingAlarm(...)` 调用点
- `cardaggProducerAlarmRouter = ""` 空字符串常量 + 注释 `TODO(战役 5.1): audit + 迁移`
- `runNightAbsenceCheck` 10min ticker 还活着，跟 sensor zonealarm 双发 NightAbsence
- AHI `CheckAH` / `AHIQueryReady` v2 改写了一半（query 切到 monitor_stream）但功能从来没启用 — 死代码改成新 schema 还在那

**Why this is bad**: 下一个会话来"清 iot_timeseries 尾巴"时，发现 cardagg 还引用 iot_timeseries（CheckAH 那段），又要顺手清；清完发现 CheckNightAbsence 也该删；清完发现 5 个 no-op stub 也是死的 — 永远在收拾上一个 cutover 没扫干净的残局。

**How to apply** — 任何 cutover/退役 PR：
1. 完成迁移后**当场删干净所有 stub/orphan**（不留 "保留 callsite 编译" 的占位空壳）
2. 不写 `TODO(战役 X): audit + 迁移` — TODO 注释是债务签字，要么这次就做，要么显式 issue 跟踪，不留在代码里发酵
3. 留下的每行死代码 = 下一个 PR 的额外摩擦 = 用户要再开一次会话重复审计

**判定标准**：迁移 PR 合并前 grep `已退役|已废弃|deprecated|TODO.*迁移|no-op stub`，命中 = 没扫干净，不要合并。
