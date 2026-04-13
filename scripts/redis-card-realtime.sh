#!/usr/bin/env bash
# Tail card realtime stream and filter by `card_id` (password taken from .env)
# Usage: ./redis_card_realtime.sh [STREAM] [CARD_ID]
# Defaults: STREAM=card:realtime:stream

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a

REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASS="${REDIS_PASSWORD:-${REDIS_PASS:-}}"

STREAM="${1:-card:realtime:stream}"
CARD_ID="${2:-}"

if [ -z "$REDIS_PASS" ]; then
  echo "Error: REDIS_PASSWORD not found in ${ENV_FILE}" >&2
  exit 1
fi

echo "Following Redis stream '${STREAM}' on ${REDIS_HOST}:${REDIS_PORT}, filter card_id='${CARD_ID:-(all)}'"

last_id=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" XREVRANGE "$STREAM" + - COUNT 1 2>/dev/null | sed -n '1p' || true)
if [ -z "$last_id" ]; then
  last_id="$"
else
  last_id="$last_id"
fi

echo "Starting from id: ${last_id} (use Ctrl+C to stop)"

while true; do
  out=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" XREAD BLOCK 0 STREAMS "$STREAM" "$last_id" 2>/dev/null || true)
  if [ -z "$out" ]; then
    continue
  fi
  echo "$out" | tr '\n' ' ' | sed 's/ *"card_id"/\n"card_id"/g' | grep --line-buffered -F "${CARD_ID:-}" || true
  new_id=$(echo "$out" | tr ' ' '\n' | grep -E '^[0-9]+-[0-9]+' | tail -n1 || true)
  if [ -n "$new_id" ]; then
    last_id="$new_id"
  fi
done
