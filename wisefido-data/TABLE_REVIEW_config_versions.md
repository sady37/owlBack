# 表结构强规范检查报告：config_versions

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE config_versions (
    version_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    config_type       VARCHAR(50) NOT NULL,
    entity_id         UUID NOT NULL,
    current_entity_id UUID,
    config_data       JSONB NOT NULL,
    valid_from        TIMESTAMPTZ NOT NULL,
    valid_to          TIMESTAMPTZ
);
```

### 字段列表：
- `version_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `config_type` (VARCHAR(50), NOT NULL)
- `entity_id` (UUID, NOT NULL)
- `current_entity_id` (UUID, nullable)
- `config_data` (JSONB, NOT NULL)
- `valid_from` (TIMESTAMPTZ, NOT NULL)
- `valid_to` (TIMESTAMPTZ, nullable)

### 约束：
- 主键：`version_id`
- 外键：`tenant_id` → `tenants.tenant_id`
- 唯一约束：`(tenant_id, config_type, entity_id, valid_from)`

---

## 2. Domain 模型检查

**状态：⚠️ 发现一个问题**

当前文件：`internal/domain/config_version.go`

### Domain 模型：
```go
type ConfigVersion struct {
    VersionID        string          `db:"version_id"`        // ✅ UUID, PRIMARY KEY
    TenantID         string          `db:"tenant_id"`         // ✅ UUID, NOT NULL
    ConfigType       string          `db:"config_type"`       // ✅ VARCHAR(50), NOT NULL
    EntityID         string          `db:"entity_id"`         // ✅ UUID, NOT NULL
    CurrentEntityID  string          `db:"current_entity_id"` // ⚠️ UUID, nullable - 应该是 *string 或 sql.NullString
    ConfigData       json.RawMessage `db:"config_data"`       // ✅ JSONB, NOT NULL
    ValidFrom        time.Time       `db:"valid_from"`        // ✅ TIMESTAMPTZ, NOT NULL
    ValidTo          *time.Time      `db:"valid_to"`          // ✅ TIMESTAMPTZ, nullable
}
```

**问题**：
- `CurrentEntityID` 字段类型是 `string`，但数据库字段是 nullable UUID。应该使用 `*string` 或 `sql.NullString` 来表示可空字段。

**建议修复**：
```go
CurrentEntityID *string `db:"current_entity_id"` // UUID, nullable
```
或者
```go
CurrentEntityID sql.NullString `db:"current_entity_id"` // UUID, nullable
```

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/config_versions_repo.go`

接口方法：
- `GetConfigVersion` - 获取配置版本
- `GetConfigVersionAtTime` - 查询某个时间点的配置（用于回放）
- `ListConfigVersions` - 查询某个实体的所有配置历史（支持分页、时间范围过滤）
- `CreateConfigVersion` - 创建新版本
- `UpdateConfigVersion` - 更新配置版本
- `DeleteConfigVersion` - 删除配置版本

---

## 4. Repository 实现检查

**状态：✅ 字段匹配正确（但需要修复 CurrentEntityID 类型）**

文件：`internal/repository/postgres_config_versions.go`

### 检查结果：

**GetConfigVersion** (第 32-44 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段（`current_entity_id`, `config_data`, `valid_to`）

**GetConfigVersionAtTime** (第 87-105 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段

**ListConfigVersions** (第 186-198 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段

**CreateConfigVersion** (第 288-296 行)：
- ✅ INSERT 字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段（`current_entity_id`, `valid_to`）
- ✅ 正确处理 JSONB 字段（`config_data`）

**UpdateConfigVersion** (第 332-339 行)：
- ✅ UPDATE 字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段

**DeleteConfigVersion** (第 382-384 行)：
- ✅ DELETE 条件：`tenant_id`, `version_id`
- ✅ 字段匹配正确

---

## 5. 问题总结

### ⚠️ 需要修复：
1. **CurrentEntityID 字段类型**：`domain.ConfigVersion.CurrentEntityID` 应该是 `*string` 或 `sql.NullString`，而不是 `string`，以正确表示可空字段。

### ✅ 其他内容正确：
1. Repository 接口：已定义，方法完整
2. Repository 实现：所有 SQL 查询字段与数据库表结构一致，正确处理可空字段

---

## 6. 修复方案

需要修复 `domain/config_version.go` 中的 `CurrentEntityID` 字段类型：

**当前**：
```go
CurrentEntityID string `db:"current_entity_id"` // UUID, nullable
```

**应该修复为**：
```go
CurrentEntityID *string `db:"current_entity_id"` // UUID, nullable
```

然后需要更新 `postgres_config_versions.go` 中的相关代码，将 `sql.NullString` 的处理改为 `*string` 的处理。

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/config_version.go`
- Repository 接口：`internal/repository/config_versions_repo.go`
- Repository 实现：`internal/repository/postgres_config_versions.go`

