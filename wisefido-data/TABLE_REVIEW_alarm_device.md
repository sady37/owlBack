# 表结构强规范检查报告：alarm_device

## 1. 数据库表结构（PostgreSQL 实际）

```sql
CREATE TABLE alarm_device (
    device_id      UUID PRIMARY KEY REFERENCES devices(device_id) ON DELETE CASCADE,
    tenant_id      UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    monitor_config JSONB NOT NULL DEFAULT '{"alarms": {}}'::jsonb,
    vendor_config  JSONB,
    metadata       JSONB
);
```

### 字段列表：
- `device_id` (UUID, PK, NOT NULL, FK → devices.device_id)
- `tenant_id` (UUID, NOT NULL, FK → tenants.tenant_id)
- `monitor_config` (JSONB, NOT NULL, DEFAULT '{"alarms": {}}'::jsonb)
- `vendor_config` (JSONB, nullable)
- `metadata` (JSONB, nullable)

### 约束：
- 主键：`device_id`（每个设备只有一条配置记录）
- 外键：`device_id` → `devices.device_id`，`tenant_id` → `tenants.tenant_id`

### 业务规则：
1. 每个设备只有一条配置记录（PRIMARY KEY device_id）
2. `monitor_config`：存储设备的完整配置 JSON（包含所有报警项的配置、睡眠时间、阈值等）
3. `vendor_config`：厂家参考值，方便前端参考（只读）
4. 初次配置时，使用 `alarm_cloud` 中的默认值作为初始配置

---

## 2. Domain 模型检查

**状态：✅ 字段匹配正确**

当前文件：`internal/domain/alarm_device.go`

### Domain 模型：
```go
type AlarmDevice struct {
    DeviceID      string          `db:"device_id"`      // ✅ UUID, PRIMARY KEY
    TenantID      string          `db:"tenant_id"`      // ✅ UUID, NOT NULL
    MonitorConfig json.RawMessage `db:"monitor_config"` // ✅ JSONB, NOT NULL
    VendorConfig  json.RawMessage `db:"vendor_config"`  // ✅ JSONB, nullable
    Metadata      json.RawMessage `db:"metadata"`       // ✅ JSONB, nullable
}
```

**所有字段的 db tag 与数据库表结构一致。**

---

## 3. Repository 接口检查

**状态：✅ 已存在**

文件：`internal/repository/alarm_device_repo.go`

接口方法：
- `GetAlarmDevice` - 获取设备的告警配置
- `UpsertAlarmDevice` - 创建或更新设备的告警配置
- `DeleteAlarmDevice` - 删除设备的告警配置
- `ListAlarmDevices` - 批量查询设备的告警配置（支持分页）

---

## 4. Repository 实现检查

**状态：✅ 字段匹配正确**

文件：`internal/repository/postgres_alarm_device.go`

### 检查结果：

**GetAlarmDevice** (第 30-39 行)：
- ✅ 查询字段：`device_id`, `tenant_id`, `monitor_config`, `vendor_config`, `metadata`
- ✅ 所有字段与数据库表结构一致

**UpsertAlarmDevice** (第 77-90 行)：
- ✅ INSERT/UPDATE 字段：`device_id`, `tenant_id`, `monitor_config`, `vendor_config`, `metadata`
- ✅ 所有字段与数据库表结构一致
- ✅ 正确处理 JSONB 类型转换

**DeleteAlarmDevice** (第 119-122 行)：
- ✅ DELETE 条件：`tenant_id`, `device_id`
- ✅ 字段匹配正确

**ListAlarmDevices** (第 166-177 行)：
- ✅ 查询字段：`device_id`, `tenant_id`, `monitor_config`, `vendor_config`, `metadata`
- ✅ 所有字段与数据库表结构一致

---

## 5. 问题总结

### ✅ 所有内容正确：
1. Domain 模型：所有字段的 db tag 与数据库表结构一致
2. Repository 接口：已定义，方法完整
3. Repository 实现：所有 SQL 查询字段与数据库表结构一致

### ✅ 数据库表结构：
- 表结构完整，字段定义正确
- 约束和索引都已创建

---

## 6. 结论

**alarm_device 表的所有代码与数据库表结构完全一致，无需修复。**

---

## 7. 相关代码位置

- Domain 模型：`internal/domain/alarm_device.go`
- Repository 接口：`internal/repository/alarm_device_repo.go`
- Repository 实现：`internal/repository/postgres_alarm_device.go`

