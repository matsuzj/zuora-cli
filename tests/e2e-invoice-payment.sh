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
RATE_PLAN_ID="${ZR_E2E_RATE_PLAN_ID:-4c6059a8d8899f453ffa0637451d0003}"

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-invoice-payment-${TIMESTAMP}.log"

exec > >(tee >(sed 's/\x1b\[[0-9;]*m//g' > "$LOG_FILE")) 2>&1

green()  { printf "\033[32m%s\033[0m\n" "$1"; }
red()    { printf "\033[31m%s\033[0m\n" "$1"; }
yellow() { printf "\033[33m%s\033[0m\n" "$1"; }

pass() { PASS=$((PASS+1)); green "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); red   "  ✗ $1"; }
skip() { SKIP=$((SKIP+1)); yellow "  ⊘ $1 (skipped)"; }

header() { printf "\n\033[1m=== %s ===\033[0m\n" "$1"; }

# run <command...> — stdout→RUN_OUT (clean), stderr→RUN_ERR, exit→RUN_RC.
RUN_OUT=""; RUN_ERR=""; RUN_RC=0
run() {
  local ef="$LOG_DIR/.run.$$.err"
  RUN_OUT=$("$@" 2>"$ef"); RUN_RC=$?
  RUN_ERR=$(cat "$ef" 2>/dev/null); rm -f "$ef"
}

# expect_fail <description> <expected-substring> -- <command...>
expect_fail() {
  local desc="$1" want="$2"; shift 2
  [ "${1:-}" = "--" ] && shift
  local out rc
  out=$("$@" 2>&1); rc=$?
  if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -qF -- "$want"; then
    pass "$desc"
  else
    fail "$desc → rc=$rc, expected '$want', got: $(printf '%s' "$out" | head -1)"
  fi
}

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
[ -x "$ZR" ] || { red "zr binary not found/executable at $ZR (build it first)"; exit 1; }
AUTH_OUT=$($ZR auth status 2>&1)
if echo "$AUTH_OUT" | grep -qE "Token:[[:space:]]+valid"; then
  pass "Auth OK"
else
  fail "Auth failed (token not valid): $(echo "$AUTH_OUT" | grep -i 'token' | head -1)"
  exit 1
fi

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
expect_fail "payment create validation → requires --body" "--body is required" -- $ZR payment create

echo "  Testing: payment apply without argument"
expect_fail "payment apply validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR payment apply

echo "  Testing: payment refund without argument"
expect_fail "payment refund validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR payment refund

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
  echo "  Testing: invoice list --account $ACCT_ID"
  for _i in 1 2 3 4 5 6; do
    run $ZR invoice list --account "$ACCT_ID" --json
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
header "Step 5: payment list"
# ─────────────────────────────────────────
# No payment gateway on this tenant → an empty .payments array is the correct
# result. Assert the shape (catches an error-object/shape regression).
PAY_ID=""
if [ -n "$ACCT_NUM" ]; then
  echo "  Testing: payment list --account $ACCT_NUM"
  run $ZR payment list --account "$ACCT_NUM" --json
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
TOTAL=$((PASS + FAIL + SKIP))
echo "  Test Account: $ACCT_NUM ($ACCT_ID)"
echo "  Passed:  $PASS / $TOTAL"
echo "  Failed:  $FAIL / $TOTAL"
echo "  Skipped: $SKIP / $TOTAL"
echo ""
echo "  Log: $LOG_FILE"
echo ""
if [ "$FAIL" -gt 0 ]; then
  echo "  RESULT: FAIL"
  exit 1
else
  echo "  RESULT: PASS"
fi
