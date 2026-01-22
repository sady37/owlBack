-- 检查 demo 租户的 device 表中是否有 status='disabled' 的设备
-- 以及 device_store 和 devices 表的数量对比

-- 1. 获取 demo 租户的 tenant_id
SELECT tenant_id, tenant_name FROM tenants WHERE tenant_name = 'demo';

-- 2. 统计 device_store 中 demo 租户的设备数量
SELECT 
    COUNT(*) as device_store_total
FROM device_store ds
WHERE ds.tenant_id = (SELECT tenant_id FROM tenants WHERE tenant_name = 'demo');

-- 3. 统计 devices 表中 demo 租户的设备数量（按 status 分组）
SELECT 
    status,
    COUNT(*) as count
FROM devices
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE tenant_name = 'demo')
GROUP BY status
ORDER BY status;

-- 4. 统计 devices 表中 demo 租户的设备总数（排除 disabled）
SELECT 
    COUNT(*) as devices_total_excluding_disabled
FROM devices
WHERE tenant_id = (SELECT tenant_id FROM tenants WHERE tenant_name = 'demo')
  AND status <> 'disabled';

-- 5. 列出所有 demo 租户的 devices（包括 disabled）
SELECT 
    d.device_id,
    d.device_uid,
    d.device_name,
    d.status,
    d.business_access,
    ds.device_code,
    ds.device_type
FROM devices d
LEFT JOIN device_store ds ON d.device_id = ds.device_id
WHERE d.tenant_id = (SELECT tenant_id FROM tenants WHERE tenant_name = 'demo')
ORDER BY d.status, d.device_uid;

-- 6. 列出所有 demo 租户的 device_store（包括未在 devices 表中的）
SELECT 
    ds.device_id,
    ds.device_uid,
    ds.device_code,
    ds.device_type,
    ds.tenant_id,
    CASE 
        WHEN d.device_id IS NULL THEN 'NOT IN devices'
        ELSE d.status
    END as status_in_devices
FROM device_store ds
LEFT JOIN devices d ON ds.device_id = d.device_id
WHERE ds.tenant_id = (SELECT tenant_id FROM tenants WHERE tenant_name = 'demo')
ORDER BY status_in_devices, ds.device_uid;
