-- 清空指定表的记录
-- 表：config_versions, iot_timeseries, alarm_cloud, alarm_device, alarm_events

-- 1. 清空 config_versions 表
TRUNCATE TABLE config_versions CASCADE;

-- 2. iot_timeseries：删除并重建（与 owlRD/db/18_iot_timeseries.sql 一致）
DROP TABLE IF EXISTS iot_timeseries CASCADE;

CREATE TABLE iot_timeseries (
    id            BIGSERIAL NOT NULL,

    tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE RESTRICT,
    device_id     UUID REFERENCES device_store(device_id) ON DELETE SET NULL,
    device_uid    VARCHAR(50) NOT NULL,
    device_type   VARCHAR(50),
    card_id       VARCHAR(100),

    "timestamp"   BIGINT NOT NULL,

    topic_type    VARCHAR(50),
    category      VARCHAR(50),

    branch_name   VARCHAR(200),
    building_name VARCHAR(200),
    unit_name     VARCHAR(200),
    room_name     VARCHAR(200),
    bed_name      VARCHAR(200),

    data_value    JSONB NOT NULL,

    CONSTRAINT chk_data_value_array CHECK (jsonb_typeof(data_value) = 'array'),
    CONSTRAINT iot_timeseries_pkey PRIMARY KEY (id, "timestamp")
);

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_tenant_device_uid_ts
    ON iot_timeseries(tenant_id, device_uid, "timestamp" DESC);

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_device_id_ts
    ON iot_timeseries(device_id, "timestamp" DESC);

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_device_uid_ts
    ON iot_timeseries(device_uid, "timestamp" DESC);

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_timestamp
    ON iot_timeseries("timestamp" DESC);

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_topic_type_ts
    ON iot_timeseries(topic_type, "timestamp" DESC)
    WHERE topic_type IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_category_ts
    ON iot_timeseries(category, "timestamp" DESC)
    WHERE category IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_iot_timeseries_data_value_gin
    ON iot_timeseries USING GIN (data_value);

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
    PERFORM create_hypertable(
      'iot_timeseries',
      'timestamp',
      chunk_time_interval => 86400000,
      if_not_exists => TRUE
    );
  END IF;
END $$;

-- 3. 清空 alarm_cloud 表
TRUNCATE TABLE alarm_cloud CASCADE;

-- 4. 清空 alarm_device 表
TRUNCATE TABLE alarm_device CASCADE;

-- 5. 清空 alarm_events 表
DELETE FROM alarm_events;

-- 6. 重置 cards 的报警计数与 pop，否则 QueryCardAlarmState 仍会返回旧值，前端 active_err 等不归零
UPDATE cards SET
  unhandled_alarm_0 = 0, unhandled_alarm_1 = 0, unhandled_alarm_2 = 0,
  unhandled_alarm_3 = 0, unhandled_alarm_4 = 0,
  pop_alarm_level = '', pop_alarm_type = '', pop_alarm_event_id = NULL;

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
    'alarm_device' AS table_name, COUNT(*) AS record_count FROM alarm_device
UNION ALL
SELECT 
    'alarm_events' AS table_name, COUNT(*) AS record_count FROM alarm_events;
