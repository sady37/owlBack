---
name: server-internal-utc-only
description: Server 内部逻辑全 UTC；TZ 转换只在 API 输出/输入边界做（2026-05-15 拍板）
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 5c181f1d-cf48-440b-bd17-e5ae711213b7
---

2026-05-15 用户拍板：

> Unit ResetTime 要转换为统一的 UTC，Server 只用统一的 UTC 来处理，简化逻辑。

## 规则

- **Why**：跨 unit/branch TZ 在 server 内部做 timezone math 复杂易错；UTC 单一时间轴让对账/比较/cron 都简单
- **How to apply**：
  - DB 列：`TIMESTAMP WITH TIME ZONE`（PG 内部存 UTC，TZ 信息保留供查询）
  - Go 内部：所有 `time.Time` 比较 / 算术都用 `.UTC()` 之后再做
  - 配置（如 ResetTime "21:00-07:00"）：load 时转 UTC、按 cron rule 处理；不要在每次检查时 `time.Now().In(loc)` 做 TZ math
  - 跨 unit 时间窗：unit 各自 TZ 但比较时统一转 UTC 后 compare
  - **唯一例外**：API 输出/输入边界。output 给 FE 同时给 RFC3339-with-offset + UTC ms (FE 选择)；input 接收 RFC3339 解析为 time.Time（带 TZ → UTC 后存）

## 当前需要 audit/改造的位置

- `wisefido-cardagg/internal/service/alarm_service.go::isWithinResetTimeWindow` — 当前每次调用时拿 unit TZ + time.Now().In(loc) 做窗口判定 → 改造为「配置 load 时算出本次窗口的 UTC 起止 ts」缓存，runtime 只比 UTC
- 其他 `*.In(loc)` 调用 grep 一遍

## 与已有的关系

- [[v2_cutover_lessons]]：v2 schema 已经 TIMESTAMP WITH TIME ZONE 全 UTC 存储
- API output：[platform_agent_addressing](platform_agent_addressing.md) "TZ 规则" 一节 — 返 RFC3339-with-offset + UTC ms
