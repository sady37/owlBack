# GetUser 方法完整分支逻辑图

## 输入参数

```
GetUserRequest {
    TenantID      string   // 必填
    UserID        string   // 可选：Edit 模式必填，Create 模式为空或 "new"
    CurrentUserID string   // 必填：当前登录用户 ID
    BranchIDs     []string // 可选：Create 模式时，如果已指定 branch，传入 branch_ids
}
```

## 完整分支流程图

```
GetUser(req GetUserRequest)
│
├─ [1] 参数验证
│   ├─ TenantID == ""? 
│   │   └─ YES → ❌ 返回错误: "tenant_id is required"
│   ├─ CurrentUserID == ""?
│   │   └─ YES → ❌ 返回错误: "current_user_id is required"
│   └─ 继续
│
├─ [2] 判断模式
│   ├─ UserID == "" || UserID == "new"?
│   │   ├─ YES → [CREATE 模式]
│   │   └─ NO  → [EDIT 模式]
│   │
│   ├─ [CREATE 模式]
│   │   │
│   │   ├─ [2.1] 获取当前登录用户信息（用于后续权限检查）
│   │   │   └─ GetUser(tenantID, currentUserID)
│   │   │       └─ 失败 → ❌ 返回错误
│   │   │
│   │   ├─ [2.2] 计算 available_tags
│   │   │   │
│   │   │   ├─ len(BranchIDs) > 0? (Branch 已指定)
│   │   │   │   ├─ YES → [分支 2.2.1]
│   │   │   │   │   │   └─ 使用指定的 BranchIDs
│   │   │   │   │   │   └─ availableTags = getAvailableTagsFromBranchIDs(tenantID, BranchIDs)
│   │   │   │   │   │
│   │   │   │   └─ NO  → [分支 2.2.2]
│   │   │   │       │   └─ Branch 未指定，使用当前登录用户的 branch
│   │   │   │       │   └─ availableTags = getAvailableTagsFromBranches(tenantID, CurrentUserID)
│   │   │   │       │
│   │   │   └─ 错误处理
│   │   │       └─ 失败 → 记录警告，availableTags = []
│   │   │
│   │   └─ [2.3] 返回结果
│   │       └─ GetUserResponse {
│   │           User:          nil 或 empty UserDTO,
│   │           AvailableTags: availableTags
│   │       }
│   │
│   └─ [EDIT 模式]
│       │
│       ├─ [3.1] 获取当前登录用户信息（用于权限检查）
│       │   └─ GetUser(tenantID, currentUserID)
│       │       └─ 失败 → ❌ 返回错误
│       │
│       ├─ [3.2] 权限检查
│       │   │
│       │   ├─ CurrentUserID == UserID? (查看自己)
│       │   │   ├─ YES → 跳过权限检查，继续
│       │   │   └─ NO  → [3.2.1]
│       │   │       │   └─ 获取目标用户信息
│       │   │       │   └─ GetUser(tenantID, userID)
│       │   │       │       └─ 失败 → ❌ 返回错误: "user not found"
│       │   │       │
│       │   │       └─ [3.2.2] 角色层级检查
│       │   │           └─ canCreateRole(currentUser.Role, targetUser.Role)?
│       │   │               ├─ NO  → ❌ 返回错误: "not allowed to view ..."
│       │   │               └─ YES → 继续
│       │   │
│       │   └─ 继续
│       │
│       ├─ [3.3] 查询用户详情
│       │   └─ GetUser(tenantID, userID)
│       │       └─ 失败 → ❌ 返回错误: "failed to get user"
│       │
│       ├─ [3.4] 查询用户的所有 branch 关联
│       │   └─ getUserBranchIDs(tenantID, userID)
│       │       └─ 失败 → ❌ 返回错误: "failed to get user branches"
│       │
│       ├─ [3.5] 转换为 DTO
│       │   └─ dto = domainUserToDTO(user)
│       │
│       ├─ [3.6] 填充 branch_ids 和主院区
│       │   └─ if len(userBranches) > 0:
│       │       └─ dto.BranchIDs = extractBranchIDs(userBranches)
│       │       └─ dto.PrimaryBranchID = userBranches[0].BranchID
│       │
│       ├─ [3.7] 计算 available_tags
│       │   │   └─ ⚠️ 修改：使用 req.UserID（被编辑用户）而不是 req.CurrentUserID
│       │   │
│       │   └─ availableTags = getAvailableTagsFromBranches(tenantID, UserID)
│       │       └─ 失败 → 记录警告，availableTags = []
│       │
│       └─ [3.8] 返回结果
│           └─ GetUserResponse {
│               User:          dto,
│               AvailableTags: availableTags
│           }
│
└─ 结束
```

## 关键修改点

### 1. GetUserRequest 结构体修改

```go
type GetUserRequest struct {
    TenantID      string   // 必填
    UserID        string   // 可选：Edit 模式必填，Create 模式为空或 "new"
    CurrentUserID string   // 必填：当前登录用户 ID
    BranchIDs     []string // 可选：Create 模式时，如果已指定 branch，传入 branch_ids
}
```

### 2. 参数验证修改（第 737 行）

```go
// 修改前：
if req.TenantID == "" || req.UserID == "" {
    return nil, fmt.Errorf("tenant_id and user_id are required")
}

// 修改后：
if req.TenantID == "" {
    return nil, fmt.Errorf("tenant_id is required")
}
if req.CurrentUserID == "" {
    return nil, fmt.Errorf("current_user_id is required")
}
// UserID 可以为空（Create 模式）
```

### 3. 模式判断（新增）

```go
// 判断是 Create 模式还是 Edit 模式
isCreateMode := req.UserID == "" || req.UserID == "new"
```

### 4. Create 模式分支（新增）

```go
if isCreateMode {
    // [2.1] 获取当前登录用户信息（用于后续可能需要的权限检查）
    currentUser, err := s.usersRepo.GetUser(ctx, req.TenantID, req.CurrentUserID)
    if err != nil {
        s.logger.Error("Failed to get current user", zap.Error(err))
        return nil, fmt.Errorf("failed to get current user: %w", err)
    }
    
    // [2.2] 计算 available_tags
    var availableTags []string
    if len(req.BranchIDs) > 0 {
        // [分支 2.2.1] Branch 已指定：使用指定的 BranchIDs
        availableTags, err = s.getAvailableTagsFromBranchIDs(ctx, req.TenantID, req.BranchIDs)
    } else {
        // [分支 2.2.2] Branch 未指定：使用当前登录用户的 branch
        availableTags, err = s.getAvailableTagsFromBranches(ctx, req.TenantID, req.CurrentUserID)
    }
    
    if err != nil {
        s.logger.Warn("Failed to get available tags", zap.Error(err))
        availableTags = []string{}
    }
    
    // [2.3] 返回结果
    return &GetUserResponse{
        User:          nil, // Create 模式下没有用户数据
        AvailableTags: availableTags,
    }, nil
}
```

### 5. Edit 模式修改（第 798-804 行）

```go
// 修改前：
availableTags, err := s.getAvailableTagsFromBranches(ctx, req.TenantID, req.CurrentUserID)

// 修改后：
availableTags, err := s.getAvailableTagsFromBranches(ctx, req.TenantID, req.UserID)
// 使用 req.UserID（被编辑用户）而不是 req.CurrentUserID（登录用户）
```

### 6. 新增辅助方法

```go
// getAvailableTagsFromBranchIDs 根据 branch_ids 直接查询 tags（不依赖 user_id）
func (s *userService) getAvailableTagsFromBranchIDs(ctx context.Context, tenantID string, branchIDs []string) ([]string, error) {
    if len(branchIDs) == 0 {
        return []string{}, nil
    }
    
    // 查询这些 branch_ids 中所有用户的 tags
    query := `
        SELECT DISTINCT tag
        FROM users u
        INNER JOIN user_branches ub ON u.user_id = ub.user_id
        CROSS JOIN LATERAL jsonb_array_elements_text(COALESCE(u.user_tags, '[]'::jsonb)) AS tag
        WHERE u.tenant_id = $1
          AND ub.branch_id = ANY($2::uuid[])
          AND tag IS NOT NULL
          AND tag != ''
        ORDER BY tag
    `
    
    rows, err := s.db.QueryContext(ctx, query, tenantID, pq.Array(branchIDs))
    if err != nil {
        return nil, fmt.Errorf("failed to query available tags: %w", err)
    }
    defer rows.Close()
    
    var tags []string
    for rows.Next() {
        var tag string
        if err := rows.Scan(&tag); err != nil {
            return nil, fmt.Errorf("failed to scan tag: %w", err)
        }
        if tag != "" {
            tags = append(tags, tag)
        }
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("failed to iterate tags: %w", err)
    }
    
    return tags, nil
}
```

## 分支总结表

| 模式 | UserID | BranchIDs | available_tags 来源 | 说明 |
|------|--------|-----------|---------------------|------|
| **Create** | 空或 "new" | 有值 | `getAvailableTagsFromBranchIDs(BranchIDs)` | Branch 已指定 |
| **Create** | 空或 "new" | 空 | `getAvailableTagsFromBranches(CurrentUserID)` | Branch 未指定，使用登录用户的 branch |
| **Edit** | 有值 | 忽略 | `getAvailableTagsFromBranches(UserID)` | 使用被编辑用户的 branch |

## 注意事项

1. **Create 模式下不需要查询用户详情**：直接返回 `User: nil` 或空的 `UserDTO`
2. **Edit 模式下必须进行权限检查**：确保当前用户有权限查看目标用户
3. **available_tags 计算逻辑**：
   - Create 模式：根据是否指定 branch 选择不同的计算方式
   - Edit 模式：始终使用被编辑用户的 branch（不是登录用户的 branch）

