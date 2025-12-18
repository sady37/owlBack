# GET /residents/:id 权限矩阵表

## 操作信息
- **路径**: `GET /admin/api/v1/residents/:id`
- **代码位置**: 行 2070-2612
- **资源类型**: `residents`
- **权限类型**: `R` (Read) - 查看详情

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 查看限制 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无限制（可查看所有住户详情） | 租户管理员，可查看所有住户详情 |
| **Manager** | `FALSE` | `TRUE` | 只能查看同 branch 的住户详情<br>如果 `branch_tag=NULL`，只能查看 `units.branch_tag IS NULL OR '-'` 的住户详情 | 分支经理，只能查看同 branch 的住户详情 |
| **IT** | `FALSE` | `FALSE` | 无限制（可查看所有住户详情） | IT支持，可查看所有住户详情 |
| **Caregiver** | `TRUE` | `FALSE` | 只能查看分配的住户详情 | 护理员，只能查看分配的住户详情 |
| **Nurse** | `TRUE` | `FALSE` | 只能查看分配的住户详情 | 护士，只能查看分配的住户详情 |
| **Resident** | N/A | N/A | 只能查看自己的详情 | 住户，只能查看自己的详情 |
| **Family** | N/A | N/A | 只能查看关联的住户详情 | 家属，只能查看关联的住户详情 |

---

## 当前实现问题

### 1. ✅ Resident/Family 自检已实现
- **当前**: 行 2104-2128，已有 Resident/Family 自检逻辑
- **逻辑**: 
  - Resident 只能查看自己的详情
  - Family（resident_contact 登录）只能查看关联的住户详情

### 2. ❌ Staff 权限检查缺失
- **当前**: 行 2129-2612，完全无 Staff 权限检查
- **问题**: 任何 Staff 角色都可以查看任何住户详情（包括 Caregiver/Nurse 查看未分配的住户详情，Manager 查看其他 branch 的住户详情）

---

## 计划实现逻辑

### 1. Resident/Family 自检（已实现，无需修改）
```go
// 行 2104-2128，已有逻辑，保持不变
```

### 2. Staff 权限检查（需要添加）
```go
// 在行 2128 之后，行 2129 之前添加
if userType != "resident" && userType != "family" {
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

    // Check R permission
    var permCheck *PermissionCheck
    if userRole.Valid && userRole.String != "" {
        var err error
        permCheck, err = GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "R")
        if err != nil {
            writeJSON(w, http.StatusOK, Fail("permission denied: failed to check permissions"))
            return
        }

        // Check if R permission record exists
        var hasRPermission bool
        err = s.DB.QueryRowContext(r.Context(),
            `SELECT EXISTS(
                SELECT 1 FROM role_permissions
                WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'R'
            )`,
            SystemTenantID(), userRole.String,
        ).Scan(&hasRPermission)
        if err != nil {
            writeJSON(w, http.StatusOK, Fail("permission denied: failed to verify permissions"))
            return
        }
        if !hasRPermission {
            writeJSON(w, http.StatusOK, Fail("permission denied: no read permission for residents"))
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
        tenantID, actualResidentID,
    ).Scan(&targetBranchTag)
    if err != nil {
        if err == sql.ErrNoRows {
            writeJSON(w, http.StatusOK, Fail("resident not found"))
        } else {
            writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to get resident info: %v", err)))
        }
        return
    }

    // Check assigned_only (Nurse/Caregiver: can only view assigned residents)
    if permCheck.AssignedOnly && userID != "" {
        var isAssigned bool
        err := s.DB.QueryRowContext(r.Context(),
            `SELECT EXISTS(
                SELECT 1 FROM resident_caregivers rc
                WHERE rc.tenant_id = $1
                  AND rc.resident_id::text = $2
                  AND (rc.userList::text LIKE $3 OR rc.userList::text LIKE $4)
            )`,
            tenantID, actualResidentID, userID, "%\""+userID+"\"%",
        ).Scan(&isAssigned)
        if err != nil {
            writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("failed to check assignment: %v", err)))
            return
        }
        if !isAssigned {
            writeJSON(w, http.StatusOK, Fail("permission denied: can only view assigned residents"))
            return
        }
    }

    // Check branch_only (Manager: can only view residents in same branch)
    if permCheck.BranchOnly {
        if !userBranchTag.Valid || userBranchTag.String == "" {
            // User branch_tag is NULL: can only view residents in units with branch_tag IS NULL OR '-'
            if targetBranchTag.Valid && targetBranchTag.String != "" && targetBranchTag.String != "-" {
                writeJSON(w, http.StatusOK, Fail("permission denied: can only view residents in units with branch_tag IS NULL or '-'"))
                return
            }
        } else {
            // User branch_tag has value: can only view residents in matching branch
            if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag.String {
                writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("permission denied: can only view residents in units with branch_tag = %s", userBranchTag.String)))
                return
            }
        }
    }
}
```

---

## 修改后的行为

| 角色 | 权限检查 | 查看限制 | 结果 |
|------|---------|---------|------|
| **Admin** | ✅ 有 R 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可查看所有住户详情 |
| **Manager** (有 branch_tag) | ✅ 有 R 权限（assigned_only=FALSE, branch_only=TRUE） | 只能查看同 branch 的住户详情 | 可查看同 branch 的住户详情 |
| **Manager** (branch_tag=NULL) | ✅ 有 R 权限（assigned_only=FALSE, branch_only=TRUE） | 只能查看 `branch_tag IS NULL OR '-'` 的住户详情 | 可查看未分配 branch 的住户详情 |
| **IT** | ✅ 有 R 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可查看所有住户详情 |
| **Caregiver** | ✅ 有 R 权限（assigned_only=TRUE, branch_only=FALSE） | 只能查看分配的住户详情 | 可查看分配的住户详情 |
| **Nurse** | ✅ 有 R 权限（assigned_only=TRUE, branch_only=FALSE） | 只能查看分配的住户详情 | 可查看分配的住户详情 |
| **Resident** | ✅ 自检 | 只能查看自己的详情 | 可查看自己的详情 |
| **Family** | ✅ 自检 | 只能查看关联的住户详情 | 可查看关联的住户详情 |

---

## 修改要点

1. ✅ Resident/Family 自检：已实现，无需修改
2. ✅ 添加 Staff 权限检查：使用 `GetResourcePermission()` 查询 R 权限（residents 资源）
3. ✅ 添加 assigned_only 检查：Nurse/Caregiver 只能查看分配的住户详情
4. ✅ 添加 branch_only 检查：Manager 只能查看同 branch 的住户详情（含空值匹配）

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

