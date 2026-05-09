#!/bin/bash
# setup_owl_v2.sh
# 在已运行的 postgres container 内创建 owl_v2 数据库 + 跑 dbv2 schema
# 不影响现有 owlrd 库（旧 schema 保留作回退）
#
# 用法（从 owlBack/ 目录执行）：
#   bash scripts/setup_owl_v2.sh                # 创建 owl_v2 + 全套 dbv2
#   bash scripts/setup_owl_v2.sh --drop         # 先 DROP 再 CREATE（重置）
#   bash scripts/setup_owl_v2.sh --check        # 仅检查 owl_v2 是否存在 + 列表

set -euo pipefail

CONTAINER=${OWL_PG_CONTAINER:-owl-postgresql}
DB=${OWL_V2_DB:-owl_v2}
DBV2_DIR=${OWL_V2_DIR:-../owlRD/dbv2}

cmd_check() {
    docker exec "$CONTAINER" psql -U postgres -tAc \
        "SELECT 1 FROM pg_database WHERE datname='$DB';" \
        | grep -q 1 && echo "[OK] $DB exists" \
        || echo "[NO] $DB does not exist"
}

cmd_drop() {
    echo "[!] DROP DATABASE $DB"
    docker exec -i "$CONTAINER" psql -U postgres <<EOF
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='$DB' AND pid != pg_backend_pid();
DROP DATABASE IF EXISTS $DB;
EOF
}

cmd_create() {
    echo "[+] CREATE DATABASE $DB"
    docker exec -i "$CONTAINER" psql -U postgres <<EOF
CREATE DATABASE $DB;
EOF
}

cmd_apply_schema() {
    echo "[+] Apply dbv2 schema to $DB"
    cd "$(dirname "$0")/.."   # owlBack/
    if [[ ! -d "$DBV2_DIR" ]]; then
        echo "[ERR] $DBV2_DIR not found（cwd=$(pwd)）"
        exit 1
    fi
    for f in "$DBV2_DIR"/*.sql; do
        echo "  ▸ $(basename "$f")"
        docker exec -i "$CONTAINER" psql -U postgres -d "$DB" -v ON_ERROR_STOP=1 < "$f"
    done
    echo "[OK] schema applied"
}

case "${1:-}" in
    --check)
        cmd_check
        ;;
    --drop)
        cmd_drop
        cmd_create
        cmd_apply_schema
        cmd_check
        ;;
    *)
        if cmd_check 2>/dev/null | grep -q "exists"; then
            echo "[!] $DB 已存在；用 --drop 重建，或 --check 仅查看"
            exit 1
        fi
        cmd_create
        cmd_apply_schema
        cmd_check
        ;;
esac

echo "[done] $DB ready"
echo
echo "切换 owl 服务到新库："
echo "  1) 编辑 .env: DB_NAME=owl_v2"
echo "  2) 重启 wisefido-* services"
echo "  3) 旧库 owlrd 保留作回退（DROP DATABASE owlrd 在确认没问题后执行）"
