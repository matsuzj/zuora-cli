#!/bin/bash
# E2E Test: Usage + Meter Commands (Phase 7)
# テナント: apac-sandbox
# 注意: Usage/Meter API はテナント設定・プロダクト設定に依存する

set -uo pipefail

ZR="./bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-usage-meter-${TIMESTAMP}.log"

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
header "Step 1: Usage Validation"
# ─────────────────────────────────────────

# 1a: usage post missing --file
echo "  Testing: usage post without --file"
UP_ERR=$($ZR usage post 2>&1) || true
if echo "$UP_ERR" | grep -q "\-\-file is required"; then
  pass "usage post validation → requires --file"
else
  fail "usage post validation → unexpected: $UP_ERR"
fi

# 1b: usage post with nonexistent file
echo "  Testing: usage post with nonexistent file"
UP_NF=$($ZR usage post --file /nonexistent/file.csv 2>&1) || true
if echo "$UP_NF" | grep -qi "no such file\|not found\|reading file"; then
  pass "usage post → correctly rejects nonexistent file"
else
  fail "usage post nonexistent → unexpected: $UP_NF"
fi

# 1c: usage create missing --body
echo "  Testing: usage create without --body"
UC_ERR=$($ZR usage create 2>&1) || true
if echo "$UC_ERR" | grep -q "\-\-body is required"; then
  pass "usage create validation → requires --body"
else
  fail "usage create validation → unexpected: $UC_ERR"
fi

# 1d: usage create rejects stray args (NoArgs)
echo "  Testing: usage create with stray arg"
UC_NA=$($ZR usage create extraArg --body '{}' 2>&1) || true
if echo "$UC_NA" | grep -qi "unknown command\|too many arg"; then
  pass "usage create → rejects stray positional arg"
else
  fail "usage create → accepted stray arg: $UC_NA"
fi

# 1e: usage get missing argument
echo "  Testing: usage get without argument"
UG_ERR=$($ZR usage get 2>&1) || true
if echo "$UG_ERR" | grep -qi "arg\|required"; then
  pass "usage get validation → requires argument"
else
  fail "usage get validation → unexpected: $UG_ERR"
fi

# 1f: usage update missing --body
echo "  Testing: usage update without --body"
UU_ERR=$($ZR usage update FAKE-ID 2>&1) || true
if echo "$UU_ERR" | grep -q "\-\-body is required"; then
  pass "usage update validation → requires --body"
else
  fail "usage update validation → unexpected: $UU_ERR"
fi

# 1g: usage delete missing --confirm
echo "  Testing: usage delete without --confirm"
UD_ERR=$($ZR usage delete FAKE-ID 2>&1) || true
if echo "$UD_ERR" | grep -q "\-\-confirm"; then
  pass "usage delete validation → requires --confirm"
else
  fail "usage delete validation → unexpected: $UD_ERR"
fi

# 1h: usage delete missing argument
echo "  Testing: usage delete without argument"
UD_ERR2=$($ZR usage delete --confirm 2>&1) || true
if echo "$UD_ERR2" | grep -qi "arg\|required"; then
  pass "usage delete validation → requires argument"
else
  fail "usage delete validation → unexpected: $UD_ERR2"
fi

# ─────────────────────────────────────────
header "Step 2: Meter Validation"
# ─────────────────────────────────────────

# 2a: meter run missing arguments (requires 2)
echo "  Testing: meter run without arguments"
MR_ERR=$($ZR meter run 2>&1) || true
if echo "$MR_ERR" | grep -qi "arg\|required"; then
  pass "meter run validation → requires 2 arguments"
else
  fail "meter run validation → unexpected: $MR_ERR"
fi

# 2b: meter debug missing arguments
echo "  Testing: meter debug without arguments"
MD_ERR=$($ZR meter debug 2>&1) || true
if echo "$MD_ERR" | grep -qi "arg\|required"; then
  pass "meter debug validation → requires 2 arguments"
else
  fail "meter debug validation → unexpected: $MD_ERR"
fi

# 2c: meter status missing arguments
echo "  Testing: meter status without arguments"
MS_ERR=$($ZR meter status 2>&1) || true
if echo "$MS_ERR" | grep -qi "arg\|required"; then
  pass "meter status validation → requires 2 arguments"
else
  fail "meter status validation → unexpected: $MS_ERR"
fi

# 2d: meter summary missing --run-type
echo "  Testing: meter summary without --run-type"
MSU_ERR=$($ZR meter summary FAKE-ID 2>&1) || true
if echo "$MSU_ERR" | grep -qi "run-type.*required\|required.*run-type"; then
  pass "meter summary validation → requires --run-type"
else
  fail "meter summary validation → unexpected: $MSU_ERR"
fi

# 2e: meter summary missing argument
echo "  Testing: meter summary without argument"
MSU_ERR2=$($ZR meter summary 2>&1) || true
if echo "$MSU_ERR2" | grep -qi "arg\|required"; then
  pass "meter summary validation → requires argument"
else
  fail "meter summary validation → unexpected: $MSU_ERR2"
fi

# 2f: meter audit missing required flags
echo "  Testing: meter audit without required flags"
MA_ERR=$($ZR meter audit FAKE-ID 2>&1) || true
if echo "$MA_ERR" | grep -qi "required\|export-type\|run-type\|from\|to"; then
  pass "meter audit validation → requires flags"
else
  fail "meter audit validation → unexpected: $MA_ERR"
fi

# 2g: meter audit missing argument
echo "  Testing: meter audit without argument"
MA_ERR2=$($ZR meter audit 2>&1) || true
if echo "$MA_ERR2" | grep -qi "arg\|required"; then
  pass "meter audit validation → requires argument"
else
  fail "meter audit validation → unexpected: $MA_ERR2"
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
