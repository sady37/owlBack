#!/usr/bin/env bash
# check_bathroom_dwell.sh — 只读核查 chair_dwell_state 里 __bathroom__ 哨兵行，
# 直接算出每个浴室的实际 floor 兜底阈（tFloorFor 浴室分支：clamp(max(20min, μ+1.5σ[N≥3], dms), ≤45min)）。
#
# sensor 不连库（铁律 sensor_asks_data_sync_not_db）；本工具走 data 侧同库只读 SELECT。
# 连接串从 owlBack/.env 读（单源，creds 变了不用改脚本）。
#
# 用法：
#   tools/check_bathroom_dwell.sh                 # 全部浴室哨兵行
#   tools/check_bathroom_dwell.sh fd00:0:3:411    # 只看匹配前缀的房（host(prefix) LIKE 'arg%'）
set -euo pipefail

ENV_FILE="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.env"
[[ -f "$ENV_FILE" ]] || { echo "找不到 $ENV_FILE"; exit 1; }
get() { grep -E "^$1=" "$ENV_FILE" | head -1 | cut -d= -f2-; }
export PGPASSWORD="$(get DB_PASSWORD)"
PSQL="psql -h $(get DB_HOST) -p $(get DB_PORT) -U $(get DB_USER) -d $(get DB_NAME) -X -q"

FILTER="${1:-}"
WHERE="object_id='__bathroom__'"
[[ -n "$FILTER" ]] && WHERE="$WHERE AND host(spatial_prefix) LIKE '${FILTER}%'"

echo "=== 浴室 floor 兜底实况（est_floor = 该房实际生效阈；dms 无条件计入，μ+1.5σ 仅 N≥3）==="
$PSQL -c "
SELECT
  host(spatial_prefix)                                                        AS room_prefix,
  round(dms::numeric)                                                         AS dms_sec,
  round((dms/60.0)::numeric,1)                                                AS dms_min,
  CASE WHEN dms>0 THEN to_char(to_timestamp(dmsm/1000),'MM-DD HH24:MI') END   AS dms_asof,
  round(dmu::numeric) AS mu, round(dsg::numeric) AS sig,
  COALESCE((SELECT SUM((b->>'n')::int) FROM jsonb_array_elements(dwin) b),0)  AS n,
  round((LEAST(2700, GREATEST(1200, dms,
        CASE WHEN COALESCE((SELECT SUM((b->>'n')::int) FROM jsonb_array_elements(dwin) b),0) >= 3
             THEN dmu+1.5*dsg ELSE 0 END))/60.0)::numeric,1)                  AS est_floor_min,
  updated_at
FROM chair_dwell_state
WHERE $WHERE
ORDER BY spatial_prefix;
"

echo "=== 汇总 ==="
$PSQL -c "
SELECT count(*)                             AS bathroom_rows,
       count(*) FILTER (WHERE dms > 0)      AS rows_with_maxsit,
       count(*) FILTER (WHERE dms = 0)      AS rows_default_20min
FROM chair_dwell_state WHERE $WHERE;
"
echo "提示：est_floor_min=20 且 dms=0 = 该浴室未被抬（走默认/μ+1.5σ），正常；要抬需 false_alarm 棘轮或补 dms。"
