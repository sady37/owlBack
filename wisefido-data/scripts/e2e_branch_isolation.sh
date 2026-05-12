#!/bin/bash
# E2E: care_team must belong to branch + resident cross-branch bind rejection.
# 1) Create care_team without branch_id → 400
# 2) Create care_team with valid branch_id → 200
# 3) Bind same-branch caregiver to resident → ok
# 4) Bind cross-branch caregiver to resident → rejected
# 5) Bind cross-branch care_team to resident → rejected
# 6) Cleanup created care_team

set -e
BASE_URL="${BASE_URL:-http://localhost:8080}"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; FAILED=1; }
info() { echo -e "${YELLOW}• $1${NC}"; }

# admin login (Denver tenant)
PASSHASH=$(printf 'Ts123@123' | sha256sum | cut -d' ' -f1)
TOKEN=$(curl -sS -X POST "$BASE_URL/auth/api/v2/login" -H 'Content-Type: application/json' \
  -d "$(jq -nc --arg u admin --arg p "$PASSHASH" '{user_account:$u, password_hash:$p}')" \
  | jq -r '.result.accessToken // .result.access_token')
[[ -n "$TOKEN" ]] || { echo "login failed"; exit 1; }
AUTH=(-H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json')

# Test fixtures (predefined in owl_v2 seed)
RESIDENT_HOA='fd00:0:3:ff01:1::'              # Denver
RESIDENT_BRANCH='fd00:0:3:100::/56'           # Denver
DENVER_BRANCH='fd00:0:3:100::/56'
SANDIEGO_BRANCH='fd00:0:3:300::/56'
DVC1_CG='3d82ee14-60ea-4f05-a876-785865de9791'  # Denver caregiver
SDC1_CG='277dacd2-a07b-4648-8c69-3cd9ae2560b2'  # SanDiego caregiver
DENVER_TEAM='f05bed62-c38f-4dbf-815a-27037d7bf67c'  # emerge (Denver)
HOA_ENC=$(printf '%s' "$RESIDENT_HOA" | jq -sRr @uri)

info "1) Create care_team without branch_id — should fail"
TS=$(date +%s)
R1=$(curl -sS -X POST "${AUTH[@]}" "$BASE_URL/admin/api/v1/care-teams" \
  -d "$(jq -nc --arg n "test_nobr_$TS" '{team_name:$n, team_kind:"specialty"}')")
MSG=$(jq -r '.message // ""' <<<"$R1")
if echo "$MSG" | grep -qi 'branch_id'; then
  pass "create without branch rejected: $MSG"
else
  fail "create without branch returned: $R1"
fi

info "2) Create care_team in Denver — should succeed"
R2=$(curl -sS -X POST "${AUTH[@]}" "$BASE_URL/admin/api/v1/care-teams" \
  -d "$(jq -nc --arg n "test_dv_$TS" --arg b "$DENVER_BRANCH" '{team_name:$n, team_kind:"specialty", branch_id:$b}')")
NEW_TEAM_ID=$(jq -r '.result.team_id // empty' <<<"$R2")
if [[ -n "$NEW_TEAM_ID" ]]; then
  pass "create Denver team ok: $NEW_TEAM_ID"
else
  fail "create Denver team failed: $R2"; exit 1
fi

info "3) Bind same-branch caregiver (Denver dvc1) to Denver resident — should succeed"
R3=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" \
  -d "$(jq -nc --arg cg "$DVC1_CG" '{caregiver_user_ids:[$cg]}')")
if [[ "$(jq -r '.code' <<<"$R3")" == "2000" ]]; then
  pass "same-branch caregiver bind ok"
else
  fail "same-branch caregiver bind failed: $R3"
fi

info "4) Bind cross-branch caregiver (SanDiego sdc1) to Denver resident — should fail"
R4=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" \
  -d "$(jq -nc --arg cg "$SDC1_CG" '{caregiver_user_ids:[$cg]}')")
MSG=$(jq -r '.message // ""' <<<"$R4")
if echo "$MSG" | grep -qi "not assigned to resident's branch"; then
  pass "cross-branch caregiver rejected: $MSG"
else
  fail "cross-branch caregiver NOT rejected: $R4"
fi

info "5) Bind Denver care_team to Denver resident — should succeed"
R5=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" \
  -d "$(jq -nc --arg tm "$DENVER_TEAM" '{care_team_ids:[$tm]}')")
if [[ "$(jq -r '.code' <<<"$R5")" == "2000" ]]; then
  pass "same-branch team bind ok"
else
  fail "same-branch team bind failed: $R5"
fi

info "6) Create SanDiego team, bind to Denver resident — should fail"
R6A=$(curl -sS -X POST "${AUTH[@]}" "$BASE_URL/admin/api/v1/care-teams" \
  -d "$(jq -nc --arg n "test_sd_$TS" --arg b "$SANDIEGO_BRANCH" '{team_name:$n, team_kind:"specialty", branch_id:$b}')")
SD_TEAM_ID=$(jq -r '.result.team_id' <<<"$R6A")
R6=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" \
  -d "$(jq -nc --arg tm "$SD_TEAM_ID" '{care_team_ids:[$tm]}')")
MSG=$(jq -r '.message // ""' <<<"$R6")
if echo "$MSG" | grep -qi 'cannot bind to resident'; then
  pass "cross-branch team rejected: $MSG"
else
  fail "cross-branch team NOT rejected: $R6"
fi

info "7) Cleanup: clear caregivers/teams + delete test care_teams"
curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" \
  -d '{"caregiver_user_ids":[], "care_team_ids":[]}' >/dev/null
curl -sS -X DELETE "${AUTH[@]}" "$BASE_URL/admin/api/v1/care-teams/$NEW_TEAM_ID" >/dev/null
curl -sS -X DELETE "${AUTH[@]}" "$BASE_URL/admin/api/v1/care-teams/$SD_TEAM_ID" >/dev/null
pass "cleanup done"

echo ""
if [[ "${FAILED:-0}" == "1" ]]; then
  echo -e "${RED}BRANCH ISOLATION TEST FAILED${NC}"
  exit 1
fi
echo -e "${GREEN}BRANCH ISOLATION TEST PASSED${NC}"
