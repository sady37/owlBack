---
name: role-naming-v1-v2-mess
description: PG users.role 列 v1/v2 命名 + case 双轨混存；backend/FE 兜底层都仍在用 v1 名，盲删会丢权限
metadata: 
  node_type: memory
  type: project
  originSessionId: 2bffdee5-d5ae-4711-879c-844ea5b2555b
---

**users.role 列实际数据**（2026-05-22 PG owl_v2 audit）：

TitleCase（v2 norm）+ lowercase（v1 残留）混存，9 种取值共 22 行：
- `Caregiver` ×6 / `Nurse` ×5 / `Admin` ×1 / `Manager` ×1 / `Family` ×1 — v2 TitleCase
- `family` ×4 / `tenant_admin` ×2 / `platform_admin` ×1 / `manager` ×1 — v1/lowercase 残留

**Why 不能盲清 FE v1 名**：backend SQL/code 仍在主动用 v1 名查询，FE 删 v1 兜底会让对应账号失去权限：
- [auth_handler.go:443-444,544](../../owlBack/wisefido-data/internal/http/auth_handler.go) `u.role='platform_admin'`
- [user_handler.go:542](../../owlBack/wisefido-data/internal/http/user_handler.go#L542) `tenant_admin`
- [admin_tenants_handlers.go:65](../../owlBack/wisefido-data/internal/http/admin_tenants_handlers.go#L65) `role='tenant_admin'`
- FE [UserList.vue:1227-1229](../../owlFront/src/views/admin/users/UserList.vue#L1227-L1229) `hasManagePermission` 用 `.toLowerCase()` 后比 6 个名兜底，**defensive 不是死代码**

**How to apply**：
- 改 role 相关代码前先 `SELECT role, COUNT(*) FROM users GROUP BY role` 看实际状态
- 加新 role 判定写法用 `.toLowerCase()` 兼容两种 case
- 想统一命名 = 架构级工作：PG migration + backend 单一源 + FE 清兜底 + 数据迁移
- v1→v2 映射（疑似，未正式拍板）：`platform_admin`→`SystemAdmin`、`tenant_admin`→`Admin`、`branchmanager`/`manager`→`Manager`、`family`→`Family`

**支线状态**：[data_v2_todo.md S11](../../owlBack/doc/data_v2_todo.md) 记账；属架构级决策 + 数据迁移，不能在 v2 cutover 边缘子任务顺手做

相关：[[v2_cutover_lessons]] 短路 vs 重写；[[multi_branch_nurse_employment]] 角色扩展待加 my_branches 时一并处理
