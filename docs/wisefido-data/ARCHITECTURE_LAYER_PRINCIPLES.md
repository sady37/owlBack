# Service 层与 Repository 层分层原则

## 一、Repository 层职责

### 1.1 核心职责
- **只负责数据访问**：对数据库表的 CRUD 操作
- **使用强类型领域模型**：所有方法使用 `domain.*` 类型，不使用 `map[string]any`
- **单一表操作**：每个 Repository 方法通常对应一个表的操作
- **数据验证**：只做基础的数据验证（必填字段、格式检查），不做业务逻辑验证

### 1.2 事务管理
- **单个操作内部可以使用事务**：如果单个 Repository 方法需要多个 SQL 操作（如先查询再插入），可以在方法内部使用事务
- **示例**：`CreateUserBranch` 方法内部使用事务来确保"检查是否存在"和"插入新记录"的原子性

```go
// ✅ 正确：Repository 方法内部使用事务
func (r *PostgresUserBranchesRepository) CreateUserBranch(ctx context.Context, tenantID string, userBranch *domain.UserBranch) (string, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    // ... 检查是否存在
    // ... 插入新记录
    return tx.Commit()
}
```

### 1.3 禁止事项
- ❌ **不包含业务逻辑**：不判断角色、权限、业务规则
- ❌ **不跨表操作**：不直接操作多个表（除非是关联表，如 `user_branches`）
- ❌ **不管理跨 Repository 的事务**：不接收外部事务对象

## 二、Service 层职责

### 2.1 核心职责
- **业务逻辑**：参数验证、权限检查、业务规则验证、数据转换
- **调用多个 Repository**：一个 Service 方法可以调用多个 Repository 方法
- **DTO 转换**：将 HTTP 请求的 DTO 转换为 Domain Model，将 Domain Model 转换为响应 DTO

### 2.2 多步数据库操作
- **原则**：Service 层应该调用多个 Repository 方法，而不是直接操作数据库
- **示例**：更新用户时，需要更新 `users` 表和 `user_branches` 表

```go
// ✅ 正确：Service 层调用多个 Repository 方法
func (s *userService) UpdateUser(ctx context.Context, req UpdateUserRequest) (*UpdateUserResponse, error) {
    // 1. 调用 Repository 删除所有 user_branches
    if err := s.userBranchesRepo.DeleteAllUserBranches(ctx, req.TenantID, req.UserID); err != nil {
        return nil, err
    }
    
    // 2. 调用 Repository 创建新的 user_branches
    for _, branchID := range req.BranchIDs {
        userBranch := &domain.UserBranch{...}
        _, err := s.userBranchesRepo.CreateUserBranch(ctx, req.TenantID, userBranch)
        if err != nil {
            return nil, err
        }
    }
    
    // 3. 调用 Repository 更新 users 表
    return s.usersRepo.UpdateUser(ctx, req.TenantID, req.UserID, &updateUser)
}
```

### 2.3 事务管理
- **跨 Repository 的事务**：如果多个 Repository 操作需要原子性，由 Service 层管理事务
- **当前实践**：大多数情况下，Service 层调用多个 Repository 方法，每个方法内部管理自己的事务
- **未来改进**：如果需要跨 Repository 的原子性，Service 层可以使用 `db.BeginTx`，但**仍然调用 Repository 方法**，而不是直接执行 SQL

```go
// ⚠️ 当前实践：Service 层直接操作数据库（不推荐，应改为调用 Repository）
func (s *userService) updateUserBranches(ctx context.Context, tenantID, userID string, branchIDs []string) error {
    // ❌ 错误：直接使用 s.db.BeginTx 和 tx.ExecContext
    tx, err := s.db.BeginTx(ctx, nil)
    _, err = tx.ExecContext(ctx, `DELETE FROM user_branches ...`)
    // ...
}

// ✅ 正确：Service 层调用 Repository 方法
func (s *userService) updateUserBranches(ctx context.Context, tenantID, userID string, branchIDs []string) error {
    // ✅ 调用 Repository 方法
    if err := s.userBranchesRepo.DeleteAllUserBranches(ctx, tenantID, userID); err != nil {
        return err
    }
    for _, branchID := range branchIDs {
        _, err := s.userBranchesRepo.CreateUserBranch(ctx, tenantID, &domain.UserBranch{...})
        if err != nil {
            return err
        }
    }
    return nil
}
```

### 2.4 禁止事项
- ❌ **不直接操作数据库**：不使用 `s.db.BeginTx`、`s.db.ExecContext`、`s.db.QueryContext` 等（除非是复杂查询，如 JOIN、权限过滤）
- ❌ **不包含数据访问逻辑**：不写 SQL 语句（除非是复杂查询，如 JOIN、权限过滤）

## 三、复杂查询的特殊情况

### 3.1 Service 层可以使用 `db` 的场景
- **复杂 JOIN 查询**：需要跨多个表 JOIN，且涉及权限过滤
- **示例**：`ListUsers` 需要 JOIN `users`、`user_branches`、`branches` 表，并根据当前用户的权限过滤

```go
// ✅ 允许：Service 层使用 db 进行复杂查询
func (s *userService) ListUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
    query := `
        SELECT u.*, b.branch_name
        FROM users u
        LEFT JOIN LATERAL (
            SELECT branch_name FROM branches b
            JOIN user_branches ub ON b.branch_id = ub.branch_id
            WHERE ub.user_id = u.user_id
            LIMIT 1
        ) b ON true
        WHERE u.tenant_id = $1
        -- 权限过滤逻辑
    `
    rows, err := s.db.QueryContext(ctx, query, ...)
    // ...
}
```

### 3.2 原则
- **简单查询**：使用 Repository 方法
- **复杂查询**：如果 Repository 接口过于复杂，Service 层可以使用 `db` 直接查询，但应该封装为 Service 层的内部方法

## 四、总结

### 4.1 分层原则总结

| 职责 | Repository 层 | Service 层 |
|------|--------------|-----------|
| **数据访问** | ✅ 负责 | ❌ 不直接操作（复杂查询除外） |
| **业务逻辑** | ❌ 不包含 | ✅ 负责 |
| **事务管理** | ✅ 单个操作内部 | ✅ 跨 Repository 操作 |
| **多步操作** | ❌ 单一表操作 | ✅ 调用多个 Repository |
| **SQL 语句** | ✅ 包含 | ❌ 不包含（复杂查询除外） |
| **DTO 转换** | ❌ 不包含 | ✅ 负责 |

### 4.2 关键原则
1. **Repository 层**：单一职责，只负责数据访问
2. **Service 层**：通过调用多个 Repository 方法实现业务逻辑
3. **事务管理**：单个操作在 Repository 内部，跨操作在 Service 层
4. **复杂查询**：允许 Service 层使用 `db`，但应封装为内部方法

### 4.3 当前代码改进方向
- ✅ 已改进：`updateUserBranches` 和 `createUserBranches` 现在调用 Repository 方法
- ⚠️ 待改进：如果需要在 Service 层管理跨 Repository 的事务，应该使用事务对象传递给 Repository（需要扩展 Repository 接口支持事务参数）
