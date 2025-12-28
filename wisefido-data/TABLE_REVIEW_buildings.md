# 表结构强规范检查报告：buildings

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE buildings (
    building_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    branch_id     UUID REFERENCES branches(branch_id) ON DELETE SET NULL,
    building_name VARCHAR(50) NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(tenant_id, branch_id, building_name) WHERE branch_id IS NOT NULL,
    UNIQUE(tenant_id, building_name) WHERE branch_id IS NULL
);
```

### 字段列表：
- `building_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `branch_id` (UUID, nullable, FK → branches.branch_id)
- `building_name` (VARCHAR(50), NOT NULL)
- `created_at` (TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP)
- `updated_at` (TIMESTAMP, nullable, DEFAULT CURRENT_TIMESTAMP)

### 约束：
- 主键：`building_id`
- 唯一约束：
  - `(tenant_id, branch_id, building_name) WHERE branch_id IS NOT NULL`
  - `(tenant_id, building_name) WHERE branch_id IS NULL`
- 额外唯一约束：`building_id` - 支持其他表通过 building_id 建立外键

### 业务规则：
1. Building 是独立的实体，可以独立创建、编辑、删除
2. Building 不依赖 units 的存在，即使没有 units 也可以创建 building
3. `units.building_name` 字段引用 `buildings.building_name`（通过 building_name + branch_id 唯一标识）
4. `building_name` 可以为 "-"（表示没有特定楼栋，使用默认值）
5. 当创建 Building 时，如果 building_name 为空，自动设置为 '-'

---

## 2. Domain 模型检查

**状态：⚠️ 存在字段不匹配问题**

当前文件：`internal/domain/building.go`

### 问题：
- `BranchTag` 字段的 db tag 是 `branch_name`，但数据库表实际字段是 `branch_id` (UUID)
- 数据库表中没有 `branch_name` 字段，需要通过 JOIN branches 表获取

### 需要修复：
```go
// 当前（错误）：
BranchTag sql.NullString `db:"branch_name"`  // ❌ 数据库没有 branch_name 字段

// 应该改为：
BranchID sql.NullString `db:"branch_id"`    // ✅ 数据库字段是 branch_id (UUID)
```

或者，如果需要显示 branch_name，可以通过 JOIN 获取：
```go
BranchID   sql.NullString `db:"branch_id"`   // 存储 branch_id
BranchName sql.NullString `db:"branch_name"` // 通过 JOIN branches 获取
```

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/units_repo.go`

接口方法：
- `ListBuildings` - 列出所有楼栋
- `GetBuilding` - 获取楼栋信息
- `CreateBuilding` - 创建楼栋
- `UpdateBuilding` - 更新楼栋信息
- `DeleteBuilding` - 删除楼栋

---

## 4. Repository 实现检查

**状态：⚠️ 存在字段不匹配问题**

文件：`internal/repository/postgres_units.go`

### 问题：
1. **GetBuilding** (第 90 行)：
   - 查询使用了 `branch_name`，但数据库表没有此字段
   - 应该查询 `branch_id`，然后 JOIN branches 表获取 `branch_name`

2. **CreateBuilding** (第 149, 161 行)：
   - INSERT 语句使用了 `branch_name`，但数据库表字段是 `branch_id`
   - 应该使用 `branch_id`，并通过 branch_name 查找 branch_id

3. **UpdateBuilding** (第 194, 233, 241 行)：
   - UPDATE 语句使用了 `branch_name`，但数据库表字段是 `branch_id`
   - 应该使用 `branch_id`，并通过 branch_name 查找 branch_id

---

## 5. 问题总结

### ❌ 需要修复的内容：
1. **Domain 模型** (`domain/building.go`)：
   - `BranchTag` 字段的 db tag 错误
   - 应该改为 `BranchID` 存储 branch_id，或添加 `BranchName` 通过 JOIN 获取

2. **Repository 实现** (`postgres_units.go`)：
   - 所有 SQL 查询中的 `branch_name` 应该改为 `branch_id`
   - 需要通过 JOIN branches 表获取 branch_name（如果需要显示）

### ✅ 已存在的内容：
- Repository 接口已定义
- Repository 实现已存在（但需要修复字段问题）

---

## 6. 修复方案

### 方案 1：只存储 branch_id（推荐）
- Domain 模型：`BranchID sql.NullString` 存储 branch_id
- Repository：直接使用 branch_id，不 JOIN branches 表
- 如果需要显示 branch_name，在 Service 层查询

### 方案 2：同时存储 branch_id 和 branch_name
- Domain 模型：`BranchID` 和 `BranchName` 两个字段
- Repository：查询时 JOIN branches 表获取 branch_name
- 更新时通过 branch_name 查找 branch_id

推荐使用方案 1，保持 Domain 模型与数据库表结构一致。

