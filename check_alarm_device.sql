-- 查询 alarm_device 表中的记录
-- device_id: 791fc634-69de-4987-b7eb-803c17e545a5
-- tenant_id: bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c

SELECT 
    device_id::text,
    tenant_id::text,
    monitor_config,
    vendor_config,
    metadata,
    created_at,
    updated_at,
    updated_by
FROM alarm_device
WHERE tenant_id = 'bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c' 
  AND device_id = '791fc634-69de-4987-b7eb-803c17e545a5';

-- 如果记录存在，显示 monitor_config 的格式化内容
SELECT 
    device_id::text,
    tenant_id::text,
    monitor_config::text as monitor_config_json,
    jsonb_pretty(monitor_config) as monitor_config_formatted
FROM alarm_device
WHERE tenant_id = 'bb045e6b-7bc2-4e59-af2e-d8b1adc77f2c' 
  AND device_id = '791fc634-69de-4987-b7eb-803c17e545a5';
