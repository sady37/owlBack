#!/usr/bin/env bash
# 检查 config_versions 中 room_layout 的版本历史与变化：列出版本、导出最新与上一版并 diff。
# 用法: ./check_room_layout_changes.sh [room_id]
#   无参数时检查所有有 room_layout 的 room；传 room_id 时只检查该房间。

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUT_DIR="$SCRIPT_DIR/room_layouts_export"
mkdir -p "$OUT_DIR"

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

RUN_PSQL() {
  if docker ps --format '{{.Names}}' 2>/dev/null | grep -q 'owl-postgresql'; then
    docker exec -i owl-postgresql psql -U "$PGUSER" -d "$PGDATABASE" "$@"
  else
    psql "$@"
  fi
}

ROOM_ID="$1"

echo "=== room_layout 版本历史 (config_versions) ==="
if [ -n "$ROOM_ID" ]; then
  RUN_PSQL -c "
    SELECT entity_id AS room_id, version_id, valid_from, valid_to,
      (valid_to IS NULL) AS is_current,
      length(config_data::text) AS config_bytes
    FROM config_versions
    WHERE config_type = 'room_layout' AND entity_id::text = '$ROOM_ID'
    ORDER BY valid_from DESC;
  "
else
  RUN_PSQL -c "
    SELECT entity_id AS room_id, version_id, valid_from, valid_to,
      (valid_to IS NULL) AS is_current,
      length(config_data::text) AS config_bytes
    FROM config_versions
    WHERE config_type = 'room_layout'
    ORDER BY entity_id, valid_from DESC;
  "
fi

echo ""
echo "=== 导出最新版与上一版并对比 ==="

# 取有 room_layout 的 entity_id 列表（若有传入 room_id 则只查该 room）
if [ -n "$ROOM_ID" ]; then
  IDS=$(RUN_PSQL -t -A -c "
    SELECT DISTINCT entity_id::text FROM config_versions
    WHERE config_type = 'room_layout' AND entity_id::text = '$ROOM_ID';
  ")
else
  IDS=$(RUN_PSQL -t -A -c "
    SELECT DISTINCT entity_id::text FROM config_versions WHERE config_type = 'room_layout';
  ")
fi

for entity_id in $IDS; do
  [ -z "$entity_id" ] && continue
  LATEST_FILE="$OUT_DIR/room_layout_${entity_id}_latest.json"
  PREV_FILE="$OUT_DIR/room_layout_${entity_id}_previous.json"
  RUN_PSQL -t -A -c "
    SELECT config_data FROM (
      SELECT config_data, row_number() OVER (ORDER BY valid_from DESC) AS rn
      FROM config_versions
      WHERE config_type = 'room_layout' AND entity_id = '$entity_id'
    ) x WHERE rn = 1;
  " > "$LATEST_FILE"
  RUN_PSQL -t -A -c "
    SELECT config_data FROM (
      SELECT config_data, row_number() OVER (ORDER BY valid_from DESC) AS rn
      FROM config_versions
      WHERE config_type = 'room_layout' AND entity_id = '$entity_id'
    ) x WHERE rn = 2;
  " 2>/dev/null > "$PREV_FILE" || true

  echo "[$entity_id]"
  echo "  最新 -> $LATEST_FILE"
  if [ -s "$PREV_FILE" ]; then
    echo "  上一版 -> $PREV_FILE"
    if command -v jq >/dev/null 2>&1; then
      echo "  --- objects 的 id / angle ---"
      echo "  上一版:"
      jq -r '.objects[]? | "    \(.id): angle=\(.angle // "null")"' "$PREV_FILE" 2>/dev/null || true
      echo "  最新:"
      jq -r '.objects[]? | "    \(.id): angle=\(.angle // "null")"' "$LATEST_FILE" 2>/dev/null || true
      echo "  --- 变更摘要 (最新相对上一版) ---"
      P_OBJ=$(jq -c '.objects | length' "$PREV_FILE" 2>/dev/null)
      L_OBJ=$(jq -c '.objects | length' "$LATEST_FILE" 2>/dev/null)
      echo "    objects 数量: ${P_OBJ:-?} -> ${L_OBJ:-?}"
      jq -r '
        .objects[]? | select(.device?.iot?.radar?.areas != null) | . as $o |
        "    雷达 \(.id): areas 数 \(.device.iot.radar.areas | length), baseline.queriedAt=\(.device.iot.radar.baseline.queriedAt // "null")"
      ' "$LATEST_FILE" 2>/dev/null | head -5
      echo "    timestamp: $(jq -r '.timestamp // "null"' "$PREV_FILE" 2>/dev/null) -> $(jq -r '.timestamp // "null"' "$LATEST_FILE" 2>/dev/null)"
    fi
    echo "  --- 全文 diff 见下方 (可忽略格式差异) ---"
    diff "$PREV_FILE" "$LATEST_FILE" 2>/dev/null || true
  else
    echo "  (仅有一版，无上一版)"
  fi
  echo ""
done

echo "导出目录: $OUT_DIR"
