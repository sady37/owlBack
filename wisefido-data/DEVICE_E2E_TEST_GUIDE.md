# Device 端点端到端测试指南

## 📋 测试目标

验证新的 `DeviceHandler` 与前端（owlFront）的集成是否正常工作。

---

## 🚀 启动服务

### 1. 启动 wisefido-data 服务

```bash
cd /Users/sady3721/project/owlBack
docker-compose up -d wisefido-data
```

或者直接运行：

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
go run cmd/wisefido-data/main.go
```

### 2. 确认服务已启动

```bash
curl http://localhost:8080/health
```

应该返回：
```json
{
  "status": "healthy",
  "timestamp": "...",
  "services": {
    "redis": "healthy",
    "database": "healthy"
  }
}
```

---

## 🔍 测试端点

### 1. GET /admin/api/v1/devices - 查询设备列表

#### 1.1 准备测试数据

确保数据库中有测试设备：

```sql
-- 创建测试租户
INSERT INTO tenants (tenant_id, tenant_name, domain, status)
VALUES ('00000000-0000-0000-0000-000000000002', 'Test Device Tenant', 'test-device.local', 'active')
ON CONFLICT (tenant_id) DO NOTHING;

-- 创建设备库存
INSERT INTO device_store (device_store_id, tenant_id, device_type, serial_number, uid, status)
VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002', 'Radar', 'TEST-SERIAL-001', 'TEST-UID-001', 'available')
ON CONFLICT (device_store_id) DO NOTHING;

-- 创建设备
INSERT INTO devices (device_id, tenant_id, device_store_id, device_name, serial_number, uid, status, business_access, monitoring_enabled)
VALUES (
  '00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000002',
  'Test Device',
  'TEST-SERIAL-001',
  'TEST-UID-001',
  'online',
  'approved',
  true
)
ON CONFLICT (device_id) DO UPDATE SET
  device_name = EXCLUDED.device_name,
  status = EXCLUDED.status,
  business_access = EXCLUDED.business_access,
  monitoring_enabled = EXCLUDED.monitoring_enabled;
```

#### 1.2 测试查询设备列表

```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices?tenant_id=00000000-0000-0000-0000-000000000002&page=1&size=20" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002"
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "items": [
      {
        "device_id": "00000000-0000-0000-0000-000000000002",
        "tenant_id": "00000000-0000-0000-0000-000000000002",
        "device_name": "Test Device",
        "status": "online",
        "business_access": "approved",
        "monitoring_enabled": true,
        "serial_number": "TEST-SERIAL-001",
        "uid": "TEST-UID-001",
        "bound_room_id": null,
        "bound_bed_id": null
      }
    ],
    "total": 1
  }
}
```

#### 1.3 测试过滤条件

**按状态过滤**：
```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices?tenant_id=00000000-0000-0000-0000-000000000002&status=online" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002"
```

**按业务访问权限过滤**：
```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices?tenant_id=00000000-0000-0000-0000-000000000002&business_access=approved" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002"
```

**搜索设备**：
```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices?tenant_id=00000000-0000-0000-0000-000000000002&search_type=device_name&search_keyword=Test" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002"
```

---

### 2. GET /admin/api/v1/devices/:id - 查询设备详情

#### 2.1 测试查询设备详情

```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000002" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002"
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "device_id": "00000000-0000-0000-0000-000000000002",
    "tenant_id": "00000000-0000-0000-0000-000000000002",
    "device_name": "Test Device",
    "status": "online",
    "business_access": "approved",
    "monitoring_enabled": true,
    "serial_number": "TEST-SERIAL-001",
    "uid": "TEST-UID-001",
    "bound_room_id": null,
    "bound_bed_id": null
  }
}
```

#### 2.2 测试设备不存在

```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000000" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002"
```

**预期响应**：
```json
{
  "code": -1,
  "type": "error",
  "message": "device not found",
  "result": null
}
```

---

### 3. PUT /admin/api/v1/devices/:id - 更新设备

#### 3.1 测试更新设备

```bash
curl -X PUT "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000002" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002" \
  -d '{
    "device_name": "Updated Device Name",
    "status": "offline",
    "business_access": "pending",
    "monitoring_enabled": false
  }'
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "success": true
  }
}
```

#### 3.2 测试设备绑定验证

**无效绑定（unit_id 但缺少 bound_room_id/bound_bed_id）**：
```bash
curl -X PUT "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000002" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002" \
  -d '{
    "unit_id": "00000000-0000-0000-0000-000000000001"
  }'
```

**预期响应**：
```json
{
  "code": -1,
  "type": "error",
  "message": "invalid binding: unit_id provided but bound_room_id/bound_bed_id missing",
  "result": null
}
```

---

### 4. DELETE /admin/api/v1/devices/:id - 删除设备

#### 4.1 测试删除设备

```bash
curl -X DELETE "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000002" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002"
```

**预期响应**：
```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "success": true
  }
}
```

**注意**：删除是软删除（禁用设备），设备状态会变为 `disabled`，不会出现在列表中。

---

## ✅ 验证清单

### 功能验证

- [ ] GET /admin/api/v1/devices - 查询设备列表成功
- [ ] GET /admin/api/v1/devices - 过滤条件正常
- [ ] GET /admin/api/v1/devices/:id - 查询设备详情成功
- [ ] GET /admin/api/v1/devices/:id - 设备不存在错误
- [ ] PUT /admin/api/v1/devices/:id - 更新设备成功
- [ ] PUT /admin/api/v1/devices/:id - 设备绑定验证
- [ ] DELETE /admin/api/v1/devices/:id - 删除设备成功

### 响应格式验证

- [ ] 成功响应格式：`{code: 2000, type: "success", message: "ok", result: {...}}`
- [ ] 错误响应格式：`{code: -1, type: "error", message: "...", result: null}`
- [ ] HTTP 状态码：200 OK（错误通过 code=-1 表示）

### 前端集成验证

- [ ] 前端设备列表功能正常
- [ ] 前端设备详情功能正常
- [ ] 前端设备更新功能正常
- [ ] 前端设备删除功能正常
- [ ] 前端错误提示正常

---

## 🔍 路由优先级验证

由于新 Handler 的路由注册在 `RegisterAdminUnitDeviceRoutes` 之后，新 Handler 会优先处理请求。

**验证方法**：
1. 检查日志，确认请求被新 Handler 处理
2. 在 Handler 中添加日志，确认请求到达新 Handler
3. 测试响应格式，确认与新 Handler 一致

---

## 📝 测试结果记录

### 测试日期：__________

### 测试环境：
- 服务地址：`http://localhost:8080`
- 数据库：PostgreSQL
- 测试租户：00000000-0000-0000-0000-000000000002

### 测试结果：

| 端点 | 测试场景 | 结果 | 备注 |
|------|---------|------|------|
| GET /admin/api/v1/devices | 查询列表 | ✅/❌ | |
| GET /admin/api/v1/devices | 过滤条件 | ✅/❌ | |
| GET /admin/api/v1/devices/:id | 查询详情 | ✅/❌ | |
| GET /admin/api/v1/devices/:id | 设备不存在 | ✅/❌ | |
| PUT /admin/api/v1/devices/:id | 更新设备 | ✅/❌ | |
| PUT /admin/api/v1/devices/:id | 绑定验证 | ✅/❌ | |
| DELETE /admin/api/v1/devices/:id | 删除设备 | ✅/❌ | |

### 前端集成测试：

- [ ] 设备列表页面正常
- [ ] 设备详情页面正常
- [ ] 设备更新功能正常
- [ ] 设备删除功能正常
- [ ] 错误提示正常

### 问题记录：

1. 
2. 
3. 

---

## 🎯 确认步骤

完成所有测试后，确认以下事项：

1. ✅ 所有端点响应格式正确
2. ✅ 所有端点 HTTP 状态码正确
3. ✅ 前端集成正常
4. ✅ 日志无异常
5. ✅ 性能无异常

**确认后，可以移除 `RegisterAdminUnitDeviceRoutes` 中的旧 Device 路由。**

