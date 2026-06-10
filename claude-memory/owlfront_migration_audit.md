---
name: owlFront 迁移 audit 完成 — artifacts 在 /owlFront-migration/
description: 2026-05-09 owlFront → owl_v2 迁移 audit 已扫完；artifacts 在 /home/wisefido/owl/owlFront-migration/；下次会话从 README.md 接续
type: project
originSessionId: 53d81dc2-b99e-4371-95da-247c2cf5c46c
---
owlFront → owl_v2 schema/API 迁移的扫描已完成。下次会话从 artifacts 接续工作。

**Why:** 后端 dbv2 + v2 spatial REST API 上线后，前端 ~85 个 v1 API endpoint 和 57 个 .vue 都需要同步迁移；用户列出 3 个 P0 问题（KMS phone/email、B2B medical 提示、中国手机号校验）；audit 是契约对齐的前提。

**How to apply:**
- 下次会话先读这 4 个文件（按顺序）：
  1. `/home/wisefido/owl/owlFront-migration/README.md` — 高层 spec + 决策点
  2. `schema-column-diff.txt` — 31 张共享表的列级 diff (v1 vs v2)
  3. `api-endpoints-v1.txt` — owlFront ~85 个 v1 endpoint inventory
  4. `views-inventory.txt` — 57 个 .vue 与其 api/ 引用
- 已有 README.md 末尾"下一会话 prompt 模板" + 3 个起步选项 (A/B/C)
- 起步前问用户选哪个 option（A=写 mapping CSV / B=补 v2 API 缺口 / C=改 owlFront API 层试点）

**关键发现：**
- owlFront：Vue 3.5 + ant-design-vue 3.2，14 个 api/ 模块，145 个 defHttp 调用，54748 行 .vue
- 6 处 `validateUSPhoneNumber` 重复（NANP 硬编码），需统一抽 libphonenumber-js
- v2 spatial REST 已实现 7 个 endpoint (commit 7dcaea0)；剩 ~50+ endpoint 待补（CRUD list/update/delete + residents/users/devices/cards）
- v1→v2 body 格式重大变化：`tenant_id: UUID` → `parent: "fd00:0:T::/48"`；前端 API 层不只是换 URL，需要重做 ts model

**3 个 P0 用户已 raise 的问题（详 README §二）：**
- #1 KMS 上线后可存 phone/email（前端表单恢复编辑 + K 服务加密）
- #2 B2B 机构停用 medical_history（v-if tenant 类型；可能要补 tenants.deployment_kind 字段）
- #3 中国手机号校验（demo tenant 已有 ShenZhen branch + 3 中国 resident，校验器需国际化）
