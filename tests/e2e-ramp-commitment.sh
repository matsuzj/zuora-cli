#!/bin/bash
# E2E Test: Ramp + Commitment + Fulfillment + Prepaid (Phase 8)
# テナント: apac-sandbox

set -uo pipefail

ZR="./bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-ramp-commitment-${TIMESTAMP}.log"

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
header "Step 1: Ramp Validation"
# ─────────────────────────────────────────

echo "  Testing: ramp get without argument"
RG_ERR=$($ZR ramp get 2>&1) || true
if echo "$RG_ERR" | grep -qi "arg\|required"; then
  pass "ramp get validation → requires argument"
else
  fail "ramp get validation → unexpected: $RG_ERR"
fi

echo "  Testing: ramp get-by-subscription without argument"
RGS_ERR=$($ZR ramp get-by-subscription 2>&1) || true
if echo "$RGS_ERR" | grep -qi "arg\|required"; then
  pass "ramp get-by-subscription validation → requires argument"
else
  fail "ramp get-by-subscription validation → unexpected: $RGS_ERR"
fi

echo "  Testing: ramp metrics without argument"
RM_ERR=$($ZR ramp metrics 2>&1) || true
if echo "$RM_ERR" | grep -qi "arg\|required"; then
  pass "ramp metrics validation → requires argument"
else
  fail "ramp metrics validation → unexpected: $RM_ERR"
fi

echo "  Testing: ramp metrics-by-subscription without argument"
RMS_ERR=$($ZR ramp metrics-by-subscription 2>&1) || true
if echo "$RMS_ERR" | grep -qi "arg\|required"; then
  pass "ramp metrics-by-subscription validation → requires argument"
else
  fail "ramp metrics-by-subscription validation → unexpected: $RMS_ERR"
fi

echo "  Testing: ramp metrics-by-order without argument"
RMO_ERR=$($ZR ramp metrics-by-order 2>&1) || true
if echo "$RMO_ERR" | grep -qi "arg\|required"; then
  pass "ramp metrics-by-order validation → requires argument"
else
  fail "ramp metrics-by-order validation → unexpected: $RMO_ERR"
fi

# ─────────────────────────────────────────
header "Step 2: Commitment Validation"
# ─────────────────────────────────────────

echo "  Testing: commitment list without --account"
CL_ERR=$($ZR commitment list 2>&1) || true
if echo "$CL_ERR" | grep -qi "required.*account\|account.*required"; then
  pass "commitment list validation → requires --account"
else
  fail "commitment list validation → unexpected: $CL_ERR"
fi

echo "  Testing: commitment get without argument"
CG_ERR=$($ZR commitment get 2>&1) || true
if echo "$CG_ERR" | grep -qi "arg\|required"; then
  pass "commitment get validation → requires argument"
else
  fail "commitment get validation → unexpected: $CG_ERR"
fi

echo "  Testing: commitment periods without --commitment"
CP_ERR=$($ZR commitment periods 2>&1) || true
if echo "$CP_ERR" | grep -qi "commitment.*required\|required.*commitment"; then
  pass "commitment periods validation → requires --commitment"
else
  fail "commitment periods validation → unexpected: $CP_ERR"
fi

echo "  Testing: commitment balance without argument"
CB_ERR=$($ZR commitment balance 2>&1) || true
if echo "$CB_ERR" | grep -qi "arg\|required"; then
  pass "commitment balance validation → requires argument"
else
  fail "commitment balance validation → unexpected: $CB_ERR"
fi

echo "  Testing: commitment schedules without argument"
CS_ERR=$($ZR commitment schedules 2>&1) || true
if echo "$CS_ERR" | grep -qi "arg\|required"; then
  pass "commitment schedules validation → requires argument"
else
  fail "commitment schedules validation → unexpected: $CS_ERR"
fi

# ─────────────────────────────────────────
header "Step 3: Fulfillment Validation"
# ─────────────────────────────────────────

echo "  Testing: fulfillment create without --body"
FC_ERR=$($ZR fulfillment create 2>&1) || true
if echo "$FC_ERR" | grep -q "\-\-body is required"; then
  pass "fulfillment create validation → requires --body"
else
  fail "fulfillment create validation → unexpected: $FC_ERR"
fi

echo "  Testing: fulfillment create rejects stray arg"
FC_NA=$($ZR fulfillment create extraArg --body '{}' 2>&1) || true
if echo "$FC_NA" | grep -qi "unknown command\|too many arg\|accepts 0 arg"; then
  pass "fulfillment create → rejects stray positional arg"
else
  fail "fulfillment create → accepted stray arg"
fi

echo "  Testing: fulfillment get without argument"
FG_ERR=$($ZR fulfillment get 2>&1) || true
if echo "$FG_ERR" | grep -qi "arg\|required"; then
  pass "fulfillment get validation → requires argument"
else
  fail "fulfillment get validation → unexpected: $FG_ERR"
fi

echo "  Testing: fulfillment update without --body"
FU_ERR=$($ZR fulfillment update FAKE-ID 2>&1) || true
if echo "$FU_ERR" | grep -q "\-\-body is required"; then
  pass "fulfillment update validation → requires --body"
else
  fail "fulfillment update validation → unexpected: $FU_ERR"
fi

echo "  Testing: fulfillment delete without --confirm"
FD_ERR=$($ZR fulfillment delete FAKE-ID 2>&1) || true
if echo "$FD_ERR" | grep -q "\-\-confirm"; then
  pass "fulfillment delete validation → requires --confirm"
else
  fail "fulfillment delete validation → unexpected: $FD_ERR"
fi

# ─────────────────────────────────────────
header "Step 4: Fulfillment Item Validation"
# ─────────────────────────────────────────

echo "  Testing: fulfillment-item create without --body"
FIC_ERR=$($ZR fulfillment-item create 2>&1) || true
if echo "$FIC_ERR" | grep -q "\-\-body is required"; then
  pass "fulfillment-item create validation → requires --body"
else
  fail "fulfillment-item create validation → unexpected: $FIC_ERR"
fi

echo "  Testing: fulfillment-item get without argument"
FIG_ERR=$($ZR fulfillment-item get 2>&1) || true
if echo "$FIG_ERR" | grep -qi "arg\|required"; then
  pass "fulfillment-item get validation → requires argument"
else
  fail "fulfillment-item get validation → unexpected: $FIG_ERR"
fi

echo "  Testing: fulfillment-item update without --body"
FIU_ERR=$($ZR fulfillment-item update FAKE-ID 2>&1) || true
if echo "$FIU_ERR" | grep -q "\-\-body is required"; then
  pass "fulfillment-item update validation → requires --body"
else
  fail "fulfillment-item update validation → unexpected: $FIU_ERR"
fi

echo "  Testing: fulfillment-item delete without --confirm"
FID_ERR=$($ZR fulfillment-item delete FAKE-ID 2>&1) || true
if echo "$FID_ERR" | grep -q "\-\-confirm"; then
  pass "fulfillment-item delete validation → requires --confirm"
else
  fail "fulfillment-item delete validation → unexpected: $FID_ERR"
fi

# ─────────────────────────────────────────
header "Step 5: Prepaid Validation"
# ─────────────────────────────────────────

echo "  Testing: prepaid rollover without --body"
PR_ERR=$($ZR prepaid rollover 2>&1) || true
if echo "$PR_ERR" | grep -q "\-\-body is required"; then
  pass "prepaid rollover validation → requires --body"
else
  fail "prepaid rollover validation → unexpected: $PR_ERR"
fi

echo "  Testing: prepaid reverse-rollover without --body"
PRR_ERR=$($ZR prepaid reverse-rollover 2>&1) || true
if echo "$PRR_ERR" | grep -q "\-\-body is required"; then
  pass "prepaid reverse-rollover validation → requires --body"
else
  fail "prepaid reverse-rollover validation → unexpected: $PRR_ERR"
fi

echo "  Testing: prepaid deplete without --body"
PD_ERR=$($ZR prepaid deplete 2>&1) || true
if echo "$PD_ERR" | grep -q "\-\-body is required"; then
  pass "prepaid deplete validation → requires --body"
else
  fail "prepaid deplete validation → unexpected: $PD_ERR"
fi

echo "  Testing: prepaid rollover rejects stray arg"
PR_NA=$($ZR prepaid rollover extraArg --body '{}' 2>&1) || true
if echo "$PR_NA" | grep -qi "unknown command\|too many arg\|accepts 0 arg"; then
  pass "prepaid rollover → rejects stray positional arg"
else
  fail "prepaid rollover → accepted stray arg"
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
