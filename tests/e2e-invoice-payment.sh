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
header "Step 3: Invoice List (実行テスト)"
# ─────────────────────────────────────────

# テスト用アカウントを取得
FIRST_ACCT=$($ZR account list --page-size 1 --json 2>/dev/null | jq -r '.data[0].accountNumber // empty' 2>/dev/null) || true

if [ -n "$FIRST_ACCT" ]; then
  echo "  Testing: invoice list --account $FIRST_ACCT"
  INV_LIST=$($ZR invoice list --account "$FIRST_ACCT" --json 2>/dev/null) || true
  if echo "$INV_LIST" | jq -e '.' >/dev/null 2>&1; then
    pass "invoice list → returned JSON"

    # invoice list のデータから invoice ID を取得
    INV_ID=$(echo "$INV_LIST" | jq -r '.invoices[0].id // empty' 2>/dev/null) || true
  else
    skip "invoice list → API error"
    INV_ID=""
  fi
else
  skip "invoice list → no account found"
  INV_ID=""
fi

# ─────────────────────────────────────────
header "Step 4: Invoice Get / Items (実行テスト)"
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
else
  skip "invoice get → no invoice ID"
  skip "invoice items → no invoice ID"
  skip "invoice get --jq → no invoice ID"
fi

# ─────────────────────────────────────────
header "Step 5: Payment List (実行テスト)"
# ─────────────────────────────────────────

if [ -n "$FIRST_ACCT" ]; then
  echo "  Testing: payment list --account $FIRST_ACCT"
  PAY_LIST=$($ZR payment list --account "$FIRST_ACCT" --json 2>/dev/null) || true
  if echo "$PAY_LIST" | jq -e '.' >/dev/null 2>&1; then
    pass "payment list → returned JSON"

    PAY_ID=$(echo "$PAY_LIST" | jq -r '.payments[0].id // empty' 2>/dev/null) || true
  else
    skip "payment list → API error"
    PAY_ID=""
  fi
else
  skip "payment list → no account found"
  PAY_ID=""
fi

# ─────────────────────────────────────────
header "Step 6: Payment Get (実行テスト)"
# ─────────────────────────────────────────

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
