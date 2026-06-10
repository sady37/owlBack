---
name: v2 cutover 项目硬约束
description: owl_v2 schema 迁移期间不修改 v1 业务逻辑；遇到业务问题只能建议，不擅自改
type: feedback
originSessionId: 6a5f2016-6082-4856-b91a-b95fb56f8a66
---
owl_v2 schema cutover 期间的两条硬约束（用户 2026-05-09 明确）：

**R-001 不修改 v1 业务逻辑/功能**
- v2 cutover 只做"管道层"改造：schema/SQL 适配、字段映射、URL 解析、prefix 派生工具
- 业务流程一律保持 v1 行为：登录顺序、字段语义、permission 判定、UI 交互
- 发现 v1 业务问题 → **只能向用户建议**，明确得到指示后才能动

**R-002 禁止硬删除业务数据**
- HIPAA 数据保留 + 业务恢复（试用期暂停后期可能恢复合作）双重要求
- 所有数据表用 `status='deleted'` 软删；list 默认过滤 `deleted`
- 例外（独立生命周期）：`device_factory_meta`（物理出厂元数据）/ `audit_log`（HIPAA 7 年保留）/ `signing_keys`（验签历史）
- **空容器允许硬删**：tenant/branch 误创建救济 — DELETE 时若下游 spatial_prefix 范围内无任何业务数据（含已 deleted 软删行）→ 真物理 DELETE。判定走 repo 内 isXxxEmpty(prefix) helper（UNION ALL EXISTS 检查 sub-spatial 表 + users/residents/devices）
- 真要彻底清理 → ROOT 运维 SQL 直接 DELETE，不走 API

**Why:** v1 系统已经在用户/资源那里运行；不必要的业务变更扩大测试面+破坏既有 SOP。HIPAA 与业务恢复对硬删都是绝对约束。

**How to apply:**
- 编 v2 repo 时只动 SQL；service/handler 业务流不动（除非纯 schema mismatch 修复）
- 遇到看似不合理的 v1 行为 → 列入 checklist BUGS 区，给用户判断
- 任何"我想顺便改一下…"的念头先停下，问用户
- DELETE 永远翻译成 UPDATE status='deleted'；下游 list/get 默认隐藏
