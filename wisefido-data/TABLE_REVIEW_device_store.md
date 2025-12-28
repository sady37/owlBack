# 表结构强规范检查报告：device_store

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE device_store (
    device_store_id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_type                 VARCHAR(50) NOT NULL,
    device_model                VARCHAR(50),
    serial_number               VARCHAR(100),
    uid                         VARCHAR(50),
    imei                        VARCHAR(50),
    comm_mode                   VARCHAR(20),
    mcu_model                   VARCHAR(50),
    firmware_version            VARCHAR(50),
    tenant_id                   UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    import_date                 TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    allocate_time               TIMESTAMPTZ,
    ota_target_firmware_version VARCHAR(50),
    ota_target_mcu_model        VARCHAR(50),
    allow_access                BOOLEAN NOT NULL DEFAULT FALSE
);
```

### 字段列表：
- `device_store_id` (UUID, PK, NOT NULL, DEFAULT gen_random_uuid())
- `device_type` (VARCHAR(50), NOT NULL)
- `device_model` (VARCHAR(50), nullable)
- `serial_number` (VARCHAR(100), nullable)
- `uid` (VARCHAR(50), nullable)
- `imei` (VARCHAR(50), nullable)
- `comm_mode` (VARCHAR(20), nullable)
- `mcu_model` (VARCHAR(50), nullable)
- `firmware_version` (VARCHAR(50), nullable)
- `tenant_id` (UUID, NOT NULL, DEFAULT '00000000-0000-0000-0000-000000000000')
- `import_date` (TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP)
- `allocate_time` (TIMESTAMPTZ, nullable)
- `ota_target_firmware_version` (VARCHAR(50), nullable)
- `ota_target_mcu_model` (VARCHAR(50), nullable)
- `allow_access` (BOOLEAN, NOT NULL, DEFAULT FALSE)

---

## 2. Domain 模型检查

**状态：⚠️ 发现一个问题**

当前文件：`internal/domain/device_store.go`

### Domain 模型：
```go
type DeviceStore struct {
    DeviceStoreID             string         `db:"device_store_id"`             // ✅ UUID, PRIMARY KEY
    DeviceType                string         `db:"device_type"`                 // ✅ VARCHAR(50), NOT NULL
    DeviceModel               sql.NullString `db:"device_model"`                // ✅ VARCHAR(50), nullable
    SerialNumber              sql.NullString `db:"serial_number"`               // ✅ VARCHAR(100), nullable
    UID                       sql.NullString `db:"uid"`                         // ✅ VARCHAR(50), nullable
    IMEI                      sql.NullString `db:"imei"`                        // ✅ VARCHAR(50), nullable
    CommMode                  sql.NullString `db:"comm_mode"`                   // ✅ VARCHAR(20), nullable
    MCUModel                  sql.NullString `db:"mcu_model"`                  // ✅ VARCHAR(50), nullable
    FirmwareVersion           sql.NullString `db:"firmware_version"`            // ✅ VARCHAR(50), nullable
    OTATargetFirmwareVersion  sql.NullString `db:"ota_target_firmware_version"` // ✅ VARCHAR(50), nullable
    OTATargetMCUModel         sql.NullString `db:"ota_target_mcu_model"`        // ✅ VARCHAR(50), nullable
    TenantID                  string         `db:"tenant_id"`                   // ✅ UUID, NOT NULL
    ImportDate                sql.NullTime   `db:"import_date"`                 // ⚠️ TIMESTAMPTZ, NOT NULL - 应该是 time.Time
    AllocateTime              sql.NullTime   `db:"allocate_time"`               // ✅ TIMESTAMPTZ, nullable
    AllowAccess               bool           `db:"allow_access"`                 // ✅ BOOLEAN, NOT NULL
    TenantName                sql.NullString `db:"tenant_name"`                 // ✅ 仅用于查询结果（JOIN获取）
}
```

**问题**：
- `ImportDate` 字段类型是 `sql.NullTime`，但数据库字段是 `NOT NULL`，应该使用 `time.Time`。

**建议修复**：
```go
ImportDate time.Time `db:"import_date"` // TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP
```

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/device_store_repo.go`

接口方法：
- `ListDeviceStores` - 查询设备库存列表
- `GetDeviceStore` - 查询单个设备库存
- `CreateDeviceStore` - 单个创建设备库存（入库操作）
- `BatchUpdateDeviceStores` - 批量更新设备库存
- `DeleteDeviceStore` - 删除设备库存
- `ImportDeviceStores` - 批量导入设备库存

---

## 4. Repository 实现检查

**状态：✅ 字段匹配正确（但需要修复 ImportDate 类型）**

文件：`internal/repository/postgres_device_store.go`

### 检查结果：

**ListDeviceStores** (第 81-103 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段和 JOIN 查询（`tenant_name`）

**GetDeviceStore** (第 141-162 行)：
- ✅ 查询字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段和 JOIN 查询

**CreateDeviceStore** (第 238-245 行)：
- ✅ INSERT 字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段
- ✅ 正确处理默认值（`tenant_id`, `allow_access`）

**BatchUpdateDeviceStores** (第 329-333 行)：
- ✅ UPDATE 字段：动态构建，支持部分更新
- ✅ 所有字段与数据库表结构一致

**DeleteDeviceStore** (第 380-383 行)：
- ✅ DELETE 条件：`device_store_id`
- ✅ 字段匹配正确

**ImportDeviceStores** (第 456-461 行)：
- ✅ INSERT 字段：所有字段与数据库表结构一致
- ✅ 正确处理可空字段

---

## 5. 问题总结

### ⚠️ 需要修复：
1. **ImportDate 字段类型**：`domain.DeviceStore.ImportDate` 应该是 `time.Time`，而不是 `sql.NullTime`，因为数据库字段是 `NOT NULL`。

### ✅ 其他内容正确：
1. Repository 接口：已定义，方法完整
2. Repository 实现：所有 SQL 查询字段与数据库表结构一致，正确处理可空字段

---

## 6. 修复方案

需要修复 `domain/device_store.go` 中的 `ImportDate` 字段类型：

**当前**：
```go
ImportDate sql.NullTime `db:"import_date"` // NOT NULL, default CURRENT_TIMESTAMP
```

**应该修复为**：
```go
ImportDate time.Time `db:"import_date"` // TIMESTAMPTZ, NOT NULL, DEFAULT CURRENT_TIMESTAMP
```

然后需要更新 `postgres_device_store.go` 中的相关代码，将 `sql.NullTime` 的处理改为 `time.Time` 的处理。

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/device_store.go`
- Repository 接口：`internal/repository/device_store_repo.go`
- Repository 实现：`internal/repository/postgres_device_store.go`

