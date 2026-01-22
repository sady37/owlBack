-- 删除 device_store 中 device_uid 与 device_code 均为 NULL 或空的记录

-- 1. 预览要删除的记录
SELECT 
    device_id,
    device_uid,
    device_code,
    device_type,
    device_model,
    tenant_id,
    import_date
FROM device_store
WHERE (device_uid IS NULL OR device_uid = '')
  AND (device_code IS NULL OR device_code = '')
ORDER BY import_date DESC;

-- 2. 统计要删除的数量
SELECT COUNT(*) AS records_to_delete
FROM device_store
WHERE (device_uid IS NULL OR device_uid = '')
  AND (device_code IS NULL OR device_code = '');

-- 3. 执行删除
BEGIN;

DELETE FROM device_store
WHERE (device_uid IS NULL OR device_uid = '')
  AND (device_code IS NULL OR device_code = '');

SELECT COUNT(*) AS remaining
FROM device_store
WHERE (device_uid IS NULL OR device_uid = '')
  AND (device_code IS NULL OR device_code = '');

COMMIT;
