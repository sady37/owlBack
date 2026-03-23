#!/bin/bash
# owlBack/docker-compose.yml 全服务：先停再起（数据卷保留）
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

docker_compose() {
  if docker compose version >/dev/null 2>&1; then
    docker compose "$@"
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose "$@"
  else
    echo "Error: 需要 docker compose 插件或 docker-compose 命令" >&2
    return 1
  fi
}

case "${1:-}" in
  down|stop)
    cd "$SCRIPT_DIR" && docker_compose down
    ;;
  up|start)
    cd "$SCRIPT_DIR" && docker_compose up -d --remove-orphans
    ;;
  cycle|restart)
    cd "$SCRIPT_DIR" && docker_compose down && docker_compose up -d --remove-orphans
    ;;
  ps|status)
    cd "$SCRIPT_DIR" && docker_compose ps -a
    ;;
  *)
    echo "Usage: $0 {down|up|cycle|status}"
    echo "  down   — docker compose down（停并删容器，卷保留）"
    echo "  up     — docker compose up -d（起全部：pgsql/redis/mqtt/mysql）"
    echo "  cycle  — 先 down 再 up（整体关掉再起）"
    exit 1
    ;;
esac
