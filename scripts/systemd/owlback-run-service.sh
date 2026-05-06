#!/usr/bin/env bash
# 供 systemd 单模块启动；与 start-owlback.sh 同源环境约定。
# 用法: owlback-run-service.sh wisefido-data|wisefido-cardagg|wisefido-qinglan|wisefido-sleepace|wisefido-iot|wisefido-sensor
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

# 为wisefido-qinglan生成唯一的MQTT客户端ID，使用无条件覆盖
if [[ "$MODULE" == "wisefido-qinglan" ]]; then
  export MQTT_CLIENT_ID="wisefido-qinglan-systemd-$(date +%s%3N)"
elif [[ "$MODULE" == "wisefido-sleepace" ]]; then
  export MQTT_CLIENT_ID="wisefido-sleepace-$(date +%s%3N)"
else
  export MQTT_CLIENT_ID="${MQTT_CLIENT_ID:-$MODULE-$(date +%s)}"
fi

export LOG_LEVEL="${LOG_LEVEL:-info}"
export LOG_FORMAT="${LOG_FORMAT:-json}"
export CARD_TRIGGER_MODE="${CARD_TRIGGER_MODE:-polling}"
export CARD_POLLING_INTERVAL="${CARD_POLLING_INTERVAL:-86400}"
export CARD_AGGREGATION_ENABLED="${CARD_AGGREGATION_ENABLED:-true}"
export CARD_AGGREGATION_INTERVAL="${CARD_AGGREGATION_INTERVAL:-2}"

# go run 在子进程里跑编译产物，systemd 停服时易漏杀监听进程；构建到 .bin 后 exec 单进程。
owlback_go_exec() {
  local dir="$1"
  local build_pkg="$2"
  local bin_name="$3"
  local log_file="$4"
  shift 4
  local bin="$dir/.bin/$bin_name"
  mkdir -p "$dir/.bin"
  cd "$dir" || exit 1
  local stale=0
  [[ ! -x "$bin" ]] && stale=1
  [[ "$dir/go.mod" -nt "$bin" ]] && stale=1
  if [[ -f "$dir/go.sum" && "$dir/go.sum" -nt "$bin" ]]; then
    stale=1
  fi
  # 检查任意 .go 文件是否比二进制新（避免只检查 main.go 漏掉其他源文件变更）
  if [[ "$stale" -eq 0 ]] && find "$dir" -name '*.go' -newer "$bin" -print -quit 2>/dev/null | grep -q .; then
    stale=1
  fi
  if [[ "$stale" -eq 1 ]]; then
    go build -o "$bin" "$build_pkg"
  fi
  # Write a prominent startup header so systemd-launched services show a clear start marker in logs
  printf "\n########################################################################\n" >>"$log_file" 2>&1
  printf "# STARTING %s (systemd)\n" "$bin_name" >>"$log_file" 2>&1
  printf "# Time: %s\n" "$(date -u '+%Y-%m-%d %H:%M:%S %Z')" >>"$log_file" 2>&1
  printf "########################################################################\n\n" >>"$log_file" 2>&1
  exec "$bin" "$@" >>"$log_file" 2>&1
}

case "$MODULE" in
  wisefido-data)
    export HTTP_ADDR="${HTTP_ADDR:-:8080}"
    export QINGLAN_API_BASE_URL="${QINGLAN_API_BASE_URL:-http://127.0.0.1:8081}"
    owlback_go_exec "$OWLBACK/wisefido-data" "./cmd/wisefido-data" "wisefido-data" "$LOG_DIR/wisefido-data.log"
    ;;
  wisefido-cardagg)
    owlback_go_exec "$OWLBACK/wisefido-cardagg" "." "wisefido-cardagg" "$LOG_DIR/wisefido-cardagg.log"
    ;;
  wisefido-qinglan)
    export HTTP_HOST="${HTTP_HOST:-0.0.0.0}"
    export HTTP_PORT="${HTTP_PORT:-8081}"
    export QINGLAN_HTTPS_PORT="${QINGLAN_HTTPS_PORT:-${RADAR_HTTPS_PORT:-8443}}"
    export QINGLAN_HTTPS_CERT_FILE="${QINGLAN_HTTPS_CERT_FILE:-$RADAR_HTTPS_CERT_FILE}"
    export QINGLAN_HTTPS_KEY_FILE="${QINGLAN_HTTPS_KEY_FILE:-$RADAR_HTTPS_KEY_FILE}"
    owlback_go_exec "$OWLBACK/wisefido-qinglan" "./cmd/wisefido-qinglan" "wisefido-qinglan" "$LOG_DIR/wisefido-qinglan.log"
    ;;
  wisefido-sleepace)
    if [[ ! -d "$OWLBACK/wisefido-sleepace" ]]; then
      echo "wisefido-sleepace directory missing, exit" >&2
      exit 1
    fi
    export MQTT_CLIENT_ID="${MQTT_CLIENT_ID:-wisefido-sleepace-2}"
    owlback_go_exec "$OWLBACK/wisefido-sleepace" "./cmd/wisefido-sleepace" "wisefido-sleepace" "$LOG_DIR/wisefido-sleepace.log" -env dev
    ;;
  wisefido-iot)
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
    owlback_go_exec "$OWLBACK/wisefido-iot" "./cmd/wisefido-iot" "wisefido-iot" "$LOG_DIR/wisefido-iot.log"
    ;;
  wisefido-sensor)
    export REDIS_DB="${REDIS_DB:-0}"
    owlback_go_exec "$OWLBACK/wisefido-sensor" "./cmd/wisefido-sensor" "wisefido-sensor" "$LOG_DIR/wisefido-sensor.log"
    ;;
  *)
    echo "unknown module: $MODULE" >&2
    exit 1
    ;;
esac
