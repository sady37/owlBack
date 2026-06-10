---
name: 多 branch 执业 user_roles 扩展（已落地）
description: 非 admin user 的 my_branches 多选 + now_branch 单选切换；2026-05 已实现
metadata:
  type: project
  originSessionId: 6a5f2016-6082-4856-b91a-b95fb56f8a66
---

## 业务场景

NurseB 在 TenantA 的 BranchC + BranchD 同时执业（养老院常见替班/轮岗）。

## 实际落地（2026-05，已完成）

### 数据模型

| 概念 | 实际表/列 | 说明 |
|---|---|---|
| my_branches | `user_roles.scope` INET 多行 | 一 user 对一 role 可有多条 scope 行 = 多 branch |
| now_branch | `user_branches.is_primary=TRUE` 的那条 | legacy 表保留作"当前激活 branch"语义来源 |

注：原设计意图把 my_branches 做成 `INET[]` 列；实际落地用 `user_roles` 多行 + scope。
最终效果等价：一 user 多 row 一 row 一 branch；is_primary 标识 current。

### 后端
- `auth_handler.go` 登录填 `Session.CurrentBranchID` from `user_branches.is_primary=TRUE`
- `scope.go` 所有 SQL 过滤用 `sc.CurrentBranchID`：units/cards/devices/residents 全链路
- `user_handler.go` `PUT /admin/api/v1/users/me/current-branch` 切换 current
- `GetAvailableBranches` 返 user 可执业 branch 列表（Sidebar 下拉用）

### 前端
- `views/admin/users/UserDetail.vue` Branches `mode="multiple"` 多选；后端 batch 写 user_roles.scope
- `components/layout/Sidebar.vue` 顶部 branch 下拉：候选 = my_branches → PUT me/current-branch + localStorage + userStore.setCurrentBranchId

### 权限规则
- session_prefix = `current_branch` /56 优先；空时 fallback `tenant_prefix` /48
- Admin/SystemAdmin my_branches 可空 → tenant 整体 scope
- 非 admin 必填至少一 branch；switch 校验目标 ∈ my_branches

## 后续可选改进（非 blocker）
- `user_branches` legacy 表完全吸收到 `user_roles`（删表 + is_primary 迁过去）
- 现状能用，不影响业务
