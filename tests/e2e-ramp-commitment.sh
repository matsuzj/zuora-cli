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

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

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
  elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "50000040"; then
    # ONLY the documented endpoint-missing error (feature off on this tenant)
    # may skip; any other Zuora API error is a real failure.
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
print_summary
