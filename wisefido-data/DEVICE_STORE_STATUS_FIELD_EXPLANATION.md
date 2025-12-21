# device_store 表的 status 字段说明

## 📋 问题

**问题**：`device_store` 表没有 `status` 字段，这个 `status` 字段是什么？

---

## ✅ 答案

**`device_store` 表确实没有 `status` 字段。**

`status` 字段存在于 **`devices` 表**中，而不是 `device_store` 表中。

---

## 📊 两个表的区别

### 1. device_store 表（设备库存表）

**用途**：系统管理员管理设备库存、分配、出库、OTA 升级

**关键字段**：
- `device_type` - 设备类型
- `serial_number` / `uid` - 序列号/UID
- `tenant_id` - 租户 ID（分配状态）
- `allow_access` - **系统级访问权限**（BOOLEAN，不是 status）
- `firmware_version` - 固件版本
- `ota_target_firmware_version` - OTA 目标版本

**没有 `status` 字段**：因为 `device_store` 是库存表，设备在库存中只有"已分配"或"未分配"的概念（通过 `tenant_id` 判断），不需要运行状态。

---

### 2. devices 表（设备业务表）

**用途**：租户管理业务设备

**关键字段**：
- `device_name` - 设备名称（用户自定义）
- `status` - **设备运行状态**（online/offline/error/disabled）
- `business_access` - 租户业务访问权限（pending/approved/rejected）
- `monitoring_enabled` - 监控启用状态
- `bound_room_id` / `bound_bed_id` - 位置绑定

**有 `status` 字段**：因为 `devices` 是业务表，需要跟踪设备的运行状态。

---

## 🔍 status 字段的含义

### devices.status

**类型**：`VARCHAR(20)`

**可能的值**：
- `online` - 设备在线
- `offline` - 设备离线
- `error` - 设备错误
- `disabled` - 设备已禁用（软删除）

**用途**：表示设备的**运行状态**，由设备连接状态决定。

---

## 🔍 allow_access 字段的含义

### device_store.allow_access

**类型**：`BOOLEAN`

**可能的值**：
- `TRUE` - 系统允许设备接入业务系统
- `FALSE` - 系统不允许设备接入业务系统

**用途**：表示**系统级访问权限**，由系统管理员控制。

---

## 📊 两个字段的对比

| 字段 | 表 | 类型 | 用途 | 控制者 |
|------|-----|------|------|--------|
| `status` | `devices` | VARCHAR(20) | 设备运行状态（online/offline/error/disabled） | 系统自动更新（基于设备连接） |
| `allow_access` | `device_store` | BOOLEAN | 系统级访问权限（是否允许接入） | 系统管理员 |

---

## 🔗 关系说明

### 设备接入业务系统的条件

设备可以接入业务系统需要**同时满足**两个条件：

1. **系统级权限**：`device_store.allow_access = TRUE`（系统管理员设置）
2. **租户级权限**：`devices.business_access = 'approved'`（租户设置）

**注意**：`devices.status` 不影响设备是否可以接入，只表示设备当前的运行状态。

---

## 📝 测试数据脚本修正

### 错误的写法（已修正）

```sql
-- ❌ 错误：device_store 表没有 status 字段
INSERT INTO device_store (..., status)
VALUES (..., 'available')
```

### 正确的写法

```sql
-- ✅ 正确：device_store 表使用 allow_access
INSERT INTO device_store (device_store_id, tenant_id, device_type, serial_number, uid)
VALUES ('...', '...', 'Radar', 'TEST-SERIAL-001', 'TEST-UID-001');

-- ✅ 正确：devices 表使用 status
INSERT INTO devices (device_id, tenant_id, device_store_id, device_name, status, business_access, monitoring_enabled)
VALUES ('...', '...', '...', 'Test Device', 'online', 'approved', true);
```

---

## ✅ 总结

1. **`device_store` 表没有 `status` 字段**
2. **`status` 字段在 `devices` 表中**，用于表示设备运行状态
3. **`device_store` 表使用 `allow_access` 字段**，用于系统级访问权限控制
4. **两个字段用途不同**：
   - `devices.status`：设备运行状态（online/offline/error/disabled）
   - `device_store.allow_access`：系统是否允许设备接入（TRUE/FALSE）

---

## 🔧 已修正的文件

以下文件已修正，移除了 `device_store` 表中的 `status` 字段：

- ✅ `scripts/prepare_device_test_data.sql`
- ✅ `DEVICE_E2E_TEST_EXECUTION.md`
- ✅ `DEVICE_E2E_TESTING_START.md`
- ✅ `DEVICE_E2E_TEST_GUIDE.md`
- ✅ `DEVICE_E2E_TEST_REPORT.md`

**注意**：`device_service_integration_test.go` 中可能还有错误的引用，需要检查并修正。

