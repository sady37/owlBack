-- 清空指定表的记录
-- 表：config_versions, iot_timeseries, alarm_cloud, alarm_device

-- 1. 清空 config_versions 表
TRUNCATE TABLE config_versions CASCADE;

-- 2. 清空 iot_timeseries 表（TimescaleDB 超表）
TRUNCATE TABLE iot_timeseries CASCADE;

-- 3. 清空 alarm_cloud 表
TRUNCATE TABLE alarm_cloud CASCADE;

-- 4. 清空 alarm_device 表
TRUNCATE TABLE alarm_device CASCADE;

-- 显示清空后的记录数（应该都是 0）
SELECT 
    'config_versions' AS table_name, COUNT(*) AS record_count FROM config_versions
UNION ALL
SELECT 
    'iot_timeseries' AS table_name, COUNT(*) AS record_count FROM iot_timeseries
UNION ALL
SELECT 
    'alarm_cloud' AS table_name, COUNT(*) AS record_count FROM alarm_cloud
UNION ALL
SELECT 
    'alarm_device' AS table_name, COUNT(*) AS record_count FROM alarm_device;
