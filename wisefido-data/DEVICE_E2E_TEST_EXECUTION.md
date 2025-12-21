# Device 端到端测试执行指南

## 📋 测试执行步骤

### 步骤 1: 启动服务

```bash
cd /Users/sady3721/project/owlBack
docker-compose up -d wisefido-data
```

等待服务启动（约 10-30 秒），然后验证：

```bash
curl http://localhost:8080/health
```

**预期输出**：
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

### 步骤 2: 准备测试数据

连接到 PostgreSQL 数据库：

```bash
# 如果使用 Docker Compose
docker-compose exec postgresql psql -U postgres -d owlrd

# 或者直接连接
psql -h localhost -U postgres -d owlrd
```

执行以下 SQL 脚本：

```sql
-- 创建测试租户
INSERT INTO tenants (tenant_id, tenant_name, domain, status)
VALUES ('00000000-0000-0000-0000-000000000002', 'Test Device Tenant', 'test-device.local', 'active')
ON CONFLICT (tenant_id) DO UPDATE SET
  tenant_name = EXCLUDED.tenant_name,
  domain = EXCLUDED.domain,
  status = EXCLUDED.status;

-- 创建设备库存
INSERT INTO device_store (device_store_id, tenant_id, device_type, serial_number, uid, status)
VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002', 'Radar', 'TEST-SERIAL-001', 'TEST-UID-001', 'available')
ON CONFLICT (device_store_id) DO UPDATE SET
  tenant_id = EXCLUDED.tenant_id,
  device_type = EXCLUDED.device_type,
  serial_number = EXCLUDED.serial_number,
  uid = EXCLUDED.uid,
  status = EXCLUDED.status;

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
  tenant_id = EXCLUDED.tenant_id,
  device_store_id = EXCLUDED.device_store_id,
  device_name = EXCLUDED.device_name,
  serial_number = EXCLUDED.serial_number,
  uid = EXCLUDED.uid,
  status = EXCLUDED.status,
  business_access = EXCLUDED.business_access,
  monitoring_enabled = EXCLUDED.monitoring_enabled;

-- 验证数据
SELECT device_id, device_name, status, business_access FROM devices WHERE tenant_id = '00000000-0000-0000-0000-000000000002';
```

**预期输出**：
```
device_id                              | device_name | status | business_access
--------------------------------------+-------------+--------+----------------
00000000-0000-0000-0000-000000000002  | Test Device | online | approved
```

---

### 步骤 3: 运行自动化测试

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
./scripts/test_device_endpoints.sh
```

**预期输出**：
```
==========================================
Device 端点端到端测试
==========================================
服务地址: http://localhost:8080
测试租户: 00000000-0000-0000-0000-000000000002
测试设备: 00000000-0000-0000-0000-000000000002
==========================================

=== 检查服务状态 ===
✓ 服务运行正常

=== 测试 GET /admin/api/v1/devices ===
HTTP 状态码: 200
响应: {
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "items": [...],
    "total": 1
  }
}
✓ 查询设备列表成功
设备数量: 1, 总数: 1

...

==========================================
测试总结
==========================================
总测试数: 7
通过: 7
失败: 0
==========================================
✓ 所有测试通过！
```

---

### 步骤 4: 手动测试（可选）

如果自动化测试通过，可以进行更详细的手动测试：

#### 4.1 测试查询设备列表

```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices?tenant_id=00000000-0000-0000-0000-000000000002&page=1&size=20" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002" | jq '.'
```

#### 4.2 测试查询设备详情

```bash
curl -X GET "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000002" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002" | jq '.'
```

#### 4.3 测试更新设备

```bash
curl -X PUT "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000002" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002" \
  -d '{
    "device_name": "Updated Device Name",
    "status": "offline",
    "business_access": "pending",
    "monitoring_enabled": false
  }' | jq '.'
```

#### 4.4 测试设备绑定验证

```bash
curl -X PUT "http://localhost:8080/admin/api/v1/devices/00000000-0000-0000-0000-000000000002" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: 00000000-0000-0000-0000-000000000002" \
  -d '{
    "unit_id": "00000000-0000-0000-0000-000000000001"
  }' | jq '.'
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

### 步骤 5: 验证前端集成

1. 打开前端应用（owlFront）
2. 登录系统
3. 导航到设备管理页面
4. 验证以下功能：
   - [ ] 设备列表正常显示
   - [ ] 设备详情正常显示
   - [ ] 设备更新功能正常
   - [ ] 设备删除功能正常
   - [ ] 错误提示正常

---

### 步骤 6: 检查日志

查看服务日志：

```bash
# Docker Compose
docker-compose logs -f wisefido-data | grep -i "device\|error"

# 或者直接查看日志文件
tail -f /path/to/wisefido-data.log | grep -i "device\|error"
```

**检查要点**：
- [ ] 无异常错误
- [ ] 请求正常处理
- [ ] 日志记录完整

---

### 步骤 7: 填写测试报告

参考 `DEVICE_E2E_TEST_REPORT.md` 填写测试结果。

---

## ⚠️ 常见问题

### 问题 1: 服务无法启动

**症状**：`curl http://localhost:8080/health` 返回错误

**解决方案**：
1. 检查 Docker 容器状态：`docker-compose ps`
2. 查看日志：`docker-compose logs wisefido-data`
3. 检查端口占用：`lsof -i :8080`
4. 重启服务：`docker-compose restart wisefido-data`

---

### 问题 2: 数据库连接失败

**症状**：服务启动但健康检查失败

**解决方案**：
1. 检查数据库容器：`docker-compose ps postgresql`
2. 检查数据库连接配置
3. 等待数据库完全启动（约 10-30 秒）

---

### 问题 3: 测试数据不存在

**症状**：查询返回空结果

**解决方案**：
1. 重新执行 SQL 脚本创建测试数据
2. 验证数据：`SELECT * FROM devices WHERE tenant_id = '00000000-0000-0000-0000-000000000002';`

---

### 问题 4: 路由未注册

**症状**：请求返回 404

**解决方案**：
1. 检查路由注册：确认 `RegisterDeviceRoutes` 已调用
2. 检查路由优先级：新 Handler 路由应在旧 Handler 之后注册
3. 查看服务日志确认路由注册

---

## ✅ 测试完成标准

所有测试通过后，确认：

1. ✅ 所有端点响应格式正确
2. ✅ 所有端点 HTTP 状态码正确
3. ✅ 前端集成正常
4. ✅ 日志无异常
5. ✅ 性能无异常

**确认后，可以移除 `RegisterAdminUnitDeviceRoutes` 中的旧 Device 路由。**

---

## 📝 下一步

测试完成后：

1. 填写测试报告
2. 记录问题和改进建议
3. 移除旧的 Device 路由（如果测试通过）
4. 更新文档

