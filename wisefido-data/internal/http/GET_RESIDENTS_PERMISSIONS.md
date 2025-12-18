# GET /residents 权限检查 - 按角色分析

## 角色权限配置（role_permissions 表）

| 角色 | assigned_only | branch_only | 说明 |
|------|--------------|-------------|------|
| **Admin** | `FALSE` | `FALSE` | 可查看所有住户 |
| **IT** | `FALSE` | `FALSE` | 可查看所有住户 |
| **Manager** | `FALSE` | `TRUE` | 只能查看同 branch 的住户 |
| **Caregiver** | `TRUE` | `FALSE` | 只能查看分配的住户 |
| **Nurse** | `TRUE` | `FALSE` | 只能查看分配的住户 |
| **Resident** | N/A | N/A | 只能查看自己 |
| **Family** | N/A | N/A | 只能查看关联的住户 |

---

## 按角色分析：当前实现 vs 计划实现

### 1. Resident / Family（住户/家属登录）

| 项目 | 当前实现 | 计划实现 | 状态 |
|------|---------|---------|------|
| **过滤逻辑** | 行158-167：仅显示自己（`resident_id = $userID`） | ✅ 保持不变 | ✅ 已正确 |
| **实现方式** | 通过 `isResidentLogin` 标志判断 | ✅ 保持不变 | ✅ 已正确 |
| **问题** | 无 | 无 | ✅ 无需修改 |

**当前代码：**
```go
// 行158-167
if isResidentLogin {
    if residentIDForSelf.Valid {
        args = append(args, residentIDForSelf.String)
        q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND r.resident_id::text = $%d`, len(args))
    } else {
        q += ` WHERE 1=0`
    }
}
```

---

### 2. Admin（管理员）

| 项目 | 当前实现 | 计划实现 | 状态 |
|------|---------|---------|------|
| **过滤逻辑** | 行198-199：显示所有住户（`WHERE tenant_id = $1`） | ✅ 保持不变 | ✅ 已正确 |
| **权限检查** | 行122-138：查询 `assigned_only`（结果为 `FALSE`） | ✅ 使用 `GetResourcePermission()` 统一查询 | ⚠️ 需要统一 |
| **问题** | 1. 未查询 `branch_only`<br>2. 未使用统一函数 | 使用 `GetResourcePermission()` 查询 `assigned_only` 和 `branch_only` | ⚠️ 需要修复 |

**当前代码：**
```go
// 行122-138：仅查询 assigned_only
err := s.DB.QueryRowContext(r.Context(),
    `SELECT assigned_only FROM role_permissions
     WHERE tenant_id = $1 AND role_code = $2 AND resource_type = 'residents' AND permission_type = 'R'
     LIMIT 1`,
    SystemTenantID(), userRole.String,
).Scan(&assignedOnly)

// 行198-199：显示所有
q += ` WHERE r.tenant_id = $1`
```

**计划代码：**
```go
// 使用 GetResourcePermission() 统一查询
permCheck, err := GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "R")
// permCheck.AssignedOnly = FALSE, permCheck.BranchOnly = FALSE
// 结果：显示所有住户（保持不变）
```

---

### 3. IT（IT支持）

| 项目 | 当前实现 | 计划实现 | 状态 |
|------|---------|---------|------|
| **过滤逻辑** | 行198-199：显示所有住户（`WHERE tenant_id = $1`） | ✅ 保持不变 | ✅ 已正确 |
| **权限检查** | 行122-138：查询 `assigned_only`（结果为 `FALSE`） | ✅ 使用 `GetResourcePermission()` 统一查询 | ⚠️ 需要统一 |
| **问题** | 1. 未查询 `branch_only`<br>2. 未使用统一函数 | 使用 `GetResourcePermission()` 查询 `assigned_only` 和 `branch_only` | ⚠️ 需要修复 |

**当前代码：**（与 Admin 相同）

**计划代码：**（与 Admin 相同）

---

### 4. Manager（分支经理）

| 项目 | 当前实现 | 计划实现 | 状态 |
|------|---------|---------|------|
| **过滤逻辑** | 行194-196：硬编码分支过滤（`u.branch_tag = $userBranchTag`） | ✅ 使用 `ApplyBranchFilter()` 实现空值匹配 | ⚠️ 需要修复 |
| **权限检查** | 行122-138：查询 `assigned_only`（结果为 `FALSE`） | ✅ 使用 `GetResourcePermission()` 统一查询 | ⚠️ 需要统一 |
| **空值匹配** | ❌ 未实现（`branch_tag=NULL` 时无法匹配） | ✅ 实现空值匹配（NULL 或 '-'） | ❌ 需要实现 |
| **问题** | 1. 硬编码 Manager 判断<br>2. 未查询 `branch_only`<br>3. 未实现空值匹配 | 1. 使用 `GetResourcePermission()` 查询 `branch_only=TRUE`<br>2. 使用 `ApplyBranchFilter()` 实现空值匹配 | ❌ 需要修复 |

**当前代码：**
```go
// 行194-196：硬编码 Manager 分支过滤
if userRole.Valid && userRole.String == "Manager" && userBranchTag.Valid && userBranchTag.String != "" {
    args = append(args, userBranchTag.String)
    q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND u.branch_tag = $%d`, len(args))
} else {
    q += ` WHERE r.tenant_id = $1`  // Manager 无 branch_tag 时显示所有（错误！）
}
```

**问题：**
- Manager 的 `branch_tag=NULL` 时，应该只能查看 `units.branch_tag IS NULL OR '-'` 的住户
- 当前实现：Manager 无 `branch_tag` 时显示所有住户（错误）

**计划代码：**
```go
// 使用 GetResourcePermission() 查询
permCheck, err := GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "R")
// permCheck.AssignedOnly = FALSE, permCheck.BranchOnly = TRUE

// 使用 ApplyBranchFilter() 实现分支过滤（含空值匹配）
if permCheck.BranchOnly {
    ApplyBranchFilter(&q, &args, userBranchTag, "u", false)  // false = 不是第一个条件（已有 WHERE tenant_id）
} else {
    q += ` WHERE r.tenant_id = $1`
}
```

---

### 5. Caregiver（护理员）

| 项目 | 当前实现 | 计划实现 | 状态 |
|------|---------|---------|------|
| **过滤逻辑** | 行168-189：根据 `alarm_scope` 过滤 | ⚠️ 需要评估：是否保留 `alarm_scope` 逻辑 | ⚠️ 需要评估 |
| **权限检查** | 行122-138：查询 `assigned_only`（结果为 `TRUE`） | ✅ 使用 `GetResourcePermission()` 统一查询 | ⚠️ 需要统一 |
| **分配过滤** | 行174-185：通过 `resident_caregivers.userList` 过滤 | ✅ 保持不变 | ✅ 已正确 |
| **分支过滤** | 行170-173：`alarm_scope='BRANCH'` 时过滤分支 | ⚠️ 需要评估：是否与 `branch_only` 合并 | ⚠️ 需要评估 |
| **问题** | 1. 未查询 `branch_only`<br>2. `alarm_scope` 逻辑复杂 | 1. 使用 `GetResourcePermission()` 查询<br>2. 评估是否保留 `alarm_scope` 逻辑 | ⚠️ 需要评估 |

**当前代码：**
```go
// 行168-189：根据 alarm_scope 过滤
if assignedOnly && userID != "" {
    if alarmScope.Valid && alarmScope.String == "BRANCH" && userBranchTag.Valid {
        // 分支过滤
        args = append(args, userBranchTag.String)
        q += fmt.Sprintf(` WHERE r.tenant_id = $1 AND u.branch_tag = $%d`, len(args))
    } else if alarmScope.Valid && alarmScope.String == "ASSIGNED_ONLY" {
        // 分配过滤
        args = append(args, userID)
        q += fmt.Sprintf(` WHERE r.tenant_id = $1
                          AND EXISTS (
                              SELECT 1 FROM resident_caregivers rc
                              WHERE rc.tenant_id = r.tenant_id
                                AND rc.resident_id = r.resident_id
                                AND (rc.userList::text LIKE $%d OR rc.userList::text LIKE $%d)
                          )`, len(args), len(args)+1)
        args = append(args, "%\""+userID+"\"%")
    } else {
        // alarm_scope='ALL' 或 NULL：显示所有（但 assigned_only=TRUE 时不应该显示所有）
        q += ` WHERE r.tenant_id = $1`
    }
}
```

**问题：**
- `assigned_only=TRUE` 但 `alarm_scope='ALL'` 时，显示所有住户（错误）
- 应该始终通过 `resident_caregivers` 过滤

**计划代码：**
```go
// 使用 GetResourcePermission() 查询
permCheck, err := GetResourcePermission(s.DB, r.Context(), userRole.String, "residents", "R")
// permCheck.AssignedOnly = TRUE, permCheck.BranchOnly = FALSE

if permCheck.AssignedOnly && userID != "" {
    // 始终通过 resident_caregivers 过滤
    args = append(args, userID)
    q += fmt.Sprintf(` WHERE r.tenant_id = $1
                      AND EXISTS (
                          SELECT 1 FROM resident_caregivers rc
                          WHERE rc.tenant_id = r.tenant_id
                            AND rc.resident_id = r.resident_id
                            AND (rc.userList::text LIKE $%d OR rc.userList::text LIKE $%d)
                      )`, len(args), len(args)+1)
    args = append(args, "%\""+userID+"\"%")
    
    // 如果 branch_only=TRUE，额外添加分支过滤
    if permCheck.BranchOnly {
        ApplyBranchFilter(&q, &args, userBranchTag, "u", false)
    }
}
```

---

### 6. Nurse（护士）

| 项目 | 当前实现 | 计划实现 | 状态 |
|------|---------|---------|------|
| **过滤逻辑** | 行168-189：根据 `alarm_scope` 过滤（与 Caregiver 相同） | ⚠️ 需要评估：是否保留 `alarm_scope` 逻辑 | ⚠️ 需要评估 |
| **权限检查** | 行122-138：查询 `assigned_only`（结果为 `TRUE`） | ✅ 使用 `GetResourcePermission()` 统一查询 | ⚠️ 需要统一 |
| **分配过滤** | 行174-185：通过 `resident_caregivers.userList` 过滤 | ✅ 保持不变 | ✅ 已正确 |
| **问题** | 与 Caregiver 相同 | 与 Caregiver 相同 | ⚠️ 需要评估 |

**当前代码：**（与 Caregiver 相同）

**计划代码：**（与 Caregiver 相同）

---

## 总结

### 当前实现问题

1. **未使用统一权限检查函数**
   - 行122-138：直接查询 `assigned_only`，未查询 `branch_only`
   - 应使用 `GetResourcePermission()` 统一查询

2. **Manager 分支过滤硬编码**
   - 行194-196：硬编码 `userRole.String == "Manager"` 判断
   - 未实现空值匹配（`branch_tag=NULL` 时错误地显示所有）

3. **Caregiver/Nurse 逻辑复杂**
   - 行168-189：依赖 `alarm_scope` 字段
   - `assigned_only=TRUE` 但 `alarm_scope='ALL'` 时显示所有（错误）

### 计划修复

1. **统一权限检查**
   - 使用 `GetResourcePermission()` 查询 `assigned_only` 和 `branch_only`
   - 移除硬编码角色判断

2. **统一分支过滤**
   - 使用 `ApplyBranchFilter()` 实现分支过滤（含空值匹配）
   - Manager `branch_tag=NULL` 时，只能查看 `units.branch_tag IS NULL OR '-'` 的住户

3. **简化 Caregiver/Nurse 逻辑**
   - `assigned_only=TRUE` 时，始终通过 `resident_caregivers` 过滤
   - 如果 `branch_only=TRUE`，额外添加分支过滤

### 修复优先级

1. **高优先级：**
   - Manager 分支过滤（空值匹配）
   - 统一权限检查函数

2. **中优先级：**
   - Caregiver/Nurse 逻辑简化

3. **低优先级：**
   - 代码重构（移除硬编码）

