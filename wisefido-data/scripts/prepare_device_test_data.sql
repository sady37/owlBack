-- Device 端到端测试数据准备脚本

-- 创建测试租户
INSERT INTO tenants (tenant_id, tenant_name, domain, status)
VALUES ('00000000-0000-0000-0000-000000000002', 'Test Device Tenant', 'test-device.local', 'active')
ON CONFLICT (tenant_id) DO UPDATE SET
  tenant_name = EXCLUDED.tenant_name,
  domain = EXCLUDED.domain,
  status = EXCLUDED.status;

-- 创建设备库存（注意：device_store 表可能没有 status 字段）
INSERT INTO device_store (device_store_id, tenant_id, device_type, serial_number, uid)
VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002', 'Radar', 'TEST-SERIAL-001', 'TEST-UID-001')
ON CONFLICT (device_store_id) DO UPDATE SET
  tenant_id = EXCLUDED.tenant_id,
  device_type = EXCLUDED.device_type,
  serial_number = EXCLUDED.serial_number,
  uid = EXCLUDED.uid;

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
SELECT 
  d.device_id,
  d.device_name,
  d.status,
  d.business_access,
  d.monitoring_enabled,
  d.serial_number,
  d.uid
FROM devices d
WHERE d.tenant_id = '00000000-0000-0000-0000-000000000002';

