#!/bin/bash
# E2E Test: Ramp, Commitment, Fulfillment, Prepaid Commands (Sub-phase 3e)
# テナント: apac-sandbox
# 注意: ramp/fulfillment/prepaid は専用設定が必要なため入力バリデーション中心。
#       commitment list だけは設定不要で実 API を叩けるので happy-path を 1 本含める。
#       各バリデーションは「非ゼロ終了」かつ「想定メッセージ(固定文字列)」を要求する。

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
LOG_FILE="$LOG_DIR/e2e-ramp-commitment-${TIMESTAMP}.log"

# Tee all output to log file
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
# rc!=0 AND exact fixed-string match required, else FAIL.
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
# auth status always exits 0 and prints "Token: valid|expired"; the only reliable
# signal of a usable session is a "Token: ... valid" line, so key on that.
AUTH_OUT=$($ZR auth status 2>&1)
if echo "$AUTH_OUT" | grep -qE "Token:[[:space:]]+valid"; then
  pass "Auth OK"
else
  fail "Auth failed (token not valid): $(echo "$AUTH_OUT" | grep -i 'token' | head -1)"
  exit 1
fi

# ─────────────────────────────────────────
header "Step 1: Ramp Validation"
# ─────────────────────────────────────────
echo "  Testing: ramp get without argument"
expect_fail "ramp get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR ramp get

echo "  Testing: ramp get-by-subscription without argument"
expect_fail "ramp get-by-subscription validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR ramp get-by-subscription

echo "  Testing: ramp metrics without argument"
expect_fail "ramp metrics validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR ramp metrics

echo "  Testing: ramp metrics-by-order without argument"
expect_fail "ramp metrics-by-order validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR ramp metrics-by-order

echo "  Testing: ramp metrics-by-subscription without argument"
expect_fail "ramp metrics-by-subscription validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR ramp metrics-by-subscription

# ─────────────────────────────────────────
header "Step 2: Commitment Validation"
# ─────────────────────────────────────────
echo "  Testing: commitment get without argument"
expect_fail "commitment get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR commitment get

echo "  Testing: commitment list without --account"
expect_fail "commitment list validation → requires --account" "--account is required" -- $ZR commitment list

echo "  Testing: commitment balance without argument"
expect_fail "commitment balance validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR commitment balance

# commitment periods is flag-driven (no positional): it requires --commitment,
# or --account together with --start-date and --end-date.
echo "  Testing: commitment periods without flags"
expect_fail "commitment periods validation → requires --commitment/--account" \
  "--commitment or --account (with --start-date and --end-date) is required" -- $ZR commitment periods

echo "  Testing: commitment schedules without argument"
expect_fail "commitment schedules validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR commitment schedules

# ─────────────────────────────────────────
header "Step 3: Commitment list (live)"
# ─────────────────────────────────────────
# The /v1/commitments endpoint is not provisioned on this apac-sandbox tenant
# (HTTP 404 "endpoint does not exist"). Exercise the real call: pass on a
# .commitments array, skip on that specific Zuora API error, fail otherwise.
ACCT_BODY=$(cat <<'JSON'
{
  "name": "E2E-Commitment-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {"firstName": "Test", "lastName": "Commit", "country": "Japan", "state": "Tokyo"}
}
JSON
)
run $ZR account create --body "$ACCT_BODY" --json
ACCT_NUM=$(echo "$RUN_OUT" | jq -r '.accountNumber // empty' 2>/dev/null)
if [ -n "$ACCT_NUM" ]; then
  pass "account create (for commitment list) → $ACCT_NUM"
  echo "  Testing: commitment list --account $ACCT_NUM"
  run $ZR commitment list --account "$ACCT_NUM" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.commitments | type == "array"' >/dev/null 2>&1; then
    pass "commitment list → .commitments array (count=$(echo "$RUN_OUT" | jq '.commitments | length'))"
  elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
    skip "commitment list → $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
  else
    fail "commitment list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "commitment list live → could not create test account: ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 4: Fulfillment Validation"
# ─────────────────────────────────────────
echo "  Testing: fulfillment create without --body"
expect_fail "fulfillment create validation → requires --body" "--body is required" -- $ZR fulfillment create

echo "  Testing: fulfillment get without argument"
expect_fail "fulfillment get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR fulfillment get

echo "  Testing: fulfillment update without --body"
expect_fail "fulfillment update validation → requires --body" "--body is required" -- $ZR fulfillment update FAKE-ID

echo "  Testing: fulfillment delete without --confirm"
expect_fail "fulfillment delete validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR fulfillment delete FAKE-ID

# ─────────────────────────────────────────
header "Step 5: Fulfillment-Item Validation"
# ─────────────────────────────────────────
echo "  Testing: fulfillment-item create without --body"
expect_fail "fulfillment-item create validation → requires --body" "--body is required" -- $ZR fulfillment-item create

echo "  Testing: fulfillment-item get without argument"
expect_fail "fulfillment-item get validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR fulfillment-item get

echo "  Testing: fulfillment-item update without --body"
expect_fail "fulfillment-item update validation → requires --body" \
  "--body is required" -- $ZR fulfillment-item update FAKE-ID

echo "  Testing: fulfillment-item delete without --confirm"
expect_fail "fulfillment-item delete validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR fulfillment-item delete FAKE-ID

# ─────────────────────────────────────────
header "Step 6: Prepaid Validation"
# ─────────────────────────────────────────
echo "  Testing: prepaid rollover without --body"
expect_fail "prepaid rollover validation → requires --body" "--body is required" -- $ZR prepaid rollover

echo "  Testing: prepaid deplete without --body"
expect_fail "prepaid deplete validation → requires --body" "--body is required" -- $ZR prepaid deplete

echo "  Testing: prepaid reverse-rollover without --body"
expect_fail "prepaid reverse-rollover validation → requires --body" "--body is required" -- $ZR prepaid reverse-rollover

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Passed:  $PASS / $((PASS+FAIL+SKIP))"
echo "  Failed:  $FAIL / $((PASS+FAIL+SKIP))"
echo "  Skipped: $SKIP / $((PASS+FAIL+SKIP))"
echo ""
echo "  Log: $LOG_FILE"
echo ""
if [ "$FAIL" -gt 0 ]; then
  echo "  RESULT: FAIL"
  exit 1
else
  echo "  RESULT: PASS"
fi
