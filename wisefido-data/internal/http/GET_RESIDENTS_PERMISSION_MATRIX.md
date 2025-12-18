# GET /residents 权限矩阵表

## 操作信息
- **路径**: `GET /admin/api/v1/residents`
- **代码位置**: 行 73-304
- **资源类型**: `residents`
- **权限类型**: `R` (Read)

---

## 权限矩阵表（基于 role_permissions 表）

| 角色 | assigned_only | branch_only | 过滤逻辑 | 说明 |
|------|--------------|-------------|---------|------|
| **Admin** | `FALSE` | `FALSE` | 无过滤（显示所有住户） | 租户管理员，可查看所有住户 |
| **IT** | `FALSE` | `FALSE` | 无过滤（显示所有住户） | IT支持，可查看所有住户 |
| **Manager** | `FALSE` | `TRUE` | 分支过滤：`units.branch_tag = users.branch_tag`<br>空值匹配：`branch_tag IS NULL` 时匹配 `units.branch_tag IS NULL OR '-'` | 分支经理，只能查看同 branch 的住户 |
| **Caregiver** | `TRUE` | `FALSE` | 分配过滤：通过 `resident_caregivers.userList` 过滤 | 护理员，只能查看分配的住户 |
| **Nurse** | `TRUE` | `FALSE` | 分配过滤：通过 `resident_caregivers.userList` 过滤 | 护士，只能查看分配的住户 |
| **Resident** | N/A | N/A | 自检过滤：`resident_id = $userID` | 住户，只能查看自己 |
| **Family** | N/A | N/A | 自检过滤：`resident_id = $linkedResidentID` | 家属，只能查看关联的住户 |

---

## 当前实现问题

### 1. 未使用统一权限检查函数
- **当前**: 行122-138，直接查询 `assigned_only`
- **问题**: 未查询 `branch_only`，未使用 `GetResourcePermission()`

### 2. Manager 分支过滤硬编码
- **当前**: 行194-196，硬编码 `userRole.String == "Manager"` 判断
- **问题**: 未实现空值匹配（`branch_tag=NULL` 时错误显示所有）

### 3. Caregiver/Nurse 逻辑复杂
- **当前**: 行168-189，依赖 `alarm_scope` 字段
- **问题**: `assigned_only=TRUE` 但 `alarm_scope='ALL'` 时显示所有（错误）

---

## 计划实现逻辑

### 1. Resident/Family（住户/家属）
```go
if isResidentLogin {
    // 保持不变：仅显示自己或关联住户
    if residentIDForSelf.Valid {
        args = append(args, residentIDForSelf.String)
        q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND r.resident_id::text = $%d`, len(args))
    } else {
        q += ` WHERE 1=0`
    }
}
```

### 2. Staff（Admin/IT/Manager/Caregiver/Nurse）
```go
// 使用 GetResourcePermission() 统一查询
permCheck, err := GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "R")
if err != nil {
    // 默认最严格权限
    permCheck = &PermissionCheck{AssignedOnly: true, BranchOnly: true}
}

// 构建基础 WHERE 条件
q += ` WHERE r.tenant_id = $1`

// 应用 assigned_only 过滤
if permCheck.AssignedOnly && userID != "" {
    // Caregiver/Nurse: 通过 resident_caregivers 过滤
    args = append(args, userID)
    q += fmt.Sprintf(` AND EXISTS (
        SELECT 1 FROM resident_caregivers rc
        WHERE rc.tenant_id = r.tenant_id
          AND rc.resident_id = r.resident_id
          AND (rc.userList::text LIKE $%d OR rc.userList::text LIKE $%d)
    )`, len(args), len(args)+1)
    args = append(args, "%\""+userID+"\"%")
}

// 应用 branch_only 过滤
if permCheck.BranchOnly {
    // Manager: 使用 ApplyBranchFilter() 实现分支过滤（含空值匹配）
    ApplyBranchFilter(&q, &args, userBranchTag, "u", false)  // false = 不是第一个条件
}
```

---

## 修改后的行为

| 角色 | assigned_only | branch_only | 过滤结果 |
|------|--------------|-------------|---------|
| **Admin** | `FALSE` | `FALSE` | 显示所有住户（`WHERE tenant_id = $1`） |
| **IT** | `FALSE` | `FALSE` | 显示所有住户（`WHERE tenant_id = $1`） |
| **Manager** (有 branch_tag) | `FALSE` | `TRUE` | 显示同 branch 的住户（`WHERE tenant_id = $1 AND u.branch_tag = $2`） |
| **Manager** (branch_tag=NULL) | `FALSE` | `TRUE` | 显示 `branch_tag IS NULL OR '-'` 的住户（空值匹配） |
| **Caregiver** | `TRUE` | `FALSE` | 显示分配的住户（通过 `resident_caregivers` 过滤） |
| **Nurse** | `TRUE` | `FALSE` | 显示分配的住户（通过 `resident_caregivers` 过滤） |
| **Resident** | N/A | N/A | 显示自己（`WHERE tenant_id = $1 AND resident_id = $2`） |
| **Family** | N/A | N/A | 显示关联住户（`WHERE tenant_id = $1 AND resident_id = $2`） |

---

## 修改要点

1. ✅ 使用 `GetResourcePermission()` 统一查询权限配置
2. ✅ 使用 `ApplyBranchFilter()` 实现分支过滤（含空值匹配）
3. ✅ 简化 Caregiver/Nurse 逻辑（移除 `alarm_scope` 依赖）
4. ✅ 移除硬编码 Manager 判断

---

## 请确认

- [ ] 权限矩阵表是否正确？
- [ ] 计划实现逻辑是否符合预期？
- [ ] 是否同意开始修改？

