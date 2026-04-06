#!/usr/bin/env bash
# 执行 diagnose_card_creation.sql 诊断卡片创建问题
# 用法:
#   ./run_diagnose_cards.sh
#   RUN_TENANT_ID=xxx ./run_diagnose_cards.sh   # 可选，替换 SQL 中的 YOUR_TENANT_ID
# 依赖: docker compose / docker-compose 已起 postgresql，或 .env 中 DB_* 可用且本机有 psql

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OWLBACK_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$OWLBACK_DIR/docker-compose.yml"
SQL_FILE="${SCRIPT_DIR}/diagnose_card_creation.sql"

if [[ ! -f "$SQL_FILE" ]]; then
  echo "Missing $SQL_FILE"
  exit 1
fi

# 可选：替换 tenant
if [[ -n "$RUN_TENANT_ID" ]]; then
  TMP_SQL=$(mktemp)
  sed "s/YOUR_TENANT_ID/$RUN_TENANT_ID/g" "$SQL_FILE" > "$TMP_SQL"
  trap 'rm -f "$TMP_SQL"' EXIT
  RUN_FILE="$TMP_SQL"
else
  RUN_FILE="$SQL_FILE"
fi

_compose_postgres_running() {
  local q
  if docker compose version >/dev/null 2>&1; then
    q=$(docker compose -f "$COMPOSE_FILE" ps -q postgresql 2>/dev/null || true)
    [[ -n "$q" ]]
  elif command -v docker-compose >/dev/null 2>&1; then
    q=$(docker-compose -f "$COMPOSE_FILE" ps -q postgresql 2>/dev/null || true)
    [[ -n "$q" ]]
  else
    return 1
  fi
}

_compose_exec_psql() {
  if docker compose version >/dev/null 2>&1; then
    docker compose -f "$COMPOSE_FILE" exec -T postgresql \
      psql -U postgres -d owlrd -v ON_ERROR_STOP=1 -f - < "$RUN_FILE"
  else
    docker-compose -f "$COMPOSE_FILE" exec -T postgresql \
      psql -U postgres -d owlrd -v ON_ERROR_STOP=1 -f - < "$RUN_FILE"
  fi
}

run_psql() {
  if _compose_postgres_running; then
    _compose_exec_psql
  elif [[ -n "$DB_HOST" ]] && [[ -n "$DB_NAME" ]]; then
    export PGPASSWORD="${DB_PASSWORD:-postgres}"
    psql -h "$DB_HOST" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "$DB_NAME" \
      -v ON_ERROR_STOP=1 -f "$RUN_FILE"
  else
    echo "Load env: source $OWLBACK_DIR/load_env.sh"
    [[ -f "$OWLBACK_DIR/.env" ]] && set -a && source "$OWLBACK_DIR/.env" && set +a
    export PGPASSWORD="${DB_PASSWORD:-postgres}"
    psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-owlrd}" \
      -v ON_ERROR_STOP=1 -f "$RUN_FILE"
  fi
}

cd "$OWLBACK_DIR"
[[ -f "$OWLBACK_DIR/load_env.sh" ]] && source "$OWLBACK_DIR/load_env.sh" || true
run_psql
