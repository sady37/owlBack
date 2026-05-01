#!/usr/bin/env bash
# 运行时切换 wisefido-cardagg AI override 模式
#
# 用法：
#   ./switch-cardagg-mode.sh sandbox
#   ./switch-cardagg-mode.sh release
#   ./switch-cardagg-mode.sh status
#
# 说明：
# - sandbox: 向 cardagg 进程发送 SIGUSR1
# - release: 向 cardagg 进程发送 SIGUSR2
# - status:  读取最近日志推断当前模式

set -u

ACTION="${1:-status}"

find_cardagg_pid() {
  local pid=""

  # 与 scripts/ServiceStatus.sh 保持一致的多模式匹配，避免漏检 go run 场景
  pid=$(pgrep -f "wisefido-cardagg/main.go" 2>/dev/null | head -1)
  [[ -z "$pid" ]] && pid=$(pgrep -f "wisefido-cardagg" 2>/dev/null | head -1)
  [[ -z "$pid" ]] && pid=$(pgrep -f "go run main.go" 2>/dev/null | head -1)

  echo "$pid"
}

print_last_mode_from_journal() {
  local line=""
  line=$(journalctl -t owlback-cardagg --since "24 hours ago" --no-pager 2>/dev/null \
    | grep -E "ai_override cache initialized|ai_override mode switched" \
    | tail -1)

  if [[ -z "$line" ]]; then
    echo "UNKNOWN (no ai_override mode logs in last 24h)"
    return 0
  fi

  if echo "$line" | grep -q '"mode":"release"'; then
    echo "release"
    return 0
  fi
  if echo "$line" | grep -q '"mode":"sandbox"'; then
    echo "sandbox"
    return 0
  fi

  echo "UNKNOWN"
}

send_mode_signal() {
  local target_mode="$1"
  local sig="$2"
  local pid

  pid=$(find_cardagg_pid)
  if [[ -z "$pid" ]]; then
    echo "[ERROR] wisefido-cardagg process not found"
    echo "Hint: run ./scripts/ServiceStatus.sh first"
    return 1
  fi

  if ! kill -"$sig" "$pid" 2>/dev/null; then
    echo "[ERROR] failed to send SIG${sig} to pid=${pid}"
    return 1
  fi

  echo "[OK] sent SIG${sig} to wisefido-cardagg pid=${pid} (target=${target_mode})"
  echo "[INFO] last mode from journal: $(print_last_mode_from_journal)"
  echo "[INFO] tail logs: journalctl -t owlback-cardagg -n 30 --no-pager"
  return 0
}

case "$ACTION" in
  sandbox)
    send_mode_signal "sandbox" "USR1"
    ;;
  release)
    send_mode_signal "release" "USR2"
    ;;
  status)
    pid=$(find_cardagg_pid)
    if [[ -z "$pid" ]]; then
      echo "[WARN] wisefido-cardagg process not found"
    else
      echo "[OK] wisefido-cardagg pid=${pid}"
    fi
    echo "[INFO] mode from journal: $(print_last_mode_from_journal)"
    ;;
  *)
    echo "Usage: $0 {sandbox|release|status}"
    exit 2
    ;;
esac
