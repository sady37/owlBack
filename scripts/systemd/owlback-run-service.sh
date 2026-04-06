#!/usr/bin/env bash
# 供 systemd 单模块启动；与 start-owlback.sh 同源环境约定。
# 用法: owlback-run-service.sh wisefido-data|wisefido-cardagg|wisefido-qinglan|wisefido-sleepace|wisefido-iot|wisefido-ai
set -euo pipefail

MODULE="${1:-}"
if [[ -z "$MODULE" ]]; then
  echo "usage: $0 <module>" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWLBACK="$(cd "$SCRIPT_DIR/../.." && pwd)"
OWL_ROOT="$(cd "$OWLBACK/.." && pwd)"
LOG_DIR="${LOG_DIR:-$OWL_ROOT/log}"
mkdir -p "$LOG_DIR"

cd "$OWLBACK"
if [[ -f .env ]]; then
  set -a
  # shellcheck source=/dev/null
  source .env
  set +a
fi

export DB_HOST="${DB_HOST:-127.0.0.1}"
export DB_PORT="${DB_PORT:-5432}"
export DB_USER="${DB_USER:-postgres}"
export DB_PASSWORD="${DB_PASSWORD:-postgres}"
export DB_NAME="${DB_NAME:-owlrd}"
export DB_SSLMODE="${DB_SSLMODE:-disable}"
export REDIS_ADDR="${REDIS_ADDR:-127.0.0.1:6379}"
export REDIS_PASSWORD="${REDIS_PASSWORD:-TeLunSu-36kr}"
export MQTT_BROKER="${MQTT_BROKER:-127.0.0.1}"
export MQTT_PORT="${MQTT_PORT:-1883}"
export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"
export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
export CARD_POLLING_INTERVAL="${CARD_POLLING_INTERVAL:-86400}"
export CARD_AGGREGATION_ENABLED="${CARD_AGGREGATION_ENABLED:-true}"
export CARD_AGGREGATION_INTERVAL="${CARD_AGGREGATION_INTERVAL:-2}"

case "$MODULE" in
  wisefido-data)
    cd "$OWLBACK/wisefido-data"
    export HTTP_ADDR="${HTTP_ADDR:-:8080}"
    export QINGLAN_API_BASE_URL="${QINGLAN_API_BASE_URL:-http://127.0.0.1:8081}"
    exec go run cmd/wisefido-data/main.go >>"$LOG_DIR/wisefido-data.log" 2>&1
    ;;
  wisefido-cardagg)
    cd "$OWLBACK/wisefido-cardagg"
    exec go run "$OWLBACK/wisefido-cardagg/main.go" >>"$LOG_DIR/wisefido-cardagg.log" 2>&1
    ;;
  wisefido-qinglan)
    cd "$OWLBACK/wisefido-qinglan"
    export HTTP_HOST="${HTTP_HOST:-0.0.0.0}"
    export HTTP_PORT="${HTTP_PORT:-8081}"
    export QINGLAN_HTTPS_PORT="${QINGLAN_HTTPS_PORT:-${RADAR_HTTPS_PORT:-8443}}"
    export QINGLAN_HTTPS_CERT_FILE="${QINGLAN_HTTPS_CERT_FILE:-$RADAR_HTTPS_CERT_FILE}"
    export QINGLAN_HTTPS_KEY_FILE="${QINGLAN_HTTPS_KEY_FILE:-$RADAR_HTTPS_KEY_FILE}"
    exec go run cmd/wisefido-qinglan/main.go >>"$LOG_DIR/wisefido-qinglan.log" 2>&1
    ;;
  wisefido-sleepace)
    if [[ ! -d "$OWLBACK/wisefido-sleepace" ]]; then
      echo "wisefido-sleepace directory missing, exit" >&2
      exit 1
    fi
    cd "$OWLBACK/wisefido-sleepace"
    export MQTT_CLIENT_ID="${MQTT_CLIENT_ID:-wisefido-sleepace-2}"
    exec go run cmd/wisefido-sleepace/main.go -env dev >>"$LOG_DIR/wisefido-sleepace.log" 2>&1
    ;;
  wisefido-iot)
    cd "$OWLBACK/wisefido-iot"
    export HTTP_ADDR="${HTTP_ADDR:-:8085}"
    export REDIS_DB="${REDIS_DB:-0}"
    export STREAM_IOT_MONITOR="${STREAM_IOT_MONITOR:-iot:monitor:stream}"
    export STREAM_IOT_STAT="${STREAM_IOT_STAT:-iot:stat:stream}"
    export STREAM_IOT_EVENT="${STREAM_IOT_EVENT:-iot:event:stream}"
    export STREAM_IOT_ALARM="${STREAM_IOT_ALARM:-iot:alarm:stream}"
    export STREAM_IOT_AUTH="${STREAM_IOT_AUTH:-iot:auth:stream}"
    export STREAM_RADAR_MONITOR="${STREAM_RADAR_MONITOR:-iot:monitor:stream}"
    export STREAM_RADAR_STAT="${STREAM_RADAR_STAT:-iot:stat:stream}"
    export STREAM_RADAR_EVENT="${STREAM_RADAR_EVENT:-iot:event:stream}"
    export STREAM_RADAR_ALARM="${STREAM_RADAR_ALARM:-iot:alarm:stream}"
    export STREAM_RADAR_AUTH="${STREAM_RADAR_AUTH:-iot:auth:stream}"
    export STREAM_SLEEPACE_MONITOR="${STREAM_SLEEPACE_MONITOR:-iot:monitor:stream}"
    export STREAM_SLEEPACE_STAT="${STREAM_SLEEPACE_STAT:-iot:stat:stream}"
    export STREAM_SLEEPACE_EVENT="${STREAM_SLEEPACE_EVENT:-iot:event:stream}"
    export STREAM_SLEEPACE_ALARM="${STREAM_SLEEPACE_ALARM:-iot:alarm:stream}"
    export STREAM_SLEEPACE_AUTH="${STREAM_SLEEPACE_AUTH:-iot:auth:stream}"
    export CONSUMER_GROUP="${CONSUMER_GROUP:-iot-timeseries-group}"
    export CONSUMER_NAME="${CONSUMER_NAME:-iot-timeseries-1}"
    exec go run cmd/wisefido-iot/main.go >>"$LOG_DIR/wisefido-iot.log" 2>&1
    ;;
  wisefido-ai)
    cd "$OWLBACK/wisefido-ai"
    export REDIS_DB="${REDIS_DB:-0}"
    exec go run cmd/wisefido-ai/main.go >>"$LOG_DIR/wisefido-ai.log" 2>&1
    ;;
  *)
    echo "unknown module: $MODULE" >&2
    exit 1
    ;;
esac
