#!/usr/bin/env bash
# FCM 部署验证辅助：设备侧测试期间反复运行，观察注册/回收/推送日志。
# 用法: ./scripts/fcm_verify.sh [db|log|tail]
#   db   — 查 apns_devices 里 android 行（步骤3注册 / 步骤5回收 后各查一次）
#   log  — grep 最近的 [FCM] 日志（推送发出 / token回收 / 失败）
#   tail — 实时跟踪 [FCM] 日志（触发真告警时开着看到达）
set -euo pipefail
LOG=/home/wisefido/owl/log/wisefido-data.log
export PGPASSWORD="${DB_PASSWORD:-postgres}"
PSQL=(psql -h "${DB_HOST:-127.0.0.1}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "${DB_NAME:-owl_v2}")

case "${1:-db}" in
  db)
    "${PSQL[@]}" -c "SELECT left(push_token,8) AS token8, platform, is_active, environment,
                            to_char(last_seen_at,'MM-DD HH24:MI') AS last_seen,
                            to_char(updated_at,'MM-DD HH24:MI')   AS updated
                     FROM apns_devices WHERE platform='android'
                     ORDER BY updated_at DESC;"
    ;;
  log)
    grep -E "\[FCM\]" "$LOG" | tail -30
    ;;
  tail)
    tail -f "$LOG" | grep --line-buffered -E "\[FCM\]|\[APNS\]"
    ;;
  *)
    echo "用法: $0 [db|log|tail]" >&2; exit 1 ;;
esac
