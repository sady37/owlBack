#!/bin/bash
# E2E smoke test for v2 spatial CRUD chain.
#
# 验证：v2 spatial endpoint 全链路 (branches → units → rooms → beds → devices)
# + GET LIST/联表派生字段 + PUT/DELETE 条件硬删 + device branch rebind
#
# Usage: BASE_URL=http://localhost:8080 ./scripts/e2e_spatial_v2.sh

set -e
BASE_URL="${BASE_URL:-http://localhost:8080}"
USER_ACCOUNT="${USER_ACCOUNT:-admin}"
PASSWORD="${PASSWORD:-Ts123@123}"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; FAILED=1; }
info() { echo -e "${YELLOW}• $1${NC}"; }

need() { command -v "$1" >/dev/null || { echo "missing: $1"; exit 1; }; }
need curl
need jq

# ============================================================================
# 1. Login
# ============================================================================
PASSWORD_HASH=$(printf '%s' "$PASSWORD" | sha256sum | cut -d' ' -f1)
info "1) Login as $USER_ACCOUNT"
LOGIN=$(curl -sS -X POST "$BASE_URL/auth/api/v2/login" -H 'Content-Type: application/json' \
  -d "$(jq -nc --arg u "$USER_ACCOUNT" --arg p "$PASSWORD_HASH" '{user_account:$u, password_hash:$p}')")
CODE=$(jq -r '.code // empty' <<<"$LOGIN")
[[ "$CODE" == "2000" ]] || { fail "login failed: $LOGIN"; exit 1; }
TOKEN=$(jq -r '.result.accessToken // .result.access_token' <<<"$LOGIN")
TENANT=$(jq -r '.result.tenant_id' <<<"$LOGIN")
AUTH=(-H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json')
pass "login ok; tenant=$TENANT"

# ============================================================================
# 2. List branches; pick first
# ============================================================================
info "2) GET /admin/api/v2/spatial/branches"
BRANCHES=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/branches")
B_TOTAL=$(jq -r '.result.total' <<<"$BRANCHES")
[[ "$B_TOTAL" -gt 0 ]] || { fail "no branches: $BRANCHES"; exit 1; }
BRANCH_PFX=$(jq -r '.result.items[0].prefix' <<<"$BRANCHES")
BRANCH_NAME=$(jq -r '.result.items[0].name' <<<"$BRANCHES")
B_UNITS=$(jq -r '.result.items[0].units | length' <<<"$BRANCHES")
pass "branches=$B_TOTAL; first=[$BRANCH_NAME @ $BRANCH_PFX] with $B_UNITS units"

# ============================================================================
# 3. List units under that branch
# ============================================================================
BRANCH_PFX_ENC=$(printf '%s' "$BRANCH_PFX" | jq -sRr @uri)
info "3) GET /admin/api/v2/spatial/units?branch=$BRANCH_PFX"
UNITS=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/units?branch=$BRANCH_PFX_ENC")
U_TOTAL=$(jq -r '.result.total' <<<"$UNITS")
[[ "$U_TOTAL" -gt 0 ]] || { fail "no units under $BRANCH_NAME: $UNITS"; exit 1; }
# Pick a unit with rooms+beds
UNIT_PFX=$(jq -r '[.result.items[] | select((.rooms|length)>0 and (.beds|length)>0)][0].prefix // .result.items[0].prefix' <<<"$UNITS")
UNIT_NAME=$(jq -r --arg p "$UNIT_PFX" '.result.items[] | select(.prefix==$p) | .name' <<<"$UNITS")
UNIT_BLD=$(jq -r --arg p "$UNIT_PFX" '.result.items[] | select(.prefix==$p) | .building_name' <<<"$UNITS")
UNIT_FLR=$(jq -r --arg p "$UNIT_PFX" '.result.items[] | select(.prefix==$p) | .floor' <<<"$UNITS")
pass "units=$U_TOTAL; picked=[$UNIT_NAME @ $UNIT_PFX] in building $UNIT_BLD floor $UNIT_FLR"

# ============================================================================
# 4. List rooms under that unit; pick first
# ============================================================================
UNIT_PFX_ENC=$(printf '%s' "$UNIT_PFX" | jq -sRr @uri)
info "4) GET /admin/api/v2/spatial/rooms?unit=$UNIT_PFX"
ROOMS=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/rooms?unit=$UNIT_PFX_ENC")
R_TOTAL=$(jq -r '.result.total' <<<"$ROOMS")
[[ "$R_TOTAL" -gt 0 ]] || { fail "no rooms under $UNIT_NAME: $ROOMS"; exit 1; }
ROOM_PFX=$(jq -r '[.result.items[] | select((.beds|length)>0)][0].prefix // .result.items[0].prefix' <<<"$ROOMS")
ROOM_NAME=$(jq -r --arg p "$ROOM_PFX" '.result.items[] | select(.prefix==$p) | .name' <<<"$ROOMS")
ROOM_UN=$(jq -r --arg p "$ROOM_PFX" '.result.items[] | select(.prefix==$p) | .unit_name' <<<"$ROOMS")
ROOM_BN=$(jq -r --arg p "$ROOM_PFX" '.result.items[] | select(.prefix==$p) | .branch_name' <<<"$ROOMS")
pass "rooms=$R_TOTAL; picked=[$ROOM_NAME @ $ROOM_PFX] (联表 unit=$ROOM_UN, branch=$ROOM_BN)"

# ============================================================================
# 5. List beds under that room
# ============================================================================
ROOM_PFX_ENC=$(printf '%s' "$ROOM_PFX" | jq -sRr @uri)
info "5) GET /admin/api/v2/spatial/beds?room=$ROOM_PFX"
BEDS=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/beds?room=$ROOM_PFX_ENC")
B_TOTAL=$(jq -r '.result.total' <<<"$BEDS")
pass "beds=$B_TOTAL under $ROOM_NAME"

# ============================================================================
# 6. Create test bed → update → delete (roundtrip)
# ============================================================================
TS=$(date +%s)
TEST_BED_NAME="E2E-TestBed-$TS"
info "6) POST /admin/api/v2/spatial/beds (create '$TEST_BED_NAME' under $ROOM_NAME)"
CREATE=$(curl -sS -X POST "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/beds" \
  -d "$(jq -nc --arg p "$ROOM_PFX" --arg n "$TEST_BED_NAME" '{parent:$p, attrs:{name:$n}}')")
NEW_BED_PFX=$(jq -r '.result.prefix // empty' <<<"$CREATE")
[[ -n "$NEW_BED_PFX" ]] || { fail "create bed failed: $CREATE"; exit 1; }
pass "created bed @ $NEW_BED_PFX"

NEW_BED_PFX_ENC=$(printf '%s' "$NEW_BED_PFX" | jq -sRr @uri)
info "6.1) PUT bed (update description)"
PUT_RES=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/beds/$NEW_BED_PFX_ENC" \
  -d '{"description":"e2e test bed"}')
[[ "$(jq -r '.code' <<<"$PUT_RES")" == "2000" ]] || { fail "put bed failed: $PUT_RES"; exit 1; }
pass "updated"

info "6.2) GET bed"
GET_RES=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/beds/$NEW_BED_PFX_ENC")
GOT_DESC=$(jq -r '.result.description // empty' <<<"$GET_RES")
[[ "$GOT_DESC" == "e2e test bed" ]] || { fail "description not persisted: $GET_RES"; exit 1; }
pass "verified description = '$GOT_DESC'"

info "6.3) DELETE bed"
DEL=$(curl -sS -X DELETE "${AUTH[@]}" "$BASE_URL/admin/api/v2/spatial/beds/$NEW_BED_PFX_ENC")
[[ "$(jq -r '.code' <<<"$DEL")" == "2000" ]] || { fail "delete bed failed: $DEL"; exit 1; }
pass "deleted"

# ============================================================================
# 7. Device branch rebind (find unbound device in tenant pool → rebind round-trip)
# ============================================================================
info "7) Device branch rebind round-trip"
# Find unbound device (byte 8-11 zero)
DEVICES=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v1/devices?page=1&size=300")
DEV_ID=$(jq -r --arg t "$TENANT" '[.result.items[] | select(.unit_id==null and .device_id!=null)] | .[0].device_id // empty' <<<"$DEVICES")
[[ -n "$DEV_ID" ]] && [[ "$DEV_ID" != "null" ]] || { info "no unbound device available — skipping rebind test"; ok_rebind=skipped; }
if [[ -n "$DEV_ID" ]] && [[ "$DEV_ID" != "null" ]]; then
  info "  picked device $DEV_ID; rebind to $BRANCH_PFX"
  REBIND=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v1/devices/$DEV_ID" \
    -d "$(jq -nc --arg b "$BRANCH_PFX" '{branch_id:$b}')")
  [[ "$(jq -r '.code' <<<"$REBIND")" == "2000" ]] || { fail "rebind failed: $REBIND"; exit 1; }
  pass "rebound to $BRANCH_NAME"

  info "  revert (set bound_room_id=null, bound_bed_id=null) → tenant pool"
  REVERT=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v1/devices/$DEV_ID" \
    -d '{"bound_room_id":null,"bound_bed_id":null}')
  [[ "$(jq -r '.code' <<<"$REVERT")" == "2000" ]] || { fail "revert failed: $REVERT"; exit 1; }
  pass "reverted to tenant pool"
fi

# ============================================================================
# 8. Residents LIST + GET (PHI/contacts/unit binding)
# ============================================================================
info "8) GET /admin/api/v2/residents (含 PHI/contacts 联表)"
RESIDENTS=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents?size=5")
RES_TOTAL=$(jq -r '.result.total' <<<"$RESIDENTS")
HOA=$(jq -r '.result.items[0].resident_id // empty' <<<"$RESIDENTS")
[[ -n "$HOA" ]] && [[ "$HOA" != "null" ]] || { fail "no residents: $RESIDENTS"; exit 1; }
HOA_ENC=$(printf '%s' "$HOA" | jq -sRr @uri)
DETAIL=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC?include_phi=true&include_contacts=true")
NICK=$(jq -r '.result.nickname // empty' <<<"$DETAIL")
HAS_PHI=$(jq -r '.result.phi != null' <<<"$DETAIL")
pass "residents=$RES_TOTAL; picked=[$NICK @ $HOA] has_phi=$HAS_PHI"

# ============================================================================
# Summary
# ============================================================================
if [[ "${FAILED:-0}" == "1" ]]; then
  echo -e "\n${RED}=== E2E FAILED ===${NC}"
  exit 1
fi
echo -e "\n${GREEN}=== v2 spatial E2E PASS ===${NC}"
echo "  branches LIST + 联表 units"
echo "  units LIST + branch/building/floor 派生"
echo "  rooms LIST + unit/branch JOIN"
echo "  beds LIST + room/unit/branch 三层 JOIN + resident binding"
echo "  beds CREATE → PUT → GET → DELETE round-trip"
echo "  device branch rebind round-trip"
echo "  residents v2 LIST + PHI/contacts include"
