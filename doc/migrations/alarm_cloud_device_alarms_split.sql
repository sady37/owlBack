-- alarm_cloud.device_alarms JSONB schema 升级：合并 string → 双字段 object
--
-- 旧格式：{"SleepPad": {"LeftBed": "WARNING"}, ...}
--   - "WARNING"/"CRITICAL"/"ERROR"/"INFORMATIONAL" → is_enabled=1, alarm_level=<这个值>
--   - "DISABLED"/"DISABLE" → is_enabled=0, alarm_level="" （等级丢失，无法挽回）
-- 新格式：{"SleepPad": {"LeftBed": {"is_enabled": 1, "alarm_level": "WARNING"}}, ...}
--
-- 必须在 wisefido-data 部署 commit splitting alarm_cloud schema 之前/之时跑一次。
-- 跑完后 wisefido-data 启动；之后 buildDomainAlarmCloudFromConfig / buildAlarmCloudConfigFromDomain
-- 仅识别 object 格式，遇 string 反序列化报错。
--
-- 幂等：jsonb_typeof(value) 已经是 'object' 的项原样保留；string 才转换。

BEGIN;

UPDATE alarm_cloud
SET device_alarms = (
    SELECT jsonb_object_agg(devType, devMap_new)
    FROM (
        SELECT
            devType,
            jsonb_object_agg(
                alarmType,
                CASE jsonb_typeof(val)
                    WHEN 'object' THEN val
                    WHEN 'string' THEN
                        CASE WHEN UPPER(val #>> '{}') IN ('DISABLED', 'DISABLE')
                            THEN jsonb_build_object('is_enabled', 0, 'alarm_level', '')
                            ELSE jsonb_build_object('is_enabled', 1, 'alarm_level', val #>> '{}')
                        END
                    ELSE val
                END
            ) AS devMap_new
        FROM jsonb_each(device_alarms) AS outer_pair(devType, devMap)
        CROSS JOIN LATERAL jsonb_each(devMap) AS inner_pair(alarmType, val)
        GROUP BY devType
    ) AS rebuilt
)
WHERE jsonb_typeof(device_alarms) = 'object'
  AND device_alarms <> '{}'::jsonb
  AND EXISTS (
      SELECT 1
      FROM jsonb_each(device_alarms) AS o(devType, devMap)
      CROSS JOIN LATERAL jsonb_each(devMap) AS i(alarmType, val)
      WHERE jsonb_typeof(val) = 'string'
  );

-- 验证：剩余还存有 string 值的行（应该为 0）
SELECT
    tenant_id,
    devType,
    alarmType,
    val
FROM alarm_cloud
CROSS JOIN LATERAL jsonb_each(device_alarms) AS o(devType, devMap)
CROSS JOIN LATERAL jsonb_each(devMap) AS i(alarmType, val)
WHERE jsonb_typeof(val) = 'string';

COMMIT;
