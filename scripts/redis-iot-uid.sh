#!/usr/bin/env bash
# Tail a Redis stream and filter by `device_uid` (password taken from .env)
# Usage: ./redis_iot_device_uid.sh [STREAM] [DEVICE_UID]
# Defaults: STREAM=iot:monitor:stream

set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/../.env"

usage() {
  cat <<'USAGE'
Usage: redis-iot-uid.sh [STREAM] [DEVICE_UID]

Tail a Redis stream and filter by `device_uid`.

Defaults:
  STREAM=iot:monitor:stream

Options:
  -h, --help    Show this help and exit

Examples:
  ./redis-iot-uid.sh iot:monitor:stream E598A2ACD5F7
  REDIS_PASSWORD=yourpass ./redis-iot-uid.sh E598A2ACD5F7

Notes:
  The script requires `REDIS_PASSWORD` to be set (from ../.env or environment).
  It uses `XREAD BLOCK` to follow new stream entries in realtime.
USAGE
}

# Show help early (before loading .env) if requested
if [ "${1:-}" = "-h" ] || [ "${1:-}" = "--help" ]; then
  usage
  exit 0
fi

[ -f "$ENV_FILE" ] && set -a && source "$ENV_FILE" && set +a

REDIS_HOST="${REDIS_HOST:-127.0.0.1}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASS="${REDIS_PASSWORD:-${REDIS_PASS:-}}"

STREAM="${1:-iot:monitor:stream}"
UID_FILTER="${2:-}"

if [ -z "$REDIS_PASS" ]; then
  echo "Error: REDIS_PASSWORD not found in ${ENV_FILE}" >&2
  exit 1
fi

echo "Following Redis stream '${STREAM}' on ${REDIS_HOST}:${REDIS_PORT}, filter device_uid='${UID_FILTER:-(all)}'"

# get last id to avoid replaying old messages
last_id=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" XREVRANGE "$STREAM" + - COUNT 1 2>/dev/null | sed -n '1p' || true)
if [ -z "$last_id" ]; then
  last_id="$"
else
  # XREAD reads entries > ID, so use the last_id we found
  last_id="$last_id"
fi

echo "Starting from id: ${last_id} (use Ctrl+C to stop)"

while true; do
  out=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASS" XREAD BLOCK 0 STREAMS "$STREAM" "$last_id" 2>/dev/null || true)
  if [ -z "$out" ]; then
    continue
  fi
  # Print and filter JSON-like payloads; keep line-buffered grep for realtime
  echo "$out" | tr '\n' ' ' | sed 's/ *"device_uid"/\n"device_uid"/g' | grep --line-buffered -F "${UID_FILTER:-}" || true
  # update last_id to last seen stream id
  new_id=$(echo "$out" | tr ' ' '\n' | grep -E '^[0-9]+-[0-9]+' | tail -n1 || true)
  if [ -n "$new_id" ]; then
    last_id="$new_id"
  fi
done
