#!/usr/bin/env bash
# 从 config_versions 导出 room_layout 的 config_data 为 JSON 文件，便于与当前画布对比。
# 用法: ./export_room_layouts.sh [输出目录]  默认输出到 scripts/room_layouts_export

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="${1:-$SCRIPT_DIR/room_layouts_export}"
mkdir -p "$OUT_DIR"

# 使用 .env 中的 DB 配置（若存在）
if [ -f "$SCRIPT_DIR/../.env" ]; then
  set -a
  source "$SCRIPT_DIR/../.env"
  set +a
fi
PGHOST="${DB_HOST:-127.0.0.1}"
PGPORT="${DB_PORT:-5432}"
PGUSER="${DB_USER:-postgres}"
PGPASSWORD="${DB_PASSWORD:-postgres}"
PGDATABASE="${DB_NAME:-owlrd}"
export PGHOST PGPORT PGUSER PGPASSWORD PGDATABASE

# 是否通过 Docker 执行 psql
RUN_PSQL() {
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q 'owl-postgresql'; then
    docker exec -i owl-postgresql psql -U "$PGUSER" -d "$PGDATABASE" "$@"
  else
    psql "$@"
  fi
}

# 先取所有 entity_id 列表
IDS=$(RUN_PSQL -t -A -c "
  SELECT DISTINCT entity_id::text FROM (
    SELECT entity_id, row_number() OVER (PARTITION BY entity_id ORDER BY valid_from DESC) AS rn
    FROM config_versions
    WHERE config_type = 'room_layout' AND (valid_to IS NULL OR valid_to > now())
  ) x WHERE rn = 1;
")

echo "Exporting room_layout from config_versions to $OUT_DIR"
count=0
for entity_id in $IDS; do
  [ -z "$entity_id" ] && continue
  outfile="$OUT_DIR/room_layout_${entity_id}.json"
  RUN_PSQL -t -A -c "
    SELECT config_data FROM (
      SELECT config_data, row_number() OVER (ORDER BY valid_from DESC) AS rn
      FROM config_versions
      WHERE config_type = 'room_layout' AND entity_id = '$entity_id' AND (valid_to IS NULL OR valid_to > now())
    ) x WHERE rn = 1;
  " > "$outfile"
  count=$((count + 1))
  echo "  $entity_id -> $outfile"
done

echo "Done. Exported $count room layout(s) to $OUT_DIR"
