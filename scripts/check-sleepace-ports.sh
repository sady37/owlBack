#!/usr/bin/env bash
# 本地端口验证：wisefido-data -> wisefido-sleepace(:8083) -> sleepace-service(:8090)
set -e
echo "=== Sleepace 链路端口检查 ==="
echo ""

check_http() {
  local name="$1"
  local url="$2"
  if curl -sS --connect-timeout 2 -o /dev/null -w "%{http_code}" "$url" 2>/dev/null | grep -q '200\|301\|302'; then
    echo "  OK  $name  $url"
    return 0
  fi
  echo "  FAIL $name  $url"
  return 1
}

check_port() {
  local name="$1"
  local host="$2"
  local port="$3"
  if command -v nc >/dev/null 2>&1; then
    if nc -z -w2 "$host" "$port" 2>/dev/null; then
      echo "  OK  $name  $host:$port"
      return 0
    fi
  else
    if curl -sS --connect-timeout 2 -o /dev/null "http://${host}:${port}/" 2>/dev/null; then
      echo "  OK  $name  $host:$port"
      return 0
    fi
  fi
  echo "  FAIL $name  $host:$port"
  return 1
}

err=0
check_http "wisefido-sleepace (网关)" "http://127.0.0.1:8083/health" || err=1
check_port "sleepace-service (Docker)" "127.0.0.1" "8090" || err=1

echo ""
if [ $err -eq 0 ]; then
  echo "端口正常。若仍报 user not found，请确认 sleepace-service 内该 device_code 已 bind。"
else
  echo "有端口不可达：请先启动 wisefido-sleepace (本机) 和 sleepace-service (docker-compose up sleepace-service)。"
fi
exit $err
