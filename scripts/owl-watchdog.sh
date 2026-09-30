#!/bin/bash
# owl-watchdog — 黑盒探活 + 自愈 + 状态机日志
#
# 缘起 2026-09-30：nginx 自 09-25 15:49 停机 4 天无人知晓，靠人工发现网页打不开。
# 同期 certbot 续期连挂 3 张证书、ota-ql 容器 PID 耗尽，全部零告警。
# 这些故障系统本已记录(systemd inactive / Failed、docker unhealthy)，缺的是送达。
#
# 设计要点：
#   ① 自愈优先于告警——凌晨三点没人看告警，先拉起能把 4 天缩成一个探测周期。
#   ② 状态机，只在"翻转"时记一条——acme.sh 那次 4 天里失败几十次，
#      每次都记会把日志刷爆、真信号被淹没。
#   ③ 自愈节流——避免对着一个起不来的服务无限重启。
#   ④ notify() 目前只落日志，是预留的通知出口；以后接
#      钉钉/Telegram/webhook 只改这一个函数，其余逻辑不动。
set -uo pipefail

LOG=/var/log/owl-watchdog.log
STATE_DIR=/var/lib/owl-watchdog
MAX_LOG_BYTES=$((5 * 1024 * 1024))   # 5MB 后自截半，防撑爆磁盘
HEAL_WINDOW=600                       # 自愈节流窗口(秒)
HEAL_MAX=2                            # 窗口内同一目标最多自愈次数
CURL_TIMEOUT=10
CERT_WARN_DAYS=14

mkdir -p "$STATE_DIR"

# ── 通知出口（当前仅落日志）───────────────────────────────────────────
# level: OK | WARN | CRIT | HEAL
notify() {
  local level="$1" target="$2" msg="$3"
  printf '%s [%-4s] %-28s %s\n' "$(date '+%F %T')" "$level" "$target" "$msg" >> "$LOG"
  # TODO 接通道时在此处补一行，例如：
  #   [ "$level" != OK ] && curl -s -m 10 -X POST "$WEBHOOK" -d "..." >/dev/null
}

rotate_log() {
  [ -f "$LOG" ] || return 0
  local sz; sz=$(stat -c %s "$LOG" 2>/dev/null || echo 0)
  if [ "$sz" -gt "$MAX_LOG_BYTES" ]; then
    tail -c $((MAX_LOG_BYTES / 2)) "$LOG" > "$LOG.tmp" && mv "$LOG.tmp" "$LOG"
    notify OK watchdog "日志超过 ${MAX_LOG_BYTES}B，已截断保留后半"
  fi
}

# ── 状态机：只在状态翻转时告警 ────────────────────────────────────────
# 用法: transition <target> <up|down> <描述>
transition() {
  local target="$1" now="$2" desc="$3"
  local f="$STATE_DIR/$(echo "$target" | tr '/:' '__').state"
  local prev; prev=$(cat "$f" 2>/dev/null || echo "unknown")
  echo "$now" > "$f"
  [ "$prev" = "$now" ] && return 1        # 状态未变 → 不告警
  case "$now" in
    down) notify CRIT "$target" "故障: $desc" ;;
    up)   [ "$prev" = "down" ] && notify OK "$target" "已恢复: $desc" ;;
  esac
  return 0
}

# ── 自愈节流 ──────────────────────────────────────────────────────────
can_heal() {
  local target="$1"
  local f="$STATE_DIR/$(echo "$target" | tr '/:' '__').heal"
  local now; now=$(date +%s)
  local kept=""; local n=0
  if [ -f "$f" ]; then
    while read -r ts; do
      [ -z "$ts" ] && continue
      if [ $((now - ts)) -lt "$HEAL_WINDOW" ]; then kept+="$ts"$'\n'; n=$((n+1)); fi
    done < "$f"
  fi
  [ "$n" -ge "$HEAL_MAX" ] && { printf '%s' "$kept" > "$f"; return 1; }
  printf '%s%s\n' "$kept" "$now" > "$f"
  return 0
}

# ── 检查项 ────────────────────────────────────────────────────────────

# HTTP 探活（用户视角：nginx / 后端 / 证书 任一环坏都会失败）
check_http() {
  local name="$1" url="$2"
  local code; code=$(curl -s -o /dev/null -m "$CURL_TIMEOUT" -w '%{http_code}' "$url" 2>/dev/null)
  if [ "$code" = "200" ] || [ "$code" = "302" ]; then
    transition "http:$name" up "HTTP $code"
  else
    transition "http:$name" down "HTTP ${code:-000}"
    heal_nginx "$name"
  fi
}

# nginx 自愈：HTTP 全线失败多半是 nginx 本身没了（09-25 即如此）
heal_nginx() {
  local trigger="$1"
  systemctl is-active --quiet nginx && return 0   # nginx 活着 → 问题在上游，不乱动
  can_heal nginx || { notify WARN nginx "自愈已达节流上限(${HEAL_MAX}次/${HEAL_WINDOW}s)，不再重试"; return 1; }
  if nginx -t >/dev/null 2>&1; then
    systemctl start nginx && notify HEAL nginx "检测到 inactive(由 $trigger 触发)，已自动拉起"
  else
    notify CRIT nginx "inactive 且配置校验失败，不敢自动拉起——需人工介入"
  fi
}

# systemd 单元
check_unit() {
  local unit="$1" auto_heal="${2:-yes}"
  if systemctl is-active --quiet "$unit"; then
    transition "unit:$unit" up "active"
  else
    transition "unit:$unit" down "$(systemctl is-active "$unit" 2>&1)"
    [ "$auto_heal" = yes ] || return 0
    can_heal "$unit" || { notify WARN "$unit" "自愈已达节流上限"; return 1; }
    systemctl start "$unit" 2>/dev/null && notify HEAL "$unit" "已自动拉起"
  fi
}

# docker 容器（exited → 自动拉起；unhealthy → 只告警，可能有状态不宜盲重启）
check_container() {
  local name="$1"
  local st; st=$(docker inspect -f '{{.State.Status}}' "$name" 2>/dev/null)
  local hl; hl=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}none{{end}}' "$name" 2>/dev/null)
  if [ -z "$st" ]; then transition "ctr:$name" down "容器不存在"; return; fi
  if [ "$st" != "running" ]; then
    transition "ctr:$name" down "status=$st"
    can_heal "ctr:$name" || return 1
    docker start "$name" >/dev/null 2>&1 && notify HEAL "ctr:$name" "status=$st，已自动 start"
  elif [ "$hl" = "unhealthy" ]; then
    # ota-ql 那次即此类：进程在、端口通，但 healthcheck 持续失败(PID 耗尽)
    transition "ctr:$name" down "running 但 unhealthy — 不自动重启，需人工判断"
  else
    transition "ctr:$name" up "running${hl:+/$hl}"
  fi
}

# 证书到期预警
check_cert() {
  local name="$1" f="/etc/letsencrypt/live/$1/cert.pem"
  [ -f "$f" ] || { transition "cert:$name" down "证书文件不存在"; return; }
  local end; end=$(openssl x509 -in "$f" -noout -enddate 2>/dev/null | cut -d= -f2)
  local days=$(( ( $(date -d "$end" +%s) - $(date +%s) ) / 86400 ))
  if [ "$days" -lt "$CERT_WARN_DAYS" ]; then
    transition "cert:$name" down "仅剩 ${days} 天到期"
  else
    transition "cert:$name" up "剩余 ${days} 天"
  fi
}

# ── 主流程 ────────────────────────────────────────────────────────────
rotate_log

check_http pilot https://pilot.wisefido.com/
check_http www   https://www.wisefido.com/
check_http ota   https://ota.wisefido.com/
check_http demo  https://demo.wisefido.com/

check_unit nginx
check_unit owlback
check_unit owlback.qinglan
check_unit owlfront

for c in ota-ql owl-mqtt owl-redis owl-postgresql; do
  # docker-compose 重建过的容器名会带前缀哈希(如 a2437c59f1e9_owl-mqtt)，
  # 故按"结尾匹配"取实际名；锚定 $ 可排除 ota-ql-broken-* 这类保留的旧容器。
  real=$(docker ps -a --format '{{.Names}}' 2>/dev/null | grep -E "(^|_)${c}$" | head -1)
  if [ -n "$real" ]; then check_container "$real"
  else notify WARN "ctr:$c" "未找到该容器——名字变了或已被删除"; fi
done

for cert in app.wisefido.com wisefido.com demo.wisefido.com ota.wisefido.com; do
  check_cert "$cert"
done

exit 0
