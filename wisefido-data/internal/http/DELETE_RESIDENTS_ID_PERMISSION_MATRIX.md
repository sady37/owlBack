# DELETE /residents/:id 权限矩阵表

## 操作信息
- **路径**: `DELETE /admin/api/v1/residents/:id`
- **代码位置**: 行 2770-2790
- **资源类型**: `residents`
- **权限类型**: `D` (Delete) - 软删除（设置 status = 'discharged'）

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 删除限制 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无限制（可删除所有住户） | 租户管理员，可删除所有住户 |
| **Manager** | `FALSE` | `TRUE` | 只能删除同 branch 的住户<br>如果 `branch_tag=NULL`，只能删除 `units.branch_tag IS NULL OR '-'` 的住户 | 分支经理，只能删除同 branch 的住户 |
| **IT** | `FALSE` | `FALSE` | 无限制（可删除所有住户） | IT支持，可删除所有住户 |
| **Caregiver** | ❌ 无 D 权限 | N/A | 不能删除住户 | 护理员，只有查看权限（R），无删除权限（D） |
| **Nurse** | `TRUE` | `FALSE` | 只能删除分配的住户 | 护士，只能删除分配的住户 |
| **Resident** | N/A | N/A | 不能删除（包括自己） | 住户，不能删除任何住户 |
| **Family** | N/A | N/A | 不能删除（包括关联的住户） | 家属，不能删除任何住户 |

---

## 当前实现问题

### 1. ❌ 完全无权限检查
- **当前**: 行 2770-2790，完全无权限检查
- **问题**: 任何角色都可以删除任何住户（包括 Caregiver/Nurse 删除未分配的住户，Manager 删除其他 branch 的住户，Resident/Family 删除住户）

---

## 计划实现逻辑

### 1. Resident/Family 拒绝删除
```go
// 拒绝 Resident/Family 删除任何住户（包括自己）
userID := r.Header.Get("X-User-Id")
userType := r.Header.Get("X-User-Type")
if userType == "resident" || userType == "family" {
    writeJSON(w, http.StatusOK, Fail("permission denied: resident/family cannot delete residents"))
    return
}
```

### 2. Staff 权限检查
```go
// Staff permission check
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

// Check D permission (Caregiver has no D permission, should be denied)
var permCheck *PermissionCheck
if userRole.Valid && userRole.String != "" {
    var err error
    permCheck, err = GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "D")
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("permission denied: failed to check permissions"))
        return
    }

    // Check if D permission record exists (Caregiver has no D permission)
    var hasDPermission bool
    err = s.DB.QueryRowContext(r.Context(),
        `SELECT EXISTS(
            SELECT 1 FROM role_permissions
            WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'D'
        )`,
        SystemTenantID(), userRole.String,
    ).Scan(&hasDPermission)
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("permission denied: failed to verify permissions"))
        return
    }
    if !hasDPermission {
        writeJSON(w, http.StatusOK, Fail("permission denied: no delete permission for residents"))
        return
    }
} else {
    writeJSON(w, http.StatusOK, Fail("permission denied: no role found"))
    return
}

// Get target resident's branch_tag
var targetBranchTag sql.NullString
err := s.DB.QueryRowContext(r.Context(),
    `SELECT COALESCE(u.branch_tag, '') as branch_tag
     FROM residents r
     LEFT JOIN units u ON u.unit_id = r.unit_id
     WHERE r.tenant_id = $1 AND r.resident_id::text = $2`,
    tenantID, id,
).Scan(&targetBranchTag)
if err != nil {
    if err == sql.ErrNoRows {
        writeJSON(w, http.StatusOK, Fail("resident not found"))
    } else {
        writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get resident info: %v", err)))
    }
    return
}

// Check assigned_only (Nurse: can only delete assigned residents)
if permCheck.AssignedOnly && userID != "" {
    var isAssigned bool
    err := s.DB.QueryRowContext(r.Context(),
        `SELECT EXISTS(
            SELECT 1 FROM resident_caregivers rc
            WHERE rc.tenant_id = $1
              AND rc.resident_id::text = $2
              AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
        )`,
        tenantID, id, userID, "%\""+userID+"\"%",
    ).Scan(&isAssigned)
    if err != nil {
        writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to check assignment: %v", err)))
        return
    }
    if !isAssigned {
        writeJSON(w, http.StatusOK, Fail("permission denied: can only delete assigned residents"))
        return
    }
}

// Check branch_only (Manager: can only delete residents in same branch)
if permCheck.BranchOnly {
    if !userBranchTag.Valid || userBranchTag.String == "" {
        // User branch_tag is NULL: can only delete residents in units with branch_tag IS NULL OR '-'
        if targetBranchTag.Valid && targetBranchTag.String != "" && targetBranchTag.String != "-" {
            writeJSON(w, http.StatusOK, Fail("permission denied: can only delete residents in units with branch_tag IS NULL or '-'"))
            return
        }
    } else {
        // User branch_tag has value: can only delete residents in matching branch
        if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag.String {
            writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("permission denied: can only delete residents in units with branch_tag = %s", userBranchTag.String)))
            return
        }
    }
}
```

---

## 修改后的行为

| 角色 | 权限检查 | 删除限制 | 结果 |
|------|---------|---------|------|
| **Admin** | ✅ 有 D 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可删除所有住户 |
| **Manager** (有 branch_tag) | ✅ 有 D 权限（assigned_only=FALSE, branch_only=TRUE） | 只能删除同 branch 的住户 | 可删除同 branch 的住户 |
| **Manager** (branch_tag=NULL) | ✅ 有 D 权限（assigned_only=FALSE, branch_only=TRUE） | 只能删除 `branch_tag IS NULL OR '-'` 的住户 | 可删除未分配 branch 的住户 |
| **IT** | ✅ 有 D 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可删除所有住户 |
| **Caregiver** | ❌ 无 D 权限 | 不能删除住户 | 返回 "permission denied" |
| **Nurse** | ✅ 有 D 权限（assigned_only=TRUE, branch_only=FALSE） | 只能删除分配的住户 | 可删除分配的住户 |
| **Resident** | ❌ 拒绝 | 不能删除任何住户 | 返回 "permission denied" |
| **Family** | ❌ 拒绝 | 不能删除任何住户 | 返回 "permission denied" |

---

## 修改要点

1. ✅ 拒绝 Resident/Family：不能删除任何住户（包括自己）
2. ✅ 添加 Staff 权限检查：使用 `GetResourcePermission()` 查询 D 权限（residents 资源）
3. ✅ 添加 assigned_only 检查：Nurse 只能删除分配的住户
4. ✅ 添加 branch_only 检查：Manager 只能删除同 branch 的住户（含空值匹配）
5. ✅ 拒绝无权限角色：Caregiver 无 D 权限，拒绝删除

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

