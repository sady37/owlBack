# PUT /residents/:id/contacts 权限矩阵表

## 操作信息
- **路径**: `PUT /admin/api/v1/residents/:id/contacts`
- **代码位置**: 行 1400-1570
- **资源类型**: `resident_contacts`
- **权限类型**: `U` (Update)

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 更新限制 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无限制（可更新所有联系人） | 租户管理员，可更新所有联系人 |
| **Manager** | `FALSE` | `TRUE` | 只能更新同 branch 的住户的联系人<br>如果 `branch_tag=NULL`，只能更新 `units.branch_tag IS NULL OR '-'` 的住户的联系人 | 分支经理，只能更新同 branch 的住户的联系人 |
| **IT** | ❌ 无 U 权限 | N/A | 不能更新联系人 | IT支持，无更新权限 |
| **Caregiver** | ❌ 无 U 权限 | N/A | 不能更新联系人 | 护理员，只有查看权限（R），无更新权限（U） |
| **Nurse** | `TRUE` | `FALSE` | 只能更新分配的住户的联系人 | 护士，只能更新分配的住户的联系人 |
| **Resident** | ❌ 无 U 权限（只有 R） | N/A | ⚠️ 业务例外：允许更新自己的联系人 | 住户，虽然数据库只有 R 权限，但业务上允许更新自己的联系人 |
| **Family** | `TRUE` | `FALSE` | 只能更新自己的 contact slot | 家属，只能更新自己的 contact slot |

---

## 当前实现问题

### 1. ⚠️ 仅 Resident/Family 自检
- **当前**: 行 1439-1369，仅检查 Resident/Family 自检
- **问题**: 缺少 Staff 权限检查（assigned_only + branch_only）

---

## 计划实现逻辑

### 1. Resident/Family 自检（保持现有逻辑，但需要完善）
```go
// 检查是否是住户/家属登录
userType := r.Header.Get("X-User-Type")
if userType == "resident" || userType == "family" {
    userID := r.Header.Get("X-User-Id")
    // 检查是否是 resident_contact 登录
    var foundResidentID sql.NullString
    var foundSlot sql.NullString
    err := s.DB.QueryRowContext(r.Context(),
        `SELECT resident_id::text, slot FROM resident_contacts 
         WHERE tenant_id = $1 AND contact_id::text = $2`,
        tenantID, userID,
    ).Scan(&foundResidentID, &foundSlot)
    if err == nil && foundResidentID.Valid {
        // 这是 resident_contact 登录 - 只能修改自己的 slot
        if foundResidentID.String != residentID {
            writeJSON(w, http.StatusOK, Fail("access denied: can only modify contacts for linked resident"))
            return
        }
        // 验证 slot 是否匹配
        if foundSlot.Valid && foundSlot.String != slot {
            writeJSON(w, http.StatusOK, Fail("access denied: can only modify own slot"))
            return
        }
    } else {
        // 这是 resident 登录 - 只能修改自己的联系人
        if userID != residentID {
            writeJSON(w, http.StatusOK, Fail("access denied: can only modify contacts for self"))
            return
        }
    }
}
```

### 2. Staff 权限检查（新增）
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

// 检查 U 权限（IT/Caregiver 没有 U 权限，应该被拒绝）
var permCheck *PermissionCheck
if userRole.Valid && userRole.String != "" {
    var err error
    permCheck, err = GetResourcePermission(s.DB, r.Context(), userRole.String, "resident_contacts", "U")
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("permission denied: failed to check permissions"))
        return
    }
    
    // 检查是否有 U 权限记录（IT/Caregiver 没有 U 权限）
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

// 检查 assigned_only（Nurse: can only update contacts for assigned residents）
if permCheck.AssignedOnly && userID != "" {
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
        writeJSON(w, http.StatusOK, Fail("permission denied: can only update contacts for assigned residents"))
        return
    }
}

// 检查 branch_only（Manager: can only update contacts for residents in same branch）
if permCheck.BranchOnly {
    if !userBranchTag.Valid || userBranchTag.String == "" {
        // User branch_tag is NULL: can only update contacts for residents in units with branch_tag IS NULL OR '-'
        if targetBranchTag.Valid && targetBranchTag.String != "" && targetBranchTag.String != "-" {
            writeJSON(w, http.StatusOK, Fail("permission denied: can only update contacts for residents in units with branch_tag IS NULL or '-'"))
            return
        }
    } else {
        // User branch_tag has value: can only update contacts for residents in matching branch
        if !targetBranchTag.Valid || targetBranchTag.String != userBranchTag.String {
            writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("permission denied: can only update contacts for residents in units with branch_tag = %s", userBranchTag.String)))
            return
        }
    }
}
```

---

## 修改后的行为

| 角色 | 权限检查 | 更新限制 | 结果 |
|------|---------|---------|------|
| **Admin** | ✅ 有 U 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可更新所有联系人 |
| **Manager** (有 branch_tag) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能更新同 branch 的住户的联系人 | 可更新同 branch 的住户的联系人 |
| **Manager** (branch_tag=NULL) | ✅ 有 U 权限（assigned_only=FALSE, branch_only=TRUE） | 只能更新 `branch_tag IS NULL OR '-'` 的住户的联系人 | 可更新未分配 branch 的住户的联系人 |
| **IT** | ❌ 无 U 权限 | 不能更新联系人 | 返回 "permission denied" |
| **Caregiver** | ❌ 无 U 权限 | 不能更新联系人 | 返回 "permission denied" |
| **Nurse** | ✅ 有 U 权限（assigned_only=TRUE, branch_only=FALSE） | 只能更新分配的住户的联系人 | 可更新分配的住户的联系人 |
| **Resident** | ✅ 自检 | 只能更新自己的联系人 | 可更新自己的联系人 |
| **Family** | ✅ 自检 | 只能更新自己的 contact slot | 可更新自己的 contact slot |

---

## 修改要点

1. ✅ 保持 Resident/Family 自检逻辑（已存在，需要完善）
2. ✅ 添加 Staff 权限检查：使用 `GetResourcePermission()` 查询 U 权限（resident_contacts 资源）
3. ✅ 添加 assigned_only 检查：Nurse 只能更新分配的住户的联系人
4. ✅ 添加 branch_only 检查：Manager 只能更新同 branch 的住户的联系人（含空值匹配）
5. ✅ 拒绝无权限角色：IT/Caregiver 无 U 权限，拒绝更新

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

