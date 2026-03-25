#!/bin/bash
# E2E Test: Invoice + Payment Commands (Phase 6)
# テナント: apac-sandbox

set -uo pipefail

ZR="./bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Log directory
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
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
header "Step 1: Invoice Validation"
# ─────────────────────────────────────────

# 1a: invoice list missing --account
echo "  Testing: invoice list without --account"
IL_ERR=$($ZR invoice list 2>&1) || true
if echo "$IL_ERR" | grep -qi "required.*account\|account.*required"; then
  pass "invoice list validation → requires --account"
else
  fail "invoice list validation → unexpected: $IL_ERR"
fi

# 1b: invoice get missing argument
echo "  Testing: invoice get without argument"
IG_ERR=$($ZR invoice get 2>&1) || true
if echo "$IG_ERR" | grep -qi "arg\|required"; then
  pass "invoice get validation → requires argument"
else
  fail "invoice get validation → unexpected: $IG_ERR"
fi

# 1c: invoice items missing argument
echo "  Testing: invoice items without argument"
II_ERR=$($ZR invoice items 2>&1) || true
if echo "$II_ERR" | grep -qi "arg\|required"; then
  pass "invoice items validation → requires argument"
else
  fail "invoice items validation → unexpected: $II_ERR"
fi

# 1d: invoice files missing argument
echo "  Testing: invoice files without argument"
IF_ERR=$($ZR invoice files 2>&1) || true
if echo "$IF_ERR" | grep -qi "arg\|required"; then
  pass "invoice files validation → requires argument"
else
  fail "invoice files validation → unexpected: $IF_ERR"
fi

# 1e: invoice email missing --body
echo "  Testing: invoice email without --body"
IE_ERR=$($ZR invoice email FAKE-ID 2>&1) || true
if echo "$IE_ERR" | grep -q "\-\-body is required"; then
  pass "invoice email validation → requires --body"
else
  fail "invoice email validation → unexpected: $IE_ERR"
fi

# 1f: invoice usage-rate-detail missing argument
echo "  Testing: invoice usage-rate-detail without argument"
IU_ERR=$($ZR invoice usage-rate-detail 2>&1) || true
if echo "$IU_ERR" | grep -qi "arg\|required"; then
  pass "invoice usage-rate-detail validation → requires argument"
else
  fail "invoice usage-rate-detail validation → unexpected: $IU_ERR"
fi

# ─────────────────────────────────────────
header "Step 2: Payment Validation"
# ─────────────────────────────────────────

# 2a: payment list missing --account
echo "  Testing: payment list without --account"
PL_ERR=$($ZR payment list 2>&1) || true
if echo "$PL_ERR" | grep -qi "required.*account\|account.*required"; then
  pass "payment list validation → requires --account"
else
  fail "payment list validation → unexpected: $PL_ERR"
fi

# 2b: payment get missing argument
echo "  Testing: payment get without argument"
PG_ERR=$($ZR payment get 2>&1) || true
if echo "$PG_ERR" | grep -qi "arg\|required"; then
  pass "payment get validation → requires argument"
else
  fail "payment get validation → unexpected: $PG_ERR"
fi

# 2c: payment create missing --body
echo "  Testing: payment create without --body"
PC_ERR=$($ZR payment create 2>&1) || true
if echo "$PC_ERR" | grep -q "\-\-body is required"; then
  pass "payment create validation → requires --body"
else
  fail "payment create validation → unexpected: $PC_ERR"
fi

# 2d: payment create rejects stray args (NoArgs)
echo "  Testing: payment create with stray arg"
PC_NA=$($ZR payment create extraArg --body '{}' 2>&1) || true
if echo "$PC_NA" | grep -qi "unknown command\|too many arg"; then
  pass "payment create → rejects stray positional arg"
else
  fail "payment create → accepted stray arg: $PC_NA"
fi

# 2e: payment apply missing --body
echo "  Testing: payment apply without --body"
PA_ERR=$($ZR payment apply FAKE-ID 2>&1) || true
if echo "$PA_ERR" | grep -q "\-\-body is required"; then
  pass "payment apply validation → requires --body"
else
  fail "payment apply validation → unexpected: $PA_ERR"
fi

# 2f: payment refund missing --body
echo "  Testing: payment refund without --body"
PR_ERR=$($ZR payment refund FAKE-ID 2>&1) || true
if echo "$PR_ERR" | grep -q "\-\-body is required"; then
  pass "payment refund validation → requires --body"
else
  fail "payment refund validation → unexpected: $PR_ERR"
fi

# ─────────────────────────────────────────
header "Step 3: テストデータ生成 (Account + Subscription + Billing)"
# ─────────────────────────────────────────

RATE_PLAN_ID="4c6059a8d8899f453ffa0637451d0003"
TODAY=$(date +%Y-%m-%d)

# Account create
ACCT_BODY=$(cat <<'JSON'
{
  "name": "E2E-Invoice-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {
    "firstName": "Test",
    "lastName": "Invoice",
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
  fail "account create failed"
  printf '\n'
  red "Cannot proceed without test account. Aborting."
  exit 1
fi

# Order create with runBilling: true to generate invoice
ORDER_BODY=$(cat <<EOF
{
  "existingAccountNumber": "$ACCT_NUM",
  "orderDate": "$TODAY",
  "subscriptions": [{
    "orderActions": [{
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
    }]
  }],
  "processingOptions": {"runBilling": true, "collectPayment": false}
}
EOF
)
ORDER_RESULT=$($ZR order create --body "$ORDER_BODY" --json 2>/dev/null) || true
ORDER_SUCCESS=$(echo "$ORDER_RESULT" | jq -r '.success // empty' 2>/dev/null)

if [ "$ORDER_SUCCESS" = "true" ]; then
  pass "order create (with billing) → $(echo "$ORDER_RESULT" | jq -r '.orderNumber // empty')"
else
  fail "order create failed: $(echo "$ORDER_RESULT" | head -3)"
  ACCT_NUM=""
fi

# Wait for billing to complete
sleep 3

# ─────────────────────────────────────────
header "Step 4: Invoice List (実行テスト)"
# ─────────────────────────────────────────

INV_ID=""
if [ -n "$ACCT_NUM" ]; then
  echo "  Testing: invoice list --account $ACCT_NUM"
  INV_LIST=$($ZR invoice list --account "$ACCT_NUM" --json 2>/dev/null) || true
  if echo "$INV_LIST" | jq -e '.invoices' >/dev/null 2>&1; then
    INV_COUNT=$(echo "$INV_LIST" | jq '.invoices | length')
    pass "invoice list → returned $INV_COUNT invoices"
    INV_ID=$(echo "$INV_LIST" | jq -r '.invoices[0].id // empty' 2>/dev/null) || true
  else
    skip "invoice list → API error"
  fi
else
  skip "invoice list → no account"
fi

# ─────────────────────────────────────────
header "Step 5: Invoice Get / Items / Files (実行テスト)"
# ─────────────────────────────────────────

if [ -n "$INV_ID" ]; then
  echo "  Testing: invoice get $INV_ID"
  INV_GET=$($ZR invoice get "$INV_ID" --json 2>/dev/null) || true
  if echo "$INV_GET" | jq -e '.id // .invoiceNumber' >/dev/null 2>&1; then
    pass "invoice get → returned invoice data"
  else
    fail "invoice get → unexpected: $(echo "$INV_GET" | head -3)"
  fi

  echo "  Testing: invoice items $INV_ID"
  INV_ITEMS=$($ZR invoice items "$INV_ID" 2>&1) || true
  if echo "$INV_ITEMS" | grep -qi "ID\|CHARGE\|invoiceItems" || echo "$INV_ITEMS" | jq -e '.' >/dev/null 2>&1; then
    pass "invoice items → returned data"
  elif echo "$INV_ITEMS" | grep -qi "error"; then
    skip "invoice items → API error"
  else
    fail "invoice items → unexpected: $(echo "$INV_ITEMS" | head -3)"
  fi

  # --jq output format test
  echo "  Testing: invoice get with --jq '.id'"
  INV_JQ=$($ZR invoice get "$INV_ID" --jq '.id' 2>/dev/null) || true
  if [ -n "$INV_JQ" ] && [ "$INV_JQ" != "null" ]; then
    pass "invoice get --jq → filtered output"
  else
    skip "invoice get --jq → no data"
  fi
  # invoice files
  echo "  Testing: invoice files $INV_ID"
  INV_FILES=$($ZR invoice files "$INV_ID" 2>&1) || true
  if echo "$INV_FILES" | jq -e '.' >/dev/null 2>&1; then
    pass "invoice files → returned JSON"
  elif echo "$INV_FILES" | grep -qi "error"; then
    skip "invoice files → API error (files may not be generated yet)"
  else
    fail "invoice files → unexpected: $(echo "$INV_FILES" | head -3)"
  fi
else
  skip "invoice get → no invoice ID"
  skip "invoice items → no invoice ID"
  skip "invoice get --jq → no invoice ID"
  skip "invoice files → no invoice ID"
fi

# ─────────────────────────────────────────
header "Step 6: Payment List (実行テスト)"
# ─────────────────────────────────────────

if [ -n "$ACCT_NUM" ]; then
  echo "  Testing: payment list --account $ACCT_NUM"
  PAY_LIST=$($ZR payment list --account "$ACCT_NUM" --json 2>/dev/null) || true
  if echo "$PAY_LIST" | jq -e '.' >/dev/null 2>&1; then
    pass "payment list → returned JSON"

    PAY_ID=$(echo "$PAY_LIST" | jq -r '.payments[0].id // empty' 2>/dev/null) || true
  else
    skip "payment list → API error"
    PAY_ID=""
  fi
else
  skip "payment list → no account"
  PAY_ID=""
fi

# ─────────────────────────────────────────
header "Step 7: Payment Get (実行テスト)"
# ─────────────────────────────────────────
# 注意: このテナントでは payment gateway が未設定のため payment 作成ができず、
# payment get のテストに必要な payment ID が取得できない。
# Gateway 設定済みテナントでは payment list から ID を取得して PASS になる。

if [ -n "$PAY_ID" ]; then
  echo "  Testing: payment get $PAY_ID"
  PAY_GET=$($ZR payment get "$PAY_ID" --json 2>/dev/null) || true
  if echo "$PAY_GET" | jq -e '.id // .paymentNumber' >/dev/null 2>&1; then
    pass "payment get → returned payment data"
  else
    fail "payment get → unexpected: $(echo "$PAY_GET" | head -3)"
  fi
else
  skip "payment get → no payment ID"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
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
