# 表结构强规范检查报告：cards

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE cards (
    card_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    card_type         VARCHAR(20) NOT NULL,
    bed_id            UUID REFERENCES beds(bed_id) ON DELETE CASCADE,
    unit_id           UUID REFERENCES units(unit_id) ON DELETE CASCADE,
    card_name         VARCHAR(255) NOT NULL,
    card_address      VARCHAR(255) NOT NULL,
    resident_id       UUID REFERENCES residents(resident_id) ON DELETE SET NULL,
    devices           JSONB NOT NULL DEFAULT '[]'::jsonb,
    residents         JSONB NOT NULL DEFAULT '[]'::jsonb,
    unhandled_alarm_0 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_1 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_2 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_3 INTEGER NOT NULL DEFAULT 0,
    unhandled_alarm_4 INTEGER NOT NULL DEFAULT 0,
    icon_alarm_level  INTEGER NOT NULL DEFAULT 3 CHECK (icon_alarm_level >= 0 AND icon_alarm_level <= 4),
    pop_alarm_emerge  INTEGER NOT NULL DEFAULT 0 CHECK (pop_alarm_emerge >= 0 AND pop_alarm_emerge <= 4)
);
```

### 字段列表：
- `card_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `card_type` (VARCHAR(20), NOT NULL, 'ActiveBed' | 'Location')
- `bed_id` (UUID, nullable, FK → beds.bed_id)
- `unit_id` (UUID, nullable, FK → units.unit_id)
- `card_name` (VARCHAR(255), NOT NULL)
- `card_address` (VARCHAR(255), NOT NULL)
- `resident_id` (UUID, nullable, FK → residents.resident_id)
- `devices` (JSONB, NOT NULL, DEFAULT '[]'::jsonb)
- `residents` (JSONB, NOT NULL, DEFAULT '[]'::jsonb)
- `unhandled_alarm_0` (INTEGER, NOT NULL, DEFAULT 0)
- `unhandled_alarm_1` (INTEGER, NOT NULL, DEFAULT 0)
- `unhandled_alarm_2` (INTEGER, NOT NULL, DEFAULT 0)
- `unhandled_alarm_3` (INTEGER, NOT NULL, DEFAULT 0)
- `unhandled_alarm_4` (INTEGER, NOT NULL, DEFAULT 0)
- `icon_alarm_level` (INTEGER, NOT NULL, DEFAULT 3, CHECK >= 0 AND <= 4)
- `pop_alarm_emerge` (INTEGER, NOT NULL, DEFAULT 0, CHECK >= 0 AND <= 4)

### 约束：
- 主键：`card_id`
- 外键：`tenant_id` → `tenants.tenant_id`，`bed_id` → `beds.bed_id`，`unit_id` → `units.unit_id`，`resident_id` → `residents.resident_id`
- 唯一约束：
  - `(tenant_id, bed_id)` WHERE `card_type = 'ActiveBed' AND bed_id IS NOT NULL`
  - `(tenant_id, unit_id)` WHERE `card_type = 'Location' AND unit_id IS NOT NULL AND bed_id IS NULL`
- CHECK 约束：
  - `card_type = 'ActiveBed' AND bed_id IS NOT NULL` OR `card_type = 'Location' AND unit_id IS NOT NULL AND bed_id IS NULL`

---

## 2. Domain 模型检查

**状态：✅ 字段匹配正确**

当前文件：`internal/domain/card.go`

### Domain 模型：
```go
type Card struct {
    CardID            string          `db:"card_id"`            // ✅ UUID, PRIMARY KEY
    TenantID          string          `db:"tenant_id"`         // ✅ UUID, NOT NULL
    CardType          string          `db:"card_type"`         // ✅ VARCHAR(20), NOT NULL
    BedID             sql.NullString `db:"bed_id"`            // ✅ UUID, nullable
    UnitID             sql.NullString `db:"unit_id"`           // ✅ UUID, nullable
    CardName           string          `db:"card_name"`         // ✅ VARCHAR(255), NOT NULL
    CardAddress        string          `db:"card_address"`      // ✅ VARCHAR(255), NOT NULL
    ResidentID         sql.NullString `db:"resident_id"`       // ✅ UUID, nullable
    Devices            json.RawMessage `db:"devices"`          // ✅ JSONB, NOT NULL
    Residents          json.RawMessage `db:"residents"`        // ✅ JSONB, NOT NULL
    UnhandledAlarm0   int             `db:"unhandled_alarm_0"` // ✅ INTEGER, NOT NULL
    UnhandledAlarm1   int             `db:"unhandled_alarm_1"` // ✅ INTEGER, NOT NULL
    UnhandledAlarm2   int             `db:"unhandled_alarm_2"` // ✅ INTEGER, NOT NULL
    UnhandledAlarm3   int             `db:"unhandled_alarm_3"` // ✅ INTEGER, NOT NULL
    UnhandledAlarm4   int             `db:"unhandled_alarm_4"` // ✅ INTEGER, NOT NULL
    IconAlarmLevel    int             `db:"icon_alarm_level"`  // ✅ INTEGER, NOT NULL
    PopAlarmEmerge    int             `db:"pop_alarm_emerge"`  // ✅ INTEGER, NOT NULL
}
```

**所有字段的 db tag 与数据库表结构一致。**

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/cards_repository.go`

接口方法：
- `ListCards` - 查询卡片列表（返回所有可见的卡片，不分页）

---

## 4. Repository 实现检查

**状态：⚠️ 发现一个问题需要修复**

文件：`internal/repository/cards_repository.go`

### 问题 1：BranchTag 过滤使用了错误的字段

**位置**：第 119 行、第 214 行、第 216 行

**问题**：
- 代码使用了 `u.branch_name`，但 `units` 表只有 `branch_id` 字段
- 需要通过 JOIN `branches` 表来获取 `branch_name`

**当前代码**（错误）：
```go
// SELECT 子句（第 119 行）
u.branch_name,

// BranchOnly 权限过滤（第 214、216 行）
query.WriteString(` AND (u.branch_name IS NULL OR u.branch_name = '-') `)
query.WriteString(` AND u.branch_name = $` + fmt.Sprintf("%d", argIdx) + ` `)
```

**应该修复为**：
```go
// SELECT 子句：需要 JOIN branches 表
LEFT JOIN branches br ON u.branch_id = br.branch_id
...
COALESCE(br.branch_name, NULL) as branch_name,

// BranchOnly 权限过滤
query.WriteString(` AND (br.branch_name IS NULL OR br.branch_name = '-') `)
query.WriteString(` AND br.branch_name = $` + fmt.Sprintf("%d", argIdx) + ` `)
```

### 其他检查结果：

**ListCards** (第 98-133 行)：
- ✅ 查询字段：所有 `cards` 表的字段与数据库表结构一致
- ⚠️ 查询字段：`u.branch_name` 需要修复（应通过 JOIN branches 表获取）
- ✅ 正确处理可空字段和 JSONB 字段

---

## 5. 问题总结

### ⚠️ 需要修复：
1. **BranchTag 过滤**：`cards_repository.go` 第 119、214、216 行使用了 `u.branch_name`，应该通过 JOIN `branches` 表获取 `branch_name`。

### ✅ 其他内容正确：
1. Domain 模型：所有字段的 db tag 与数据库表结构一致
2. Repository 接口：已定义，方法完整
3. Repository 实现：除 BranchTag 过滤外，所有 SQL 查询字段与数据库表结构一致

---

## 6. 修复方案

需要修复 `cards_repository.go` 中的三处 `branch_name` 使用：
1. 第 119 行：SELECT 子句中的 `u.branch_name`
2. 第 131 行：需要添加 `LEFT JOIN branches br ON u.branch_id = br.branch_id`
3. 第 214、216 行：BranchOnly 权限过滤中的 `u.branch_name`

修复方法：在 JOIN `units` 表后，再 JOIN `branches` 表，然后使用 `br.branch_name` 进行查询和过滤。

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/card.go`
- Repository 接口：`internal/repository/cards_repository.go`
- Repository 实现：`internal/repository/cards_repository.go`

