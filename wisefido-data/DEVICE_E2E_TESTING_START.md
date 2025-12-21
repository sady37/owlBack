# Device 端到端测试快速开始

## 🚀 快速开始

### 1. 启动服务

```bash
cd /Users/sady3721/project/owlBack
docker-compose up -d wisefido-data
```

### 2. 准备测试数据

连接到数据库并执行：

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

### 3. 运行自动化测试

```bash
cd /Users/sady3721/project/owlBack/wisefido-data
./scripts/test_device_endpoints.sh
```

### 4. 手动测试

参考 `DEVICE_E2E_TEST_GUIDE.md` 进行详细的手动测试。

---

## 📋 测试清单

### 自动化测试

- [ ] 运行 `test_device_endpoints.sh` 脚本
- [ ] 检查所有测试是否通过
- [ ] 查看测试输出和日志

### 手动测试

- [ ] GET /admin/api/v1/devices - 查询设备列表
- [ ] GET /admin/api/v1/devices - 过滤条件
- [ ] GET /admin/api/v1/devices/:id - 查询设备详情
- [ ] GET /admin/api/v1/devices/:id - 设备不存在
- [ ] PUT /admin/api/v1/devices/:id - 更新设备
- [ ] PUT /admin/api/v1/devices/:id - 绑定验证
- [ ] DELETE /admin/api/v1/devices/:id - 删除设备

### 前端集成测试

- [ ] 设备列表页面正常
- [ ] 设备详情页面正常
- [ ] 设备更新功能正常
- [ ] 设备删除功能正常
- [ ] 错误提示正常

---

## 📝 测试报告

填写 `DEVICE_E2E_TEST_REPORT.md` 记录测试结果。

---

## 🎯 完成标准

所有测试通过后：

1. ✅ 所有端点响应格式正确
2. ✅ 所有端点 HTTP 状态码正确
3. ✅ 前端集成正常
4. ✅ 日志无异常
5. ✅ 性能无异常

**确认后，可以移除 `RegisterAdminUnitDeviceRoutes` 中的旧 Device 路由。**

