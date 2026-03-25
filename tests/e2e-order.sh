#!/bin/bash
# E2E Test: Order Commands (Phase 4)
# テナント: apac-sandbox (Orders 有効)
# 注意: Orders 有効テナントでのテスト。非 Orders テナントでは多くが期待エラーとなる

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
LOG_FILE="$LOG_DIR/e2e-order-${TIMESTAMP}.log"

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
  "name": "E2E-Order-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {
    "firstName": "Test",
    "lastName": "OrderUser",
    "country": "Japan",
    "state": "Tokyo"
  }
}
JSON
)

ACCT_RESULT=$($ZR account create --body "$ACCT_BODY" --json 2>/dev/null) || true
ACCT_NUM=$(echo "$ACCT_RESULT" | jq -r '.accountNumber // empty' 2>/dev/null)

if [ -n "$ACCT_NUM" ]; then
  pass "account create → $ACCT_NUM"
else
  fail "account create failed: $ACCT_RESULT"
  printf '\n'
  red "Cannot proceed without a dedicated test account. Aborting."
  exit 1
fi

echo "  Account: $ACCT_NUM"

# ─────────────────────────────────────────
header "Step 2: Order Create Validation"
# ─────────────────────────────────────────

# 2a: Missing --body
echo "  Testing: order create without --body"
CREATE_ERR=$($ZR order create 2>&1) || true
if echo "$CREATE_ERR" | grep -q "\-\-body is required"; then
  pass "order create validation → requires --body"
else
  fail "order create validation → unexpected: $CREATE_ERR"
fi

# ─────────────────────────────────────────
header "Step 3: Order Create"
# ─────────────────────────────────────────
TODAY=$(date +%Y-%m-%d)

ORDER_BODY=$(cat <<EOF
{
  "existingAccountNumber": "$ACCT_NUM",
  "orderDate": "$TODAY",
  "subscriptions": [
    {
      "orderActions": [
        {
          "type": "CreateSubscription",
          "triggerDates": [
            {"name": "ServiceActivation", "triggerDate": "$TODAY"},
            {"name": "CustomerAcceptance", "triggerDate": "$TODAY"}
          ],
          "createSubscription": {
            "terms": {
              "initialTerm": {
                "period": 12,
                "periodType": "Month",
                "termType": "TERMED",
                "startDate": "$TODAY"
              },
              "renewalTerms": [
                {
                  "period": 12,
                  "periodType": "Month"
                }
              ],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": false
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

CREATE_RESULT=$($ZR order create --body "$ORDER_BODY" --json 2>/dev/null) || true
ORDER_NUM=$(echo "$CREATE_RESULT" | jq -r '.orderNumber // empty' 2>/dev/null)
CREATE_SUCCESS=$(echo "$CREATE_RESULT" | jq -r '.success // empty' 2>/dev/null)

if [ "$CREATE_SUCCESS" = "true" ] && [ -n "$ORDER_NUM" ]; then
  pass "order create → $ORDER_NUM"
else
  fail "order create → $(echo "$CREATE_RESULT" | head -3)"
  ORDER_NUM=""
fi

echo "  Order: $ORDER_NUM"

# ─────────────────────────────────────────
header "Step 4: Order Get"
# ─────────────────────────────────────────

# 4a: Missing argument
echo "  Testing: order get without argument"
GET_ERR=$($ZR order get 2>&1) || true
if echo "$GET_ERR" | grep -qi "arg\|required"; then
  pass "order get validation → requires argument"
else
  fail "order get validation → unexpected: $GET_ERR"
fi

# 4b: Actual get
if [ -n "$ORDER_NUM" ]; then
  echo "  Testing: order get $ORDER_NUM"
  GET_RESULT=$($ZR order get "$ORDER_NUM" --json 2>/dev/null) || true
  GET_ORDER=$(echo "$GET_RESULT" | jq -r '.order.orderNumber // .orderNumber // empty' 2>/dev/null)
  if [ "$GET_ORDER" = "$ORDER_NUM" ]; then
    pass "order get → orderNumber=$ORDER_NUM"
  else
    fail "order get → unexpected: $(echo "$GET_RESULT" | head -3)"
  fi
else
  skip "order get → no order created"
fi

# ─────────────────────────────────────────
header "Step 5: Order List"
# ─────────────────────────────────────────
echo "  Testing: order list"
LIST_RESULT=$($ZR order list 2>&1) || true
if echo "$LIST_RESULT" | grep -q "ORDER_NUMBER\|orderNumber\|$ORDER_NUM"; then
  pass "order list → returned orders"
else
  # May return empty list or JSON
  if echo "$LIST_RESULT" | jq -e '.orders' >/dev/null 2>&1; then
    pass "order list → returned JSON (may be empty)"
  else
    fail "order list → unexpected: $(echo "$LIST_RESULT" | head -3)"
  fi
fi

# 5b: List with --status filter
echo "  Testing: order list --status Completed"
LIST_STATUS=$($ZR order list --status Completed --json 2>/dev/null) || true
if echo "$LIST_STATUS" | jq -e '.' >/dev/null 2>&1; then
  pass "order list --status → returned JSON"
else
  fail "order list --status → unexpected"
fi

# ─────────────────────────────────────────
header "Step 6: Order Query Commands"
# ─────────────────────────────────────────

# 6a: list-by-subscription-owner
echo "  Testing: order list-by-subscription-owner $ACCT_NUM"
LSO_RESULT=$($ZR order list-by-subscription-owner "$ACCT_NUM" --json 2>/dev/null) || true
if echo "$LSO_RESULT" | jq -e '.' >/dev/null 2>&1; then
  pass "order list-by-subscription-owner → returned JSON"
else
  fail "order list-by-subscription-owner → unexpected"
fi

# 6b: list-by-invoice-owner
echo "  Testing: order list-by-invoice-owner $ACCT_NUM"
LIO_RESULT=$($ZR order list-by-invoice-owner "$ACCT_NUM" --json 2>/dev/null) || true
if echo "$LIO_RESULT" | jq -e '.' >/dev/null 2>&1; then
  pass "order list-by-invoice-owner → returned JSON"
else
  fail "order list-by-invoice-owner → unexpected"
fi

# 6c: list-by-subscription (need subscription key from order)
if [ -n "$ORDER_NUM" ]; then
  SUB_NUM=$(echo "$CREATE_RESULT" | jq -r '.subscriptions[0].subscriptionNumber // empty' 2>/dev/null)
  if [ -n "$SUB_NUM" ]; then
    echo "  Testing: order list-by-subscription $SUB_NUM"
    LBS_RESULT=$($ZR order list-by-subscription "$SUB_NUM" --json 2>/dev/null) || true
    if echo "$LBS_RESULT" | jq -e '.' >/dev/null 2>&1; then
      pass "order list-by-subscription → returned JSON"
    else
      fail "order list-by-subscription → unexpected"
    fi

    # 6d: list-pending
    echo "  Testing: order list-pending $SUB_NUM"
    LP_RESULT=$($ZR order list-pending "$SUB_NUM" --json 2>/dev/null) || true
    if echo "$LP_RESULT" | jq -e '.' >/dev/null 2>&1; then
      pass "order list-pending → returned JSON"
    else
      fail "order list-pending → unexpected"
    fi
  else
    skip "order list-by-subscription → no subscription number from order create"
    skip "order list-pending → no subscription number from order create"
  fi
else
  skip "order list-by-subscription → no order created"
  skip "order list-pending → no order created"
fi

# ─────────────────────────────────────────
header "Step 7: Order Preview Validation"
# ─────────────────────────────────────────

# 7a: Missing --body
echo "  Testing: order preview without --body"
PREV_ERR=$($ZR order preview 2>&1) || true
if echo "$PREV_ERR" | grep -q "\-\-body is required"; then
  pass "order preview validation → requires --body"
else
  fail "order preview validation → unexpected: $PREV_ERR"
fi

# 7b: Actual preview
echo "  Testing: order preview"
PREVIEW_BODY=$(cat <<EOF
{
  "existingAccountNumber": "$ACCT_NUM",
  "orderDate": "$TODAY",
  "subscriptions": [
    {
      "orderActions": [
        {
          "type": "CreateSubscription",
          "triggerDates": [
            {"name": "ServiceActivation", "triggerDate": "$TODAY"},
            {"name": "CustomerAcceptance", "triggerDate": "$TODAY"}
          ],
          "createSubscription": {
            "terms": {
              "initialTerm": {
                "period": 12,
                "periodType": "Month",
                "termType": "TERMED",
                "startDate": "$TODAY"
              },
              "renewalTerms": [{"period": 12, "periodType": "Month"}],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": false
            },
            "subscribeToRatePlans": [
              {"productRatePlanId": "$RATE_PLAN_ID"}
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

PREV_RESULT=$($ZR order preview --body "$PREVIEW_BODY" 2>&1) || true
if echo "$PREV_RESULT" | jq -e '.' >/dev/null 2>&1; then
  pass "order preview → returned JSON"
elif echo "$PREV_RESULT" | grep -qi "error\|not.*support\|invalid"; then
  skip "order preview → API error (may not be supported on this tenant)"
else
  fail "order preview → $(echo "$PREV_RESULT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 8: Order Delete Validation"
# ─────────────────────────────────────────

# 8a: Missing --confirm
echo "  Testing: order delete without --confirm"
DEL_ERR=$($ZR order delete O-NONEXIST 2>&1) || true
if echo "$DEL_ERR" | grep -q "\-\-confirm"; then
  pass "order delete validation → requires --confirm"
else
  fail "order delete validation → unexpected: $DEL_ERR"
fi

# 8b: Missing argument
echo "  Testing: order delete without argument"
DEL_ERR2=$($ZR order delete --confirm 2>&1) || true
if echo "$DEL_ERR2" | grep -qi "arg\|required"; then
  pass "order delete validation → requires argument"
else
  fail "order delete validation → unexpected: $DEL_ERR2"
fi

# ─────────────────────────────────────────
header "Step 9: Order Revert Validation"
# ─────────────────────────────────────────

# 9a: Missing --body
echo "  Testing: order revert without --body"
REV_ERR=$($ZR order revert O-NONEXIST 2>&1) || true
if echo "$REV_ERR" | grep -q "\-\-body is required"; then
  pass "order revert validation → requires --body"
else
  fail "order revert validation → unexpected: $REV_ERR"
fi

# ─────────────────────────────────────────
header "Step 10: Async Command Validation"
# ─────────────────────────────────────────

# 10a: create-async missing --body
echo "  Testing: order create-async without --body"
CA_ERR=$($ZR order create-async 2>&1) || true
if echo "$CA_ERR" | grep -q "\-\-body is required"; then
  pass "order create-async validation → requires --body"
else
  fail "order create-async validation → unexpected: $CA_ERR"
fi

# 10b: preview-async missing --body
echo "  Testing: order preview-async without --body"
PA_ERR=$($ZR order preview-async 2>&1) || true
if echo "$PA_ERR" | grep -q "\-\-body is required"; then
  pass "order preview-async validation → requires --body"
else
  fail "order preview-async validation → unexpected: $PA_ERR"
fi

# 10c: delete-async missing argument
echo "  Testing: order delete-async without argument"
DA_ERR=$($ZR order delete-async 2>&1) || true
if echo "$DA_ERR" | grep -qi "arg\|required"; then
  pass "order delete-async validation → requires argument"
else
  fail "order delete-async validation → unexpected: $DA_ERR"
fi

# 10d: job-status missing argument
echo "  Testing: order job-status without argument"
JS_ERR=$($ZR order job-status 2>&1) || true
if echo "$JS_ERR" | grep -qi "arg\|required"; then
  pass "order job-status validation → requires argument"
else
  fail "order job-status validation → unexpected: $JS_ERR"
fi

# ─────────────────────────────────────────
header "Step 11: Order Action / Line Item Validation"
# ─────────────────────────────────────────

# 11a: order-action update missing --body
echo "  Testing: order-action update without --body"
OA_ERR=$($ZR order-action update 2>&1) || true
if echo "$OA_ERR" | grep -qi "arg\|required\|body"; then
  pass "order-action update validation → requires argument and --body"
else
  fail "order-action update validation → unexpected: $OA_ERR"
fi

# 11b: order-line-item get missing argument
echo "  Testing: order-line-item get without argument"
OLI_ERR=$($ZR order-line-item get 2>&1) || true
if echo "$OLI_ERR" | grep -qi "arg\|required"; then
  pass "order-line-item get validation → requires argument"
else
  fail "order-line-item get validation → unexpected: $OLI_ERR"
fi

# 11c: order-line-item update missing --body
echo "  Testing: order-line-item update without --body"
OLI_ERR2=$($ZR order-line-item update someId 2>&1) || true
if echo "$OLI_ERR2" | grep -q "\-\-body is required"; then
  pass "order-line-item update validation → requires --body"
else
  fail "order-line-item update validation → unexpected: $OLI_ERR2"
fi

# 11d: order-line-item bulk-update missing --body
echo "  Testing: order-line-item bulk-update without --body"
OLI_ERR3=$($ZR order-line-item bulk-update 2>&1) || true
if echo "$OLI_ERR3" | grep -q "\-\-body is required"; then
  pass "order-line-item bulk-update validation → requires --body"
else
  fail "order-line-item bulk-update validation → unexpected: $OLI_ERR3"
fi

# ─────────────────────────────────────────
header "Step 12: Output format tests (--json, --jq)"
# ─────────────────────────────────────────

if [ -n "$ORDER_NUM" ]; then
  # 12a: --json output test
  echo "  Testing: order get with --json"
  JSON_OUT=$($ZR order get "$ORDER_NUM" --json 2>/dev/null) || true
  if echo "$JSON_OUT" | jq -e '.order.orderNumber // .orderNumber' >/dev/null 2>&1; then
    pass "order get --json → valid JSON with orderNumber"
  else
    fail "order get --json → unexpected: $(echo "$JSON_OUT" | head -3)"
  fi

  # 12b: --jq output test
  echo "  Testing: order get with --jq '.order.orderNumber'"
  JQ_OUT=$($ZR order get "$ORDER_NUM" --jq '.order.orderNumber // .orderNumber' 2>/dev/null) || true
  if echo "$JQ_OUT" | grep -q "$ORDER_NUM"; then
    pass "order get --jq → filtered output correct"
  else
    fail "order get --jq → unexpected: $JQ_OUT"
  fi
else
  skip "output format tests → no order created"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Test Account: $ACCT_NUM"
echo "  Order: $ORDER_NUM"
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
