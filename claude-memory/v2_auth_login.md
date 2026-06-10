---
name: v2 IPv6 auth/login 路径已落地
description: owl_v2 cutover 第一步 — auth/login 全栈 v2 化，双层 hash + IPv6 prefix tenant，e2e 已跑通
type: project
originSessionId: 6a5f2016-6082-4856-b91a-b95fb56f8a66
---
owl_v2 cutover 的 auth/login 已端到端跑通（2026-05-09）。是 v2 迁移真正起点，后续 vue 鉴权用同一套基础设施。

**Why:** owlBack `.env` DB 切到 owl_v2 后 v1 SQL 全断（无 user_account_hash / user_branches 等），登录都登不上；这是后续所有 v2 业务的前置条件。

**架构决策（不向后兼容）：**
- 凭证形态：双层 hash — 前端 `sha256(password)` hex 上送；DB 存 `bcrypt(sha256(password))`；后端 `bcrypt.CompareHashAndPassword(stored, sha256_hex)`。bcrypt 抗暴力 + sha256 防止网络/日志层明文泄漏。
- tenant 标识：从 UUID 切到 IPv6 prefix CIDR `fd00:0:T::/48`，权限边界用前缀掩码计算（不再多表 JOIN）。
- v1 auth 完全删除（前端 loginApi/searchInstitutionsApi/Institution 类型/utils/crypto.ts hashAccount 全删）；v1 后端 handler 保留作冷备但不被前端调用。
- 鉴权工具：`owlFront/src/utils/spatial.ts` 镜像 `owl-common/spatial` — `parsePrefix/contains/tenantOf/branchOf/scopeOf/slotsOf`，后续所有 vue 权限判断走这个。
- session: `SessionData` 加 `TenantPrefix` + `HoA` 字段；middleware 注入 `X-Tenant-Prefix` + `X-User-HoA` header；`X-Tenant-Id` 为兼容字段（也装 prefix CIDR）。
- forgot-password / verify-pin 仍走 v1 endpoint（独立功能；v2 化在后续 PR 里做）。

**关键文件：**
- backend: [auth_v2_handler.go](../../../owl/owlBack/wisefido-data/internal/http/auth_v2_handler.go) + [session_store.go](../../../owl/owlBack/wisefido-data/internal/http/session_store.go) + [auth_middleware.go](../../../owl/owlBack/wisefido-data/internal/http/auth_middleware.go)
- frontend: [spatial.ts](../../../owl/owlFront/src/utils/spatial.ts) / [auth.ts](../../../owl/owlFront/src/api/auth/auth.ts) / [authModel.ts](../../../owl/owlFront/src/api/auth/model/authModel.ts) / [user.ts store](../../../owl/owlFront/src/store/modules/user.ts) / [LoginForm.vue](../../../owl/owlFront/src/views/login/LoginForm.vue) / [axios index.ts](../../../owl/owlFront/src/utils/http/axios/index.ts)
- seed: [seed_demo_data.sql](../../../owl/owlBack/scripts/seed_demo_data.sql) — `crypt(encode(sha256('xxx'::bytea), 'hex'), gen_salt('bf'))` 模板

**E2E 验证（2026-05-09）：**
- POST `/auth/api/v2/search-tenants` {username, password_hash} → matches[]
- POST `/auth/api/v2/login` {username, password_hash, tenant_prefix?} → {accessToken, tenant_prefix, hoa, ...}
- admin@demo / caregiver_denver_1 / 错密码 / 错用户名 — 4 个 case 全通过

**配置 quirk:** `wisefido-data/config.yaml` 的 `database` 字段会**覆盖**环境变量 `DB_NAME`（YAML 存在则不读 env）；想切 owl_v2 必须同步改 yaml。已改 yaml→owl_v2。

**下一步：**
- 在 [README](../../../owl/owlFront-migration/README.md) Option C（branch.ts 试点）继续：现在 login 通了，前端能拿到 tenant_prefix，可以正式切 spatial v2 endpoint
- 后端 v2 admin API 还需补 GET LIST / PUT / DELETE branches（spatial_v2_handler 当前只有 POST）
- v1 service/repo（card_sync_service、user_handler 等）SQL 仍跑 v1 schema — 大部分会在 owl_v2 上 SQL 错误，但只在 token 鉴权后调用业务 endpoint 时才暴露；逐个 v2 化
