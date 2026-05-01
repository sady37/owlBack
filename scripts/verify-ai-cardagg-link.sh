#!/usr/bin/env bash
# 三期任务 B 第 1 项：AI emit + cardagg cache 链路实证。
#
# 目的：一次性 dump 4 个 tap 点的日志/wire，确认 wisefido-ai → iot:event:stream
# → wisefido-cardagg 的 track_verdict 链路是否完整（含 sandbox/release 模式诊断）。
#
# 4 个 tap 点：
#   ① wisefido-ai     ai_emit                   （AI 决策 emit + would_publish_to + published）
#   ② iot:event:stream                          （wire 实证：track_verdict 落地 redis）
#   ③ wisefido-cardagg ai_verdict_cached/cleared（cache Set 与 device 清理）
#   ④ wisefido-cardagg ai_verdict_applied       （monitor 流合并：mode + merged 标志）
#
# 用法：
#   ./verify-ai-cardagg-link.sh                # 最近 5 分钟
#   ./verify-ai-cardagg-link.sh '15 minutes ago'
#   ./verify-ai-cardagg-link.sh '2026-05-01 13:00:00'
#
# 期望：
#   sandbox 模式：① ② ③ 都有，④ 出现且 merged=false（fields 不被改写）
#   release 模式：① ② ③ ④ 都有，④ merged=true 时 monitor 流的 track_confidence
#                 应被 AI 值覆写（前端拿到低饱和 ghost UI）

set -u

SINCE="${1:-5 minutes ago}"
PASS="${REDIS_PASSWORD:-TeLunSu-36kr}"
HOST="${REDIS_HOST:-127.0.0.1}"
PORT="${REDIS_PORT:-6379}"

echo "================================================================"
echo "AI ↔ cardagg 链路实证  (since: ${SINCE})"
echo "================================================================"
echo

# 0. cardagg 启动配置确认（看 mode/ttl/gc——决定 ④ 期望行为）
echo "── [0] cardagg AI override 配置（启动日志）─────────────────────"
journalctl -t owlback-cardagg --since "${SINCE}" --no-pager 2>/dev/null \
  | grep -E "ai_override|AIOverride|track_verdict 收口|PR6" | tail -5
echo

# 1. AI 侧 emit
echo "── [1] wisefido-ai  ai_emit (category=track_verdict) ──────────"
ai_emit_count=$(journalctl -t owlback-ai --since "${SINCE}" --no-pager 2>/dev/null \
  | grep -c '"ai_emit"' || true)
journalctl -t owlback-ai --since "${SINCE}" --no-pager 2>/dev/null \
  | grep '"ai_emit"' | grep 'track_verdict' | tail -10
echo "  → emit 总数: ${ai_emit_count}"
echo

# 2. wire 实证：iot:event:stream 中的 track_verdict 消息
echo "── [2] iot:event:stream  最近 track_verdict 消息 (XREVRANGE) ──"
redis-cli -h "${HOST}" -p "${PORT}" -a "${PASS}" --no-auth-warning \
  XREVRANGE iot:event:stream + - COUNT 30 2>/dev/null \
  | awk '/track_verdict/,/^[0-9]+\)/' | head -40
echo

# 3. cardagg cache Set / Clear
echo "── [3] wisefido-cardagg  ai_verdict_cached / _cleared ─────────"
cached_count=$(journalctl -t owlback-cardagg --since "${SINCE}" --no-pager 2>/dev/null \
  | grep -c 'ai_verdict_cached' || true)
cleared_count=$(journalctl -t owlback-cardagg --since "${SINCE}" --no-pager 2>/dev/null \
  | grep -c 'ai_verdict_cleared' || true)
journalctl -t owlback-cardagg --since "${SINCE}" --no-pager 2>/dev/null \
  | grep -E 'ai_verdict_(cached|cleared)' | tail -10
echo "  → cached 总数: ${cached_count}    cleared 总数: ${cleared_count}"
echo

# 4. cardagg cache Apply（monitor 流合并）
echo "── [4] wisefido-cardagg  ai_verdict_applied (monitor 合并) ────"
applied_count=$(journalctl -t owlback-cardagg --since "${SINCE}" --no-pager 2>/dev/null \
  | grep -c 'ai_verdict_applied' || true)
merged_true=$(journalctl -t owlback-cardagg --since "${SINCE}" --no-pager 2>/dev/null \
  | grep 'ai_verdict_applied' | grep -c '"merged":true' || true)
journalctl -t owlback-cardagg --since "${SINCE}" --no-pager 2>/dev/null \
  | grep 'ai_verdict_applied' | tail -10
echo "  → applied 总数: ${applied_count}    其中 merged=true: ${merged_true}"
echo

# 5. 一致性总结
echo "================================================================"
echo "一致性诊断"
echo "================================================================"
echo "  AI emit (track_verdict): ${ai_emit_count}"
echo "  cardagg cached:          ${cached_count}"
echo "  cardagg applied:         ${applied_count}（merged=true: ${merged_true}）"
echo
if [ "${ai_emit_count}" -gt 0 ] && [ "${cached_count}" -eq 0 ]; then
  echo "  ⚠ AI emit 但 cardagg 未 cache —— stream 未落地 / event_handler case 未命中 / DeviceUID 空"
fi
if [ "${cached_count}" -gt 0 ] && [ "${applied_count}" -eq 0 ]; then
  echo "  ⚠ cache 已 Set 但无 Apply —— monitor 流可能未路由到此 device / track_id 无效（88）"
fi
if [ "${applied_count}" -gt 0 ] && [ "${merged_true}" -eq 0 ]; then
  echo "  ℹ apply 全为 merged=false —— 当前 sandbox 模式（fields 不动）"
  echo "    切 release：编辑 cardagg 配置 ai_override.mode=release 后 systemctl restart owlback.cardagg"
fi
if [ "${ai_emit_count}" -eq 0 ]; then
  echo "  ℹ 时间窗内无 AI emit —— 没有 ghost / track_verdict 触发；可重放 case fixture 或扩大时间窗"
fi
