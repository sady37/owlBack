# PUT /residents/:id 权限矩阵表

## 操作信息
- **路径**: `PUT /admin/api/v1/residents/:id`
- **代码位置**: 行 2548-2685
- **资源类型**: `residents`
- **权限类型**: `U` (Update) - 更新住户信息

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 更新限制 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无限制（可更新所有住户信息） | 租户管理员，可更新所有住户信息 |
| **Manager** | `FALSE` | `TRUE` | 只能更新同 branch 的住户信息<br>如果 `branch_tag=NULL`，只能更新 `units.branch_tag IS NULL OR '-'` 的住户信息 | 分支经理，只能更新同 branch 的住户信息 |
| **IT** | `FALSE` | `FALSE` | 无限制（可更新所有住户信息） | IT支持，可更新所有住户信息 |
| **Caregiver** | ❌ 无 U 权限 | N/A | 不能更新住户信息 | 护理员，只有查看权限（R），无更新权限（U） |
| **Nurse** | `TRUE` | `FALSE` | 只能更新分配的住户信息 | 护士，只能更新分配的住户信息 |
| **Resident** | N/A | N/A | 只能更新自己的信息 | 住户，只能更新自己的信息 |
| **Family** | N/A | N/A | 只能更新关联的住户信息 | 家属，只能更新关联的住户信息 |

---

## 当前实现问题

### 1. ✅ Resident/Family 自检已实现
- **当前**: 行 2556-2580，已有 Resident/Family 自检逻辑
- **逻辑**: 
  - Resident 只能更新自己的信息
  - Family（resident_contact 登录）只能更新关联的住户信息

### 2. ❌ Staff 权限检查缺失
- **当前**: 行 2580-2685，完全无 Staff 权限检查
- **问题**: 任何 Staff 角色都可以更新任何住户信息（包括 Caregiver 更新住户信息，Manager 更新其他 branch 的住户信息）

---

## 计划实现逻辑

### 1. Resident/Family 自检（已实现，无需修改）
```go
// 行 2556-2580，已有逻辑，保持不变
```

### 2. Staff 权限检查（需要添加）
```go
// 在行 2580 之后，行 2581 之前添加
} else {
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

    // Check U permission (Caregiver has no U permission, should be denied)
    var permCheck *PermissionCheck
    if userRole.Valid && userRole.String != "" {
        var err error
        permCheck, err = GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "U")
        if err != nil {
            writeJSON(w, http.StatusOK, Fail("permission denied: failed to check permissions"))
            return
        }

        // Check if U permission record exists (Caregiver has no U permission)
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

    // Check assigned_only (Nurse: can only update assigned residents)
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
            writeJSON(w, http.StatusOK, Fail("permission denied: can only update assigned residents"))
            return
        }
    }

    // Check branch_only (Manager: can only update residents in same branch)
    if permCheck.BranchOnly {
        if !userBranchTag.Valid || userBranchTag.String == "" {
            // User branch_tag is NULL: can only update residents in units with branch_tag IS NULL OR '-'
            if targetBranchTag.Valid && targetBranchTag.String != "" && targetBranchTag.String != "-" {
                writeJSON(w, http.StatusOK, Fail("permission denied: can only update residents in units with branch_tag IS NULL or '-'"))
                return
            }
        } else {
            // User branch_tag has value: can only update residents in matching branch
            if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag.String {
                writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("permission denied: can only update residents in units with branch_tag = %s", userBranchTag.String)))
                return
            }
        }
    }
}
```

---

## 修改后的行为

| 角色 | 权限检查 | 更新限制 | 结果 |
|------|---------|---------|------|
| **Admin** | ✅ 有 U 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可更新所有住户信息 |
| **Manager** (有 branch_tag) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能更新同 branch 的住户信息 | 可更新同 branch 的住户信息 |
| **Manager** (branch_tag=NULL) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能更新 `branch_tag IS NULL OR '-'` 的住户信息 | 可更新未分配 branch 的住户信息 |
| **IT** | ✅ 有 U 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可更新所有住户信息 |
| **Caregiver** | ❌ 无 U 权限 | 不能更新住户信息 | 返回 "permission denied" |
| **Nurse** | ✅ 有 U 权限（assigned_only=TRUE, branch_only=FALSE） | 只能更新分配的住户信息 | 可更新分配的住户信息 |
| **Resident** | ✅ 自检 | 只能更新自己的信息 | 可更新自己的信息 |
| **Family** | ✅ 自检 | 只能更新关联的住户信息 | 可更新关联的住户信息 |

---

## 修改要点

1. ✅ Resident/Family 自检：已实现，无需修改
2. ✅ 添加 Staff 权限检查：使用 `GetResourcePermission()` 查询 U 权限（residents 资源）
3. ✅ 添加 assigned_only 检查：Nurse 只能更新分配的住户信息
4. ✅ 添加 branch_only 检查：Manager 只能更新同 branch 的住户信息（含空值匹配）
5. ✅ 拒绝无权限角色：Caregiver 无 U 权限，拒绝更新

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

