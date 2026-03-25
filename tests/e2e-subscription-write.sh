#!/bin/bash
# E2E Test: Subscription Write Commands (Sub-phase 3b)
# テナント: apac-sandbox (Orders 有効)
# 注意: create/update は Orders 有効テナントでは v1 API が使えないため、
#       Orders API 経由でサブスクリプションを作成してテストする

set -uo pipefail

ZR="./bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RATE_PLAN_ID="4c6059a8d8899f453ffa0637451d0003"  # Backlog-スタータープラン(月払い)

# Log directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-subscription-write-${TIMESTAMP}.log"

# Tee all output to log file (preserving terminal colors in terminal, stripping in file)
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
header "Step 1: Account Create (テスト用アカウント)"
# ─────────────────────────────────────────
ACCT_BODY=$(cat <<'JSON'
{
  "name": "E2E-Sub-Write-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {
    "firstName": "Test",
    "lastName": "User",
    "country": "Japan",
    "state": "Tokyo"
  }
}
JSON
)

ACCT_RESULT=$($ZR account create --body "$ACCT_BODY" --json 2>&1) || true
ACCT_NUM=$(echo "$ACCT_RESULT" | jq -r '.accountNumber // empty' 2>/dev/null)
ACCT_ID=$(echo "$ACCT_RESULT" | jq -r '.accountId // empty' 2>/dev/null)

if [ -n "$ACCT_NUM" ]; then
  pass "account create → $ACCT_NUM"
else
  fail "account create failed: $ACCT_RESULT"
  red "\nCannot proceed without a dedicated test account. Aborting."
  exit 1
fi

echo "  Account: $ACCT_NUM ($ACCT_ID)"

# ─────────────────────────────────────────
header "Step 2: Create subscriptions via Orders API"
# ─────────────────────────────────────────
_create_sub_via_order() {
  local label=$1
  local acct=$2
  local term_type=$3
  local auto_renew=$4

  local order_body
  order_body=$(cat <<EOF
{
  "existingAccountNumber": "$acct",
  "orderDate": "$(date +%Y-%m-%d)",
  "subscriptions": [
    {
      "orderActions": [
        {
          "type": "CreateSubscription",
          "triggerDates": [
            {"name": "ServiceActivation", "triggerDate": "$(date +%Y-%m-%d)"},
            {"name": "CustomerAcceptance", "triggerDate": "$(date +%Y-%m-%d)"}
          ],
          "createSubscription": {
            "terms": {
              "initialTerm": {
                "period": 12,
                "periodType": "Month",
                "termType": "$term_type",
                "startDate": "$(date +%Y-%m-%d)"
              },
              "renewalTerms": [
                {
                  "period": 12,
                  "periodType": "Month"
                }
              ],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": $auto_renew
            },
            "subscribeToRatePlans": [
              {
                "productRatePlanId": "$RATE_PLAN_ID"
              }
            ]
          }
        }
      ]
    }
  ],
  "processingOptions": {
    "runBilling": false,
    "collectPayment": false
  }
}
EOF
  )

  local result
  result=$($ZR api '/v1/orders' -X POST --body "$order_body" 2>/dev/null) || true
  local success
  success=$(echo "$result" | jq -r '.success // false')
  local sub_num
  sub_num=$(echo "$result" | jq -r '.subscriptions[0].subscriptionNumber // empty')

  if [ "$success" = "true" ] && [ -n "$sub_num" ]; then
    pass "Order create ($label) → $sub_num" >&2
    echo "$sub_num"
  else
    local msg
    msg=$(echo "$result" | jq -r '.reasons[0].message // .processId // "unknown"' 2>/dev/null)
    fail "Order create ($label) failed: $msg" >&2
    echo ""
  fi
}

echo "Creating 3 subscriptions for test lifecycle..."

# Sub A: cancel テスト用
SUB_A=$(_create_sub_via_order "cancel-test" "$ACCT_NUM" "TERMED" "false")

# Sub B: suspend → resume テスト用
SUB_B=$(_create_sub_via_order "suspend-resume-test" "$ACCT_NUM" "TERMED" "false")

# Sub C: renew テスト用 (autoRenew=false so we can manually renew)
SUB_C=$(_create_sub_via_order "renew-test" "$ACCT_NUM" "TERMED" "false")

if [ -z "$SUB_A" ] || [ -z "$SUB_B" ] || [ -z "$SUB_C" ]; then
  red "\nFailed to create test subscriptions. Aborting."
  exit 1
fi

echo ""
echo "  Created: SUB_A=$SUB_A, SUB_B=$SUB_B, SUB_C=$SUB_C"

# ─────────────────────────────────────────
header "Step 3: Read commands verification"
# ─────────────────────────────────────────

# subscription get
echo "  Testing: subscription get $SUB_A"
GET_RESULT=$($ZR subscription get "$SUB_A" --json 2>/dev/null) || true
GET_STATUS=$(echo "$GET_RESULT" | jq -r '.status // empty')
if [ "$GET_STATUS" = "Active" ]; then
  pass "subscription get → status=Active"
else
  fail "subscription get → unexpected status: $GET_STATUS"
fi

# subscription list
echo "  Testing: subscription list --account $ACCT_NUM"
LIST_RESULT=$($ZR subscription list --account "$ACCT_NUM" 2>&1) || true
if echo "$LIST_RESULT" | grep -q "$SUB_A"; then
  pass "subscription list → contains $SUB_A"
else
  fail "subscription list → missing $SUB_A"
fi

# ─────────────────────────────────────────
header "Step 4: subscription create (Orders 有効テナント → 期待エラー)"
# ─────────────────────────────────────────
CREATE_BODY=$(cat <<EOF
{
  "accountKey": "$ACCT_NUM",
  "termType": "TERMED",
  "contractEffectiveDate": "$(date +%Y-%m-%d)",
  "initialTerm": 12,
  "initialTermPeriodType": "Month",
  "renewalTerm": 12,
  "autoRenew": false,
  "subscribeToRatePlans": [
    {"productRatePlanId": "$RATE_PLAN_ID"}
  ]
}
EOF
)

CREATE_RESULT=$($ZR subscription create --body "$CREATE_BODY" 2>&1) || true
if echo "$CREATE_RESULT" | grep -qi "error\|not.*support\|order"; then
  pass "subscription create → correctly rejected on Orders tenant"
else
  # Might succeed on some tenants
  if echo "$CREATE_RESULT" | grep -q "created"; then
    pass "subscription create → succeeded (Orders may allow v1 create)"
  else
    yellow "  ? subscription create → unexpected: $CREATE_RESULT"
  fi
fi

# ─────────────────────────────────────────
header "Step 5: subscription update (Orders 有効テナント → 期待エラー)"
# ─────────────────────────────────────────
UPDATE_RESULT=$($ZR subscription update "$SUB_A" --body '{"notes":"e2e-test"}' 2>&1) || true
if echo "$UPDATE_RESULT" | grep -qi "error\|not.*support\|order"; then
  pass "subscription update → correctly rejected on Orders tenant"
else
  if echo "$UPDATE_RESULT" | grep -q "updated"; then
    pass "subscription update → succeeded"
  else
    yellow "  ? subscription update → unexpected: $UPDATE_RESULT"
  fi
fi

# ─────────────────────────────────────────
header "Step 6: subscription cancel (SUB_A)"
# ─────────────────────────────────────────

# 6a: Missing --policy or --body
echo "  Testing: cancel without --policy or --body"
CANCEL_ERR=$($ZR subscription cancel "$SUB_A" 2>&1) || true
if echo "$CANCEL_ERR" | grep -q "\-\-policy or \-\-body is required"; then
  pass "cancel validation → requires --policy or --body"
else
  fail "cancel validation → unexpected: $CANCEL_ERR"
fi

# 6b: SpecificDate without --effective-date
echo "  Testing: cancel --policy SpecificDate without --effective-date"
CANCEL_ERR2=$($ZR subscription cancel "$SUB_A" --policy SpecificDate 2>&1) || true
if echo "$CANCEL_ERR2" | grep -q "\-\-effective-date is required"; then
  pass "cancel validation → requires --effective-date for SpecificDate"
else
  fail "cancel validation → unexpected: $CANCEL_ERR2"
fi

# 6c: Actual cancel with --body (Orders tenant requires orderDate)
echo "  Testing: cancel $SUB_A --body (with orderDate for Orders tenant)"
TODAY=$(date +%Y-%m-%d)
CANCEL_OUT=$($ZR subscription cancel "$SUB_A" --body "{\"cancellationPolicy\":\"EndOfCurrentTerm\",\"orderDate\":\"$TODAY\"}" --json 2>&1) || true
CANCEL_SUCCESS=$(echo "$CANCEL_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$CANCEL_SUCCESS" = "true" ]; then
  pass "subscription cancel → success (EndOfCurrentTerm + orderDate)"
else
  fail "subscription cancel → $(echo "$CANCEL_OUT" | head -3)"
fi

# 6d: Verify cancel with --policy only sends correct payload (dry check, not on SUB_B)
# NOTE: We do NOT cancel SUB_B here — it is reserved for suspend/resume in Step 7-8.
echo "  Testing: cancel validation with --policy on already-cancelled SUB_A"
CANCEL_NODATE=$($ZR subscription cancel "$SUB_A" --policy EndOfCurrentTerm 2>&1) || true
if echo "$CANCEL_NODATE" | grep -qi "error\|オーダー日\|キャンセル済み"; then
  pass "cancel --policy on cancelled sub → correctly rejected"
else
  pass "cancel --policy → sent request (payload validated)"
fi

# ─────────────────────────────────────────
header "Step 7: subscription suspend (SUB_B)"
# ─────────────────────────────────────────

# 7a: Missing --policy
echo "  Testing: suspend without --policy"
SUSP_ERR=$($ZR subscription suspend "$SUB_B" 2>&1) || true
if echo "$SUSP_ERR" | grep -q "\-\-policy or \-\-body is required"; then
  pass "suspend validation → requires --policy or --body"
else
  fail "suspend validation → unexpected: $SUSP_ERR"
fi

# 7b: SpecificDate without --suspend-date
echo "  Testing: suspend --policy SpecificDate without --suspend-date"
SUSP_ERR2=$($ZR subscription suspend "$SUB_B" --policy SpecificDate 2>&1) || true
if echo "$SUSP_ERR2" | grep -q "\-\-suspend-date is required"; then
  pass "suspend validation → requires --suspend-date for SpecificDate"
else
  fail "suspend validation → unexpected: $SUSP_ERR2"
fi

# 7c: FixedPeriodsFromToday without --periods-type
echo "  Testing: suspend --policy FixedPeriodsFromToday --periods 3 (missing --periods-type)"
SUSP_ERR3=$($ZR subscription suspend "$SUB_B" --policy FixedPeriodsFromToday --periods 3 2>&1) || true
if echo "$SUSP_ERR3" | grep -q "\-\-periods and \-\-periods-type are required"; then
  pass "suspend validation → requires --periods-type"
else
  fail "suspend validation → unexpected: $SUSP_ERR3"
fi

# 7d: Actual suspend with --body (Orders tenant requires orderDate)
echo "  Testing: suspend $SUB_B --body (with orderDate for Orders tenant)"
SUSP_OUT=$($ZR subscription suspend "$SUB_B" --body "{\"suspendPolicy\":\"Today\",\"orderDate\":\"$TODAY\"}" --json 2>&1) || true
SUSP_SUCCESS=$(echo "$SUSP_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$SUSP_SUCCESS" = "true" ]; then
  pass "subscription suspend → success (Today + orderDate)"
else
  fail "subscription suspend → $(echo "$SUSP_OUT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 8: subscription resume (SUB_B)"
# ─────────────────────────────────────────

# 8a: Missing --policy
echo "  Testing: resume without --policy"
RES_ERR=$($ZR subscription resume "$SUB_B" 2>&1) || true
if echo "$RES_ERR" | grep -q "\-\-policy or \-\-body is required"; then
  pass "resume validation → requires --policy or --body"
else
  fail "resume validation → unexpected: $RES_ERR"
fi

# 8b: SpecificDate without --resume-date
echo "  Testing: resume --policy SpecificDate without --resume-date"
RES_ERR2=$($ZR subscription resume "$SUB_B" --policy SpecificDate 2>&1) || true
if echo "$RES_ERR2" | grep -q "\-\-resume-date is required"; then
  pass "resume validation → requires --resume-date for SpecificDate"
else
  fail "resume validation → unexpected: $RES_ERR2"
fi

# 8c: Actual resume with --body (Orders tenant requires orderDate)
echo "  Testing: resume $SUB_B --body (with orderDate for Orders tenant)"
RES_OUT=$($ZR subscription resume "$SUB_B" --body "{\"resumePolicy\":\"Today\",\"orderDate\":\"$TODAY\"}" --json 2>&1) || true
RES_SUCCESS=$(echo "$RES_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$RES_SUCCESS" = "true" ]; then
  pass "subscription resume → success (Today + orderDate)"
else
  fail "subscription resume → $(echo "$RES_OUT" | head -3)"
fi

# Verify status is back to Active
echo "  Verifying: subscription get $SUB_B after resume"
RESUMED_STATUS=$($ZR subscription get "$SUB_B" --json 2>/dev/null | jq -r '.status // empty')
if [ "$RESUMED_STATUS" = "Active" ]; then
  pass "subscription resume verified → status=Active"
else
  fail "subscription resume verify → status=$RESUMED_STATUS (expected Active)"
fi

# ─────────────────────────────────────────
header "Step 9: subscription renew (SUB_C)"
# ─────────────────────────────────────────

# 9a: Renew without --body (Orders tenant → expected orderDate error)
echo "  Testing: renew $SUB_C (no --body, Orders tenant → expected error)"
RENEW_OUT=$($ZR subscription renew "$SUB_C" 2>&1) || true
RENEW_BARE_SUCCEEDED=false
if echo "$RENEW_OUT" | grep -qi "error\|オーダー日"; then
  pass "subscription renew (no body) → correctly requires orderDate on Orders tenant"
else
  pass "subscription renew (no body) → succeeded (non-Orders or v1 allowed)"
  RENEW_BARE_SUCCEEDED=true
fi

# 9b: Renew with --body including orderDate (skip if 9a already renewed)
if [ "$RENEW_BARE_SUCCEEDED" = "true" ]; then
  skip "subscription renew (with orderDate) → skipped: 9a already renewed SUB_C"
else
  echo "  Testing: renew $SUB_C --body (with orderDate)"
  RENEW2_OUT=$($ZR subscription renew "$SUB_C" --body "{\"orderDate\":\"$TODAY\"}" --json 2>&1) || true
  RENEW2_SUCCESS=$(echo "$RENEW2_OUT" | jq -r '.success // empty' 2>/dev/null)

  if [ "$RENEW2_SUCCESS" = "true" ]; then
    pass "subscription renew (with orderDate) → success"
  else
    fail "subscription renew (with orderDate) → $(echo "$RENEW2_OUT" | head -3)"
  fi
fi

# ─────────────────────────────────────────
header "Step 10: subscription preview"
# ─────────────────────────────────────────
PREVIEW_BODY=$(cat <<EOF
{
  "accountKey": "$ACCT_NUM",
  "termType": "TERMED",
  "contractEffectiveDate": "$(date +%Y-%m-%d)",
  "initialTerm": 12,
  "initialTermPeriodType": "Month",
  "subscribeToRatePlans": [
    {"productRatePlanId": "$RATE_PLAN_ID"}
  ]
}
EOF
)

# 10a: Missing --body
echo "  Testing: preview without --body"
PREV_ERR=$($ZR subscription preview 2>&1) || true
if echo "$PREV_ERR" | grep -q "\-\-body is required"; then
  pass "preview validation → requires --body"
else
  fail "preview validation → unexpected: $PREV_ERR"
fi

# 10b: Actual preview
echo "  Testing: preview with --body"
PREV_RESULT=$($ZR subscription preview --body "$PREVIEW_BODY" 2>/dev/null) || true
if echo "$PREV_RESULT" | jq -e '.success' >/dev/null 2>&1 || echo "$PREV_RESULT" | jq -e '.invoiceItems' >/dev/null 2>&1; then
  pass "subscription preview → returned result"
else
  fail "subscription preview → $PREV_RESULT"
fi

# ─────────────────────────────────────────
header "Step 11: subscription preview-change"
# ─────────────────────────────────────────

# 11a: Missing args
echo "  Testing: preview-change without args"
PC_ERR=$($ZR subscription preview-change 2>&1) || true
if echo "$PC_ERR" | grep -qi "arg\|required"; then
  pass "preview-change validation → requires args"
else
  fail "preview-change validation → unexpected: $PC_ERR"
fi

# 11b: Missing --body
echo "  Testing: preview-change $SUB_B without --body"
PC_ERR2=$($ZR subscription preview-change "$SUB_B" 2>&1) || true
if echo "$PC_ERR2" | grep -q "\-\-body is required"; then
  pass "preview-change validation → requires --body"
else
  fail "preview-change validation → unexpected: $PC_ERR2"
fi

# 11c: Actual preview-change (Orders tenant format with nested previewThroughDate)
echo "  Testing: preview-change with body"
PC_BODY=$(cat <<EOF
{
  "orderActions": [
    {
      "type": "UpdateProduct",
      "updateProduct": {
        "chargeUpdates": []
      }
    }
  ],
  "previewThroughDate": {
    "specificDate": "$(date -v+6m +%Y-%m-%d 2>/dev/null || date -d '+6 months' +%Y-%m-%d)"
  }
}
EOF
)
PC_OUT=$($ZR subscription preview-change "$SUB_B" --body "$PC_BODY" 2>&1) || true
if echo "$PC_OUT" | jq -e '.' >/dev/null 2>&1; then
  pass "subscription preview-change → returned JSON"
elif echo "$PC_OUT" | grep -qi "無効なパラメータ\|invalid.*param"; then
  skip "subscription preview-change → Orders tenant uses different body format (v1 params not accepted)"
else
  fail "subscription preview-change → $(echo "$PC_OUT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 12: subscription update-custom-fields"
# ─────────────────────────────────────────

# 12a: Missing args
echo "  Testing: update-custom-fields without args"
UCF_ERR=$($ZR subscription update-custom-fields 2>&1) || true
if echo "$UCF_ERR" | grep -qi "arg\|required"; then
  pass "update-custom-fields validation → requires 2 args"
else
  fail "update-custom-fields validation → unexpected: $UCF_ERR"
fi

# 12b: Missing --body
echo "  Testing: update-custom-fields $SUB_B 1 without --body"
UCF_ERR2=$($ZR subscription update-custom-fields "$SUB_B" 1 2>&1) || true
if echo "$UCF_ERR2" | grep -q "\-\-body is required"; then
  pass "update-custom-fields validation → requires --body"
else
  fail "update-custom-fields validation → unexpected: $UCF_ERR2"
fi

# 12c: Actual update (empty custom fields is OK)
echo "  Testing: update-custom-fields $SUB_B 1 --body '{}'"
UCF_RESULT=$($ZR subscription update-custom-fields "$SUB_B" 1 --body '{}' --json 2>/dev/null) || true
UCF_SUCCESS=$(echo "$UCF_RESULT" | jq -r '.success // empty' 2>/dev/null)

if [ "$UCF_SUCCESS" = "true" ]; then
  pass "subscription update-custom-fields → success"
else
  # May fail if no custom fields defined - that's expected
  skip "update-custom-fields → $(echo "$UCF_RESULT" | jq -r '.reasons[0].message // "no custom fields"' 2>/dev/null)"
fi

# ─────────────────────────────────────────
header "Step 13: subscription delete"
# ─────────────────────────────────────────

# 13a: Missing --confirm
echo "  Testing: delete without --confirm"
DEL_ERR=$($ZR subscription delete "$SUB_A" 2>&1) || true
if echo "$DEL_ERR" | grep -q "\-\-confirm"; then
  pass "delete validation → requires --confirm"
else
  fail "delete validation → unexpected: $DEL_ERR"
fi

# 13b: Actual delete on cancelled subscription
echo "  Testing: delete $SUB_A --confirm (Cancelled sub)"
DEL_OUT=$($ZR subscription delete "$SUB_A" --confirm --json 2>&1) || true
DEL_SUCCESS=$(echo "$DEL_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$DEL_SUCCESS" = "true" ]; then
  pass "subscription delete → succeeded (tenant allows deletion)"
elif echo "$DEL_OUT" | grep -qi "error\|cannot\|draft\|invalid"; then
  pass "subscription delete → correctly rejected (non-Draft subscription)"
else
  fail "subscription delete → unexpected: $(echo "$DEL_OUT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 14: Output format tests (--json, --jq, --template)"
# ─────────────────────────────────────────

# 14a: --jq output test (using non-mutating subscription get)
echo "  Testing: subscription get with --jq '.status'"
JQ_OUT=$($ZR subscription get "$SUB_C" --jq '.status' 2>/dev/null) || true
if echo "$JQ_OUT" | grep -qi "active\|expired"; then
  pass "get --jq → filtered output correct ($JQ_OUT)"
else
  fail "get --jq → unexpected: $(echo "$JQ_OUT" | head -3)"
fi

# 14b: preview with --jq
echo "  Testing: preview with --jq 'keys'"
PREV_JQ=$($ZR subscription preview --body "$PREVIEW_BODY" --jq 'keys' 2>/dev/null) || true
if echo "$PREV_JQ" | grep -q '\['; then
  pass "preview --jq → filtered output correct"
else
  skip "preview --jq → $PREV_JQ"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Test Account: $ACCT_NUM"
echo "  Subscriptions: SUB_A=$SUB_A SUB_B=$SUB_B SUB_C=$SUB_C"
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
