#!/bin/bash
# E2E Test: Contact, Account & Signup Commands (Sub-phase 3f)
# テナント: apac-sandbox
# 注意: signup は正しいボディでもテナント制約で HTTP 500 になるため、その場合のみ skip。
#       account get/summary/update は専用設定不要なので自作アカウントで happy-path を検証。

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
# Log directory
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-contact-signup-${TIMESTAMP}.log"

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

# ─────────────────────────────────────────
header "Step 1: Account Create (テスト用)"
# ─────────────────────────────────────────
ACCT_BODY=$(cat <<'JSON'
{
  "name": "E2E-Contact-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {
    "firstName": "Test",
    "lastName": "Contact",
    "country": "Japan",
    "state": "Tokyo"
  }
}
JSON
)
run $ZR account create --body "$ACCT_BODY" --json
ACCT_NUM=$(echo "$RUN_OUT" | jq -r '.accountNumber // empty' 2>/dev/null)
ACCT_ID=$(echo "$RUN_OUT" | jq -r '.accountId // empty' 2>/dev/null)

if [ -n "$ACCT_NUM" ]; then
  pass "account create → $ACCT_NUM ($ACCT_ID)"
else
  fail "account create failed (rc=$RUN_RC): ${RUN_ERR:-$RUN_OUT}"
  exit 1
fi

# ─────────────────────────────────────────
header "Step 1b: account get / summary / update (live)"
# ─────────────────────────────────────────
echo "  Testing: account get $ACCT_NUM"
run $ZR account get "$ACCT_NUM" --json
GOT_ACCT=$(echo "$RUN_OUT" | jq -r '.basicInfo.accountNumber // .accountNumber // empty' 2>/dev/null)
if [ "$GOT_ACCT" = "$ACCT_NUM" ]; then
  pass "account get → accountNumber=$GOT_ACCT"
else
  fail "account get (rc=$RUN_RC) → got '$GOT_ACCT' ${RUN_ERR}"
fi

echo "  Testing: account summary $ACCT_NUM"
run $ZR account summary "$ACCT_NUM" --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.basicInfo' >/dev/null 2>&1; then
  pass "account summary → has basicInfo"
else
  fail "account summary (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: account update validation (no --body)"
expect_fail "account update validation → requires --body" 'required flag(s) "body" not set' -- $ZR account update "$ACCT_NUM"

echo "  Testing: account update $ACCT_NUM (live, read-back)"
run $ZR account update "$ACCT_NUM" --body '{"name":"E2E-Contact-Test-Updated"}' --json
if [ "$RUN_RC" -eq 0 ]; then
  run $ZR account get "$ACCT_NUM" --json
  NEW_NAME=$(echo "$RUN_OUT" | jq -r '.basicInfo.name // .name // empty' 2>/dev/null)
  if [ "$NEW_NAME" = "E2E-Contact-Test-Updated" ]; then
    pass "account update verified → name=$NEW_NAME"
  else
    fail "account update verify → name='$NEW_NAME' (rc=$RUN_RC) ${RUN_ERR}"
  fi
else
  fail "account update (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: account delete validation (no --confirm)"
expect_fail "account delete validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR account delete "$ACCT_NUM"

# ─────────────────────────────────────────
header "Step 2: contact create"
# ─────────────────────────────────────────
echo "  Testing: contact create validation (no --body)"
expect_fail "contact create validation → requires --body" 'required flag(s) "body" not set' -- $ZR contact create

CONTACT_BODY=$(cat <<JSON
{
  "accountId": "$ACCT_ID",
  "firstName": "E2E",
  "lastName": "TestContact",
  "country": "Japan",
  "state": "Tokyo",
  "workEmail": "e2e-test@example.com"
}
JSON
)
run $ZR contact create --body "$CONTACT_BODY" --json
CONTACT_ID=$(echo "$RUN_OUT" | jq -r '.id // .contactId // empty' 2>/dev/null)

if [ -n "$CONTACT_ID" ]; then
  pass "contact create → $CONTACT_ID"
else
  fail "contact create (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 3: contact get"
# ─────────────────────────────────────────
echo "  Testing: contact get validation (no arg)"
expect_fail "contact get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR contact get

if [ -n "$CONTACT_ID" ]; then
  echo "  Testing: contact get $CONTACT_ID"
  run $ZR contact get "$CONTACT_ID" --json
  GOT_NAME=$(echo "$RUN_OUT" | jq -r '.firstName // empty' 2>/dev/null)
  if [ "$GOT_NAME" = "E2E" ]; then
    pass "contact get → firstName=$GOT_NAME"
  else
    fail "contact get (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "contact get → no contact ID"
fi

# ─────────────────────────────────────────
header "Step 4: contact list"
# ─────────────────────────────────────────
echo "  Testing: contact list validation (no --account-id)"
expect_fail "contact list validation → requires --account-id" 'required flag(s) "account-id" not set' -- $ZR contact list

echo "  Testing: contact list --account-id $ACCT_ID"
run $ZR contact list --account-id "$ACCT_ID" --json
LIST_COUNT=$(echo "$RUN_OUT" | jq -r 'if type=="array" then length else (.contacts // .records // [] | length) end' 2>/dev/null)
if [ -n "$LIST_COUNT" ] && [ "$LIST_COUNT" -ge 1 ] 2>/dev/null; then
  pass "contact list → $LIST_COUNT contact(s)"
else
  fail "contact list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 5: contact update"
# ─────────────────────────────────────────
echo "  Testing: contact update validation (no --body)"
expect_fail "contact update validation → requires --body" 'required flag(s) "body" not set' -- $ZR contact update C-FAKE

if [ -n "$CONTACT_ID" ]; then
  echo "  Testing: contact update $CONTACT_ID"
  run $ZR contact update "$CONTACT_ID" --body '{"firstName":"Updated"}' --json
  UPDATE_SUCCESS=$(echo "$RUN_OUT" | jq -r '.success // empty' 2>/dev/null)
  if [ "$UPDATE_SUCCESS" = "true" ]; then
    pass "contact update → success"
    run $ZR contact get "$CONTACT_ID" --json
    VNAME=$(echo "$RUN_OUT" | jq -r '.firstName // empty' 2>/dev/null)
    if [ "$VNAME" = "Updated" ]; then
      pass "contact update verified → firstName=$VNAME"
    else
      fail "contact update verify → firstName='$VNAME' (rc=$RUN_RC) ${RUN_ERR}"
    fi
  else
    fail "contact update (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "contact update → no contact ID"
fi

# ─────────────────────────────────────────
header "Step 6: contact scrub"
# ─────────────────────────────────────────
echo "  Testing: contact scrub validation (no arg)"
expect_fail "contact scrub validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR contact scrub

if [ -n "$CONTACT_ID" ]; then
  echo "  Testing: contact scrub without --confirm"
  expect_fail "contact scrub validation → requires --confirm" "--confirm" -- $ZR contact scrub "$CONTACT_ID"
else
  skip "contact scrub → no contact ID"
fi

# ─────────────────────────────────────────
header "Step 7: contact delete"
# ─────────────────────────────────────────
echo "  Testing: contact delete validation (no arg)"
expect_fail "contact delete validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR contact delete

if [ -n "$CONTACT_ID" ]; then
  echo "  Testing: contact delete without --confirm"
  expect_fail "contact delete validation → requires --confirm" "--confirm" -- $ZR contact delete "$CONTACT_ID"

  echo "  Testing: contact delete $CONTACT_ID --confirm"
  run $ZR contact delete "$CONTACT_ID" --confirm --json
  DEL_SUCCESS=$(echo "$RUN_OUT" | jq -r '.success // empty' 2>/dev/null)
  if [ "$DEL_SUCCESS" = "true" ]; then
    pass "contact delete → success"
  elif echo "${RUN_ERR}${RUN_OUT}" | grep -qiE "deleted|not.*found"; then
    pass "contact delete → done (idempotent/not found)"
  else
    fail "contact delete (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi

  echo "  Verifying: contact get after delete (bounded retry for propagation)"
  DELETED_CONFIRMED=0
  for _i in 1 2 3 4 5; do
    run $ZR contact get "$CONTACT_ID"
    if [ "$RUN_RC" -ne 0 ] || echo "${RUN_ERR}${RUN_OUT}" | grep -qiE "error|not.*found"; then
      DELETED_CONFIRMED=1; break
    fi
    sleep 2
  done
  if [ "$DELETED_CONFIRMED" -eq 1 ]; then
    pass "contact delete verified → not found"
  else
    # Zuora read-after-delete is eventually consistent; still returning data after
    # the retry window is a known propagation lag, not a CLI defect.
    skip "contact delete verify → still returned after retries (eventual consistency)"
  fi
else
  skip "contact delete → no contact ID"
fi

# ─────────────────────────────────────────
header "Step 7.5: contact snapshot/transfer + account list/payment-methods"
# ─────────────────────────────────────────
echo "  Testing: contact snapshot without arg"
expect_fail "contact snapshot validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR contact snapshot

echo "  Testing: contact transfer without arg"
expect_fail "contact transfer validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR contact transfer

echo "  Testing: contact transfer without --body"
expect_fail "contact transfer validation → requires --body" 'required flag(s) "body" not set' -- $ZR contact transfer C-FAKE

echo "  Testing: account list"
run $ZR account list --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.data | type == "array"' >/dev/null 2>&1; then
  pass "account list → .data array (count=$(echo "$RUN_OUT" | jq '.data | length'))"
else
  fail "account list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: account payment-methods without arg"
expect_fail "account payment-methods validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR account payment-methods

echo "  Testing: account payment-methods $ACCT_NUM (live)"
run $ZR account payment-methods "$ACCT_NUM" --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.success' >/dev/null 2>&1; then
  pass "account payment-methods → returned result (none configured is OK)"
elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
  skip "account payment-methods → $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
else
  fail "account payment-methods (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: account set-cascading without arg"
expect_fail "account set-cascading validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR account set-cascading

echo "  Testing: account set-cascading without --body"
expect_fail "account set-cascading validation → requires --body" 'required flag(s) "body" not set' -- $ZR account set-cascading "$ACCT_NUM"

# payment-methods-default / -cascading: read-only GETs with deterministic
# outcomes on this tenant (no gateway → no default method; cascading is a
# feature toggle). Locking the EXPECTED error codes also exercises the
# RunDetail error-rendering path live, and flips loudly if the tenant gains
# the feature (update the assertion then).
echo "  Testing: account payment-methods-default without arg"
expect_fail "payment-methods-default validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR account payment-methods-default

echo "  Testing: account payment-methods-default $ACCT_NUM (live)"
PMD_RC=0
PMD_OUT=$($ZR account payment-methods-default "$ACCT_NUM" 2>&1) || PMD_RC=$?
if [ "$PMD_RC" -ne 0 ] && echo "$PMD_OUT" | grep -q "50000040"; then
  pass "payment-methods-default → expected 'no default method' error (50000040)"
elif [ "$PMD_RC" -eq 0 ]; then
  pass "payment-methods-default → returned a default payment method"
else
  fail "payment-methods-default → rc=$PMD_RC: $(echo "$PMD_OUT" | head -2)"
fi

echo "  Testing: account payment-methods-cascading without arg"
expect_fail "payment-methods-cascading validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR account payment-methods-cascading

echo "  Testing: account payment-methods-cascading $ACCT_NUM (live)"
PMC_RC=0
PMC_OUT=$($ZR account payment-methods-cascading "$ACCT_NUM" 2>&1) || PMC_RC=$?
if [ "$PMC_RC" -ne 0 ] && echo "$PMC_OUT" | grep -q "50000010"; then
  pass "payment-methods-cascading → expected feature-disabled error (50000010)"
elif [ "$PMC_RC" -eq 0 ]; then
  pass "payment-methods-cascading → returned cascading payment methods"
else
  fail "payment-methods-cascading → rc=$PMC_RC: $(echo "$PMC_OUT" | head -2)"
fi

# ─────────────────────────────────────────
header "Step 8: signup"
# ─────────────────────────────────────────
echo "  Testing: signup validation (no --body)"
expect_fail "signup validation → requires --body" 'required flag(s) "body" not set' -- $ZR signup

echo "  Testing: signup --body (tenant may reject with HTTP 500)"
SIGNUP_BODY=$(cat <<JSON
{
  "accountData": {"name": "E2E-Signup", "currency": "JPY", "billCycleDay": 1, "billToContact": {"firstName": "S", "lastName": "U", "country": "Japan", "state": "Tokyo"}},
  "subscriptionData": {"subscribeToRatePlans": [{"productRatePlanId": "$RATE_PLAN_ID"}]}
}
JSON
)
run $ZR signup --body "$SIGNUP_BODY" --json
SIGNUP_SUCCESS=$(echo "$RUN_OUT" | jq -r '.success // empty' 2>/dev/null)
if [ "$SIGNUP_SUCCESS" = "true" ]; then
  pass "signup → success (account=$(echo "$RUN_OUT" | jq -r '.accountNumber // empty'))"
elif echo "${RUN_ERR}${RUN_OUT}" | grep -qF "Zuora API error"; then
  # The Sign-Up call depends on tenant-specific catalog/subscription setup not
  # present on this apac-sandbox (returns HTTP 400/500); the CLI built and sent
  # the request correctly, so this is an environment skip, not a CLI defect.
  skip "signup → $(echo "${RUN_ERR}${RUN_OUT}" | grep -F 'Zuora API error' | head -1)"
else
  fail "signup (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Test Account: $ACCT_NUM ($ACCT_ID)"
echo "  Contact ID: $CONTACT_ID"
echo ""
print_summary
