# 表结构强规范检查报告：user_branches

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE user_branches (
    user_branch_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id      UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    user_id        UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    branch_id      UUID NOT NULL REFERENCES branches(branch_id) ON DELETE CASCADE,
    is_primary     BOOLEAN NOT NULL DEFAULT FALSE,
    
    UNIQUE(tenant_id, user_id, branch_id)
);
```

### 字段列表：
- `user_branch_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `user_id` (UUID, NOT NULL, FK → users.user_id)
- `branch_id` (UUID, NOT NULL, FK → branches.branch_id)
- `is_primary` (BOOLEAN, NOT NULL, DEFAULT FALSE)

### 约束：
- 主键：`user_branch_id`
- 唯一约束：`(tenant_id, user_id, branch_id)`
- 部分唯一索引：`(tenant_id, user_id) WHERE is_primary = TRUE`（确保一个用户只有一个主院区）

---

## 2. Domain 模型检查

**状态：❌ 未找到 domain 模型**

需要创建：`internal/domain/user_branch.go`

### 建议的 Domain 模型：

```go
package domain

// UserBranch 用户-院区关联领域模型（对应 user_branches 表）
type UserBranch struct {
    UserBranchID string `db:"user_branch_id"` // UUID, PRIMARY KEY
    TenantID     string `db:"tenant_id"`      // UUID, NOT NULL
    UserID       string `db:"user_id"`       // UUID, NOT NULL
    BranchID     string `db:"branch_id"`     // UUID, NOT NULL
    IsPrimary    bool   `db:"is_primary"`    // BOOLEAN, NOT NULL, DEFAULT FALSE
}
```

---

## 3. Repository 接口检查

**状态：❌ 未找到 Repository 接口**

需要创建：`internal/repository/user_branches_repo.go`

### 建议的 Repository 接口：

```go
package repository

import (
    "context"
    "wisefido-data/internal/domain"
)

// UserBranchesRepository 用户-院区关联Repository接口
type UserBranchesRepository interface {
    // 查询接口
    GetUserBranches(ctx context.Context, tenantID, userID string) ([]*domain.UserBranch, error)
    GetUserPrimaryBranch(ctx context.Context, tenantID, userID string) (*domain.UserBranch, error)
    GetBranchUsers(ctx context.Context, tenantID, branchID string) ([]*domain.UserBranch, error)
    
    // 创建接口
    CreateUserBranch(ctx context.Context, tenantID string, userBranch *domain.UserBranch) (string, error)
    
    // 更新接口
    UpdateUserBranch(ctx context.Context, tenantID, userBranchID string, userBranch *domain.UserBranch) error
    SetPrimaryBranch(ctx context.Context, tenantID, userID, branchID string) error
    
    // 删除接口
    DeleteUserBranch(ctx context.Context, tenantID, userBranchID string) error
    DeleteUserBranchByUserAndBranch(ctx context.Context, tenantID, userID, branchID string) error
}
```

---

## 4. Repository 实现检查

**状态：❌ 未找到 Repository 实现**

需要创建：`internal/repository/postgres_user_branches.go`

### 需要实现的方法：
1. `GetUserBranches` - 查询用户的所有院区关联
2. `GetUserPrimaryBranch` - 查询用户的主院区
3. `GetBranchUsers` - 查询院区的所有用户关联
4. `CreateUserBranch` - 创建用户-院区关联
5. `UpdateUserBranch` - 更新用户-院区关联
6. `SetPrimaryBranch` - 设置主院区（需要确保只有一个主院区）
7. `DeleteUserBranch` - 删除用户-院区关联
8. `DeleteUserBranchByUserAndBranch` - 根据用户和院区删除关联

### 关键业务逻辑：
- **主院区约束**：设置主院区时，需要先将该用户的其他主院区设置为 FALSE
- **唯一性约束**：创建时检查 `(tenant_id, user_id, branch_id)` 是否已存在
- **外键约束**：确保 `user_id` 和 `branch_id` 存在

---

## 5. 问题总结

### ❌ 缺失的内容：
1. Domain 模型：`domain/user_branch.go`
2. Repository 接口：`repository/user_branches_repo.go`
3. Repository 实现：`repository/postgres_user_branches.go`

### ✅ 数据库表结构：
- 表结构完整，字段定义正确
- 约束和索引都已创建

---

## 6. 下一步行动

1. **创建 Domain 模型**：`internal/domain/user_branch.go`
2. **创建 Repository 接口**：`internal/repository/user_branches_repo.go`
3. **创建 Repository 实现**：`internal/repository/postgres_user_branches.go`
4. **实现业务逻辑**：特别是主院区设置的逻辑
5. **添加单元测试**：确保约束和业务逻辑正确

---

## 7. 相关代码位置

当前代码中使用了 `user_branch_tag` 的概念，但这是通过 JOIN `branches` 表获取的，不是直接查询 `user_branches` 表。

需要检查以下位置是否需要使用 `user_branches` 表：
- `internal/service/resident_service.go` - `getUserBranchTag` 方法
- `internal/http/resident_handler.go` - 多处使用 `userBranchTag`
- `internal/service/alarm_event_service.go` - 使用 `userBranchTag` 进行过滤

