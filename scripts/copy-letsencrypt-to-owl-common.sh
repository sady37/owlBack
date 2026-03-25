#!/usr/bin/env bash
# 将 Certbot 签发的证书覆盖到 owl-common，供 qinglan（等设备直连 HTTPS）使用。
# 用法（在已跑过 certbot 的服务器上）：
#   sudo LE_DOMAIN=test.wisefido.com bash owlBack/scripts/copy-letsencrypt-to-owl-common.sh
# 可选：LE_ROOT=/etc/letsencrypt/live  LE_COPY_DEST=/path/to/owl-common

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" && pwd)"
OWL_COMMON="$(cd "$SCRIPT_DIR/../owl-common" && pwd)"
DOMAIN="${LE_DOMAIN:-test.wisefido.com}"
LE_ROOT="${LE_ROOT:-/etc/letsencrypt/live}"
SRC="$LE_ROOT/$DOMAIN"
DEST="${LE_COPY_DEST:-$OWL_COMMON}"
KEY_GROUP="${LE_KEY_GROUP:-wisefido}"

CHAIN="$SRC/fullchain.pem"
KEY="$SRC/privkey.pem"

if [ ! -r "$CHAIN" ] || [ ! -r "$KEY" ]; then
	echo "无法读取：$CHAIN 或 $KEY" >&2
	echo "请先 certbot 申请 ${DOMAIN}，或用 sudo 执行。" >&2
	exit 1
fi

install -d -m 0755 "$DEST"
# live/ 下为指向 archive/ 的相对符号链接，须解引用否则复制到 owl-common 会断链
cp -fL "$CHAIN" "$DEST/server.crt"
cp -fL "$KEY" "$DEST/server.key"
chmod 0644 "$DEST/server.crt"
if getent group "$KEY_GROUP" >/dev/null 2>&1; then
	chgrp "$KEY_GROUP" "$DEST/server.key"
	chmod 0640 "$DEST/server.key"
else
	# 无目标组时回退到 0644，避免服务用户无法读取 key
	chmod 0644 "$DEST/server.key"
fi
echo "已写入：$DEST/server.crt 与 $DEST/server.key（来自 $SRC）"
