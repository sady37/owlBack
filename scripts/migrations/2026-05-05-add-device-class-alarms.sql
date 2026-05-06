-- 给 alarm_device.monitor_config.items 补 5 项设备健康类 alarm
-- (Offline / DeviceFailure / SignalPoor / AngleException / SensorDetached)
-- is_enabled=1, alarm_level=ERROR, display_setting=1 (DisplayAlarmCloud)
--
-- 配合 owl-common AlarmDef.SkipUnhandledCount=true：
--  - alarm_events 落库做审计/历史查询
--  - cards.unhandled_alarm_N 不累加（icon 红点不被设备类拉起）
--  - card.pop_alarm 不写（不弹 handle / 不闪烁 / 不响 popAlarm 链路声音）
--  - UI 设备状态由 device:status hash 独立通道驱动（卡片 sleepace-offline / radar-offline 图标）
--
-- 幂等：同一 alarm_type 已存在则跳过，可重复执行。

DO $$
DECLARE
  rec RECORD;
  new_items JSONB;
  type_to_add TEXT;
  default_item JSONB;
  device_class_types TEXT[] := ARRAY['Offline', 'DeviceFailure', 'SignalPoor', 'AngleException', 'SensorDetached'];
BEGIN
  FOR rec IN SELECT device_id, monitor_config FROM alarm_device LOOP
    new_items := COALESCE(rec.monitor_config->'items', '[]'::jsonb);

    FOREACH type_to_add IN ARRAY device_class_types LOOP
      -- 检查是否已存在
      IF NOT (new_items @> jsonb_build_array(jsonb_build_object('alarm_type', type_to_add))) THEN
        default_item := jsonb_build_object(
          'alarm_type',      type_to_add,
          'is_enabled',      1,
          'alarm_level',     'ERROR',
          'display_setting', 1
        );
        new_items := new_items || jsonb_build_array(default_item);
      END IF;
    END LOOP;

    -- 仅在确实新增时 UPDATE
    IF new_items <> COALESCE(rec.monitor_config->'items', '[]'::jsonb) THEN
      UPDATE alarm_device
        SET monitor_config = jsonb_set(monitor_config, '{items}', new_items),
            updated_at = CURRENT_TIMESTAMP
        WHERE device_id = rec.device_id;
    END IF;
  END LOOP;
END $$;

-- 验证：5 项配置覆盖率应该都到 100%
SELECT COUNT(*) AS total_devices,
       COUNT(*) FILTER (WHERE monitor_config @> '{"items": [{"alarm_type": "Offline"}]}') AS has_offline,
       COUNT(*) FILTER (WHERE monitor_config @> '{"items": [{"alarm_type": "DeviceFailure"}]}') AS has_device_failure,
       COUNT(*) FILTER (WHERE monitor_config @> '{"items": [{"alarm_type": "SignalPoor"}]}') AS has_signal_poor,
       COUNT(*) FILTER (WHERE monitor_config @> '{"items": [{"alarm_type": "AngleException"}]}') AS has_angle,
       COUNT(*) FILTER (WHERE monitor_config @> '{"items": [{"alarm_type": "SensorDetached"}]}') AS has_sensor_detached
FROM alarm_device;
