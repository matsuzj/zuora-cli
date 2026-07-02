#!/bin/bash
# E2E Test: Invoice & Payment Commands (Sub-phase 3g)
# テナント: apac-sandbox
# 注意: payment gateway 未設定のため payment は作成できず list は空配列(正常)。
#       invoice は runBilling:true のオーダーで必ず生成されるので postcondition として要求する。

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-invoice-payment-${TIMESTAMP}.log"

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# Sandbox resources this suite creates (NO auto-teardown — prune manually after
# a broken run; see docs/e2e-test-skips.md "Manual cleanup after a broken run"):
#   Account "E2E-InvoicePay-Test" (+ its billed invoice/payment).
#   Prune: zr account list  →  zr account delete <account-key> --confirm

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

# ─────────────────────────────────────────
header "Step 1: Validation"
# ─────────────────────────────────────────
echo "  Testing: invoice get without argument"
expect_fail "invoice get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR invoice get

echo "  Testing: invoice items without argument"
expect_fail "invoice items validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR invoice items

echo "  Testing: invoice files without argument"
expect_fail "invoice files validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR invoice files

echo "  Testing: invoice email without argument"
expect_fail "invoice email validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR invoice email

echo "  Testing: payment create without --body"
expect_fail "payment create validation → requires --body" 'required flag(s) "body" not set' -- $ZR payment create

echo "  Testing: payment apply without argument"
expect_fail "payment apply validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR payment apply

echo "  Testing: payment refund without argument"
expect_fail "payment refund validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR payment refund

# payment cancel/unapply/transfer (#430): validation gates only — a live
# cancel/unapply/transfer would mutate real payment records, so like refund
# these are exercised at the validation layer, not with a live mutation.
echo "  Testing: payment cancel without argument"
expect_fail "payment cancel validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR payment cancel

echo "  Testing: payment cancel without --confirm"
expect_fail "payment cancel validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR payment cancel pay-FAKE

echo "  Testing: payment unapply without --body"
expect_fail "payment unapply validation → requires --body" 'required flag(s) "body" not set' -- $ZR payment unapply pay-FAKE

echo "  Testing: payment transfer without --body"
expect_fail "payment transfer validation → requires --body" 'required flag(s) "body" not set' -- $ZR payment transfer pay-FAKE

echo "  Testing: payment get without argument"
expect_fail "payment get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR payment get

echo "  Testing: invoice usage-rate-detail without argument"
expect_fail "invoice usage-rate-detail validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR invoice usage-rate-detail

# ─────────────────────────────────────────
header "Step 2: Account + billed order (creates invoice)"
# ─────────────────────────────────────────
ACCT_BODY=$(cat <<'JSON'
{
  "name": "E2E-InvoicePay-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {"firstName": "Inv", "lastName": "Payer", "country": "Japan", "state": "Tokyo"}
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
  exit 1
fi

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
              "initialTerm": {"period": 12, "periodType": "Month", "termType": "TERMED", "startDate": "$TODAY"},
              "renewalTerms": [{"period": 12, "periodType": "Month"}],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": false
            },
            "subscribeToRatePlans": [{"productRatePlanId": "$RATE_PLAN_ID"}]
          }
        }
      ]
    }
  ],
  "processingOptions": {"runBilling": true, "collectPayment": false}
}
EOF
)
run $ZR order create --body "$ORDER_BODY" --json
BILLING_ORDER=$(echo "$RUN_OUT" | jq -r '.orderNumber // empty' 2>/dev/null)
if [ -n "$BILLING_ORDER" ]; then
  pass "order create with billing → $BILLING_ORDER"
else
  fail "order create with billing (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 3: invoice list (billing is a required postcondition)"
# ─────────────────────────────────────────
# runBilling:true must produce at least one invoice; poll briefly for any
# propagation lag, then FAIL (not skip) if none appears.
INV_ID=""
if [ -n "$ACCT_ID" ]; then
  echo "  Testing: invoice list --account-key $ACCT_ID"
  for _i in 1 2 3 4 5 6; do
    run $ZR invoice list --account-key "$ACCT_ID" --json
    INV_N=$(echo "$RUN_OUT" | jq -r '.invoices | length' 2>/dev/null)
    [ "${INV_N:-0}" -ge 1 ] 2>/dev/null && break
    sleep 2
  done
  if [ "${INV_N:-0}" -ge 1 ] 2>/dev/null; then
    INV_ID=$(echo "$RUN_OUT" | jq -r '.invoices[0].id // empty')
    pass "invoice list → $INV_N invoice(s) after billing"
  else
    fail "invoice list → no invoice after runBilling (rc=$RUN_RC) ${RUN_ERR}"
  fi
else
  skip "invoice list → no account ID"
fi

# ─────────────────────────────────────────
header "Step 4: invoice get / items / files"
# ─────────────────────────────────────────
if [ -n "$INV_ID" ]; then
  echo "  Testing: invoice get $INV_ID"
  run $ZR invoice get "$INV_ID" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.id // .invoiceNumber' >/dev/null 2>&1; then
    pass "invoice get → returned invoice"
  else
    fail "invoice get (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi

  echo "  Testing: invoice items $INV_ID"
  run $ZR invoice items "$INV_ID" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.invoiceItems | type == "array"' >/dev/null 2>&1; then
    pass "invoice items → .invoiceItems array (n=$(echo "$RUN_OUT" | jq '.invoiceItems | length'))"
  else
    fail "invoice items (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi

  echo "  Testing: invoice files $INV_ID"
  run $ZR invoice files "$INV_ID" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.invoiceFiles | type == "array"' >/dev/null 2>&1; then
    pass "invoice files → .invoiceFiles array (PDF generation may be async; empty is OK)"
  else
    fail "invoice files (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "invoice get → no invoice ID"
  skip "invoice items → no invoice ID"
  skip "invoice files → no invoice ID"
fi

# ─────────────────────────────────────────
header "Step 4.5: invoice lifecycle (post / reverse / writeoff guards)"
# ─────────────────────────────────────────
echo "  Testing: invoice post without arg"
expect_fail "invoice post validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR invoice post

echo "  Testing: invoice reverse without arg"
expect_fail "invoice reverse validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR invoice reverse

echo "  Testing: invoice writeoff without arg"
expect_fail "invoice writeoff validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR invoice writeoff

if [ -n "$INV_ID" ]; then
  echo "  Testing: invoice reverse/writeoff --confirm guards (no API call)"
  expect_fail "invoice reverse → requires --confirm" "Use --confirm to proceed" -- $ZR invoice reverse "$INV_ID"
  expect_fail "invoice writeoff → requires --confirm" "Use --confirm to proceed" -- $ZR invoice writeoff "$INV_ID"

  # Live post of the suite's own invoice. runBilling creates Draft invoices
  # (autoPost defaults false), and posting is the normal forward operation.
  # This is the permanent live gate for the #220 415 fix on invoice post:
  # before the fix every bodyless invoice lifecycle PUT failed with HTTP 415.
  INV_STATUS=$($ZR invoice get "$INV_ID" --jq '.status' 2>/dev/null | tr -d '"')
  echo "  Testing: invoice post $INV_ID (status=$INV_STATUS)"
  if [ "$INV_STATUS" = "Draft" ]; then
    POST_RC=0
    POST_OUT=$($ZR invoice post "$INV_ID" --confirm 2>&1) || POST_RC=$?
    if [ "$POST_RC" -eq 0 ] && echo "$POST_OUT" | grep -qF "posted."; then
      pass "invoice post → Draft invoice posted (bodyless PUT carries Content-Type + {})"
    elif echo "$POST_OUT" | grep -q "HTTP 415\|50000045"; then
      fail "invoice post → 415 REGRESSION (empty-JSON body lost, cf. #220): $(echo "$POST_OUT" | head -2)"
    else
      fail "invoice post → rc=$POST_RC: $(echo "$POST_OUT" | head -2)"
    fi
  else
    skip "invoice post → invoice already '$INV_STATUS' (tenant auto-post?); 415 contract still guarded by billrun suite"
  fi
else
  skip "invoice lifecycle live checks → no invoice ID"
fi

# ─────────────────────────────────────────
header "Step 4.6: creditmemo / debitmemo (read paths)"
# ─────────────────────────────────────────
echo "  Testing: creditmemo get without arg"
expect_fail "creditmemo get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR creditmemo get

echo "  Testing: debitmemo get without arg"
expect_fail "debitmemo get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR debitmemo get

if [ -n "$ACCT_NUM" ]; then
  # Fresh suite-created account → empty memo arrays are the correct result;
  # the shape assertion still catches error-object/shape regressions in the
  # listcmd-migrated commands (P3-2: zero E2E contact until now).
  echo "  Testing: creditmemo list --account-number $ACCT_NUM"
  run $ZR creditmemo list --account-number "$ACCT_NUM" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.creditmemos | type == "array"' >/dev/null 2>&1; then
    pass "creditmemo list → .creditmemos array (n=$(echo "$RUN_OUT" | jq '.creditmemos | length'))"
  else
    fail "creditmemo list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi

  echo "  Testing: debitmemo list --account-number $ACCT_NUM"
  run $ZR debitmemo list --account-number "$ACCT_NUM" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.debitmemos | type == "array"' >/dev/null 2>&1; then
    pass "debitmemo list → .debitmemos array (n=$(echo "$RUN_OUT" | jq '.debitmemos | length'))"
  else
    fail "debitmemo list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "creditmemo list → no account number"
  skip "debitmemo list → no account number"
fi

# ─────────────────────────────────────────
header "Step 5: payment list"
# ─────────────────────────────────────────
# No payment gateway on this tenant → an empty .payments array is the correct
# result. Assert the shape (catches an error-object/shape regression).
PAY_ID=""
if [ -n "$ACCT_NUM" ]; then
  echo "  Testing: payment list --account-key $ACCT_NUM"
  run $ZR payment list --account-key "$ACCT_NUM" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.payments | type == "array"' >/dev/null 2>&1; then
    PAY_N=$(echo "$RUN_OUT" | jq -r '.payments | length')
    pass "payment list → .payments array (n=$PAY_N)"
    PAY_ID=$(echo "$RUN_OUT" | jq -r '.payments[0].id // empty')
  else
    fail "payment list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "payment list → no account"
fi

# ─────────────────────────────────────────
header "Step 6: payment get (if a payment exists)"
# ─────────────────────────────────────────
if [ -n "$PAY_ID" ]; then
  echo "  Testing: payment get $PAY_ID"
  run $ZR payment get "$PAY_ID" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.id' >/dev/null 2>&1; then
    pass "payment get → returned payment"
  else
    fail "payment get (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  # Expected on this tenant: no gateway → no payments to fetch.
  skip "payment get → no payment available (gateway not configured on sandbox)"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Test Account: $ACCT_NUM ($ACCT_ID)"
echo ""
print_summary
