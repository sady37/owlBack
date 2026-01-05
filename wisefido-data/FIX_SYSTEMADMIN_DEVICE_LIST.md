# 修复 SystemAdmin 查看 Device Management 为空的问题

## 📋 问题描述

**现象**：
- sysadmin 登录后，查看 Device Management 页面，记录为空
- 但数据库中 `devices` 表有 8 条记录（Demo 租户）

**原因分析**：
1. sysadmin 的 `tenant_id` 是 `00000000-0000-0000-0000-000000000001` (System 租户)
2. System 租户下没有 `devices` 记录（0 条）
3. 所有设备都在 Demo 租户下（8 条）
4. 后端 `ListDevices` 方法强制要求 `tenant_id`，查询条件是 `d.tenant_id = $1`
5. 所以当 sysadmin 查询时，使用的是 System 租户的 tenant_id，但 System 租户下没有设备

## ✅ 修复方案

### 1. 修改 `DeviceFilters` 结构体

添加 `IsSystemAdmin` 字段：

```go
type DeviceFilters struct {
	IsSystemAdmin  bool     // SystemAdmin 查看所有租户的设备（不限制 tenant_id）
	Status         []string
	BusinessAccess string
	DeviceType     string
	SearchType     string
	SearchKeyword  string
}
```

### 2. 修改 `ListDevicesRequest` 结构体

添加 `IsSystemAdmin` 字段：

```go
type ListDevicesRequest struct {
	TenantID       string
	IsSystemAdmin  bool     // SystemAdmin 查看所有租户的设备
	Status         []string
	// ...
}
```

### 3. 修改 `DeviceHandler.ListDevices`

检测 SystemAdmin 角色，并传递 `IsSystemAdmin` 标记：

```go
userRole := r.Header.Get("X-User-Role")
isSystemAdmin := strings.EqualFold(userRole, "SystemAdmin")

req := service.ListDevicesRequest{
	TenantID:      tenantID,
	IsSystemAdmin: isSystemAdmin && tenantID == SystemTenantID(),
	// ...
}
```

### 4. 修改 `PostgresDevicesRepository.ListDevices`

当 `IsSystemAdmin = true` 时，不限制 `tenant_id`：

```go
// SystemAdmin 查看所有设备时，不限制 tenant_id
where := []string{"d.status <> 'disabled'"}
args := []any{}
argN := 1

if !filters.IsSystemAdmin {
	// 普通租户：限制 tenant_id
	where = append(where, "d.tenant_id = $1")
	args = append(args, tenantID)
	argN = 2
}
```

## 📝 修改的文件

1. `/Users/sady3721/project/owlBack/wisefido-data/internal/repository/devices_repo.go`
   - 添加 `IsSystemAdmin` 字段到 `DeviceFilters`

2. `/Users/sady3721/project/owlBack/wisefido-data/internal/repository/postgres_devices.go`
   - 修改 `ListDevices` 查询逻辑，支持 SystemAdmin 查看所有设备

3. `/Users/sady3721/project/owlBack/wisefido-data/internal/service/device_service.go`
   - 添加 `IsSystemAdmin` 字段到 `ListDevicesRequest`
   - 传递 `IsSystemAdmin` 到 Repository

4. `/Users/sady3721/project/owlBack/wisefido-data/internal/http/device_handler.go`
   - 检测 SystemAdmin 角色
   - 传递 `IsSystemAdmin` 标记到 Service

## 🚀 验证步骤

### 1. 重启后端服务

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go run cmd/wisefido-data/main.go
```

### 2. 测试 API

```bash
# sysadmin 查看所有设备（应该返回 8 条）
curl -H "X-User-Role: SystemAdmin" \
     -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000001" \
     http://localhost:8080/admin/api/v1/devices
```

### 3. 前端验证

1. 使用 sysadmin 登录
2. 访问 Device Management 页面
3. 应该能看到所有 8 个设备（来自 Demo 租户）

## 📊 数据状态

**当前设备分布**：
- System 租户 (`00000000-0000-0000-0000-000000000001`): 0 个设备
- Demo 租户 (`bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c`): 8 个设备
  - Radar: 4 个
  - Sleepad: 1 个
  - sleepad: 3 个

**修复后**：
- SystemAdmin 可以查看所有租户的设备（8 个）
- 普通租户只能查看自己租户的设备

## ⚠️ 注意事项

1. **权限控制**：只有 SystemAdmin 角色才能查看所有设备
2. **数据隔离**：普通租户仍然只能查看自己租户的设备
3. **性能考虑**：SystemAdmin 查询所有设备时，数据量可能较大，建议添加分页

## 🔍 相关查询

### 查看所有设备（SystemAdmin）

```sql
SELECT 
    d.device_id,
    d.device_name,
    d.tenant_id,
    t.tenant_name,
    ds.device_type,
    ds.device_model
FROM devices d
JOIN device_store ds ON d.device_store_id = ds.device_store_id
LEFT JOIN tenants t ON d.tenant_id = t.tenant_id
WHERE d.status <> 'disabled'
ORDER BY t.tenant_name, d.device_name;
```

### 查看各租户设备数量

```sql
SELECT 
    t.tenant_name,
    COUNT(*) as device_count
FROM devices d
JOIN tenants t ON d.tenant_id = t.tenant_id
WHERE d.status <> 'disabled'
GROUP BY t.tenant_id, t.tenant_name
ORDER BY t.tenant_name;
```

