#!/bin/bash
# E2E test for Family role scope enforcement.
# Verifies a Family user cannot see / modify residents they are not linked to.
#
# Usage: BASE_URL=http://localhost:8080 ./scripts/e2e_family_scope.sh

set -e
BASE_URL="${BASE_URL:-http://localhost:8080}"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; FAILED=1; }
info() { echo -e "${YELLOW}• $1${NC}"; }

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing dep: $1"; exit 1; }; }
need curl; need jq

# Family user 测试账号；在 owl_v2 测试库 f01 绑定到 fd00:0:3:ff01:a:: (resident 'Ton')
FAM_ACCOUNT="${FAM_ACCOUNT:-f01}"
FAM_PASSWORD="${FAM_PASSWORD:-Ts123@123}"
LINKED_HOA="${LINKED_HOA:-fd00:0:3:ff01:a::}"
OTHER_HOA="${OTHER_HOA:-fd00:0:3:ff01:1::}"

info "1) Login as Family user $FAM_ACCOUNT"
PASSHASH=$(printf '%s' "$FAM_PASSWORD" | sha256sum | cut -d' ' -f1)
LOGIN=$(curl -sS -X POST "$BASE_URL/auth/api/v2/login" \
  -H 'Content-Type: application/json' \
  -d "$(jq -nc --arg u "$FAM_ACCOUNT" --arg p "$PASSHASH" '{user_account:$u, password_hash:$p}')")
CODE=$(jq -r '.code // empty' <<<"$LOGIN")
if [[ "$CODE" != "2000" ]]; then
  echo "FAMILY LOGIN FAILED — check FAM_PASSWORD env var or seed data: $LOGIN"
  exit 1
fi
TOKEN=$(jq -r '.result.accessToken // .result.access_token' <<<"$LOGIN")
ROLE=$(jq -r '.result.role // empty' <<<"$LOGIN")
USER_ID=$(jq -r '.result.user_id // empty' <<<"$LOGIN")
pass "login ok; role=$ROLE user_id=$USER_ID"

[[ "$ROLE" == "Family" ]] || fail "expected role=Family got '$ROLE'"

AUTH=(-H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json')

info "2) List residents — should only see linked resident(s)"
LIST=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents?size=50")
TOTAL=$(jq -r '.result.total' <<<"$LIST")
ITEM_HOAS=$(jq -r '.result.items[].resident_id' <<<"$LIST" | sort -u)
echo "  total=$TOTAL"
echo "  hoas:"; echo "$ITEM_HOAS" | sed 's/^/    /'

if echo "$ITEM_HOAS" | grep -q "^$LINKED_HOA$"; then
  pass "linked resident $LINKED_HOA visible"
else
  fail "linked resident $LINKED_HOA NOT in list — scope query bug"
fi
if echo "$ITEM_HOAS" | grep -q "^$OTHER_HOA$"; then
  fail "OTHER resident $OTHER_HOA visible — Family scope leaked!"
else
  pass "other resident $OTHER_HOA filtered out"
fi

info "3) GET linked resident — should succeed"
LINKED_ENC=$(printf '%s' "$LINKED_HOA" | jq -sRr @uri)
GET_OK=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$LINKED_ENC")
[[ "$(jq -r '.code' <<<"$GET_OK")" == "2000" ]] \
  && pass "GET linked $LINKED_HOA ok" \
  || fail "GET linked $LINKED_HOA failed: $(jq -r '.message' <<<"$GET_OK")"

info "4) GET other resident — should be denied"
OTHER_ENC=$(printf '%s' "$OTHER_HOA" | jq -sRr @uri)
GET_OTHER=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$OTHER_ENC")
CODE=$(jq -r '.code' <<<"$GET_OTHER")
MSG=$(jq -r '.message // ""' <<<"$GET_OTHER")
if [[ "$CODE" != "2000" ]] && echo "$MSG" | grep -qi 'permission'; then
  pass "GET other $OTHER_HOA denied: $MSG"
else
  fail "GET other $OTHER_HOA returned $CODE / $MSG — scope bypass!"
fi

info "5) PUT linked resident — should be denied (Family cannot edit)"
PUT_LINK=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$LINKED_ENC" \
  -d '{"nickname":"hacked"}')
CODE=$(jq -r '.code' <<<"$PUT_LINK")
MSG=$(jq -r '.message // ""' <<<"$PUT_LINK")
if [[ "$CODE" != "2000" ]] && echo "$MSG" | grep -qi 'permission'; then
  pass "PUT linked denied: $MSG"
else
  fail "PUT linked returned $CODE / $MSG — Family edit not blocked!"
fi

info "6) PUT other resident — should be denied"
PUT_OTHER=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$OTHER_ENC" \
  -d '{"nickname":"hacked"}')
CODE=$(jq -r '.code' <<<"$PUT_OTHER")
MSG=$(jq -r '.message // ""' <<<"$PUT_OTHER")
if [[ "$CODE" != "2000" ]] && echo "$MSG" | grep -qi 'permission'; then
  pass "PUT other denied: $MSG"
else
  fail "PUT other returned $CODE / $MSG — Family edit not blocked!"
fi

echo ""
if [[ "${FAILED:-0}" == "1" ]]; then
  echo -e "${RED}FAMILY SCOPE TEST FAILED${NC}"
  exit 1
fi
echo -e "${GREEN}FAMILY SCOPE TEST PASSED${NC}"
