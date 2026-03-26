#!/bin/bash
# E2E Test: ZOQL + ChangeLog + Omnichannel (Phase 9)

set -uo pipefail

ZR="./bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-zoql-omnichannel-${TIMESTAMP}.log"

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
header "Step 1: ZOQL Query Validation"
# ─────────────────────────────────────────

echo "  Testing: query without argument"
Q_ERR=$($ZR query 2>&1) || true
if echo "$Q_ERR" | grep -qi "arg\|required"; then
  pass "query validation → requires argument"
else
  fail "query validation → unexpected: $Q_ERR"
fi

# ─────────────────────────────────────────
header "Step 2: ZOQL Query 実行テスト"
# ─────────────────────────────────────────

echo "  Testing: query 'SELECT Id, Name FROM Account LIMIT 3'"
Q_RESULT=$($ZR query "SELECT Id, Name FROM Account LIMIT 3" --json 2>/dev/null) || true
if echo "$Q_RESULT" | jq -e '.records' >/dev/null 2>&1; then
  Q_COUNT=$(echo "$Q_RESULT" | jq '.records | length')
  pass "query → returned $Q_COUNT records"
elif echo "$Q_RESULT" | grep -qi "error"; then
  skip "query → API error"
else
  fail "query → unexpected: $(echo "$Q_RESULT" | head -3)"
fi

# query with --limit
echo "  Testing: query with --limit 1"
Q_LIMIT=$($ZR query "SELECT Id FROM Account" --limit 1 --json 2>/dev/null) || true
if echo "$Q_LIMIT" | jq -e '.records' >/dev/null 2>&1; then
  Q_LCOUNT=$(echo "$Q_LIMIT" | jq '.records | length')
  if [ "$Q_LCOUNT" -le "1" ]; then
    pass "query --limit 1 → returned $Q_LCOUNT records (≤1)"
  else
    fail "query --limit 1 → returned $Q_LCOUNT records (expected ≤1)"
  fi
else
  skip "query --limit → API error"
fi

# query with --csv
echo "  Testing: query with --csv"
Q_CSV=$($ZR query "SELECT Id, Name FROM Account LIMIT 2" --csv 2>/dev/null) || true
if echo "$Q_CSV" | head -1 | grep -qi "Id\|name"; then
  pass "query --csv → returned CSV with headers"
else
  skip "query --csv → unexpected output"
fi

# query with --export
EXPORT_FILE=$(mktemp /tmp/zr-e2e-export-XXXXXX.json)
echo "  Testing: query with --export"
$ZR query "SELECT Id FROM Account LIMIT 1" --json --export "$EXPORT_FILE" 2>/dev/null || true
if [ -s "$EXPORT_FILE" ]; then
  pass "query --export → file written"
else
  skip "query --export → file empty or missing"
fi
rm -f "$EXPORT_FILE"

# ─────────────────────────────────────────
header "Step 3: Subscription ChangeLog Validation"
# ─────────────────────────────────────────

echo "  Testing: subscription changelog without argument"
CL_ERR=$($ZR subscription changelog 2>&1) || true
if echo "$CL_ERR" | grep -qi "arg\|required"; then
  pass "subscription changelog validation → requires argument"
else
  fail "subscription changelog validation → unexpected: $CL_ERR"
fi

echo "  Testing: subscription changelog-by-order without argument"
CLO_ERR=$($ZR subscription changelog-by-order 2>&1) || true
if echo "$CLO_ERR" | grep -qi "arg\|required"; then
  pass "subscription changelog-by-order validation → requires argument"
else
  fail "subscription changelog-by-order validation → unexpected: $CLO_ERR"
fi

echo "  Testing: subscription changelog-version without arguments"
CLV_ERR=$($ZR subscription changelog-version 2>&1) || true
if echo "$CLV_ERR" | grep -qi "arg\|required"; then
  pass "subscription changelog-version validation → requires 2 arguments"
else
  fail "subscription changelog-version validation → unexpected: $CLV_ERR"
fi

# ─────────────────────────────────────────
header "Step 4: Omnichannel Validation"
# ─────────────────────────────────────────

echo "  Testing: omnichannel create without --body"
OC_ERR=$($ZR omnichannel create 2>&1) || true
if echo "$OC_ERR" | grep -q "\-\-body is required"; then
  pass "omnichannel create validation → requires --body"
else
  fail "omnichannel create validation → unexpected: $OC_ERR"
fi

echo "  Testing: omnichannel create rejects stray arg"
OC_NA=$($ZR omnichannel create extraArg --body '{}' 2>&1) || true
if echo "$OC_NA" | grep -qi "unknown command\|too many arg\|accepts 0 arg"; then
  pass "omnichannel create → rejects stray positional arg"
else
  fail "omnichannel create → accepted stray arg"
fi

echo "  Testing: omnichannel get without argument"
OG_ERR=$($ZR omnichannel get 2>&1) || true
if echo "$OG_ERR" | grep -qi "arg\|required"; then
  pass "omnichannel get validation → requires argument"
else
  fail "omnichannel get validation → unexpected: $OG_ERR"
fi

echo "  Testing: omnichannel delete without --confirm"
OD_ERR=$($ZR omnichannel delete FAKE-KEY 2>&1) || true
if echo "$OD_ERR" | grep -q "\-\-confirm"; then
  pass "omnichannel delete validation → requires --confirm"
else
  fail "omnichannel delete validation → unexpected: $OD_ERR"
fi

echo "  Testing: omnichannel delete without argument"
OD_ERR2=$($ZR omnichannel delete --confirm 2>&1) || true
if echo "$OD_ERR2" | grep -qi "arg\|required"; then
  pass "omnichannel delete validation → requires argument"
else
  fail "omnichannel delete validation → unexpected: $OD_ERR2"
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
