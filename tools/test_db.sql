SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'iot_timeseries' 
ORDER BY ordinal_position;
