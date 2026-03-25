#!/bin/bash
# E2E Test: Commerce Commands (Phase 5)
# テナント: apac-sandbox
# 注意: Commerce API はテナントの Product Catalog 設定に依存する

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
LOG_FILE="$LOG_DIR/e2e-commerce-${TIMESTAMP}.log"

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
header "Step 1: Product Command Validation"
# ─────────────────────────────────────────

# 1a: product create missing --body
echo "  Testing: product create without --body"
PC_ERR=$($ZR product create 2>&1) || true
if echo "$PC_ERR" | grep -q "\-\-body is required"; then
  pass "product create validation → requires --body"
else
  fail "product create validation → unexpected: $PC_ERR"
fi

# 1b: product update missing --body
echo "  Testing: product update without --body"
PU_ERR=$($ZR product update 2>&1) || true
if echo "$PU_ERR" | grep -q "\-\-body is required"; then
  pass "product update validation → requires --body"
else
  fail "product update validation → unexpected: $PU_ERR"
fi

# 1c: product get missing argument
echo "  Testing: product get without argument"
PG_ERR=$($ZR product get 2>&1) || true
if echo "$PG_ERR" | grep -qi "arg\|required"; then
  pass "product get validation → requires argument"
else
  fail "product get validation → unexpected: $PG_ERR"
fi

# 1d: product list-legacy missing --body
echo "  Testing: product list-legacy without --body"
PL_ERR=$($ZR product list-legacy 2>&1) || true
if echo "$PL_ERR" | grep -q "\-\-body is required"; then
  pass "product list-legacy validation → requires --body"
else
  fail "product list-legacy validation → unexpected: $PL_ERR"
fi

# ─────────────────────────────────────────
header "Step 2: Plan Command Validation"
# ─────────────────────────────────────────

# 2a: plan create missing --body
echo "  Testing: plan create without --body"
PLCR_ERR=$($ZR plan create 2>&1) || true
if echo "$PLCR_ERR" | grep -q "\-\-body is required"; then
  pass "plan create validation → requires --body"
else
  fail "plan create validation → unexpected: $PLCR_ERR"
fi

# 2b: plan update missing --body
echo "  Testing: plan update without --body"
PLUP_ERR=$($ZR plan update 2>&1) || true
if echo "$PLUP_ERR" | grep -q "\-\-body is required"; then
  pass "plan update validation → requires --body"
else
  fail "plan update validation → unexpected: $PLUP_ERR"
fi

# 2c: plan get missing --key
echo "  Testing: plan get without --key"
PLG_ERR=$($ZR plan get 2>&1) || true
if echo "$PLG_ERR" | grep -q "\-\-key is required"; then
  pass "plan get validation → requires --key"
else
  fail "plan get validation → unexpected: $PLG_ERR"
fi

# 2d: plan list missing --body
echo "  Testing: plan list without --body"
PLL_ERR=$($ZR plan list 2>&1) || true
if echo "$PLL_ERR" | grep -q "\-\-body is required"; then
  pass "plan list validation → requires --body"
else
  fail "plan list validation → unexpected: $PLL_ERR"
fi

# 2e: plan purchase-options missing --plan
echo "  Testing: plan purchase-options without --plan"
PLPO_ERR=$($ZR plan purchase-options 2>&1) || true
if echo "$PLPO_ERR" | grep -q "\-\-plan is required"; then
  pass "plan purchase-options validation → requires --plan"
else
  fail "plan purchase-options validation → unexpected: $PLPO_ERR"
fi

# ─────────────────────────────────────────
header "Step 3: Charge Command Validation"
# ─────────────────────────────────────────

# 3a: charge create missing --body
echo "  Testing: charge create without --body"
CCR_ERR=$($ZR charge create 2>&1) || true
if echo "$CCR_ERR" | grep -q "\-\-body is required"; then
  pass "charge create validation → requires --body"
else
  fail "charge create validation → unexpected: $CCR_ERR"
fi

# 3b: charge update missing --body
echo "  Testing: charge update without --body"
CUP_ERR=$($ZR charge update 2>&1) || true
if echo "$CUP_ERR" | grep -q "\-\-body is required"; then
  pass "charge update validation → requires --body"
else
  fail "charge update validation → unexpected: $CUP_ERR"
fi

# 3c: charge get missing --key
echo "  Testing: charge get without --key"
CG_ERR=$($ZR charge get 2>&1) || true
if echo "$CG_ERR" | grep -q "\-\-key is required"; then
  pass "charge get validation → requires --key"
else
  fail "charge get validation → unexpected: $CG_ERR"
fi

# 3d: charge update-tiers missing --body
echo "  Testing: charge update-tiers without --body"
CT_ERR=$($ZR charge update-tiers 2>&1) || true
if echo "$CT_ERR" | grep -q "\-\-body is required"; then
  pass "charge update-tiers validation → requires --body"
else
  fail "charge update-tiers validation → unexpected: $CT_ERR"
fi

# ─────────────────────────────────────────
header "Step 4: RatePlan Command Validation"
# ─────────────────────────────────────────

# 4a: rateplan get missing argument
echo "  Testing: rateplan get without argument"
RP_ERR=$($ZR rateplan get 2>&1) || true
if echo "$RP_ERR" | grep -qi "arg\|required"; then
  pass "rateplan get validation → requires argument"
else
  fail "rateplan get validation → unexpected: $RP_ERR"
fi

# ─────────────────────────────────────────
header "Step 5: Product List Legacy (実行テスト)"
# ─────────────────────────────────────────
echo "  Testing: product list-legacy with empty filter"
PL_RESULT=$($ZR product list-legacy --body '{}' 2>&1) || true
if echo "$PL_RESULT" | jq -e '.' >/dev/null 2>&1; then
  pass "product list-legacy → returned JSON"
elif echo "$PL_RESULT" | grep -qi "error"; then
  skip "product list-legacy → API error (Commerce API may not be enabled)"
else
  fail "product list-legacy → unexpected: $(echo "$PL_RESULT" | head -3)"
fi

# ─────────────────────────────────────────
header "Step 6: Plan List (実行テスト)"
# ─────────────────────────────────────────
echo "  Testing: plan list with empty filter"
PLAN_LIST=$($ZR plan list --body '{}' 2>&1) || true
if echo "$PLAN_LIST" | jq -e '.' >/dev/null 2>&1; then
  pass "plan list → returned JSON"
elif echo "$PLAN_LIST" | grep -qi "error"; then
  skip "plan list → API error (Commerce API may not be enabled)"
else
  fail "plan list → unexpected: $(echo "$PLAN_LIST" | head -3)"
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
