#!/bin/bash
# E2E Test: Order Commands (Sub-phase 3c)
# テナント: apac-sandbox (Orders 有効)
#
# Order ライフサイクル: create → get → list → preview、および mutating コマンドの
# 入力バリデーション。happy-path は stdout(JSON) / stderr / 終了コードを分離して捕捉し、
# 失敗時に必ず原因が見えるようにする。
#
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
LOG_FILE="$LOG_DIR/e2e-order-${TIMESTAMP}.log"
source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

# ─────────────────────────────────────────
header "Step 1: Account Setup (テスト用アカウント)"
# ─────────────────────────────────────────
ACCT_BODY=$(cat <<'JSON'
{
  "name": "E2E-Order-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {
    "firstName": "Test",
    "lastName": "Order",
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
  pass "account create → $ACCT_NUM"
else
  fail "account create failed (rc=$RUN_RC): ${RUN_ERR:-$RUN_OUT}"
  printf '\n'
  red "Cannot proceed without a test account. Aborting."
  exit 1
fi

echo "  Account: $ACCT_NUM ($ACCT_ID)"

# ─────────────────────────────────────────
header "Step 2: order create"
# ─────────────────────────────────────────
ORDER_BODY=$(cat <<EOF
{
  "existingAccountNumber": "$ACCT_NUM",
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
              "initialTerm": {"period": 12, "periodType": "Month", "termType": "TERMED", "startDate": "$(date +%Y-%m-%d)"},
              "renewalTerms": [{"period": 12, "periodType": "Month"}],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": false
            },
            "subscribeToRatePlans": [{"productRatePlanId": "$RATE_PLAN_ID"}]
          }
        }
      ]
    }
  ]
}
EOF
)
run $ZR order create --body "$ORDER_BODY" --json
ORDER_NUM=$(echo "$RUN_OUT" | jq -r '.orderNumber // empty' 2>/dev/null)
SUB_NUM=$(echo "$RUN_OUT" | jq -r '.subscriptions[0].subscriptionNumber // empty' 2>/dev/null)

if [ -n "$ORDER_NUM" ]; then
  pass "order create → $ORDER_NUM (sub: $SUB_NUM)"
else
  fail "order create (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 3: order get"
# ─────────────────────────────────────────
if [ -n "$ORDER_NUM" ]; then
  echo "  Testing: order get $ORDER_NUM"
  run $ZR order get "$ORDER_NUM" --json
  GET_STATUS=$(echo "$RUN_OUT" | jq -r '.order.status // .status // empty' 2>/dev/null)
  if [ -n "$GET_STATUS" ]; then
    pass "order get → status=$GET_STATUS"
  else
    fail "order get (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "order get → no order number"
fi

# ─────────────────────────────────────────
header "Step 4: order list"
# ─────────────────────────────────────────
# order list has no account filter (only --status/--page); assert it returns a
# tenant-wide .orders array. The specific created order is verified in Step 3.
echo "  Testing: order list"
run $ZR order list --json
LIST_COUNT=$(echo "$RUN_OUT" | jq -r '.orders | length' 2>/dev/null)
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
  pass "order list → .orders array (count=$LIST_COUNT)"
else
  fail "order list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 5: order list-pending"
# ─────────────────────────────────────────
# list-pending takes a <subscription-key> argument.
echo "  Testing: order list-pending validation (no arg)"
expect_fail "order list-pending validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-pending

if [ -n "$SUB_NUM" ]; then
  echo "  Testing: order list-pending $SUB_NUM"
  run $ZR order list-pending "$SUB_NUM" --json
  if echo "$RUN_OUT" | jq -e '.' >/dev/null 2>&1; then
    pass "order list-pending → returned JSON"
  elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
    skip "order list-pending → Zuora API error: $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
  else
    fail "order list-pending (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "order list-pending → no subscription number from order create"
fi

# ─────────────────────────────────────────
header "Step 6: Output formats"
# ─────────────────────────────────────────
if [ -n "$ORDER_NUM" ]; then
  echo "  Testing: order get --jq '.order.orderNumber'"
  # --jq emits the JSON value verbatim (a quoted string), so assert containment.
  run $ZR order get "$ORDER_NUM" --jq '.order.orderNumber'
  if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | grep -qF "$ORDER_NUM"; then
    pass "order get --jq → $RUN_OUT"
  else
    fail "order get --jq → expected to contain $ORDER_NUM, got '$RUN_OUT' (rc=$RUN_RC) ${RUN_ERR}"
  fi

  echo "  Testing: order get --template '{{.order.orderNumber}}'"
  # --template renders a Go text/template against the JSON; the rendered value
  # is emitted raw (unquoted), so assert exact equality.
  run $ZR order get "$ORDER_NUM" --template '{{.order.orderNumber}}'
  if [ "$RUN_RC" -eq 0 ] && [ "$RUN_OUT" = "$ORDER_NUM" ]; then
    pass "order get --template → $RUN_OUT"
  else
    fail "order get --template → expected $ORDER_NUM, got '$RUN_OUT' (rc=$RUN_RC) ${RUN_ERR}"
  fi
else
  skip "order get --jq → no order number"
  skip "order get --template → no order number"
fi

# ─────────────────────────────────────────
header "Step 6.5: --body resolution (@file / stdin / literal)"
# ─────────────────────────────────────────
# cmdutil.ResolveBody: "@file" reads a file, "-" reads stdin, else literal JSON.
# Every mutating command shares this path, so exercise all three forms (plus the
# @missing error) via account create, which needs no special tenant setup.
BODY_JSON='{"name":"E2E-BodyResolve","currency":"JPY","billCycleDay":1,"autoPay":false,"billToContact":{"firstName":"B","lastName":"R","country":"Japan","state":"Tokyo"}}'

echo "  Testing: --body @file"
BODY_FILE="$LOG_DIR/.body.$$.json"
printf '%s' "$BODY_JSON" > "$BODY_FILE"
run $ZR account create --body "@$BODY_FILE" --json
rm -f "$BODY_FILE"
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.accountNumber' >/dev/null 2>&1; then
  pass "--body @file → created $(echo "$RUN_OUT" | jq -r '.accountNumber')"
else
  fail "--body @file (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: --body - (stdin)"
RUN_OUT=$(printf '%s' "$BODY_JSON" | $ZR account create --body - --json 2>"$LOG_DIR/.berr.$$"); RUN_RC=$?
RUN_ERR=$(cat "$LOG_DIR/.berr.$$" 2>/dev/null); rm -f "$LOG_DIR/.berr.$$"
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.accountNumber' >/dev/null 2>&1; then
  pass "--body - (stdin) → created $(echo "$RUN_OUT" | jq -r '.accountNumber')"
else
  fail "--body - (stdin) (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: --body @nonexistent (error path)"
expect_fail "--body @missing → reading file error" "reading body file:" -- $ZR account create --body @/nonexistent/body.json

# ─────────────────────────────────────────
header "Step 7: order preview (read-only)"
# ─────────────────────────────────────────
PREVIEW_BODY=$(cat <<EOF
{
  "existingAccountNumber": "$ACCT_NUM",
  "orderDate": "$(date +%Y-%m-%d)",
  "previewOptions": {"previewThruType": "SpecificDate", "specificPreviewThruDate": "$(date -v+1m +%Y-%m-%d 2>/dev/null || date -d '+1 month' +%Y-%m-%d)", "previewTypes": ["BillingDocs"]},
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
              "initialTerm": {"period": 12, "periodType": "Month", "termType": "TERMED", "startDate": "$(date +%Y-%m-%d)"},
              "renewalTerms": [{"period": 12, "periodType": "Month"}],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": false
            },
            "subscribeToRatePlans": [{"productRatePlanId": "$RATE_PLAN_ID"}]
          }
        }
      ]
    }
  ]
}
EOF
)
echo "  Testing: order preview"
run $ZR order preview --body "$PREVIEW_BODY" --json
if echo "$RUN_OUT" | jq -e '.success == true' >/dev/null 2>&1; then
  pass "order preview → success ($(echo "$RUN_OUT" | jq -r '.previewResult.invoices | length // 0') invoice(s))"
elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
  # Account/rate-plan-specific rejection — narrow, status-specific skip
  # (not a blanket 'grep -qi error' that would also swallow a CLI bug).
  skip "order preview → Zuora API error: $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
else
  fail "order preview (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 8: Validation (read-only checks)"
# ─────────────────────────────────────────
# All checks below require non-zero exit AND the exact CLI/cobra message. They
# never mutate anything (bad/missing args abort before any API call).

echo "  Testing: order get without arg"
expect_fail "order get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order get

echo "  Testing: order create without --body"
expect_fail "order create validation → requires --body" "--body is required" -- $ZR order create

echo "  Testing: order activate without arg"
expect_fail "order activate validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order activate

echo "  Testing: order cancel without arg"
expect_fail "order cancel validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order cancel

echo "  Testing: order cancel without --confirm"
expect_fail "order cancel validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR order cancel O-FAKE

echo "  Testing: order update without arg"
expect_fail "order update validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order update

echo "  Testing: order update without --body"
expect_fail "order update validation → requires --body" "--body is required" -- $ZR order update O-FAKE

echo "  Testing: order update-custom-fields without arg"
expect_fail "order update-custom-fields validation → requires arg" \
  "accepts 1 arg(s), received 0" -- $ZR order update-custom-fields

echo "  Testing: order update-custom-fields without --body"
expect_fail "order update-custom-fields validation → requires --body" \
  "--body is required" -- $ZR order update-custom-fields O-FAKE

echo "  Testing: order update-trigger-dates without arg"
expect_fail "order update-trigger-dates validation → requires arg" \
  "accepts 1 arg(s), received 0" -- $ZR order update-trigger-dates

echo "  Testing: order update-trigger-dates without --body"
expect_fail "order update-trigger-dates validation → requires --body" \
  "--body is required" -- $ZR order update-trigger-dates O-FAKE

echo "  Testing: order revert without arg"
expect_fail "order revert validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order revert

echo "  Testing: order create-async without --body"
expect_fail "order create-async validation → requires --body" "--body is required" -- $ZR order create-async

echo "  Testing: order preview-async without --body"
expect_fail "order preview-async validation → requires --body" "--body is required" -- $ZR order preview-async

echo "  Testing: order delete-async without arg"
expect_fail "order delete-async validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order delete-async

echo "  Testing: order delete without arg"
expect_fail "order delete validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order delete

echo "  Testing: order delete without --confirm"
expect_fail "order delete validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR order delete O-FAKE

echo "  Testing: order job-status without arg"
expect_fail "order job-status validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order job-status

echo "  Testing: order-action update without arg"
expect_fail "order-action update validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order-action update

echo "  Testing: order-line-item get without arg"
expect_fail "order-line-item get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order-line-item get

echo "  Testing: order-line-item update without arg"
expect_fail "order-line-item update validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order-line-item update

echo "  Testing: order-line-item bulk-update without --body"
expect_fail "order-line-item bulk-update validation → requires --body" "--body is required" -- $ZR order-line-item bulk-update

# ─────────────────────────────────────────
header "Step 9: order list-by-* (live reads)"
# ─────────────────────────────────────────
echo "  Testing: order list-by-subscription validation (no arg)"
expect_fail "order list-by-subscription validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-by-subscription

echo "  Testing: order list-by-invoice-owner validation (no arg)"
expect_fail "order list-by-invoice-owner validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-by-invoice-owner

echo "  Testing: order list-by-subscription-owner validation (no arg)"
expect_fail "order list-by-subscription-owner validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-by-subscription-owner

if [ -n "$SUB_NUM" ]; then
  echo "  Testing: order list-by-subscription $SUB_NUM"
  run $ZR order list-by-subscription "$SUB_NUM" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
    pass "order list-by-subscription → .orders array"
  else
    fail "order list-by-subscription (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "order list-by-subscription → no subscription number"
fi

echo "  Testing: order list-by-subscription-owner $ACCT_NUM"
run $ZR order list-by-subscription-owner "$ACCT_NUM" --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
  pass "order list-by-subscription-owner → .orders array"
else
  fail "order list-by-subscription-owner (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: order list-by-invoice-owner $ACCT_NUM"
run $ZR order list-by-invoice-owner "$ACCT_NUM" --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
  pass "order list-by-invoice-owner → .orders array"
else
  fail "order list-by-invoice-owner (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Test Account: $ACCT_NUM"
echo "  Order: $ORDER_NUM"
echo ""
print_summary
