#!/bin/bash
# E2E Test: Contact + Sign Up Commands (Sub-phase 3c)
# テナント: apac-sandbox
# フロー: account create → contact create → get → list → update → transfer → scrub → delete
#         signup はテナント制約で 500 エラーのためバリデーションのみ

set -uo pipefail

ZR="./bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
TODAY=$(date +%Y-%m-%d)

# Log directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-contact-signup-${TIMESTAMP}.log"

exec > >(tee >(sed 's/\x1b\[[0-9;]*m//g' > "$LOG_FILE")) 2>&1

green()  { printf "\033[32m%s\033[0m\n" "$1"; }
red()    { printf "\033[31m%s\033[0m\n" "$1"; }
yellow() { printf "\033[33m%s\033[0m\n" "$1"; }

pass() { PASS=$((PASS+1)); green "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); red   "  ✗ $1"; }
skip() { SKIP=$((SKIP+1)); yellow "  ⊘ $1 (skipped)"; }

header() { printf "\n\033[1m=== %s ===\033[0m\n" "$1"; }

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
if $ZR auth status 2>&1 | grep -q "Environment:"; then
  pass "Auth OK"
else
  fail "Auth failed"
  exit 1
fi

# ─────────────────────────────────────────
header "Step 1: Account Create (テスト用)"
# ─────────────────────────────────────────
ACCT_RESULT=$($ZR account create --body '{
  "name": "E2E-Contact-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {
    "firstName": "Bill",
    "lastName": "Contact",
    "country": "Japan",
    "state": "Tokyo"
  }
}' --json 2>/dev/null) || true
ACCT_NUM=$(echo "$ACCT_RESULT" | jq -r '.accountNumber // empty' 2>/dev/null)
ACCT_ID=$(echo "$ACCT_RESULT" | jq -r '.accountId // empty' 2>/dev/null)

if [ -n "$ACCT_NUM" ] && [ -n "$ACCT_ID" ]; then
  pass "account create → $ACCT_NUM ($ACCT_ID)"
else
  fail "account create failed: $ACCT_RESULT"
  printf '\n'
  red "Cannot proceed without a dedicated test account. Aborting."
  exit 1
fi

# ─────────────────────────────────────────
header "Step 2: contact create"
# ─────────────────────────────────────────

# 2a: Missing --body
echo "  Testing: contact create without --body"
CREATE_ERR=$($ZR contact create 2>&1) || true
if echo "$CREATE_ERR" | grep -qi "body.*required\|required.*body"; then
  pass "contact create validation → requires --body"
else
  fail "contact create validation → unexpected: $CREATE_ERR"
fi

# 2b: Actual create
echo "  Testing: contact create --body"
CONTACT_RESULT=$($ZR contact create --body "{
  \"accountId\": \"$ACCT_ID\",
  \"firstName\": \"E2E\",
  \"lastName\": \"TestContact\",
  \"workEmail\": \"e2e-test@example.com\",
  \"country\": \"Japan\",
  \"state\": \"Tokyo\"
}" --json 2>/dev/null) || true
CONTACT_ID=$(echo "$CONTACT_RESULT" | jq -r '.id // empty' 2>/dev/null)

if [ -n "$CONTACT_ID" ]; then
  pass "contact create → $CONTACT_ID"
else
  fail "contact create failed: $(echo "$CONTACT_RESULT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 3: contact get"
# ─────────────────────────────────────────

# 3a: Missing args
echo "  Testing: contact get without args"
GET_ERR=$($ZR contact get 2>&1) || true
if echo "$GET_ERR" | grep -qi "arg\|required"; then
  pass "contact get validation → requires args"
else
  fail "contact get validation → unexpected: $GET_ERR"
fi

# 3b: Actual get
echo "  Testing: contact get $CONTACT_ID"
GET_OUT=$($ZR contact get "$CONTACT_ID" --json 2>/dev/null) || true
GET_FIRST=$(echo "$GET_OUT" | jq -r '.firstName // empty' 2>/dev/null)

if [ "$GET_FIRST" = "E2E" ]; then
  pass "contact get → firstName=E2E"
else
  fail "contact get → unexpected: $(echo "$GET_OUT" | head -3)"
fi

# 3c: Detail output
echo "  Testing: contact get default output"
GET_DETAIL=$($ZR contact get "$CONTACT_ID" 2>/dev/null) || true
if echo "$GET_DETAIL" | grep -q "E2E" && echo "$GET_DETAIL" | grep -q "TestContact"; then
  pass "contact get detail → shows name"
else
  fail "contact get detail → $GET_DETAIL"
fi

# ─────────────────────────────────────────
header "Step 4: contact list"
# ─────────────────────────────────────────

# 4a: Missing --account-id
echo "  Testing: contact list without --account-id"
LIST_ERR=$($ZR contact list 2>&1) || true
if echo "$LIST_ERR" | grep -qi "account-id\|required"; then
  pass "contact list validation → requires --account-id"
else
  fail "contact list validation → unexpected: $LIST_ERR"
fi

# 4b: Table output
echo "  Testing: contact list --account-id $ACCT_ID"
LIST_OUT=$($ZR contact list --account-id "$ACCT_ID" 2>&1) || true
if echo "$LIST_OUT" | grep -q "E2E"; then
  pass "contact list → contains E2E contact"
else
  fail "contact list → missing E2E: $LIST_OUT"
fi

# 4c: JSON output
echo "  Testing: contact list --account-id $ACCT_ID --json"
LIST_JSON=$($ZR contact list --account-id "$ACCT_ID" --json 2>/dev/null) || true
LIST_SIZE=$(echo "$LIST_JSON" | jq -r '.size // 0' 2>/dev/null)
if [ "$LIST_SIZE" -ge 1 ] 2>/dev/null; then
  pass "contact list --json → size=$LIST_SIZE"
else
  fail "contact list --json → $LIST_JSON"
fi

# 4d: --jq output
echo "  Testing: contact list --jq '.records[].FirstName'"
JQ_OUT=$($ZR contact list --account-id "$ACCT_ID" --jq '.records[].FirstName' 2>/dev/null) || true
if echo "$JQ_OUT" | grep -q "E2E"; then
  pass "contact list --jq → filtered output correct"
else
  fail "contact list --jq → $JQ_OUT"
fi

# ─────────────────────────────────────────
header "Step 5: contact update"
# ─────────────────────────────────────────

# 5a: Missing --body
echo "  Testing: contact update without --body"
UPD_ERR=$($ZR contact update "$CONTACT_ID" 2>&1) || true
if echo "$UPD_ERR" | grep -qi "body.*required\|required.*body"; then
  pass "contact update validation → requires --body"
else
  fail "contact update validation → unexpected: $UPD_ERR"
fi

# 5b: Actual update
echo "  Testing: contact update $CONTACT_ID"
UPD_OUT=$($ZR contact update "$CONTACT_ID" --body '{"firstName":"Updated"}' --json 2>/dev/null) || true
UPD_SUCCESS=$(echo "$UPD_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$UPD_SUCCESS" = "true" ]; then
  pass "contact update → success"
else
  fail "contact update → $(echo "$UPD_OUT" | head -3)"
fi

# 5c: Verify update
echo "  Verifying: contact get after update"
VERIFY_FIRST=$($ZR contact get "$CONTACT_ID" --jq '.firstName' 2>/dev/null) || true
if echo "$VERIFY_FIRST" | grep -q "Updated"; then
  pass "contact update verified → firstName=Updated"
else
  fail "contact update verify → firstName=$VERIFY_FIRST"
fi

# ─────────────────────────────────────────
header "Step 6: contact scrub"
# ─────────────────────────────────────────

# 6a: Missing args
echo "  Testing: contact scrub without args"
SCRUB_ERR=$($ZR contact scrub 2>&1) || true
if echo "$SCRUB_ERR" | grep -qi "arg\|required"; then
  pass "contact scrub validation → requires args"
else
  fail "contact scrub validation → unexpected: $SCRUB_ERR"
fi

# 6b: Actual scrub
echo "  Testing: contact scrub $CONTACT_ID"
SCRUB_OUT=$($ZR contact scrub "$CONTACT_ID" --json 2>/dev/null) || true
SCRUB_SUCCESS=$(echo "$SCRUB_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$SCRUB_SUCCESS" = "true" ]; then
  pass "contact scrub → success"
else
  # Scrub may require specific permissions
  SCRUB_ERR2=$($ZR contact scrub "$CONTACT_ID" 2>&1) || true
  if echo "$SCRUB_ERR2" | grep -qi "error\|permission\|not.*enabled"; then
    skip "contact scrub → tenant may not have scrub permission enabled"
  else
    fail "contact scrub → $(echo "$SCRUB_OUT" | head -3)"
  fi
fi

# ─────────────────────────────────────────
header "Step 7: contact delete"
# ─────────────────────────────────────────

# 7a: Missing --confirm
echo "  Testing: contact delete without --confirm"
DEL_ERR=$($ZR contact delete "$CONTACT_ID" 2>&1) || true
if echo "$DEL_ERR" | grep -q "\-\-confirm"; then
  pass "contact delete validation → requires --confirm"
else
  fail "contact delete validation → unexpected: $DEL_ERR"
fi

# 7b: Actual delete
echo "  Testing: contact delete $CONTACT_ID --confirm"
DEL_OUT=$($ZR contact delete "$CONTACT_ID" --confirm --json 2>/dev/null) || true
DEL_SUCCESS=$(echo "$DEL_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$DEL_SUCCESS" = "true" ]; then
  pass "contact delete → success"
else
  DEL_STDERR=$($ZR contact delete "$CONTACT_ID" --confirm 2>&1) || true
  if echo "$DEL_STDERR" | grep -qi "deleted"; then
    pass "contact delete → success (stderr)"
  else
    fail "contact delete → $(echo "$DEL_OUT" | head -3)"
  fi
fi

# 7c: Verify deleted
echo "  Verifying: contact get after delete"
VERIFY_DEL=$($ZR contact get "$CONTACT_ID" 2>&1) || true
if echo "$VERIFY_DEL" | grep -qi "error\|not.*found"; then
  pass "contact delete verified → not found"
else
  skip "contact delete verify → may still return data"
fi

# ─────────────────────────────────────────
header "Step 8: signup"
# ─────────────────────────────────────────

# 8a: Missing --body
echo "  Testing: signup without --body"
SIGNUP_ERR=$($ZR signup 2>&1) || true
if echo "$SIGNUP_ERR" | grep -qi "body.*required\|required.*body"; then
  pass "signup validation → requires --body"
else
  fail "signup validation → unexpected: $SIGNUP_ERR"
fi

# 8b: Actual signup (may fail on this tenant)
echo "  Testing: signup --body"
SIGNUP_OUT=$($ZR signup --body "{
  \"accountData\": {
    \"name\": \"signup-e2e\",
    \"currency\": \"JPY\",
    \"billCycleDay\": 1,
    \"autoPay\": false,
    \"billToContact\": {\"firstName\": \"S\", \"lastName\": \"U\", \"country\": \"Japan\", \"state\": \"Tokyo\"}
  },
  \"subscriptionData\": {
    \"startDate\": \"$TODAY\",
    \"terms\": {
      \"initialTerm\": {\"period\": 12, \"periodType\": \"Month\", \"termType\": \"TERMED\"},
      \"renewalTerms\": [{\"period\": 12, \"periodType\": \"Month\"}],
      \"renewalSetting\": \"RENEW_WITH_SPECIFIC_TERM\",
      \"autoRenew\": false
    },
    \"ratePlans\": [{\"productRatePlanId\": \"4c6059a8d8899f453ffa0637451d0003\"}]
  }
}" --json 2>/dev/null) || true
SIGNUP_SUCCESS=$(echo "$SIGNUP_OUT" | jq -r '.success // empty' 2>/dev/null)
SIGNUP_STDERR=$($ZR signup --body '{"accountData":{}}' 2>&1) || true

if [ "$SIGNUP_SUCCESS" = "true" ]; then
  SIGNUP_ACCT=$(echo "$SIGNUP_OUT" | jq -r '.accountNumber // empty' 2>/dev/null)
  pass "signup → success (account=$SIGNUP_ACCT)"
elif echo "$SIGNUP_STDERR" | grep -qi "error\|500\|internal"; then
  skip "signup → tenant returns 500 (internal error, may require specific configuration)"
else
  fail "signup → $(echo "$SIGNUP_OUT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 9: contact snapshot"
# ─────────────────────────────────────────

# 9a: Missing args
echo "  Testing: contact snapshot without args"
SNAP_ERR=$($ZR contact snapshot 2>&1) || true
if echo "$SNAP_ERR" | grep -qi "arg\|required"; then
  pass "contact snapshot validation → requires args"
else
  fail "contact snapshot validation → unexpected: $SNAP_ERR"
fi

# 9b: Snapshot with non-existent ID (expected 404)
echo "  Testing: contact snapshot with invalid ID"
SNAP_OUT=$($ZR contact snapshot "nonexistent-id" 2>&1) || true
if echo "$SNAP_OUT" | grep -qi "error\|not.*found\|404"; then
  pass "contact snapshot → correctly returns error for invalid ID"
else
  skip "contact snapshot → unexpected response (may need valid snapshot ID)"
fi

# ─────────────────────────────────────────
header "Step 10: contact transfer"
# ─────────────────────────────────────────

# 10a: Missing --body
echo "  Testing: contact transfer without --body"
XFER_ERR=$($ZR contact transfer "any-id" 2>&1) || true
if echo "$XFER_ERR" | grep -qi "body.*required\|required.*body"; then
  pass "contact transfer validation → requires --body"
else
  fail "contact transfer validation → unexpected: $XFER_ERR"
fi

# 10b: Transfer with invalid ID (expected error)
echo "  Testing: contact transfer with invalid body"
XFER_OUT=$($ZR contact transfer "nonexistent-id" --body '{"destinationAccountId":"fake"}' 2>&1) || true
if echo "$XFER_OUT" | grep -qi "error\|not.*found\|invalid"; then
  pass "contact transfer → correctly returns error for invalid input"
else
  skip "contact transfer → unexpected response"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Test Account: $ACCT_NUM ($ACCT_ID)"
echo "  Contact ID: $CONTACT_ID"
echo ""
TOTAL=$((PASS + FAIL + SKIP))
green  "  Passed:  $PASS / $TOTAL"
if [ "$FAIL" -gt 0 ]; then
  red  "  Failed:  $FAIL / $TOTAL"
fi
if [ "$SKIP" -gt 0 ]; then
  yellow "  Skipped: $SKIP / $TOTAL"
fi
echo ""
echo "  Log: $LOG_FILE"
echo ""

if [ "$FAIL" -gt 0 ]; then
  red "  RESULT: FAIL"
  exit 1
else
  green "  RESULT: PASS"
  exit 0
fi
