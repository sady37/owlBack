# Sleepace Report device_code 字段说明

## 📋 关键概念

### device_code 与 serial_number/uid 的等价关系

在数据库层，`device_code`、`serial_number` 和 `uid` 是等价的：

- **Sleepace 厂家**：使用 `device_code` 作为设备标识符
- **其他厂家**：可能使用 `serial_number`（序列号）或 `uid`（唯一标识符）
- **数据库层等价性**：`sleepace_report.device_code` 可以通过 `devices.serial_number` 或 `devices.uid` 来匹配

### 数据库表结构

**`devices` 表**：
```sql
serial_number  VARCHAR(100),  -- 厂家出厂序列号（可空）
uid            VARCHAR(50),   -- 厂家或平台提供的唯一 UID（可空）
```

**`sleepace_report` 表**：
```sql
device_id       UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
device_code     VARCHAR(100) NOT NULL,  -- 设备编码（来自厂家，等价于 devices.serial_number 或 devices.uid）
```

### 索引支持

为了支持通过 `device_code` 查询，添加了索引：
```sql
CREATE INDEX IF NOT EXISTS idx_sleepace_report_device_code ON sleepace_report(tenant_id, device_code);
```

---

## 🔄 数据匹配逻辑

### 保存报告时的设备匹配

当保存报告时，如果只有 `device_code` 而没有 `device_id`，系统会通过以下逻辑匹配设备：

1. **优先使用 device_id**：如果提供了 `device_id`，直接使用
2. **通过 device_code 匹配**：如果 `device_id` 为空，通过 `device_code` 匹配 `devices` 表：
   ```sql
   SELECT device_id::text
   FROM devices
   WHERE tenant_id = $1::uuid
     AND (serial_number = $2 OR uid = $2)
     AND status <> 'disabled'
   LIMIT 1
   ```

### Repository 方法

**`GetDeviceIDByDeviceCode`**：
```go
// GetDeviceIDByDeviceCode 根据 device_code 获取 device_id
// device_code 等价于 devices.serial_number 或 devices.uid
func (r *PostgresSleepaceReportsRepository) GetDeviceIDByDeviceCode(
    ctx context.Context, 
    tenantID, deviceCode string,
) (string, error)
```

**`SaveReport`**：
- 如果 `report.DeviceID` 为空，会自动调用 `GetDeviceIDByDeviceCode` 来获取 `device_id`
- 如果 `device_code` 也无法匹配到设备，返回错误

---

## 📝 使用示例

### 场景 1：保存报告时只有 device_code

```go
report := &domain.SleepaceReport{
    TenantID:   "tenant-uuid",
    DeviceID:   "",  // 为空
    DeviceCode: "SP001",  // Sleepace 厂家的 device_code
    Date:       20240820,
    // ... 其他字段
}

// SaveReport 会自动通过 device_code 匹配 devices 表
err := repo.SaveReport(ctx, tenantID, report)
// 如果 devices 表中有 serial_number='SP001' 或 uid='SP001' 的设备，会自动获取 device_id
```

### 场景 2：保存报告时已有 device_id

```go
report := &domain.SleepaceReport{
    TenantID:   "tenant-uuid",
    DeviceID:   "device-uuid",  // 已提供
    DeviceCode: "SP001",  // 仍然保存 device_code 用于追溯
    Date:       20240820,
    // ... 其他字段
}

// SaveReport 直接使用 device_id，不会查询 devices 表
err := repo.SaveReport(ctx, tenantID, report)
```

### 场景 3：查询报告

```go
// 通过 device_id 查询（标准方式）
report, err := repo.GetReport(ctx, tenantID, deviceID, date)

// 如果需要通过 device_code 查询，需要先获取 device_id
deviceID, err := repo.GetDeviceIDByDeviceCode(ctx, tenantID, deviceCode)
if err != nil {
    return err
}
report, err := repo.GetReport(ctx, tenantID, deviceID, date)
```

---

## ⚠️ 注意事项

1. **唯一性约束**：`sleepace_report` 表的唯一性约束是 `(tenant_id, device_id, date)`，不是 `(tenant_id, device_code, date)`
   - 这意味着同一个设备（`device_id`）在同一天只能有一条报告
   - `device_code` 仅用于匹配和追溯，不参与唯一性约束

2. **设备匹配规则**：
   - 优先匹配 `devices.serial_number`
   - 如果 `serial_number` 不匹配，再匹配 `devices.uid`
   - 如果都不匹配，返回错误

3. **租户隔离**：所有查询都必须在 `tenant_id` 范围内进行，确保数据隔离

4. **设备状态**：只匹配 `status <> 'disabled'` 的设备，已禁用的设备不会匹配

---

## 🔍 相关代码位置

- **表结构**：`owlRD/db/26_sleepace_report.sql`
- **Repository 接口**：`owlBack/wisefido-data/internal/repository/sleepace_reports_repo.go`
- **Repository 实现**：`owlBack/wisefido-data/internal/repository/postgres_sleepace_reports.go`
- **Domain 模型**：`owlBack/wisefido-data/internal/domain/sleepace_report.go`

