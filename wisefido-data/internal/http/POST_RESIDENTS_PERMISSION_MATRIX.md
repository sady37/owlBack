# POST /residents 权限矩阵表

## 操作信息
- **路径**: `POST /admin/api/v1/residents`
- **代码位置**: 行 305-638
- **资源类型**: `residents`
- **权限类型**: `C` (Create)

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 创建限制 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无限制（可创建所有住户） | 租户管理员，可创建所有住户 |
| **Manager** | `FALSE` | `TRUE` | 只能创建同 branch 的住户<br>如果 `branch_tag=NULL`，只能创建 `units.branch_tag IS NULL OR '-'` 的住户 | 分支经理，只能创建同 branch 的住户 |
| **IT** | ❌ 无 C 权限 | N/A | 不能创建住户 | IT支持，无创建权限 |
| **Caregiver** | ❌ 无 C 权限 | N/A | 不能创建住户 | 护理员，无创建权限 |
| **Nurse** | ❌ 无 C 权限 | N/A | 不能创建住户 | 护士，无创建权限 |
| **Resident** | N/A | N/A | 不能创建住户 | 住户，无创建权限 |
| **Family** | N/A | N/A | 不能创建住户 | 家属，无创建权限 |

---

## 当前实现问题

### 1. ❌ 无权限检查
- **当前**: 行 305-638，完全无权限检查
- **问题**: 任何角色都可以创建住户（包括 Caregiver/Nurse/Resident/Family）

---

## 计划实现逻辑

### 1. 权限检查
```go
// 获取当前用户角色
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

// 检查 C 权限
var permCheck *PermissionCheck
if userRole.Valid && userRole.String != "" {
    var err error
    permCheck, err = GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "C")
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("failed to check permissions"))
        return
    }
} else {
    // 无角色信息：拒绝创建
    writeJSON(w, http.StatusOK, Fail("permission denied: no role found"))
    return
}

// 检查是否有 C 权限（通过 assigned_only 和 branch_only 判断）
// 如果 role_permissions 表中没有 C 权限记录，GetResourcePermission 会返回最严格权限
// 但我们需要明确检查：如果查询失败或记录不存在，应该拒绝创建
// 注意：GetResourcePermission 在记录不存在时返回 assigned_only=true, branch_only=true
// 我们需要额外检查：如果查询返回错误（非 ErrNoRows），说明查询失败，应该拒绝
```

### 2. 分支过滤（创建时）
```go
// 如果 branch_only=TRUE，需要检查创建的住户是否在允许的 branch 范围内
// 创建时，住户的 branch_tag 由 unit_id 决定（通过 units.branch_tag）
// 需要验证：如果 Manager 的 branch_tag=NULL，只能创建 unit.branch_tag IS NULL OR '-' 的住户
// 如果 Manager 的 branch_tag="BranchA"，只能创建 unit.branch_tag="BranchA" 的住户

// 获取 unit_id 对应的 branch_tag
if unitID != "" {
    var unitBranchTag sql.NullString
    err := s.DB.QueryRowContext(r.Context(),
        `SELECT branch_tag FROM units WHERE tenant_id = $1 AND unit_id::text = $2`,
        tenantID, unitID,
    ).Scan(&unitBranchTag)
    if err != nil {
        writeJSON(w, http.StatusOK, Fail("unit not found"))
        return
    }
    
    // 如果 branch_only=TRUE，检查 branch_tag 是否匹配
    if permCheck.BranchOnly {
        if !userBranchTag.Valid || userBranchTag.String == "" {
            // 用户 branch_tag 为 NULL：只能创建 branch_tag IS NULL OR '-' 的 unit
            if unitBranchTag.Valid && unitBranchTag.String != "" && unitBranchTag.String != "-" {
                writeJSON(w, http.StatusOK, Fail("permission denied: can only create residents in units with branch_tag IS NULL or '-'"))
                return
            }
        } else {
            // 用户 branch_tag 有值：只能创建匹配的 branch
            if !unitBranchTag.Valid || unitBranchTag.String != userBranchTag.String {
                writeJSON(w, http.StatusOK, Fail(fmt.Sprintf("permission denied: can only create residents in units with branch_tag = %s", userBranchTag.String)))
                return
            }
        }
    }
}
```

---

## 修改后的行为

| 角色 | 权限检查 | 创建限制 | 结果 |
|------|---------|---------|------|
| **Admin** | ✅ 有 C 权限（assigned_only=FALSE, branch_only=FALSE） | 无限制 | 可创建所有住户 |
| **Manager** (有 branch_tag) | ✅ 有 C 权限（assigned_only=FALSE, branch_only=TRUE） | 只能创建同 branch 的住户 | 可创建同 branch 的住户 |
| **Manager** (branch_tag=NULL) | ✅ 有 C 权限（assigned_only=FALSE, branch_only=TRUE） | 只能创建 `branch_tag IS NULL OR '-'` 的住户 | 可创建未分配 branch 的住户 |
| **IT** | ❌ 无 C 权限 | 拒绝创建 | 返回 "permission denied" |
| **Caregiver** | ❌ 无 C 权限 | 拒绝创建 | 返回 "permission denied" |
| **Nurse** | ❌ 无 C 权限 | 拒绝创建 | 返回 "permission denied" |
| **Resident** | ❌ 无角色信息 | 拒绝创建 | 返回 "permission denied" |
| **Family** | ❌ 无角色信息 | 拒绝创建 | 返回 "permission denied" |

---

## 修改要点

1. ✅ 添加权限检查：使用 `GetResourcePermission()` 查询 C 权限
2. ✅ 添加分支过滤：如果 `branch_only=TRUE`，检查创建的住户是否在允许的 branch 范围内
3. ✅ 拒绝无权限角色：如果无 C 权限或查询失败，拒绝创建

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

