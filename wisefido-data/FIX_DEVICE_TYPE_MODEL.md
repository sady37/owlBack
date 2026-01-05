# 修复 Device Type 和 Device Model 显示问题

## 📋 问题描述

前端设备列表页面显示：
- Device Type: 空
- Device Model: 空

但数据库中 `device_store` 表有数据，且 `devices` 表通过 `device_store_id` 正确关联。

## 🔍 问题分析

### 根本原因

后端 API 查询 `devices` 表时：
1. ✅ 虽然做了 `LEFT JOIN device_store`，但 **SELECT 语句中没有选择** `device_type` 和 `device_model` 字段
2. ✅ `GetDevice` 方法甚至没有 JOIN `device_store` 表
3. ✅ `domain.Device` 结构体中没有 `DeviceType` 和 `DeviceModel` 字段
4. ✅ `ToJSON()` 方法也没有包含这些字段

### 数据流

```
前端请求
  ↓
后端 API (wisefido-data)
  ↓
PostgresDevicesRepository.ListDevices()
  ↓
SELECT ... FROM devices d LEFT JOIN device_store ds ...
  ❌ 没有 SELECT ds.device_type, ds.device_model
  ↓
domain.Device (结构体中没有这些字段)
  ↓
ToJSON() (没有包含这些字段)
  ↓
前端收到空值
```

## ✅ 修复方案

### 1. 修改 `domain.Device` 结构体

添加物理属性字段（从 `device_store` 表获取）：

```go
// 物理属性（从 device_store 表获取，只读）
DeviceType  sql.NullString `db:"device_type"`  // from device_store.device_type
DeviceModel sql.NullString `db:"device_model"` // from device_store.device_model
IMEI        sql.NullString `db:"imei"`         // from device_store.imei
CommMode    sql.NullString `db:"comm_mode"`    // from device_store.comm_mode
MCUModel    sql.NullString `db:"mcu_model"`    // from device_store.mcu_model
FirmwareVersion sql.NullString `db:"firmware_version"` // from device_store.firmware_version
```

### 2. 修改 `ListDevices` 查询

在 SELECT 语句中添加 `device_store` 表的字段：

```go
SELECT
    d.device_id::text,
    d.tenant_id::text,
    d.device_store_id,
    d.device_name,
    d.serial_number,
    d.uid,
    d.bound_room_id,
    d.bound_bed_id,
    d.status,
    d.business_access,
    d.monitoring_enabled,
    d.metadata,
    ds.device_type,        // ✅ 新增
    ds.device_model,       // ✅ 新增
    ds.imei,               // ✅ 新增
    ds.comm_mode,          // ✅ 新增
    ds.mcu_model,          // ✅ 新增
    ds.firmware_version    // ✅ 新增
FROM devices d
LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id
```

### 3. 修改 `GetDevice` 查询

添加 JOIN 和字段：

```go
SELECT
    ...
    ds.device_type,
    ds.device_model,
    ds.imei,
    ds.comm_mode,
    ds.mcu_model,
    ds.firmware_version
FROM devices d
LEFT JOIN device_store ds ON d.device_store_id = ds.device_store_id  // ✅ 新增 JOIN
WHERE d.tenant_id = $1 AND d.device_id = $2
```

### 4. 修改 `ToJSON()` 方法

在 JSON 响应中包含物理属性：

```go
if d.DeviceType.Valid {
    m["device_type"] = d.DeviceType.String
}
if d.DeviceModel.Valid {
    m["device_model"] = d.DeviceModel.String
}
// ... 其他字段
```

## 📝 修改的文件

1. `/Users/sady3721/project/owlBack/wisefido-data/internal/domain/device.go`
   - 添加物理属性字段
   - 更新 `ToJSON()` 方法

2. `/Users/sady3721/project/owlBack/wisefido-data/internal/repository/postgres_devices.go`
   - 修改 `ListDevices()` 查询
   - 修改 `GetDevice()` 查询

## 🚀 验证步骤

### 1. 重启后端服务

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go run cmd/wisefido-data/main.go
```

### 2. 测试 API

```bash
# 获取设备列表
curl http://localhost:8080/admin/api/v1/devices?tenant_id=bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c

# 应该返回包含 device_type 和 device_model 的数据
```

### 3. 前端验证

访问设备列表页面，应该能看到：
- ✅ Device Type: Radar / Sleepad
- ✅ Device Model: HC2 / BM8701-2

## 📊 device_store 表状态

**当前状态**：
- 总设备数: 23
- 已分配: 8 (Demo 租户)
- 未分配: 15

**设备类型分布**：
- Radar: 11 个
- Sleepad: 1 个
- sleepad: 11 个

**注意**：`device_store` 表不是空的，它是所有设备的后台管理表，包含所有设备的物理属性和固件版本信息。

## ⚠️ 注意事项

1. **数据一致性**：确保 `devices.device_store_id` 正确关联到 `device_store.device_store_id`
2. **NULL 值处理**：使用 `sql.NullString` 处理可能为 NULL 的字段
3. **LEFT JOIN**：使用 LEFT JOIN 确保即使没有关联 `device_store` 的设备也能返回

