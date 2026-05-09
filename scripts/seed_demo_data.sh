#!/bin/bash
# seed_demo_data.sh
# 跑 seed_demo_data.sql 应用 demo tenant 完整测试数据。
# 幂等：可多次跑（ON CONFLICT DO NOTHING）。

set -euo pipefail
CONTAINER=${OWL_PG_CONTAINER:-owl-postgresql}
DB=${OWL_V2_DB:-owl_v2}
SQL_FILE="$(dirname "$0")/seed_demo_data.sql"

if [[ ! -f "$SQL_FILE" ]]; then
    echo "[ERR] $SQL_FILE not found"
    exit 1
fi

echo "[+] Apply seed_demo_data.sql → $DB"
docker exec -i "$CONTAINER" psql -U postgres -d "$DB" -v ON_ERROR_STOP=1 < "$SQL_FILE"
echo "[done] demo seed applied"
