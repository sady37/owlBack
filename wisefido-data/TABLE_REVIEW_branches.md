# 表结构强规范检查报告：branches

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE branches (
    branch_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    branch_name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(tenant_id, branch_name)
);
```

### 字段列表：
- `branch_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `branch_name` (VARCHAR(255), NOT NULL)
- `description` (TEXT, nullable)
- `created_at` (TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP)
- `updated_at` (TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP)

### 约束：
- 主键：`branch_id`
- 唯一约束：`(tenant_id, branch_name)` - 同一租户内院区名称唯一
- 额外唯一约束：`branch_id` - 支持其他表通过 branch_id 建立外键

### 业务规则：
1. Branch 是独立的实体，可以独立创建、编辑、删除
2. Branch 不依赖 buildings/units/users 的存在，即使没有关联数据也可以创建 branch
3. 用于统一管理院区信息，避免 branch_name 在多个表中重复
4. 支持权限过滤（branch_only）：users.branch_id = units.branch_id
5. `branch_name` 可以为 "-"（表示没有特定院区，使用默认值）
6. 当创建 Branch 时，如果 branch_name 为空，自动设置为 '-'，确保所有记录都通过 branch 关联

---

## 2. Domain 模型检查

**状态：❌ 未找到 domain 模型**

需要创建：`internal/domain/branch.go`

### 建议的 Domain 模型：

```go
package domain

import (
    "database/sql"
)

// Branch 院区领域模型（对应 branches 表）
// 业务规则：
//  1. Branch 是独立的实体，可以独立创建、编辑、删除
//  2. Branch 不依赖 buildings/units/users 的存在
//  3. 用于统一管理院区信息，避免 branch_name 在多个表中重复
//  4. branch_name 可以为 "-"（表示没有特定院区，使用默认值）
type Branch struct {
    BranchID    string         `db:"branch_id"`    // UUID, PRIMARY KEY
    TenantID    string         `db:"tenant_id"`    // UUID, NOT NULL
    BranchName  string         `db:"branch_name"`  // VARCHAR(255), NOT NULL
    Description sql.NullString `db:"description"`  // TEXT, nullable
    CreatedAt   sql.NullTime   `db:"created_at"`  // TIMESTAMP, nullable
    UpdatedAt   sql.NullTime   `db:"updated_at"`  // TIMESTAMP, nullable
}
```

---

## 3. Repository 接口检查

**状态：❌ 未找到 Repository 接口**

需要创建：`internal/repository/branches_repo.go`

### 建议的 Repository 接口：

```go
package repository

import (
    "context"
    "wisefido-data/internal/domain"
)

// BranchesRepository 院区Repository接口
// 使用强类型领域模型，不使用map[string]any
// 设计原则：从底层（数据库）向上设计，Repository层只负责数据访问
type BranchesRepository interface {
    // ========== 查询接口 ==========
    
    // GetBranch 获取院区信息（通过 branch_id）
    GetBranch(ctx context.Context, tenantID, branchID string) (*domain.Branch, error)
    
    // GetBranchByName 获取院区信息（通过 branch_name）
    // 注意：同一租户内 branch_name 唯一
    GetBranchByName(ctx context.Context, tenantID, branchName string) (*domain.Branch, error)
    
    // ListBranches 列出所有院区
    // 支持分页和搜索
    ListBranches(ctx context.Context, tenantID string, page, size int) ([]*domain.Branch, int, error)
    
    // ========== 创建接口 ==========
    
    // CreateBranch 创建院区
    // 注意：
    //   - 如果 branch_name 为空，自动设置为 '-'
    //   - 唯一性约束：同一租户内 branch_name 唯一
    CreateBranch(ctx context.Context, tenantID string, branch *domain.Branch) (string, error)
    
    // ========== 更新接口 ==========
    
    // UpdateBranch 更新院区信息
    // 注意：
    //   - 如果更新 branch_name，需要检查唯一性约束
    //   - 更新时自动更新 updated_at
    UpdateBranch(ctx context.Context, tenantID, branchID string, branch *domain.Branch) error
    
    // ========== 删除接口 ==========
    
    // DeleteBranch 删除院区
    // 注意：
    //   - 删除时，关联的 buildings/units/users 的 branch_id 会被设置为 NULL（ON DELETE SET NULL）
    //   - 关联的 user_branches 记录会被删除（ON DELETE CASCADE）
    DeleteBranch(ctx context.Context, tenantID, branchID string) error
}
```

---

## 4. Repository 实现检查

**状态：❌ 未找到 Repository 实现**

需要创建：`internal/repository/postgres_branches.go`

### 需要实现的方法：
1. `GetBranch` - 通过 branch_id 查询院区
2. `GetBranchByName` - 通过 branch_name 查询院区
3. `ListBranches` - 列出所有院区（支持分页）
4. `CreateBranch` - 创建院区（自动处理 branch_name 为空的情况）
5. `UpdateBranch` - 更新院区信息（自动更新 updated_at）
6. `DeleteBranch` - 删除院区

### 关键业务逻辑：
- **默认值处理**：创建时如果 branch_name 为空，自动设置为 '-'
- **唯一性约束**：创建/更新时检查 `(tenant_id, branch_name)` 是否已存在
- **更新时间**：更新时自动更新 `updated_at` 字段

---

## 5. 问题总结

### ❌ 缺失的内容：
1. Domain 模型：`domain/branch.go`
2. Repository 接口：`repository/branches_repo.go`
3. Repository 实现：`repository/postgres_branches.go`

### ✅ 数据库表结构：
- 表结构完整，字段定义正确
- 约束和索引都已创建

---

## 6. 下一步行动

1. **创建 Domain 模型**：`internal/domain/branch.go`
2. **创建 Repository 接口**：`internal/repository/branches_repo.go`
3. **创建 Repository 实现**：`internal/repository/postgres_branches.go`
4. **实现业务逻辑**：特别是默认值处理和唯一性检查
5. **添加单元测试**：确保约束和业务逻辑正确

---

## 7. 相关代码位置

当前代码中多处使用了 `branch_name` 或 `branch_tag`，但都是通过 JOIN `branches` 表获取的，不是直接查询 `branches` 表。

需要检查以下位置是否需要使用 `branches` 表的 Repository：
- `internal/repository/postgres_users.go` - JOIN branches 获取 branch_name
- `internal/repository/postgres_units.go` - 可能使用 branches
- `internal/repository/postgres_buildings.go` - 可能使用 branches

