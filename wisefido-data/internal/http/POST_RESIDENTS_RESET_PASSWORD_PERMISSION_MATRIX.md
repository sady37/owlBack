# POST /residents/:id/reset-password 权限矩阵表

## 操作信息
- **路径**: `POST /admin/api/v1/residents/:id/reset-password`
- **代码位置**: 行 720-787
- **资源类型**: `residents`
- **权限类型**: `U` (Update) - 重置密码属于更新操作

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 重置限制 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无限制（可重置所有住户密码） | 租户管理员，可重置所有住户密码 |
| **Manager** | `FALSE` | `TRUE` | 只能重置同 branch 的住户密码<br>如果 `branch_tag=NULL`，只能重置 `units.branch_tag IS NULL OR '-'` 的住户密码 | 分支经理，只能重置同 branch 的住户密码 |
| **IT** | `FALSE` | `FALSE` | 无限制（可重置所有住户密码） | IT支持，可重置所有住户密码 |
| **Caregiver** | ❌ 无 U 权限 | N/A | 不能重置住户密码 | 护理员，只有查看权限（R），无更新权限（U） |
| **Nurse** | `TRUE` | `FALSE` | 只能重置分配的住户密码 | 护士，只能重置分配的住户密码 |
| **Resident** | N/A | N/A | 只能重置自己的密码<br>（相连的 contact 密码通过 `/contacts/:contact_id/reset-password` 重置） | 住户，只能重置自己的密码，以及通过 `/contacts/:contact_id/reset-password` 重置相连的 contact 密码 |
| **Family** | N/A | N/A | ❌ 不通过此端点 | 家属，通过 `/contacts/:contact_id/reset-password` 重置自己的 contact 密码 |

---

## 当前实现问题

### 1. ❌ 无权限检查
- **当前**: 行 720-787，完全无权限检查
- **问题**: 任何角色都可以重置任何住户的密码（包括 Caregiver/Nurse 重置其他住户的密码）

---

## 计划实现逻辑

### 1. Resident 自检（Family 不通过此端点）
```go
// 检查是否是住户登录（Family 通过 /contacts/:contact_id/reset-password 重置自己的 contact 密码）
userType := r.Header.Get("X-User-Type")
if userType == "resident" {
    userID := r.Header.Get("X-User-Id")
    // Resident 登录 - 只能重置自己的密码
    if userID != residentID {
        writeJSON(w, http.StatusOK, Fail("access denied: can only reset own password"))
        return
    }
} else if userType == "family" {
    // Family 不通过此端点重置密码，应该通过 /contacts/:contact_id/reset-password
    writeJSON(w, http.StatusOK, Fail("access denied: family should use /contacts/:contact_id/reset-password to reset contact password"))
    return
}
```

### 2. Staff 权限检查
```go
// 获取当前用户角色和 branch_tag
userID := r.Header.Get("X-User-Id")
var userRole, userBranchTag sql.NullString
if userID != "" {
    err := s.DB.QueryRowContext(r.Context(),
        `SELECT role, branch_tag FROM users WHERE tenant_id = $1 AND user_id::text = $2`,
        tenantID, userID,
    ).Scan(&userRole, &userBranchTag)
    if err != nil && err != sql.ErrNoRows {
        fmt.Printf("[AdminResidents] Failed to get user info: %v\n", err)
    }
}

// 检查 U 权限（Caregiver 没有 U 权限，应该被拒绝）
var permCheck *PermissionCheck
if userRole.Valid && userRole.String != "" {
    var err error
    permCheck, err = GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "U")
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("permission denied: failed to check permissions"))
        return
    }
    
    // 检查是否有 U 权限记录（Caregiver 没有 U 权限，GetResourcePermission 会返回最严格权限）
    var hasUPermission bool
    err = s.DB.QueryRowContext(r.Context(),
        `SELECT EXISTS(
            SELECT 1 FROM role_permissions
            WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'U'
        )`,
        SystemTenantID(), userRole.String,
    ).Scan(&hasUPermission)
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("permission denied: failed to verify permissions"))
        return
    }
    if !hasUPermission {
        writeJSON(w, http.StatusOK, Fail("permission denied: no update permission for residents"))
        return
    }
} else {
    writeJSON(w, http.StatusOK, Fail("permission denied: no role found"))
    return
}

// 检查目标住户是否在权限范围内
// 1. 如果 assigned_only=TRUE，检查住户是否分配给当前用户
// 2. 如果 branch_only=TRUE，检查住户的 branch_tag 是否匹配

// 获取目标住户的 unit_id 和 branch_tag
var targetUnitID sql.NullString
var targetBranchTag sql.NullString
err := s.DB.QueryRowContext(r.Context(),
    `SELECT r.unit_id::text, COALESCE(u.branch_tag, '') as branch_tag
     FROM residents r
     LEFT JOIN units u ON u.unit_id = r.unit_id
     WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
    tenantID, residentID,
).Scan(&targetUnitID, &targetBranchTag)
if err != nil {
    if err == sql.ErrNoRows {
        writeJSON(w, http.StatusOK, Fail("resident not found"))
    } else {
        writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get resident info: %v", err)))
    }
    return
}

// 检查 assigned_only
if permCheck.AssignedOnly && userID != "" {
    // 检查住户是否分配给当前用户
    var isAssigned bool
    err := s.DB.QueryRowContext(r.Context(),
        `SELECT EXISTS(
            SELECT 1 FROM resident_caregivers rc
            WHERE rc.tenant_id = $1
              AND rc.resident_id::text = $2
              AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
        )`,
        tenantID, residentID, userID, "%\""+userID+"\"%",
    ).Scan(&isAssigned)
    if err != nil {
        writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to check assignment: %v", err)))
        return
    }
    if !isAssigned {
        writeJSON(w, http.StatusOK, Fail("permission denied: can only reset password for assigned residents"))
        return
    }
}

// 检查 branch_only
if permCheck.BranchOnly {
    // 检查 branch_tag 是否匹配（含空值匹配）
    if !userBranchTag.Valid || userBranchTag.String == "" {
        // 用户 branch_tag 为 NULL：只能重置 branch_tag IS NULL OR '-' 的住户
        if targetBranchTag.Valid && targetBranchTag.String != "" && targetBranchTag.String != "-" {
            writeJSON(w, http.StatusOK, Fail("permission denied: can only reset password for residents in units with branch_tag IS NULL or '-'"))
            return
        }
    } else {
        // 用户 branch_tag 有值：只能重置匹配的 branch
        if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag.String {
            writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("permission denied: can only reset password for residents in units with branch_tag = %s", userBranchTag.String)))
            return
        }
    }
}
```

---

## 修改后的行为

| 角色 | 权限检查 | 重置限制 | 结果 |
|------|---------|---------|------|
| **Admin** | ✅ 有 U 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可重置所有住户密码 |
| **IT** | ✅ 有 U 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可重置所有住户密码 |
| **Manager** (有 branch_tag) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能重置同 branch 的住户密码 | 可重置同 branch 的住户密码 |
| **Manager** (branch_tag=NULL) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能重置 `branch_tag IS NULL OR '-'` 的住户密码 | 可重置未分配 branch 的住户密码 |
| **Caregiver** | ❌ 无 U 权限 | 不能重置住户密码 | 返回 "permission denied" |
| **Nurse** | ✅ 有 U 权限（assigned_only=TRUE, branch_only=FALSE） | 只能重置分配的住户密码 | 可重置分配的住户密码 |
| **Resident** | ✅ 自检 | 只能重置自己的密码<br>（相连的 contact 密码通过 `/contacts/:contact_id/reset-password` 重置） | 可重置自己的密码 |
| **Family** | ❌ 不通过此端点 | 通过 `/contacts/:contact_id/reset-password` 重置 | 不适用 |

---

## 修改要点

1. ✅ 添加 Resident/Family 自检：只能重置自己或关联住户的密码
2. ✅ 添加 Staff 权限检查：使用 `GetResourcePermission()` 查询 U 权限
3. ✅ 添加 assigned_only 检查：Caregiver/Nurse 只能重置分配的住户密码
4. ✅ 添加 branch_only 检查：Manager 只能重置同 branch 的住户密码（含空值匹配）

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

