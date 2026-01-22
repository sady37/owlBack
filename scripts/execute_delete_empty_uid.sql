BEGIN;

UPDATE device_store ds
SET tenant_id = '00000000-0000-0000-0000-000000000000'
WHERE ds.device_id IN (
    SELECT d.device_id
    FROM devices d
    WHERE (d.device_uid IS NULL OR d.device_uid = '')
      AND NOT is_device_used(d.device_id)
);

DELETE FROM devices
WHERE (device_uid IS NULL OR device_uid = '')
  AND NOT is_device_used(device_id);

COMMIT;
