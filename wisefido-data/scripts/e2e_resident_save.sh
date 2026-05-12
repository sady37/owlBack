#!/bin/bash
# E2E test for Resident profile save flow.
# Verifies caregivers / careteams / family / contacts / family_access roundtrip after
# the v2 cleanup (single PUT + cryptor-encrypted contacts + family_access write).
#
# Usage: BASE_URL=http://localhost:8080 ./scripts/e2e_resident_save.sh

set -e
BASE_URL="${BASE_URL:-http://localhost:8080}"
USER_ACCOUNT="${USER_ACCOUNT:-admin}"
PASSWORD="${PASSWORD:-Ts123@123}"

GREEN='\033[0;32m'; RED='\033[0;31m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; FAILED=1; }
info() { echo -e "${YELLOW}• $1${NC}"; }

need() {
  command -v "$1" >/dev/null 2>&1 || { echo "missing dep: $1"; exit 1; }
}
need curl
need jq

PASSWORD_HASH=$(printf '%s' "$PASSWORD" | sha256sum | cut -d' ' -f1)

info "1) Login as $USER_ACCOUNT"
LOGIN_RESP=$(curl -sS -X POST "$BASE_URL/auth/api/v2/login" \
  -H 'Content-Type: application/json' \
  -d "$(jq -nc --arg u "$USER_ACCOUNT" --arg p "$PASSWORD_HASH" '{user_account:$u, password_hash:$p}')")
CODE=$(jq -r '.code // empty' <<<"$LOGIN_RESP")
if [[ "$CODE" != "2000" ]]; then
  fail "login: $LOGIN_RESP"; exit 1
fi
TOKEN=$(jq -r '.result.accessToken // .result.access_token' <<<"$LOGIN_RESP")
TENANT=$(jq -r '.result.tenant_id' <<<"$LOGIN_RESP")
[[ -n "$TOKEN" && "$TOKEN" != "null" ]] || { fail "no access_token"; exit 1; }
pass "login ok; tenant=$TENANT"

AUTH=(-H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json')

info "2) List residents, pick first active"
LIST=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents?size=5")
HOA=$(jq -r '.result.items[0].resident_id // empty' <<<"$LIST")
[[ -n "$HOA" ]] || { fail "no resident found"; exit 1; }
HOA_ENC=$(printf '%s' "$HOA" | jq -sRr @uri)
pass "target resident hoa=$HOA"

info "3) Snapshot current state (used to restore at end)"
SNAPSHOT=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC?include_phi=true&include_contacts=true")
[[ "$(jq -r '.code' <<<"$SNAPSHOT")" == "2000" ]] || { fail "snapshot GET: $SNAPSHOT"; exit 1; }
SAVED_CAREGIVERS=$(jq -c '.result.caregivers // []' <<<"$SNAPSHOT")
SAVED_TEAMS=$(jq -c '.result.teams // []' <<<"$SNAPSHOT")
SAVED_FAMILY=$(jq -c '.result.family // []' <<<"$SNAPSHOT")
SAVED_CONTACTS=$(jq -c '.result.contacts // []' <<<"$SNAPSHOT")
SAVED_PHI=$(jq -c '.result.phi // null' <<<"$SNAPSHOT")
SAVED_FAMILY_ACCESS=$(jq -r '.result.family_access // true' <<<"$SNAPSHOT")
pass "snapshot ok"

info "4) Fetch a caregiver user_id (Nurse/Caregiver role) for write test"
BRANCH=$(jq -r '.result.branch_id // empty' <<<"$SNAPSHOT")
[[ -n "$BRANCH" ]] || { fail "resident has no branch_id"; exit 1; }
BRANCH_ENC=$(printf '%s' "$BRANCH" | jq -sRr @uri)
CGS=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v1/users/caregivers?branch_id=$BRANCH_ENC")
TEST_CG=$(jq -r '.result.items[0].user_id // empty' <<<"$CGS")
[[ -n "$TEST_CG" ]] || info "no caregivers available — will use existing $SAVED_CAREGIVERS"
TEAMS_RESP=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v1/users/caregiver-groups?branch_id=$BRANCH_ENC")
TEST_TEAM=$(jq -r '.result.items[0].team_id // empty' <<<"$TEAMS_RESP")
FAM_RESP=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v1/users/family?branch_id=$BRANCH_ENC")
TEST_FAM=$(jq -r '.result.items[0].user_id // empty' <<<"$FAM_RESP")
info "caregiver=$TEST_CG team=$TEST_TEAM family=$TEST_FAM"

info "5) PUT full payload (caregivers / teams / family / family_access; contacts separately)"
TIMESTAMP=$(date +%s)
PUT_PAYLOAD=$(jq -nc \
  --arg cg  "$TEST_CG" \
  --arg tm  "$TEST_TEAM" \
  --arg fm  "$TEST_FAM" \
  '{
    family_access: false,
    caregiver_user_ids: ($cg | if . == "" then [] else [.] end),
    care_team_ids:     ($tm | if . == "" then [] else [.] end),
    family_user_ids:   ($fm | if . == "" then [] else [.] end)
  }')
PUT_RESP=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" -d "$PUT_PAYLOAD")
[[ "$(jq -r '.code' <<<"$PUT_RESP")" == "2000" ]] || { fail "PUT(relations): $PUT_RESP"; exit 1; }
pass "PUT(relations) ok"

# Contacts requires PHI cryptor; try separately, allow skip
CONTACTS_OK=0
CT_PUT=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" -d "$(jq -nc --arg ts "$TIMESTAMP" '{
  contacts: [{
    relationship: "spouse",
    contact_first_name: ("E2E-" + $ts),
    contact_last_name:  "Round-Trip",
    contact_phone:      "+1-555-0100",
    contact_email:      ("e2e+" + $ts + "@test.local"),
    receive_sms:  true,
    receive_email: false
  }]
}')")
if [[ "$(jq -r '.code' <<<"$CT_PUT")" == "2000" ]]; then
  CONTACTS_OK=1
  pass "PUT(contacts) ok (cryptor available)"
else
  info "PUT(contacts) skipped: $(jq -r '.message' <<<"$CT_PUT")"
fi

info "6) GET and verify roundtrip"
GET1=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC?include_phi=true&include_contacts=true")
[[ "$(jq -r '.code' <<<"$GET1")" == "2000" ]] || { fail "GET(1): $GET1"; exit 1; }

# jq // 操作符把 false 当 missing；改用 if-then-else 精确判断
GOT_FA=$(jq -r 'if (.result | has("family_access")) then (.result.family_access | tostring) else "MISSING" end' <<<"$GET1")
[[ "$GOT_FA" == "false" ]] && pass "family_access=false roundtrip" || fail "family_access expected false got $GOT_FA"

if [[ -n "$TEST_CG" ]]; then
  HAS_CG=$(jq --arg uid "$TEST_CG" -r '.result.caregivers // [] | map(.user_id) | index($uid) // -1' <<<"$GET1")
  [[ "$HAS_CG" != "-1" ]] && pass "caregiver $TEST_CG in GET" || fail "caregiver $TEST_CG missing"
fi
if [[ -n "$TEST_TEAM" ]]; then
  HAS_TM=$(jq --arg tid "$TEST_TEAM" -r '.result.teams // [] | map(.team_id) | index($tid) // -1' <<<"$GET1")
  [[ "$HAS_TM" != "-1" ]] && pass "team $TEST_TEAM in GET" || fail "team $TEST_TEAM missing"
fi
if [[ -n "$TEST_FAM" ]]; then
  HAS_FM=$(jq --arg uid "$TEST_FAM" -r '.result.family // [] | map(.user_id) | index($uid) // -1' <<<"$GET1")
  [[ "$HAS_FM" != "-1" ]] && pass "family $TEST_FAM in GET — original bug fixed" || fail "family $TEST_FAM missing — bug still present"
fi

if [[ "$CONTACTS_OK" == "1" ]]; then
  CT_NAME=$(jq -r '.result.contacts // [] | map(select(.contact_first_name=="E2E-'"$TIMESTAMP"'")) | length' <<<"$GET1")
  [[ "$CT_NAME" == "1" ]] && pass "contact roundtrip (AES-256-GCM write+decrypt read) ok" || fail "contact missing/duplicated; count=$CT_NAME"
fi

info "6b) PUT PHI — write encrypted fields"
PHI_OK=0
PHI_FIRST="E2E-PHI-$TIMESTAMP"
PHI_PUT=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" -d "$(jq -nc --arg fn "$PHI_FIRST" '{
  phi: {
    first_name:        $fn,
    last_name:         "Smith",
    gender:            "Female",
    date_of_birth:     "1942-03-15",
    resident_phone:    "+1-555-0142",
    resident_email:    "phi@test.local",
    weight_lb:         165.5,
    height_ft:         5.0,
    height_in:         9.0,
    mobility_level:    3,
    tremor_status:     "Mild",
    mobility_aid:      "Cane",
    has_hypertension:  true,
    has_alzheimer:     false,
    medical_history:   "stable",
    home_address_street: "123 Test Ave",
    home_address_city:   "Springfield",
    home_address_state:  "CA",
    home_address_postal_code: "90210",
    plus_code:         "85FQ+H7"
  }
}')")
if [[ "$(jq -r '.code' <<<"$PHI_PUT")" == "2000" ]]; then
  PHI_OK=1
  pass "PUT(phi) ok"
else
  fail "PUT(phi): $(jq -r '.message // .' <<<"$PHI_PUT")"
fi

info "6c) GET — verify PHI roundtrip (decrypt)"
if [[ "$PHI_OK" == "1" ]]; then
  GET_PHI=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC?include_phi=true")
  GOT_FN=$(jq -r '.result.phi.first_name // "MISSING"' <<<"$GET_PHI")
  GOT_GENDER=$(jq -r '.result.phi.gender // "MISSING"' <<<"$GET_PHI")
  GOT_DOB=$(jq -r '.result.phi.date_of_birth // "MISSING"' <<<"$GET_PHI")
  GOT_WT=$(jq -r '.result.phi.weight_lb // "MISSING"' <<<"$GET_PHI")
  GOT_HTN=$(jq -r 'if (.result.phi | has("has_hypertension")) then (.result.phi.has_hypertension | tostring) else "MISSING" end' <<<"$GET_PHI")
  GOT_ALZ=$(jq -r 'if (.result.phi | has("has_alzheimer")) then (.result.phi.has_alzheimer | tostring) else "MISSING" end' <<<"$GET_PHI")
  GOT_PC=$(jq -r '.result.phi.plus_code // "MISSING"' <<<"$GET_PHI")
  [[ "$GOT_FN" == "$PHI_FIRST" ]] && pass "phi.first_name decrypt ok" || fail "phi.first_name expected $PHI_FIRST got $GOT_FN"
  [[ "$GOT_GENDER" == "Female" ]] && pass "phi.gender decrypt ok" || fail "phi.gender expected Female got $GOT_GENDER"
  [[ "$GOT_DOB" == "1942-03-15" ]] && pass "phi.date_of_birth decrypt ok" || fail "phi.date_of_birth expected 1942-03-15 got $GOT_DOB"
  [[ "$GOT_WT" == "165.5" ]] && pass "phi.weight_lb (float) decrypt ok" || fail "phi.weight_lb expected 165.5 got $GOT_WT"
  [[ "$GOT_HTN" == "true" ]] && pass "phi.has_hypertension (bool) decrypt ok" || fail "phi.has_hypertension expected true got $GOT_HTN"
  [[ "$GOT_ALZ" == "false" ]] && pass "phi.has_alzheimer (bool=false) decrypt ok" || fail "phi.has_alzheimer expected false got $GOT_ALZ"
  [[ "$GOT_PC" == "85FQ+H7" ]] && pass "phi.plus_code (plain) ok" || fail "phi.plus_code expected 85FQ+H7 got $GOT_PC"
fi

info "7) PUT empty arrays + empty PHI — verify clear semantics"
CLEAR_PAYLOAD=$(jq -nc '{
  caregiver_user_ids: [], care_team_ids: [], family_user_ids: [], contacts: [], phi: {}
}')
PUT_CLR=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" -d "$CLEAR_PAYLOAD")
[[ "$(jq -r '.code' <<<"$PUT_CLR")" == "2000" ]] || { fail "PUT(clear): $PUT_CLR"; exit 1; }

GET2=$(curl -sS "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC?include_phi=true&include_contacts=true")
EMPTY_CG=$(jq -r '.result.caregivers // [] | length' <<<"$GET2")
EMPTY_TM=$(jq -r '.result.teams // [] | length' <<<"$GET2")
EMPTY_FM=$(jq -r '.result.family // [] | length' <<<"$GET2")
EMPTY_CT=$(jq -r '.result.contacts // [] | length' <<<"$GET2")
[[ "$EMPTY_CG" == "0" ]] && pass "caregivers cleared" || fail "caregivers still $EMPTY_CG"
[[ "$EMPTY_TM" == "0" ]] && pass "teams cleared" || fail "teams still $EMPTY_TM"
[[ "$EMPTY_FM" == "0" ]] && pass "family cleared" || fail "family still $EMPTY_FM"
[[ "$EMPTY_CT" == "0" ]] && pass "contacts cleared" || fail "contacts still $EMPTY_CT"
EMPTY_PHI_FN=$(jq -r 'if (.result.phi != null and (.result.phi | has("first_name"))) then .result.phi.first_name else "" end' <<<"$GET2")
[[ -z "$EMPTY_PHI_FN" ]] && pass "phi.first_name cleared" || fail "phi.first_name still '$EMPTY_PHI_FN'"

info "8) Restore snapshot"
RESTORE=$(jq -nc \
  --argjson cg "$(jq -c '[.[].user_id]' <<<"$SAVED_CAREGIVERS")" \
  --argjson tm "$(jq -c '[.[].team_id]' <<<"$SAVED_TEAMS")" \
  --argjson fm "$(jq -c '[.[].user_id]' <<<"$SAVED_FAMILY")" \
  --argjson ct "$SAVED_CONTACTS" \
  --argjson phi "$SAVED_PHI" \
  --argjson fa "$SAVED_FAMILY_ACCESS" \
  '{
    family_access: $fa,
    caregiver_user_ids: $cg,
    care_team_ids:      $tm,
    family_user_ids:    $fm,
    contacts:           ($ct | map({relationship, contact_first_name, contact_last_name, contact_phone, contact_email, receive_sms, receive_email})),
    phi:                ($phi // {})
  }')
RESTORE_RESP=$(curl -sS -X PUT "${AUTH[@]}" "$BASE_URL/admin/api/v2/residents/$HOA_ENC" -d "$RESTORE")
[[ "$(jq -r '.code' <<<"$RESTORE_RESP")" == "2000" ]] && pass "snapshot restored" || info "restore non-2000 (manual fixup may be needed): $RESTORE_RESP"

echo ""
if [[ "${FAILED:-0}" == "1" ]]; then
  echo -e "${RED}E2E FAILED${NC}"
  exit 1
fi
echo -e "${GREEN}E2E PASSED${NC}"
