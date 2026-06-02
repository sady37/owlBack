#!/usr/bin/env bash
# 批量导出标注测试 case → doc/cases/，并生成 oracle 条目（paste-ready）。
# 是 export_case_v2.sh 的 manifest 包装：把"测试时间窗 + 人工标签"一次性变成 fixture + 测试条目。
#
# 用法:
#   ./export_cases_manifest.sh <manifest_file>
#
# manifest 每行（'|' 分隔，# 开头注释，空行忽略）:
#   case_name | device_uid | start_local | end_local | tz | label | kind | note
#
#   case_name    doc/cases/ 下目录名（kebab-case，如 101-bath-ghost-0602）
#   device_uid   设备 UID（如 4D8710F41797）；脚本自查 addr
#   start_local  本地时刻 "YYYY-MM-DD HH:MM[:SS]"（按 tz 解释）
#   end_local    同上
#   tz           IANA 时区（America/Los_Angeles / America/Denver）
#   label        real | fp        —— 人工 ground-truth（real=真摔 must-fire / fp=假报 must-not-fire）
#   kind         ghost | lost | firmware-fall | bedside-fall —— 决定进哪个 oracle + 默认期望
#   note         自由文本（镜像/边缘静坐/床边 等，写进 label.json + 注释）
#
# 输出:
#   doc/cases/<case_name>/{room_layout.json, window.json, label.json}
#   末尾打印 TestReplayOracle / TestTrackLayerOracle 的 paste-ready Go 行（含 verify-note）。
#
# 时间：monitor_stream.ts 是 UTC；本脚本用 `TZ=<tz> date` 把本地时刻转 epoch-ms，无需手算。

set -euo pipefail

if [[ $# -lt 1 ]]; then sed -n '2,30p' "$0"; exit 1; fi
MANIFEST="$1"
[[ -f "$MANIFEST" ]] || { echo "ERROR: manifest 不存在: $MANIFEST" >&2; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
EXPORT="$SCRIPT_DIR/export_case_v2.sh"
[[ -x "$EXPORT" ]] || { echo "ERROR: 缺 $EXPORT（或不可执行）" >&2; exit 1; }
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

to_ms() { # <local> <tz>  -> epoch ms
  TZ="$2" date -d "$1" +%s%3N 2>/dev/null || { echo "ERROR: 时间解析失败: '$1' (tz=$2)" >&2; exit 3; }
}

REPLAY_LINES=()
TRACK_LINES=()

while IFS='|' read -r case_name uid start_local end_local tz label kind note || [[ -n "$case_name" ]]; do
  # trim
  case_name="$(echo "$case_name" | xargs)"
  [[ -z "$case_name" || "$case_name" == \#* ]] && continue
  uid="$(echo "$uid" | xargs)"; start_local="$(echo "$start_local" | xargs)"
  end_local="$(echo "$end_local" | xargs)"; tz="$(echo "$tz" | xargs)"
  label="$(echo "$label" | xargs)"; kind="$(echo "$kind" | xargs)"; note="$(echo "${note:-}" | xargs)"

  start_ms="$(to_ms "$start_local" "$tz")"
  end_ms="$(to_ms "$end_local" "$tz")"

  echo "=============================================================="
  echo ">> $case_name  [$label/$kind]  $start_local..$end_local ($tz)"
  "$EXPORT" "$uid" "$start_ms" "$end_ms" "$case_name"

  out="$ROOT_DIR/doc/cases/$case_name"
  cat > "$out/label.json" <<EOF
{
  "case_name": "$case_name",
  "device_uid": "$uid",
  "start_local": "$start_local",
  "end_local": "$end_local",
  "tz": "$tz",
  "start_ms": $start_ms,
  "end_ms": $end_ms,
  "label": "$label",
  "kind": "$kind",
  "note": "$note"
}
EOF

  # ---- oracle 条目（默认值 + verify-note；assert 由人确认后开）----
  win="$case_name/window.json"; lay="$case_name/room_layout.json"
  if [[ "$label" == "real" ]]; then want_confirm="true"; else want_confirm="false"; fi

  # Room 层 TestReplayOracle：wantConfirm = (label==real)。
  # firmware-fall / bedside-fall 是新类（模型未标定）→ 默认 assert=false 诊断，标定后再开。
  case "$kind" in
    firmware-fall|bedside-fall) replay_assert="false"; rnote="新类(Room 层 firmware/near-bed geom 未标定)→先诊断" ;;
    *)                          replay_assert="true";  rnote="P1 类" ;;
  esac
  REPLAY_LINES+=("		{\"$case_name\", \"$win\", \"$lay\", $want_confirm, $replay_assert}, // $label/$kind $note —— $rnote")

  # Track 层 TestTrackLayerOracle：仅 ghost/lost 类有意义。wantLost 需看 diagnostic 后定 → assert=false。
  case "$kind" in
    ghost)  TRACK_LINES+=("		{\"$case_name\", \"$win\", \"$lay\", false, 0.2, false}, // ghost $note —— 期望 →None(not-Lost)，验证后 assert=true") ;;
    lost)   TRACK_LINES+=("		{\"$case_name\", \"$win\", \"$lay\", false, 0.2, false}, // lost $note —— 走出?返回?→None / 走动消失→Lost；看 maxTLost 后定 wantLost+assert") ;;
  esac

done < "$MANIFEST"

echo ""
echo "############################################################"
echo "# paste-ready oracle 条目（确认 assert 前先跑 -v 看诊断值）"
echo "############################################################"
echo ""
echo "### TestReplayOracle (Room 层 confirm) ###"
printf '%s\n' "${REPLAY_LINES[@]}"
echo ""
echo "### TestTrackLayerOracle (Track 层 maxTLost) ###"
if [[ ${#TRACK_LINES[@]} -gt 0 ]]; then printf '%s\n' "${TRACK_LINES[@]}"; else echo "  (无 ghost/lost 类)"; fi
echo ""
echo "Done. fixture 在 doc/cases/；label.json 已写。下一步：跑 oracle -v 看诊断 → 定 assert。"
