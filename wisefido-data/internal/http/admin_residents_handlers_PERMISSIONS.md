# admin_residents_handlers.go 权限检查表

## residents 资源权限配置（基于 role_permissions 表）

| 角色 | 权限类型 | assigned_only | branch_only | 说明 |
|------|---------|--------------|-------------|------|
| **Admin** | R/C/U/D | `FALSE` | `FALSE` | 可查看/管理所有住户 |
| **IT** | R | `FALSE` | `FALSE` | 可查看所有住户 |
| **Manager** | R/C/U/D | `FALSE` | `TRUE` | 只能查看/管理同 branch 的住户 |
| **Caregiver** | R | `TRUE` | `FALSE` | 只能查看分配的住户 |
| **Nurse** | R/U | `TRUE` | `FALSE` | 只能查看/更新分配的住户 |

## resident_contacts 资源权限配置

| 角色 | 权限类型 | assigned_only | branch_only | 说明 |
|------|---------|--------------|-------------|------|
| **Admin** | R/C/U/D | `FALSE` | `FALSE` | 可查看/管理所有联系人 |
| **Manager** | R/C/U/D | `FALSE` | `TRUE` | 只能查看/管理同 branch 的联系人 |
| **Caregiver** | R | `TRUE` | `FALSE` | 只能查看分配住户的联系人 |
| **Nurse** | R/U | `TRUE` | `FALSE` | 只能查看/更新分配住户的联系人 |
| **Family** | R/U | `TRUE` | `FALSE` | 只能查看/更新自己的联系人信息 |

## resident_phi 资源权限配置

| 角色 | 权限类型 | assigned_only | branch_only | 说明 |
|------|---------|--------------|-------------|------|
| **Admin** | R/C/U/D | `FALSE` | `FALSE` | 可查看/管理所有PHI |
| **Manager** | R/C/U/D | `FALSE` | `TRUE` | 只能查看/管理同 branch 的PHI |
| **Caregiver** | R | `TRUE` | `FALSE` | 只能查看分配住户的PHI |
| **Nurse** | R | `TRUE` | `FALSE` | 只能查看分配住户的PHI |

---

## 所有住户操作权限检查实现状态

| 操作 | 行号 | 权限检查 | 状态 | 问题 |
|------|------|---------|------|------|
| **GET /residents** | 73-304 | ⚠️ 仅检查 `assigned_only`，硬编码 Manager 分支过滤 | ⚠️ 部分实现 | 1. 未使用 `GetResourcePermission()` 统一查询<br>2. 未检查 `branch_only` 标志<br>3. Manager 分支过滤硬编码（行194）<br>4. 未实现空值匹配逻辑 |
| **POST /residents** | 305-638 | ❌ 无权限检查 | ❌ 未实现 | 需要添加 C 权限检查（assigned_only + branch_only） |
| **POST /residents/:id/reset-password** | 642-709 | ❌ 无权限检查 | ❌ 未实现 | 需要添加权限检查（只能重置分配住户的密码） |
| **PUT /residents/:id/phi** | 710-1235 | ✅ Resident/Family 自检 | ⚠️ 部分实现 | 1. 仅检查 Resident/Family 自检<br>2. 未检查 Staff 权限（assigned_only + branch_only） |
| **PUT /residents/:id/contacts** | 1236-1570 | ✅ Resident/Family 自检 | ⚠️ 部分实现 | 1. 仅检查 Resident/Family 自检<br>2. 未检查 Staff 权限（assigned_only + branch_only） |
| **POST /contacts/:contact_id/reset-password** | 17-67, 1483-1564 | ❌ 无权限检查 | ❌ 未实现 | 需要添加权限检查 |
| **GET /residents/:id** | 1571-1947 | ✅ Resident/Family 自检 | ⚠️ 部分实现 | 1. 仅检查 Resident/Family 自检<br>2. 未检查 Staff 权限（assigned_only + branch_only） |
| **PUT /residents/:id** | 1948-2085 | ✅ Resident/Family 自检 | ⚠️ 部分实现 | 1. 仅检查 Resident/Family 自检<br>2. 未检查 Staff 权限（assigned_only + branch_only） |
| **DELETE /residents/:id** | 2086-2109 | ❌ 无权限检查 | ❌ 未实现 | 需要添加 D 权限检查（assigned_only + branch_only） |

---

## 详细分析

### 1. GET /residents (列表查询) - 行 73-304

**当前实现：**
- ✅ 检查 `assigned_only`（行121-138）
- ⚠️ 硬编码 Manager 分支过滤（行194-196）
- ❌ 未使用 `GetResourcePermission()` 统一查询
- ❌ 未检查 `branch_only` 标志
- ❌ 未实现空值匹配逻辑（NULL 或 '-'）

**问题：**
```go
// 行121-138: 仅查询 assigned_only，未查询 branch_only
err := s.DB.QueryRowContext(r.Context(),
    `SELECT assigned_only FROM role_permissions
     WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'R'
     LIMIT 1`,
    SystemTenantID(), userRole.String,
).Scan(&assignedOnly)

// 行194-196: 硬编码 Manager 分支过滤
if userRole.Valid && userRole.String == "Manager" && userBranchTag.Valid && userBranchTag.String != "" {
    args = append(args, userBranchTag.String)
    q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND u.branch_tag = $%d`, len(args))
}
```

**需要修复：**
1. 使用 `GetResourcePermission()` 统一查询 `assigned_only` 和 `branch_only`
2. 使用 `ApplyBranchFilter()` 实现分支过滤（含空值匹配）

---

### 2. POST /residents (创建住户) - 行 305-638

**当前实现：**
- ❌ 无权限检查

**需要实现：**
1. 使用 `GetResourcePermission()` 查询 C 权限
2. 检查 `assigned_only`（Caregiver/Nurse 不能创建）
3. 检查 `branch_only`（Manager 只能创建同 branch 的住户）

---

### 3. POST /residents/:id/reset-password (重置密码) - 行 642-709

**当前实现：**
- ❌ 无权限检查

**需要实现：**
1. 检查目标住户是否在权限范围内（assigned_only + branch_only）
2. 或允许 Resident 重置自己的密码

---

### 4. PUT /residents/:id/phi (更新PHI) - 行 710-1235

**当前实现：**
- ✅ Resident/Family 自检（行727-751）
- ❌ 未检查 Staff 权限

**问题：**
```go
// 行727-751: 仅检查 Resident/Family 自检
if (userType == "resident" || userType == "family") && userID != "" {
    // ... 自检逻辑
}
// 缺少 Staff 权限检查
```

**需要修复：**
1. 添加 Staff 权限检查（assigned_only + branch_only）
2. 使用 `GetResourcePermission()` 查询 U 权限

---

### 5. PUT /residents/:id/contacts (更新联系人) - 行 1236-1570

**当前实现：**
- ✅ Resident/Family 自检（行1265-1295）
- ❌ 未检查 Staff 权限

**需要修复：**
1. 添加 Staff 权限检查（assigned_only + branch_only）
2. 使用 `GetResourcePermission()` 查询 U 权限（resident_contacts 资源）

---

### 6. POST /contacts/:contact_id/reset-password (重置联系人密码) - 行 17-67, 1483-1564

**当前实现：**
- ❌ 无权限检查

**需要实现：**
1. 检查目标联系人是否在权限范围内（assigned_only + branch_only）
2. 或允许 Family 重置自己的密码

---

### 7. GET /residents/:id (查看详情) - 行 1571-1947

**当前实现：**
- ✅ Resident/Family 自检（行1605-1629）
- ❌ 未检查 Staff 权限

**需要修复：**
1. 添加 Staff 权限检查（assigned_only + branch_only）
2. 使用 `GetResourcePermission()` 查询 R 权限

---

### 8. PUT /residents/:id (更新住户) - 行 1948-2085

**当前实现：**
- ✅ Resident/Family 自检（行1956-1980）
- ❌ 未检查 Staff 权限

**需要修复：**
1. 添加 Staff 权限检查（assigned_only + branch_only）
2. 使用 `GetResourcePermission()` 查询 U 权限

---

### 9. DELETE /residents/:id (删除住户) - 行 2086-2109

**当前实现：**
- ❌ 无权限检查

**需要实现：**
1. 使用 `GetResourcePermission()` 查询 D 权限
2. 检查 `assigned_only`（Caregiver/Nurse 不能删除）
3. 检查 `branch_only`（Manager 只能删除同 branch 的住户）

---

## 总结

### 待修复问题

1. **GET /residents** - 需要统一权限检查函数，实现 `branch_only` 和空值匹配
2. **POST /residents** - 需要添加 C 权限检查
3. **POST /residents/:id/reset-password** - 需要添加权限检查
4. **PUT /residents/:id/phi** - 需要添加 Staff 权限检查
5. **PUT /residents/:id/contacts** - 需要添加 Staff 权限检查
6. **POST /contacts/:contact_id/reset-password** - 需要添加权限检查
7. **GET /residents/:id** - 需要添加 Staff 权限检查
8. **PUT /residents/:id** - 需要添加 Staff 权限检查
9. **DELETE /residents/:id** - 需要添加 D 权限检查

### 统一修复方案

1. **使用 `GetResourcePermission()`** 统一查询权限配置
2. **使用 `ApplyBranchFilter()`** 实现分支过滤（含空值匹配）
3. **统一权限检查逻辑**：先检查 Resident/Family 自检，再检查 Staff 权限

