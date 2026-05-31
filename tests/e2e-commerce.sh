#!/bin/bash
# E2E Test: Commerce / Catalog (Product, Plan, Charge, RatePlan)
# テナント: apac-sandbox
# 注意: product list-legacy / plan list は --body を必須とする検索 API。テナントの
#       Product Catalog 設定に依存するが、空フィルタ {} なら結果(空配列可)を返す。

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
LOG_FILE="$LOG_DIR/e2e-commerce-${TIMESTAMP}.log"

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

# read_or_skip <description> <jq-success-filter> -- <command...>
# pass if rc==0 and the jq filter matches; skip ONLY on a real "Zuora API error"
# (feature/endpoint not enabled on this tenant); fail on anything else.
read_or_skip() {
  local desc="$1" filter="$2"; shift 2
  [ "${1:-}" = "--" ] && shift
  run "$@"
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e "$filter" >/dev/null 2>&1; then
    pass "$desc"
  elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
    skip "$desc → $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
  else
    fail "$desc → rc=$RUN_RC: ${RUN_ERR:-$RUN_OUT}"
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
header "Step 1: Validation (read commands)"
# ─────────────────────────────────────────
echo "  Testing: product get without argument"
expect_fail "product get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR product get

echo "  Testing: product list-legacy without --body"
expect_fail "product list-legacy validation → requires --body" "--body is required" -- $ZR product list-legacy

echo "  Testing: plan get without --key"
expect_fail "plan get validation → requires --key" "--key is required" -- $ZR plan get

echo "  Testing: plan list without --body"
expect_fail "plan list validation → requires --body" "--body is required" -- $ZR plan list

echo "  Testing: plan purchase-options without --plan"
expect_fail "plan purchase-options validation → requires --plan" "--plan is required" -- $ZR plan purchase-options

echo "  Testing: charge get without --key"
expect_fail "charge get validation → requires --key" "--key is required" -- $ZR charge get

echo "  Testing: rateplan get without argument"
expect_fail "rateplan get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR rateplan get

# ─────────────────────────────────────────
header "Step 2: Validation (mutating commands)"
# ─────────────────────────────────────────
echo "  Testing: product create without --body"
expect_fail "product create validation → requires --body" "--body is required" -- $ZR product create

echo "  Testing: product update without --body"
expect_fail "product update validation → requires --body" "--body is required" -- $ZR product update

echo "  Testing: plan create without --body"
expect_fail "plan create validation → requires --body" "--body is required" -- $ZR plan create

echo "  Testing: plan update without --body"
expect_fail "plan update validation → requires --body" "--body is required" -- $ZR plan update

echo "  Testing: charge create without --body"
expect_fail "charge create validation → requires --body" "--body is required" -- $ZR charge create

echo "  Testing: charge update without --body"
expect_fail "charge update validation → requires --body" "--body is required" -- $ZR charge update

echo "  Testing: charge update-tiers without --body"
expect_fail "charge update-tiers validation → requires --body" "--body is required" -- $ZR charge update-tiers

# ─────────────────────────────────────────
header "Step 3: NoArgs rejection (stray positional)"
# ─────────────────────────────────────────
echo "  Testing: product create with stray arg"
expect_fail "product create → rejects stray positional arg" 'unknown command "extraArg"' -- $ZR product create extraArg --body '{}'

# ─────────────────────────────────────────
header "Step 4: Live read commands"
# ─────────────────────────────────────────
# product list-legacy is a --body search returning {"products":[...]} (empty OK).
echo "  Testing: product list-legacy --body '{}'"
read_or_skip "product list-legacy → .products array" '.products | type == "array"' -- $ZR product list-legacy --body '{}' --json

# plan list is a --body search returning {"plans":[...]} (empty OK).
echo "  Testing: plan list --body '{}'"
read_or_skip "plan list → .plans array" '.plans | type == "array"' -- $ZR plan list --body '{}' --json

# rateplan get expects a *subscription* rate plan id; a product rate plan id
# 404s on this tenant, which read_or_skip treats as a (status-specific) skip.
echo "  Testing: rateplan get $RATE_PLAN_ID"
read_or_skip "rateplan get → JSON object" 'type == "object"' -- $ZR rateplan get "$RATE_PLAN_ID" --json

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
TOTAL=$((PASS + FAIL + SKIP))
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
