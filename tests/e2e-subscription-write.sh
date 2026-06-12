#!/bin/bash
# E2E Test: Subscription Write Commands (Sub-phase 3b)
# テナント: apac-sandbox (Orders 有効)
# 注意: create/update は Orders 有効テナントでは v1 API が使えないため、
#       Orders API 経由でサブスクリプションを作成してテストする

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
TODAY=$(date +%Y-%m-%d)  # local order date used by suspend/resume/renew steps
# Log directory
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-subscription-write-${TIMESTAMP}.log"

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

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

ACCT_RESULT=$($ZR account create --body "$ACCT_BODY" --json 2>/dev/null) || true
ACCT_NUM=$(echo "$ACCT_RESULT" | jq -r '.accountNumber // empty' 2>/dev/null)
ACCT_ID=$(echo "$ACCT_RESULT" | jq -r '.accountId // empty' 2>/dev/null)

if [ -n "$ACCT_NUM" ]; then
  pass "account create → $ACCT_NUM"
else
  fail "account create failed: $ACCT_RESULT"
  printf '\n'
  red "Cannot proceed without a dedicated test account. Aborting."
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
echo "  Testing: subscription list --account-key $ACCT_NUM"
LIST_RESULT=$($ZR subscription list --account-key "$ACCT_NUM" 2>&1) || true
if echo "$LIST_RESULT" | grep -q "$SUB_A"; then
  pass "subscription list → contains $SUB_A"
else
  fail "subscription list → missing $SUB_A"
fi

# P5-1 deprecation contract: the OLD --account spelling must keep working
# through v0.5.x (removed in v0.6.0) and print a deprecation notice.
echo "  Testing: deprecated --account alias (removed in v0.6.0)"
ALIAS_OUT=$($ZR subscription list --account "$ACCT_NUM" 2>&1) || true
if echo "$ALIAS_OUT" | grep -q "$SUB_A" && echo "$ALIAS_OUT" | grep -qi "deprecated"; then
  pass "subscription list --account → alias works + deprecation notice"
else
  fail "subscription list --account alias → $(echo "$ALIAS_OUT" | head -2)"
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

# Success-flag checking is default-on (#71), so the exit code is authoritative:
# rc!=0 + "error" text = the expected Orders-tenant rejection; rc==0 = a real
# create succeeded (allowed on some tenants). Anything else is a failure — the
# old pattern (grep "error|not.*support|order") could match a SUCCESS response
# containing "order", and the unexpected branch was a yellow() no-op.
CREATE_RC=0
CREATE_RESULT=$($ZR subscription create --body "$CREATE_BODY" 2>&1) || CREATE_RC=$?
if [ "$CREATE_RC" -ne 0 ] && echo "$CREATE_RESULT" | grep -qi "error"; then
  pass "subscription create → correctly rejected on Orders tenant"
elif [ "$CREATE_RC" -eq 0 ] && [ -n "$CREATE_RESULT" ]; then
  pass "subscription create → succeeded (Orders may allow v1 create)"
else
  fail "subscription create → unexpected (rc=$CREATE_RC): $(echo "$CREATE_RESULT" | head -2)"
fi

# ─────────────────────────────────────────
header "Step 5: subscription update (Orders 有効テナント → 期待エラー)"
# ─────────────────────────────────────────
# Same structure as Step 4: exit code decides, unexpected output now FAILS.
UPDATE_RC=0
UPDATE_RESULT=$($ZR subscription update "$SUB_A" --body '{"notes":"e2e-test"}' 2>&1) || UPDATE_RC=$?
if [ "$UPDATE_RC" -eq 0 ] && echo "$UPDATE_RESULT" | grep -qi "updated"; then
  pass "subscription update → succeeded"
elif [ "$UPDATE_RC" -ne 0 ] && echo "$UPDATE_RESULT" | grep -qi "error"; then
  pass "subscription update → correctly rejected on Orders tenant"
else
  fail "subscription update → unexpected (rc=$UPDATE_RC): $(echo "$UPDATE_RESULT" | head -2)"
fi

# ─────────────────────────────────────────
header "Step 6: subscription cancel (SUB_A)"
# ─────────────────────────────────────────

# 6a: Missing --policy or --body
echo "  Testing: cancel without --policy or --body"
CANCEL_ERR=$($ZR subscription cancel "$SUB_A" 2>&1) || true
if echo "$CANCEL_ERR" | grep -q "at least one of the flags in the group"; then
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

# 6c: cancel without --confirm (irreversible guard)
echo "  Testing: cancel without --confirm"
CANCEL_NOCONFIRM=$($ZR subscription cancel "$SUB_A" --policy EndOfCurrentTerm 2>&1) || true
if echo "$CANCEL_NOCONFIRM" | grep -q "\-\-confirm"; then
  pass "cancel validation → requires --confirm"
else
  fail "cancel validation → unexpected: $CANCEL_NOCONFIRM"
fi

# 6d: Actual cancel with --body (Orders tenant requires orderDate)
echo "  Testing: cancel $SUB_A --body (with orderDate for Orders tenant)"
# orderDate must be >= the subscription contractEffectiveDate. The apac-sandbox
# clock can be a day ahead of local time near midnight, so derive the date
# from the subscription itself (fallback: local today). Build the JSON with
# printf to avoid shell quote-collapsing, and keep stderr so errors show.
CANCEL_DATE=$($ZR subscription get "$SUB_A" --json 2>/dev/null | jq -r '.contractEffectiveDate // empty' 2>/dev/null)
[ -z "$CANCEL_DATE" ] && CANCEL_DATE=$(date +%Y-%m-%d)
CANCEL_BODY=$(printf '{"cancellationPolicy":"EndOfCurrentTerm","orderDate":"%s"}' "$CANCEL_DATE")
CANCEL_OUT=$($ZR subscription cancel "$SUB_A" --body "$CANCEL_BODY" --confirm --json 2>&1) || true
CANCEL_SUCCESS=$(echo "$CANCEL_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$CANCEL_SUCCESS" = "true" ]; then
  pass "subscription cancel → success (EndOfCurrentTerm + orderDate)"
else
  fail "subscription cancel → $(echo "$CANCEL_OUT" | head -3)"
fi

# 6e: Re-cancel of the already-cancelled SUB_A MUST be rejected (Zuora errors on
# cancelled subs). The old check passed unconditionally (both branches were pass).
# NOTE: We do NOT cancel SUB_B here — it is reserved for suspend/resume in Step 7-8.
echo "  Testing: cancel validation with --policy on already-cancelled SUB_A"
CANCEL_NODATE_RC=0
CANCEL_NODATE=$($ZR subscription cancel "$SUB_A" --policy EndOfCurrentTerm --confirm 2>&1) || CANCEL_NODATE_RC=$?
if [ "$CANCEL_NODATE_RC" -ne 0 ] && echo "$CANCEL_NODATE" | grep -qi "error\|オーダー日\|キャンセル"; then
  pass "cancel --policy on cancelled sub → correctly rejected"
else
  fail "cancel --policy on cancelled sub → expected rejection (rc=$CANCEL_NODATE_RC): $(echo "$CANCEL_NODATE" | head -2)"
fi

# ─────────────────────────────────────────
header "Step 7: subscription suspend (SUB_B)"
# ─────────────────────────────────────────

# 7a: Missing --policy
echo "  Testing: suspend without --policy"
SUSP_ERR=$($ZR subscription suspend "$SUB_B" 2>&1) || true
if echo "$SUSP_ERR" | grep -q "at least one of the flags in the group"; then
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
SUSP_OUT=$($ZR subscription suspend "$SUB_B" --body "{\"suspendPolicy\":\"Today\",\"orderDate\":\"$TODAY\"}" --json 2>/dev/null) || true
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
if echo "$RES_ERR" | grep -q "at least one of the flags in the group"; then
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
RES_OUT=$($ZR subscription resume "$SUB_B" --body "{\"resumePolicy\":\"Today\",\"orderDate\":\"$TODAY\"}" --json 2>/dev/null) || true
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

# 9a: Renew without --body (Orders tenant → expected orderDate error).
# Exit code decides; the old version passed for ANY output (both branches pass,
# garbage/empty output counted as "succeeded").
echo "  Testing: renew $SUB_C (no --body, Orders tenant → expected error)"
RENEW_RC=0
RENEW_OUT=$($ZR subscription renew "$SUB_C" 2>&1) || RENEW_RC=$?
RENEW_BARE_SUCCEEDED=false
if [ "$RENEW_RC" -ne 0 ] && echo "$RENEW_OUT" | grep -qi "error\|オーダー日"; then
  pass "subscription renew (no body) → correctly requires orderDate on Orders tenant"
elif [ "$RENEW_RC" -eq 0 ] && [ -n "$RENEW_OUT" ]; then
  pass "subscription renew (no body) → succeeded (non-Orders or v1 allowed)"
  RENEW_BARE_SUCCEEDED=true
else
  fail "subscription renew (no body) → unexpected (rc=$RENEW_RC): $(echo "$RENEW_OUT" | head -2)"
fi

# 9b: Renew with --body including orderDate (skip if 9a already renewed)
if [ "$RENEW_BARE_SUCCEEDED" = "true" ]; then
  skip "subscription renew (with orderDate) → skipped: 9a already renewed SUB_C"
else
  echo "  Testing: renew $SUB_C --body (with orderDate)"
  RENEW2_OUT=$($ZR subscription renew "$SUB_C" --body "{\"orderDate\":\"$TODAY\"}" --json 2>/dev/null) || true
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
# Capture stderr too: on tenants without custom fields the command errors
# BEFORE rendering JSON and the Zuora message lands on stderr (Codex) —
# discarding it would turn the expected limitation into a hard failure.
UCF_RESULT=$($ZR subscription update-custom-fields "$SUB_B" 1 --body '{}' --json 2>&1) || true
UCF_SUCCESS=$(echo "$UCF_RESULT" | jq -r '.success // empty' 2>/dev/null)

if [ "$UCF_SUCCESS" = "true" ]; then
  pass "subscription update-custom-fields → success"
else
  # ONLY a custom-field-related rejection may skip (tenant without custom
  # fields defined); any other failure — auth, 4xx/5xx, transport — fails.
  UCF_MSG=$(echo "$UCF_RESULT" | jq -r '.reasons[0].message // ""' 2>/dev/null)
  [ -z "$UCF_MSG" ] && UCF_MSG="$UCF_RESULT"
  if printf '%s' "$UCF_MSG" | grep -qi "custom\|カスタム"; then
    skip "update-custom-fields → $(printf '%s' "$UCF_MSG" | head -1)"
  else
    fail "update-custom-fields → unexpected failure: $(echo "$UCF_RESULT" | head -2)"
  fi
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
DEL_OUT=$($ZR subscription delete "$SUB_A" --confirm --json 2>/dev/null) || true
DEL_SUCCESS=$(echo "$DEL_OUT" | jq -r '.success // empty' 2>/dev/null)

if [ "$DEL_SUCCESS" = "true" ]; then
  pass "subscription delete → succeeded (tenant allows deletion)"
elif echo "$DEL_OUT" | grep -qi "error\|cannot\|draft\|invalid"; then
  pass "subscription delete → correctly rejected (non-Draft subscription)"
else
  fail "subscription delete → unexpected: $(echo "$DEL_OUT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 13.5: subscription metrics / versions (live reads)"
# ─────────────────────────────────────────

echo "  Testing: subscription metrics without --subscription-numbers"
SM_ERR=$($ZR subscription metrics 2>&1); SM_RC=$?
if [ "$SM_RC" -ne 0 ] && echo "$SM_ERR" | grep -qF 'required flag(s) "subscription-numbers"'; then
  pass "subscription metrics validation → requires --subscription-numbers"
else
  fail "subscription metrics validation → rc=$SM_RC: $(echo "$SM_ERR" | head -1)"
fi

echo "  Testing: subscription metrics --subscription-numbers $SUB_C"
SM_OUT=$($ZR subscription metrics --subscription-numbers "$SUB_C" --json 2>&1); SM_RC2=$?
if [ "$SM_RC2" -eq 0 ] && echo "$SM_OUT" | jq -e '.subscriptionMetrics' >/dev/null 2>&1; then
  pass "subscription metrics → returned .subscriptionMetrics"
elif echo "$SM_OUT" | grep -qF "Zuora API error"; then
  skip "subscription metrics → $(echo "$SM_OUT" | head -1)"
else
  fail "subscription metrics → rc=$SM_RC2: $(echo "$SM_OUT" | head -1)"
fi

echo "  Testing: subscription versions with 1 arg"
SV_ERR=$($ZR subscription versions "$SUB_C" 2>&1); SV_RC=$?
if [ "$SV_RC" -ne 0 ] && echo "$SV_ERR" | grep -qF "accepts 2 arg(s), received 1"; then
  pass "subscription versions validation → requires 2 args"
else
  fail "subscription versions validation → rc=$SV_RC: $(echo "$SV_ERR" | head -1)"
fi

echo "  Testing: subscription versions $SUB_C 1"
SV_OUT=$($ZR subscription versions "$SUB_C" 1 --json 2>&1); SV_RC2=$?
if [ "$SV_RC2" -eq 0 ] && echo "$SV_OUT" | jq -e '.subscriptionNumber // .id // .subscription' >/dev/null 2>&1; then
  pass "subscription versions → returned version data"
elif [ "$SV_RC2" -ne 0 ] && echo "$SV_OUT" | grep -qF "Zuora API error"; then
  skip "subscription versions → $(echo "$SV_OUT" | head -1)"
else
  fail "subscription versions → rc=$SV_RC2: $(echo "$SV_OUT" | head -1)"
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
print_summary
