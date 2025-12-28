# 表结构强规范检查报告：devices

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE devices (
    device_id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id          UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_store_id    UUID,
    device_name        VARCHAR(100) NOT NULL,
    serial_number      VARCHAR(100),
    uid                VARCHAR(50),
    bound_room_id      UUID REFERENCES rooms(room_id) ON DELETE SET NULL,
    bound_bed_id       UUID REFERENCES beds(bed_id) ON DELETE SET NULL,
    status             VARCHAR(20) NOT NULL DEFAULT 'offline',
    business_access    VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (...),
    monitoring_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    metadata           JSONB
);
```

### 字段列表：
- `device_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `device_store_id` (UUID, nullable, FK → device_store.device_store_id)
- `device_name` (VARCHAR(100), NOT NULL)
- `serial_number` (VARCHAR(100), nullable)
- `uid` (VARCHAR(50), nullable)
- `bound_room_id` (UUID, nullable, FK → rooms.room_id)
- `bound_bed_id` (UUID, nullable, FK → beds.bed_id)
- `status` (VARCHAR(20), NOT NULL, DEFAULT 'offline')
- `business_access` (VARCHAR(20), NOT NULL, DEFAULT 'pending', CHECK IN ('pending', 'approved', 'rejected'))
- `monitoring_enabled` (BOOLEAN, NOT NULL, DEFAULT FALSE)
- `metadata` (JSONB, nullable)

### 约束：
- 主键：`device_id`
- 外键：`tenant_id` → `tenants.tenant_id`，`device_store_id` → `device_store.device_store_id`，`bound_room_id` → `rooms.room_id`，`bound_bed_id` → `beds.bed_id`
- 唯一约束：`(tenant_id, serial_number)`，`(tenant_id, uid)`
- CHECK 约束：`bound_room_id` 和 `bound_bed_id` 互斥

---

## 2. Domain 模型检查

**状态：⚠️ 发现一个问题**

当前文件：`internal/domain/device.go`

### Domain 模型：
```go
type Device struct {
    DeviceID          string          `db:"device_id"`          // ✅ UUID, PRIMARY KEY
    TenantID          string          `db:"tenant_id"`          // ✅ UUID, NOT NULL
    DeviceStoreID    sql.NullString  `db:"device_store_id"`    // ✅ UUID, nullable
    DeviceName        string          `db:"device_name"`        // ✅ VARCHAR(100), NOT NULL
    SerialNumber      sql.NullString `db:"serial_number"`      // ✅ VARCHAR(100), nullable
    UID               sql.NullString `db:"uid"`                // ✅ VARCHAR(50), nullable
    BoundRoomID       sql.NullString `db:"bound_room_id"`      // ✅ UUID, nullable
    BoundBedID        sql.NullString `db:"bound_bed_id"`       // ✅ UUID, nullable
    Status            string         `db:"status"`             // ✅ VARCHAR(20), NOT NULL
    BusinessAccess    string         `db:"business_access"`    // ✅ VARCHAR(20), NOT NULL
    MonitoringEnabled bool           `db:"monitoring_enabled"` // ✅ BOOLEAN, NOT NULL
    Metadata          sql.NullString `db:"metadata"`           // ⚠️ JSONB, nullable - 应该是 json.RawMessage
}
```

**问题**：
- `Metadata` 字段类型是 `sql.NullString`，但数据库字段是 JSONB。应该使用 `json.RawMessage` 来正确处理 JSONB 数据。

**建议修复**：
```go
Metadata json.RawMessage `db:"metadata"` // JSONB, nullable
```

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/devices_repo.go`

接口方法：
- `ListDevices` - 查询设备列表
- `GetDevice` - 查询单个设备
- `GetDeviceRelations` - 获取设备关联关系
- `CreateDevice` - 手动创建设备与位置的绑定关系（出库操作）
- `UpdateDevice` - 更新设备信息
- `DeleteDevice` - 删除设备（物理删除，仅当设备未使用时）
- `DisableDevice` - 软删除（禁用设备）
- `GetOrCreateDeviceFromStore` - 自动创建（设备首次连接时自动创建）

---

## 4. Repository 实现检查

**状态：✅ 字段匹配正确（但需要修复 Metadata 类型）**

文件：`internal/repository/postgres_devices.go`

### 检查结果：

**ListDevices** (第 95-113 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段
- ⚠️ `metadata` 字段使用 `sql.NullString` 扫描，应该使用 `json.RawMessage`

**GetDevice** (第 147-163 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段
- ⚠️ `metadata` 字段使用 `sql.NullString` 扫描，应该使用 `json.RawMessage`

**CreateDevice** (第 309-314 行)：
- ✅ INSERT 字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段
- ⚠️ `metadata` 字段未在 INSERT 中，如果需要插入 metadata，应该使用 JSONB 类型

**UpdateDevice** (第 483 行)：
- ✅ UPDATE 字段：动态构建，支持部分更新
- ✅ 所有字段与数据库表结构一致

**DeleteDevice** (第 518 行)：
- ✅ DELETE 条件：`tenant_id`, `device_id`
- ✅ 字段匹配正确

---

## 5. 问题总结

### ⚠️ 需要修复：
1. **Metadata 字段类型**：`domain.Device.Metadata` 应该是 `json.RawMessage`，而不是 `sql.NullString`，以正确处理 JSONB 数据。

### ✅ 其他内容正确：
1. Repository 接口：已定义，方法完整
2. Repository 实现：所有 SQL 查询字段与数据库表结构一致，正确处理可空字段

---

## 6. 修复方案

需要修复 `domain/device.go` 中的 `Metadata` 字段类型：

**当前**：
```go
Metadata sql.NullString `db:"metadata"` // nullable, JSONB
```

**应该修复为**：
```go
Metadata json.RawMessage `db:"metadata"` // JSONB, nullable
```

然后需要更新 `postgres_devices.go` 中的相关代码，将 `sql.NullString` 的处理改为 `json.RawMessage` 的处理。

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/device.go`
- Repository 接口：`internal/repository/devices_repo.go`
- Repository 实现：`internal/repository/postgres_devices.go`

