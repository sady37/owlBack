-- 检查表结构
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'iot_timeseries' 
ORDER BY ordinal_position;

-- 检查是否有数据
SELECT COUNT(*) as total_records FROM iot_timeseries;
