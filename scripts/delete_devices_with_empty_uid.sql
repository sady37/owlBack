-- 删除 device_uid 为 NULL 或空的设备记录
-- 功能：批量删除 devices 表中 device_uid 为 NULL 或空字符串的设备
-- 注意：只删除未使用的设备（通过 is_device_used() 检查），并将 device_store.tenant_id 设为未分配

-- 1. 查看将要删除的设备（预览）
SELECT 
    d.device_id,
    d.device_name,
    d.device_uid,
    d.tenant_id,
    ds.device_type,
    ds.device_code,
    is_device_used(d.device_id) AS is_used,
    CASE 
        WHEN is_device_used(d.device_id) THEN '已使用，将跳过'
        ELSE '未使用，将被删除'
    END AS action
FROM devices d
JOIN device_store ds ON d.device_id = ds.device_id
WHERE d.device_uid IS NULL OR d.device_uid = ''
ORDER BY d.tenant_id, d.device_name;

-- 2. 统计将要删除的设备数量
SELECT 
    COUNT(*) AS total_devices_with_empty_uid,
    COUNT(*) FILTER (WHERE NOT is_device_used(d.device_id)) AS devices_to_delete,
    COUNT(*) FILTER (WHERE is_device_used(d.device_id)) AS devices_to_skip
FROM devices d
WHERE d.device_uid IS NULL OR d.device_uid = '';

-- 3. 执行删除（使用事务保证原子性）
-- 注意：执行前请先运行上面的预览查询确认要删除的设备
BEGIN;

-- 3.1. 更新 device_store 的 tenant_id 为未分配（仅针对未使用的设备）
UPDATE device_store ds
SET tenant_id = '00000000-0000-0000-0000-000000000000'
WHERE ds.device_id IN (
    SELECT d.device_id
    FROM devices d
    WHERE (d.device_uid IS NULL OR d.device_uid = '')
      AND NOT is_device_used(d.device_id)
);

-- 3.2. 物理删除 devices 表中的记录（仅删除未使用的设备）
DELETE FROM devices
WHERE (device_uid IS NULL OR device_uid = '')
  AND NOT is_device_used(device_id);

-- 查看删除结果
SELECT 
    'Deleted devices count' AS info,
    COUNT(*) AS count
FROM devices
WHERE device_uid IS NULL OR device_uid = '';

-- 提交事务（如果确认无误，取消下面的注释）
-- COMMIT;

-- 如果发现问题，可以回滚
-- ROLLBACK;
