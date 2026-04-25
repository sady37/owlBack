#!/usr/bin/env bash
# Query recent radar track records from iot_timeseries (PostgreSQL).
# Useful for diagnosing pose/position/z signal quality of specific devices.
#
# Usage: ./radar-tracks.sh <device_uid> [options]
#
# 背景：雷达 track 数据以 JSONB 存在 iot_timeseries.data_value 里，
# 每条是一个数组（一帧），数组元素含 track_id/pose/position_x/y/z 等字段。
# 本脚本封装常见查询：逐帧列表 / pose 聚合统计 / 多 track 帧过滤。
#
# 依赖：psql 客户端 + owlBack/.env（含 DB_* 变量）

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"

usage() {
  cat <<'USAGE'
Usage: radar-tracks.sh <device_uid> [options]

查询指定雷达设备近期的 track 记录（来自 iot_timeseries 表）。

Options:
  -m, --minutes N      时间窗口（分钟，默认 5）
  -s, --stats          按 pose 聚合统计（帧数 / x,y,z 范围 / z 均值&标准差）
  -M, --multi-only     仅显示包含多个 track 的帧（检测有人经过）
  -l, --limit N        单帧列表模式下最多行数（默认 40）
  -p, --poses "1,2,5"  过滤 pose 值（逗号分隔；不传则全部）
  -h, --help           显示此帮助

Examples:
  # 近 5 分钟 E598A2ACD523 逐帧列表
  ./radar-tracks.sh E598A2ACD523

  # 近 10 分钟 pose+z 分布统计
  ./radar-tracks.sh E598A2ACD523 -m 10 -s

  # 只看多 track 的帧（偶有人经过）
  ./radar-tracks.sh E598A2ACD523 -M -m 30

  # 只看跌倒相关 pose（2=疑似跌倒, 5=确认跌倒）
  ./radar-tracks.sh 9923003AB197 -p 2,5 -m 60

Notes:
  - DB 连接参数从 owlBack/.env 读 DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME
  - 数据来自 public.iot_timeseries，category='track' 过滤
  - 时间字段 timestamp 是毫秒（UNIX ms）
USAGE
}

# 早期 help 处理（在 .env 加载之前）
if [ $# -eq 0 ] || [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi

DEVICE_UID="$1"
shift

# 默认参数
MINUTES=5
MODE="list"           # list | stats | multi
LIMIT=40
POSE_FILTER=""

while [ $# -gt 0 ]; do
  case "$1" in
    -m|--minutes)
      MINUTES="$2"; shift 2 ;;
    -s|--stats)
      MODE="stats"; shift ;;
    -M|--multi-only)
      MODE="multi"; shift ;;
    -l|--limit)
      LIMIT="$2"; shift 2 ;;
    -p|--poses)
      POSE_FILTER="$2"; shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      echo "Unknown option: $1" >&2
      usage; exit 1 ;;
  esac
done

# 加载 .env
[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a

DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-owlrd}"

if [ -z "$DB_PASSWORD" ]; then
  echo "Error: DB_PASSWORD not set (check ${ENV_FILE})" >&2
  exit 1
fi

# 时间窗口（毫秒）
WINDOW_MS=$(( MINUTES * 60 * 1000 ))

# pose 过滤 SQL 片段
POSE_WHERE=""
if [ -n "$POSE_FILTER" ]; then
  POSE_WHERE="AND EXISTS (SELECT 1 FROM jsonb_array_elements(data_value) AS e WHERE (e->>'pose')::int = ANY(ARRAY[${POSE_FILTER}]))"
fi

export PGPASSWORD="$DB_PASSWORD"
PSQL="psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -X"

echo "Device: $DEVICE_UID  |  Window: last ${MINUTES} min  |  Mode: $MODE"
echo "---"

case "$MODE" in
  list)
    $PSQL <<SQL
SELECT to_timestamp(timestamp/1000)::timestamp(0) AS ts,
       jsonb_array_length(data_value) AS n,
       jsonb_path_query_array(data_value, '\$[*].pose') AS poses,
       jsonb_path_query_array(data_value, '\$[*].position_x') AS xs,
       jsonb_path_query_array(data_value, '\$[*].position_y') AS ys,
       jsonb_path_query_array(data_value, '\$[*].position_z') AS zs,
       jsonb_path_query_array(data_value, '\$[*].track_id') AS tids
FROM iot_timeseries
WHERE device_uid='${DEVICE_UID}'
  AND timestamp > (extract(epoch from now())*1000 - ${WINDOW_MS})::bigint
  AND category='track'
  ${POSE_WHERE}
ORDER BY timestamp DESC
LIMIT ${LIMIT};
SQL
    ;;

  multi)
    $PSQL <<SQL
SELECT to_timestamp(timestamp/1000)::timestamp(0) AS ts,
       jsonb_array_length(data_value) AS n,
       jsonb_path_query_array(data_value, '\$[*].pose') AS poses,
       jsonb_path_query_array(data_value, '\$[*].position_x') AS xs,
       jsonb_path_query_array(data_value, '\$[*].position_y') AS ys,
       jsonb_path_query_array(data_value, '\$[*].position_z') AS zs,
       jsonb_path_query_array(data_value, '\$[*].track_id') AS tids
FROM iot_timeseries
WHERE device_uid='${DEVICE_UID}'
  AND timestamp > (extract(epoch from now())*1000 - ${WINDOW_MS})::bigint
  AND category='track'
  AND jsonb_array_length(data_value) > 1
  ${POSE_WHERE}
ORDER BY timestamp DESC
LIMIT ${LIMIT};
SQL
    ;;

  stats)
    $PSQL <<SQL
WITH tracks AS (
  SELECT (elem->>'track_id')::int AS tid,
         (elem->>'pose')::int     AS pose,
         (elem->>'position_x')::int AS x,
         (elem->>'position_y')::int AS y,
         (elem->>'position_z')::int AS z
  FROM iot_timeseries,
       jsonb_array_elements(data_value) AS elem
  WHERE device_uid='${DEVICE_UID}'
    AND timestamp > (extract(epoch from now())*1000 - ${WINDOW_MS})::bigint
    AND category='track'
    ${POSE_WHERE}
)
SELECT pose,
       CASE pose
         WHEN 0 THEN 'Init/Unknown'
         WHEN 1 THEN 'Walking'
         WHEN 2 THEN 'SuspectedFall'
         WHEN 3 THEN 'Sitting (蹲坐)'
         WHEN 4 THEN 'Standing'
         WHEN 5 THEN 'Fall'
         WHEN 6 THEN 'Lying'
         WHEN 7 THEN 'SuspectedSitGround'
         WHEN 8 THEN 'SitGround'
         WHEN 9 THEN 'BedSitUp'
         WHEN 10 THEN 'SuspectedBedSitUp'
         WHEN 11 THEN 'ConfirmedBedSitUp'
         WHEN 12 THEN 'Running'
         ELSE '?' END AS label,
       COUNT(*) AS n,
       MIN(x) AS x_min, MAX(x) AS x_max,
       MIN(y) AS y_min, MAX(y) AS y_max,
       MIN(z) AS z_min, MAX(z) AS z_max,
       ROUND(AVG(z)::numeric, 1)        AS z_avg,
       ROUND(STDDEV_POP(z)::numeric, 1) AS z_std
FROM tracks
GROUP BY pose
ORDER BY pose;
SQL
    ;;
esac
