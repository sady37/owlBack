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
  # 跨 module stale-check：解析 go.mod 的 `replace xxx => ../yyy` 本地依赖
  # （e.g. wisefido-cardagg 依赖 ../owl-common；owl-common 改了 cardagg 也要重建）
  if [[ "$stale" -eq 0 && -f "$dir/go.mod" ]]; then
    local replace_targets
    replace_targets=$(awk '
      /^replace[[:space:]]+/ {
        for (i=1;i<=NF;i++) if ($i=="=>") { print $(i+1); break }
      }
    ' "$dir/go.mod")
    local target resolved
    for target in $replace_targets; do
      # 只处理本地相对路径（以 . 或 / 开头）
      [[ "$target" =~ ^(\.|/) ]] || continue
      resolved="$dir/$target"
      [[ -d "$resolved" ]] || continue
      if find "$resolved" -name '*.go' -newer "$bin" -print -quit 2>/dev/null | grep -q .; then
        stale=1
        break
      fi
    done
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
    export REDIS_DB="${REDIS_DB:-0}"
    export STREAM_IOT_MONITOR="${STREAM_IOT_MONITOR:-iot:monitor:stream}"
    export STREAM_IOT_EVENT="${STREAM_IOT_EVENT:-iot:event:stream}"
    export CONSUMER_GROUP="${CONSUMER_GROUP:-wisefido-iot}"
    export CONSUMER_NAME="${CONSUMER_NAME:-wisefido-iot-1}"
    owlback_go_exec "$OWLBACK/wisefido-iot" "./cmd/wisefido-iot" "wisefido-iot" "$LOG_DIR/wisefido-iot.log"
    ;;
  wisefido-sensor)
    export REDIS_DB="${REDIS_DB:-0}"
    # sensor_v2 cutover 完成（2026-05-17 Phase C + PR-Bootstrap）：默认指 v2 (wisefido-sensor)；
    # 回退 v1 显式覆盖：SENSOR_DIR=wisefido-sensor-v1 systemctl restart owlback.sensor
    SENSOR_DIR="${SENSOR_DIR:-wisefido-sensor}"
    owlback_go_exec "$OWLBACK/$SENSOR_DIR" "./cmd/wisefido-sensor" "wisefido-sensor" "$LOG_DIR/wisefido-sensor.log"
    ;;
  *)
    echo "unknown module: $MODULE" >&2
    exit 1
    ;;
esac
