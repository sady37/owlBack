# POST /contacts/:contact_id/reset-password 权限矩阵表

## 操作信息
- **路径**: `POST /admin/api/v1/contacts/:contact_id/reset-password`
- **代码位置**: 行 14-70, 1630-1700
- **资源类型**: `resident_contacts`
- **权限类型**: `U` (Update) - 重置密码属于更新操作

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 重置限制 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无限制（可重置所有联系人密码） | 租户管理员，可重置所有联系人密码 |
| **Manager** | `FALSE` | `TRUE` | 只能重置同 branch 的住户的联系人密码<br>如果 `branch_tag=NULL`，只能重置 `units.branch_tag IS NULL OR '-'` 的住户的联系人密码 | 分支经理，只能重置同 branch 的住户的联系人密码 |
| **IT** | `FALSE` | `FALSE` | 无限制（可重置所有联系人密码） | IT支持，可重置所有联系人密码 |
| **Caregiver** | ❌ 无 U 权限 | N/A | 不能重置联系人密码 | 护理员，只有查看权限（R），无更新权限（U） |
| **Nurse** | `TRUE` | `FALSE` | 只能重置分配的住户的联系人密码 | 护士，只能重置分配的住户的联系人密码 |
| **Resident** | N/A | N/A | 只能重置自己的 contact 密码 | 住户，只能重置自己的 contact 密码 |
| **Family** | `TRUE` | `FALSE` | 只能重置自己的 contact 密码 | 家属，只能重置自己的 contact 密码 |

---

## 当前实现问题

### 1. ❌ 无权限检查
- **当前**: 行 14-70, 1630-1700，完全无权限检查
- **问题**: 任何角色都可以重置任何联系人的密码（包括 Caregiver/Nurse 重置其他住户的联系人密码）

---

## 计划实现逻辑

### 1. Resident/Family 自检
```go
// 检查是否是住户/家属登录
userType := r.Header.Get("X-User-Type")
if userType == "resident" || userType == "family" {
    userID := r.Header.Get("X-User-Id")
    // 检查是否是 resident_contact 登录
    var foundContactID sql.NullString
    err := s.DB.QueryRowContext(r.Context(),
        `SELECT contact_id::text FROM resident_contacts 
         WHERE tenant_id = $1 AND contact_id::text = $2`,
        tenantID, userID,
    ).Scan(&foundContactID)
    if err == nil && foundContactID.Valid {
        // 这是 resident_contact 登录 - 只能重置自己的密码
        if foundContactID.String != contactID {
            writeJSON(w, http.StatusOK, Fail("access denied: can only reset own password"))
            return
        }
    } else {
        // 这是 resident 登录 - 只能重置自己的 contact 密码
        // 需要检查 contact_id 是否属于该 resident
        var linkedResidentID sql.NullString
        err := s.DB.QueryRowContext(r.Context(),
            `SELECT resident_id::text FROM resident_contacts 
             WHERE tenant_id = $1 AND contact_id::text = $2`,
            tenantID, contactID,
        ).Scan(&linkedResidentID)
        if err != nil || !linkedResidentID.Valid || linkedResidentID.String != userID {
            writeJSON(w, http.StatusOK, Fail("access denied: can only reset password for own contacts"))
            return
        }
    }
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

// 检查 U 权限（Caregiver 没有 U 权限，应该被拒绝；IT 有 U 权限，可以重置）
var permCheck *PermissionCheck
if userRole.Valid && userRole.String != "" {
    var err error
    permCheck, err = GetResourcePermission(s.DB, r.Context(), userRole.String, "resident_contacts", "U")
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("permission denied: failed to check permissions"))
        return
    }
    
    // 检查是否有 U 权限记录（Caregiver 没有 U 权限）
    var hasUPermission bool
    err = s.DB.QueryRowContext(r.Context(),
        `SELECT EXISTS(
            SELECT 1 FROM role_permissions
            WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'resident_contacts' AND permission_type = 'U'
        )`,
        SystemTenantID(), userRole.String,
    ).Scan(&hasUPermission)
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("permission denied: failed to verify permissions"))
        return
    }
    if !hasUPermission {
        writeJSON(w, http.StatusOK, Fail("permission denied: no update permission for resident_contacts"))
        return
    }
} else {
    writeJSON(w, http.StatusOK, Fail("permission denied: no role found"))
    return
}

// 获取目标联系人的 resident_id 和 branch_tag
var targetResidentID sql.NullString
var targetBranchTag sql.NullString
err := s.DB.QueryRowContext(r.Context(),
    `SELECT rc.resident_id::text, COALESCE(u.branch_tag, '') as branch_tag
     FROM resident_contacts rc
     LEFT JOIN residents r ON r.resident_id = rc.resident_id
     LEFT JOIN units u ON u.unit_id = r.unit_id
     WHERE rc.tenant_id = $1 AND rc.contact_id::text = $2`,
    tenantID, contactID,
).Scan(&targetResidentID, &targetBranchTag)
if err != nil {
    if err == sql.ErrNoRows {
        writeJSON(w, http.StatusOK, Fail("contact not found"))
    } else {
        writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get contact info: %v", err)))
    }
    return
}

// 检查 assigned_only（Nurse: can only reset password for contacts of assigned residents）
if permCheck.AssignedOnly && userID != "" {
    var isAssigned bool
    err := s.DB.QueryRowContext(r.Context(),
        `SELECT EXISTS(
            SELECT 1 FROM resident_caregivers rc
            WHERE rc.tenant_id = $1
              AND rc.resident_id::text = $2
              AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
        )`,
        tenantID, targetResidentID.String, userID, "%\""+userID+"\"%",
    ).Scan(&isAssigned)
    if err != nil {
        writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to check assignment: %v", err)))
        return
    }
    if !isAssigned {
        writeJSON(w, http.StatusOK, Fail("permission denied: can only reset password for contacts of assigned residents"))
        return
    }
}

// 检查 branch_only（Manager: can only reset password for contacts of residents in same branch）
if permCheck.BranchOnly {
    if !userBranchTag.Valid || userBranchTag.String == "" {
        // User branch_tag is NULL: can only reset password for contacts of residents in units with branch_tag IS NULL OR '-'
        if targetBranchTag.Valid && targetBranchTag.String != "" && targetBranchTag.String != "-" {
            writeJSON(w, http.StatusOK, Fail("permission denied: can only reset password for contacts of residents in units with branch_tag IS NULL or '-'"))
            return
        }
    } else {
        // User branch_tag has value: can only reset password for contacts of residents in matching branch
        if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag.String {
            writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("permission denied: can only reset password for contacts of residents in units with branch_tag = %s", userBranchTag.String)))
            return
        }
    }
}
```

---

## 修改后的行为

| 角色 | 权限检查 | 重置限制 | 结果 |
|------|---------|---------|------|
| **Admin** | ✅ 有 U 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可重置所有联系人密码 |
| **Manager** (有 branch_tag) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能重置同 branch 的住户的联系人密码 | 可重置同 branch 的住户的联系人密码 |
| **Manager** (branch_tag=NULL) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能重置 `branch_tag IS NULL OR '-'` 的住户的联系人密码 | 可重置未分配 branch 的住户的联系人密码 |
| **IT** | ✅ 有 U 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可重置所有联系人密码 |
| **Caregiver** | ❌ 无 U 权限 | 不能重置联系人密码 | 返回 "permission denied" |
| **Nurse** | ✅ 有 U 权限（assigned_only=TRUE, branch_only=FALSE） | 只能重置分配的住户的联系人密码 | 可重置分配的住户的联系人密码 |
| **Resident** | ✅ 自检 | 只能重置自己的 contact 密码 | 可重置自己的 contact 密码 |
| **Family** | ✅ 自检 | 只能重置自己的 contact 密码 | 可重置自己的 contact 密码 |

---

## 修改要点

1. ✅ 添加 Resident/Family 自检：只能重置自己或自己的 contact 密码
2. ✅ 添加 Staff 权限检查：使用 `GetResourcePermission()` 查询 U 权限（resident_contacts 资源）
3. ✅ 添加 assigned_only 检查：Nurse 只能重置分配的住户的联系人密码
4. ✅ 添加 branch_only 检查：Manager 只能重置同 branch 的住户的联系人密码（含空值匹配）
5. ✅ 拒绝无权限角色：Caregiver 无 U 权限，拒绝重置；IT 有 U 权限，可以重置

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

