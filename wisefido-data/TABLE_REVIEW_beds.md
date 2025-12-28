# 表结构强规范检查报告：beds

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE beds (
    bed_id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    room_id           UUID NOT NULL REFERENCES rooms(room_id) ON DELETE CASCADE,
    bed_name          VARCHAR(50) NOT NULL,
    mattress_material VARCHAR(50),
    mattress_thickness VARCHAR(20)
);
```

### 字段列表：
- `bed_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `room_id` (UUID, NOT NULL, FK → rooms.room_id)
- `bed_name` (VARCHAR(50), NOT NULL)
- `mattress_material` (VARCHAR(50), nullable)
- `mattress_thickness` (VARCHAR(20), nullable)

### 约束：
- 主键：`bed_id`
- 外键：`tenant_id` → `tenants.tenant_id`，`room_id` → `rooms.room_id`
- 唯一约束：`(tenant_id, room_id, bed_name)` - 同一租户 + 房间内，床名唯一

### 业务规则：
1. Bed 必须绑定到 Room（room_id NOT NULL）
2. 同一租户 + 房间内，床名唯一
3. `bed_type` 字段已删除，ActiveBed 判断由应用层动态计算
4. `bound_device_count` 字段已删除，可以通过 COUNT(devices) 计算

---

## 2. Domain 模型检查

**状态：✅ 字段匹配正确**

当前文件：`internal/domain/bed.go`

### Domain 模型：
```go
type Bed struct {
    BedID            string         `db:"bed_id"`            // ✅ UUID, PRIMARY KEY
    TenantID         string         `db:"tenant_id"`         // ✅ UUID, NOT NULL
    RoomID           string         `db:"room_id"`           // ✅ UUID, NOT NULL
    BedName          string         `db:"bed_name"`          // ✅ VARCHAR(50), NOT NULL
    MattressMaterial sql.NullString `db:"mattress_material"` // ✅ VARCHAR(50), nullable
    MattressThickness sql.NullString `db:"mattress_thickness"` // ✅ VARCHAR(20), nullable
}
```

**所有字段的 db tag 与数据库表结构一致。**

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/units_repo.go`

接口方法（在 `UnitsRepository` 接口中）：
- `ListBeds` - 查询 beds 列表
- `GetBed` - 获取单个 bed
- `CreateBed` - 创建 bed
- `UpdateBed` - 更新 bed
- `DeleteBed` - 删除 bed

**注意**：beds 表的操作包含在 `UnitsRepository` 接口中，因为 beds 是 rooms 的子实体。

---

## 4. Repository 实现检查

**状态：✅ 字段匹配正确**

文件：`internal/repository/postgres_units.go`

### 检查结果：

**ListBeds** (第 1091-1102 行)：
- ✅ 查询字段：`bed_id`, `tenant_id`, `room_id`, `bed_name`, `mattress_material`, `mattress_thickness`
- ✅ 所有字段与数据库表结构一致
- ✅ 正确处理可空字段（`mattress_material`, `mattress_thickness`）

**GetBed** (第 1137-1147 行)：
- ✅ 查询字段：`bed_id`, `tenant_id`, `room_id`, `bed_name`, `mattress_material`, `mattress_thickness`
- ✅ 所有字段与数据库表结构一致
- ✅ 正确处理可空字段

**CreateBed** (第 1204-1208 行)：
- ✅ INSERT 字段：`tenant_id`, `room_id`, `bed_name`, `mattress_material`, `mattress_thickness`
- ✅ 所有字段与数据库表结构一致
- ✅ 通过 `SELECT tenant_id FROM rooms WHERE room_id = $1` 获取 `tenant_id`，确保数据一致性
- ✅ 正确处理可空字段

**UpdateBed** (第 1259 行)：
- ✅ UPDATE 字段：动态构建，支持 `bed_name`, `mattress_material`, `mattress_thickness`
- ✅ 所有字段与数据库表结构一致
- ✅ 正确处理可空字段（支持设置为 NULL）

**DeleteBed** (第 1270 行)：
- ✅ DELETE 条件：`tenant_id`, `bed_id`
- ✅ 字段匹配正确

---

## 5. 问题总结

### ✅ 所有内容正确：
1. Domain 模型：所有字段的 db tag 与数据库表结构一致
2. Repository 接口：已定义，方法完整（在 `UnitsRepository` 接口中）
3. Repository 实现：所有 SQL 查询字段与数据库表结构一致
4. 业务逻辑：正确处理可空字段，通过 JOIN rooms 表获取 tenant_id 确保数据一致性

---

## 6. 结论

**beds 表的所有代码与数据库表结构完全一致，无需修复。**

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/bed.go`
- Repository 接口：`internal/repository/units_repo.go`（`UnitsRepository` 接口中的 Bed 操作方法）
- Repository 实现：`internal/repository/postgres_units.go`（第 1084-1272 行）

