-- 删除 device_store 表中 device_uid 为 NULL 或空的记录

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
WHERE device_uid IS NULL OR device_uid = ''
ORDER BY import_date DESC;

-- 2. 统计要删除的数量
SELECT COUNT(*) AS records_to_delete
FROM device_store
WHERE device_uid IS NULL OR device_uid = '';

-- 3. 执行删除
BEGIN;

DELETE FROM device_store
WHERE device_uid IS NULL OR device_uid = '';

-- 查看删除结果
SELECT COUNT(*) AS remaining_empty_uid_records
FROM device_store
WHERE device_uid IS NULL OR device_uid = '';

COMMIT;
