#!/usr/bin/env bash
# owl-kms-run.sh — systemd-friendly wrapper for owl-kms --recover.
#
# 为什么有这个脚本：
#   - KMS 没有 systemd auto-start（init 错误使用会重新生成 master key，毁掉 PHI 解密）；
#   - 每次 reboot 必须手动跑一次 recover，原始命令参数多、敏感；
#   - 此脚本把"挑最新 archive + 喂 token"封装好，用户只需 `sudo systemctl start owl-kms`。
#
# 安全边界：
#   - secrets 文件 chmod 600，限 wisefido 用户。
#   - --init 已硬编码拒绝。
#   - WARNING: owl-kms CLI 参数中包含 token；ps/proc 仍可见。属已知妥协。
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OWLBACK="$(cd "$SCRIPT_DIR/../.." && pwd)"
KMS_DIR="$OWLBACK/kms"
LOG_DIR="${LOG_DIR:-$(cd "$OWLBACK/.." && pwd)/log}"
SECRETS="$KMS_DIR/.kms-secrets"

mkdir -p "$LOG_DIR"

if [[ ! -f "$SECRETS" ]]; then
  echo "ERROR: secrets file missing: $SECRETS" >&2
  echo "expected lines: MW_TOKEN=... / MASTER_PIN=..." >&2
  exit 1
fi
# 严防误配 0644
perm=$(stat -c '%a' "$SECRETS")
if [[ "$perm" != "600" ]]; then
  echo "ERROR: $SECRETS must be chmod 600 (currently $perm)" >&2
  exit 1
fi
# shellcheck source=/dev/null
source "$SECRETS"
: "${MW_TOKEN:?MW_TOKEN required in $SECRETS}"
: "${MASTER_PIN:?MASTER_PIN required in $SECRETS}"

# 选最新的 archive（mtime desc）
ARCHIVE=$(ls -1t "$KMS_DIR"/master_key-*.json 2>/dev/null | head -1)
if [[ -z "$ARCHIVE" ]]; then
  echo "ERROR: no master_key-*.json archive found in $KMS_DIR" >&2
  exit 1
fi

# 硬性拒绝 --init / -init（用户改 unit 时的护栏）
for arg in "$@"; do
  case "$arg" in
    --init|-init)
      echo "ERROR: --init is forbidden via this wrapper (would re-generate master key)" >&2
      exit 1
      ;;
  esac
done

cd "$OWLBACK"
echo "owl-kms recover: archive=$(basename "$ARCHIVE")"
exec "$KMS_DIR/owl-kms" \
  --recover \
  --archive "$ARCHIVE" \
  --vault-dir "$KMS_DIR" \
  --mw-token "$MW_TOKEN" \
  --master-pin "$MASTER_PIN"
