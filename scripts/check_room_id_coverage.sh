#!/bin/bash
# check_room_id_coverage.sh — 监测 alarm_events / iot_timeseries 中 room_id/unit_id 漏写情况
#
# 基线（2026-05-06 20:45 Denver）：
#   alarm_events (since 17:42): 2 total, 0 no_room, 0 no_unit
#   iot_timeseries alarm/event (since 19:15): 1872 total, 1351 no_room
#
# 用法: ./check_room_id_coverage.sh
# 期望: alarm_events no_room/no_unit = 0；iot_timeseries 增量 no_room 比例 < 5%

PGPASSWORD=postgres psql -h localhost -U postgres -d owlrd -c "
SELECT 'snapshot' AS type, NOW() AT TIME ZONE 'America/Denver' AS ts;

SELECT
  'alarm_events (since PR-1 17:42)' AS scope,
  COUNT(*) AS total_rows,
  COUNT(*) FILTER (WHERE room_id IS NULL) AS no_room,
  COUNT(*) FILTER (WHERE unit_id IS NULL) AS no_unit
FROM alarm_events
WHERE triggered_at > '2026-05-06 17:42 America/Denver';

SELECT
  'iot_timeseries alarm/event (since PR-3 19:15)' AS scope,
  COUNT(*) AS total_rows,
  COUNT(*) FILTER (WHERE room_id IS NULL) AS no_room
FROM iot_timeseries
WHERE \"timestamp\" > EXTRACT(EPOCH FROM TIMESTAMP '2026-05-06 19:15' AT TIME ZONE 'America/Denver') * 1000
  AND topic_type IN ('alarm','event');

-- 1 小时滑窗（最新 1h 覆盖率，用于查看稳态）
SELECT
  'iot_timeseries alarm/event (last 1h)' AS scope,
  COUNT(*) AS total_rows,
  COUNT(*) FILTER (WHERE room_id IS NULL) AS no_room,
  ROUND(100.0 * COUNT(*) FILTER (WHERE room_id IS NOT NULL) / NULLIF(COUNT(*), 0), 1) AS coverage_pct
FROM iot_timeseries
WHERE \"timestamp\" > EXTRACT(EPOCH FROM NOW() - INTERVAL '1 hour') * 1000
  AND topic_type IN ('alarm','event');

-- 出现 0 填的 device（找新 publisher path 漏）
SELECT
  device_uid, device_type, COUNT(*) AS rows, COUNT(room_id) AS has_room
FROM iot_timeseries
WHERE \"timestamp\" > EXTRACT(EPOCH FROM NOW() - INTERVAL '1 hour') * 1000
  AND topic_type IN ('alarm','event')
GROUP BY device_uid, device_type
HAVING COUNT(room_id) = 0
ORDER BY rows DESC LIMIT 10;
"
