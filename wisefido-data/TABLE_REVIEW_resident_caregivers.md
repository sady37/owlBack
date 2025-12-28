# 表结构强规范检查报告：resident_caregivers

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE resident_caregivers (
    caregiver_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    resident_id  UUID NOT NULL REFERENCES residents(resident_id) ON DELETE CASCADE,
    group_list   JSONB,  -- nullable
    user_list    JSONB,  -- nullable
    UNIQUE (tenant_id, resident_id)
);
```

### 字段列表（共 5 个字段）：
- `caregiver_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `resident_id` (UUID, NOT NULL, FK → residents.resident_id)
- `group_list` (JSONB, nullable)
- `user_list` (JSONB, nullable)

---

## 2. Domain 模型检查

**状态：✅ 正确**

当前文件：`internal/domain/resident_caregiver.go`

### Domain 模型：
```go
type ResidentCaregiver struct {
    CaregiverID string          `db:"caregiver_id"` // ✅ UUID, PRIMARY KEY
    TenantID    string          `db:"tenant_id"`    // ✅ UUID, NOT NULL
    ResidentID  string          `db:"resident_id"` // ✅ UUID, NOT NULL
    GroupList   json.RawMessage `db:"group_list"`  // ✅ JSONB, nullable
    UserList    json.RawMessage `db:"user_list"`   // ✅ JSONB, nullable
    Source      string          `db:"-"`            // ✅ 不在数据库表中，仅用于返回结果
}
```

**验证结果**：
- ✅ 所有字段的 db tag 与数据库表结构一致
- ✅ 可空字段使用 `json.RawMessage` 类型，正确处理 JSONB 数据
- ✅ `Source` 字段使用 `db:"-"` 标记，不在数据库表中

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/residents_repo.go`

接口方法：
- `GetResidentCaregivers` - 获取住户的护理人员关联
- `UpsertResidentCaregiver` - 创建或更新护理人员关联

**注意**：
- `GetResidentCaregivers` 返回数组，包含两类配置：
  1. 首先：通过所绑定的unit，unit指定的caregiver/caregiver_group（从units表获取）
  2. 其次：通过直接绑定的caregiver/caregiver_group（从resident_caregivers表获取）
- `UpsertResidentCaregiver` 使用 UPSERT 语义（ON CONFLICT DO UPDATE）

---

## 4. Repository 实现检查

**状态：❌ 发现字段名不匹配问题**

文件：`internal/repository/postgres_residents.go`

### 问题 1：字段名不匹配

**GetResidentCaregivers** (第 1765-1773 行)：
```go
query := `
    SELECT 
        caregiver_id::text,
        CASE WHEN groupList IS NULL THEN NULL ELSE groupList::text END as groupList,
        CASE WHEN userList IS NULL THEN NULL ELSE userList::text END as userList
    FROM resident_caregivers
    WHERE tenant_id = $1 AND resident_id = $2
`
```

**问题**：SQL 查询使用了 `groupList` 和 `userList`（驼峰命名），但数据库表结构使用的是 `group_list` 和 `user_list`（下划线命名）。

**UpsertResidentCaregiver** (第 1818-1826 行)：
```go
query := `
    INSERT INTO resident_caregivers (
        tenant_id, resident_id, groupList, userList
    ) VALUES ($1, $2, $3::jsonb, $4::jsonb)
    ON CONFLICT (tenant_id, resident_id)
    DO UPDATE SET
        groupList = EXCLUDED.groupList,
        userList = EXCLUDED.userList
`
```

**问题**：INSERT 和 UPDATE 语句使用了 `groupList` 和 `userList`（驼峰命名），但数据库表结构使用的是 `group_list` 和 `user_list`（下划线命名）。

---

## 5. 问题总结

### ❌ 需要修复：
1. **字段名不匹配**：`postgres_residents.go` 中的 SQL 查询使用了 `groupList` 和 `userList`（驼峰命名），但数据库表结构使用的是 `group_list` 和 `user_list`（下划线命名）。
   - 需要修复的位置：
     - `GetResidentCaregivers` 方法中的 SELECT 语句（第 1768-1769 行）
     - `UpsertResidentCaregiver` 方法中的 INSERT/UPDATE 语句（第 1820, 1824-1825 行）

### ✅ 其他内容正确：
1. Domain 模型：所有字段的 db tag 与数据库表结构一致
2. Repository 接口：已定义，方法完整
3. Repository 实现：逻辑正确，但字段名需要修复

---

## 6. 修复方案

需要修复 `postgres_residents.go` 中的 SQL 查询，将 `groupList` 和 `userList` 改为 `group_list` 和 `user_list`：

1. **GetResidentCaregivers** 方法：
   - 将 `groupList` 改为 `group_list`
   - 将 `userList` 改为 `user_list`

2. **UpsertResidentCaregiver** 方法：
   - 将 `groupList` 改为 `group_list`
   - 将 `userList` 改为 `user_list`

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/resident_caregiver.go`
- Repository 接口：`internal/repository/residents_repo.go`
- Repository 实现：`internal/repository/postgres_residents.go` (第 1733-1834 行)

