#!/bin/bash
# E2E Test: Order Commands (Sub-phase 3c)
# テナント: apac-sandbox (Orders 有効)
#
# Order ライフサイクル: create → get → list → preview、および mutating コマンドの
# 入力バリデーション。happy-path は stdout(JSON) / stderr / 終了コードを分離して捕捉し、
# 失敗時に必ず原因が見えるようにする。
#
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
# 月払いスタータープラン。ZR_E2E_RATE_PLAN_ID で上書き可能。カタログ変更で stale に
# なったときは Step 2 の order create が fail し、原因が CLI 退行と区別できる。
RATE_PLAN_ID="${ZR_E2E_RATE_PLAN_ID:-4c6059a8d8899f453ffa0637451d0003}"

# Log directory
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-order-${TIMESTAMP}.log"

# Tee all output to log file
exec > >(tee >(sed 's/\x1b\[[0-9;]*m//g' > "$LOG_FILE")) 2>&1
LOG_TEE_PID=$!
# Drain the tee/sed log pipeline on exit (sed block-buffers to a file;
# without a clean EOF + wait the tail of the log is silently truncated).
_drain_log() { exec 1>&- 2>&-; wait "$LOG_TEE_PID" 2>/dev/null || true; }
trap _drain_log EXIT

green()  { printf "\033[32m%s\033[0m\n" "$1"; }
red()    { printf "\033[31m%s\033[0m\n" "$1"; }
yellow() { printf "\033[33m%s\033[0m\n" "$1"; }

pass() { PASS=$((PASS+1)); green "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); red   "  ✗ $1"; }
skip() { SKIP=$((SKIP+1)); yellow "  ⊘ $1 (skipped)"; }

header() { printf "\n\033[1m=== %s ===\033[0m\n" "$1"; }

# run <command...> — capture stdout→RUN_OUT (clean, for jq), stderr→RUN_ERR
# (shown only on failure), exit code→RUN_RC. Keeps JSON parsing reliable while
# making every failure diagnosable (the old 2>/dev/null discarded the reason).
RUN_OUT=""; RUN_ERR=""; RUN_RC=0
run() {
  local ef="$LOG_DIR/.run.$$.err"
  RUN_OUT=$("$@" 2>"$ef"); RUN_RC=$?
  RUN_ERR=$(cat "$ef" 2>/dev/null); rm -f "$ef"
}

# expect_fail <description> <expected-substring> -- <command...>
# Passes only when the command exits non-zero AND output contains the exact
# expected substring (fixed-string). Catches regressions that drop validation,
# print help, or exit 0 — which a loose 'grep -qi arg|required' would not.
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
header "Step 1: Account Setup (テスト用アカウント)"
# ─────────────────────────────────────────
ACCT_BODY=$(cat <<'JSON'
{
  "name": "E2E-Order-Test",
  "currency": "JPY",
  "billCycleDay": 1,
  "autoPay": false,
  "billToContact": {
    "firstName": "Test",
    "lastName": "Order",
    "country": "Japan",
    "state": "Tokyo"
  }
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
  printf '\n'
  red "Cannot proceed without a test account. Aborting."
  exit 1
fi

echo "  Account: $ACCT_NUM ($ACCT_ID)"

# ─────────────────────────────────────────
header "Step 2: order create"
# ─────────────────────────────────────────
ORDER_BODY=$(cat <<EOF
{
  "existingAccountNumber": "$ACCT_NUM",
  "orderDate": "$(date +%Y-%m-%d)",
  "subscriptions": [
    {
      "orderActions": [
        {
          "type": "CreateSubscription",
          "triggerDates": [
            {"name": "ServiceActivation", "triggerDate": "$(date +%Y-%m-%d)"},
            {"name": "CustomerAcceptance", "triggerDate": "$(date +%Y-%m-%d)"}
          ],
          "createSubscription": {
            "terms": {
              "initialTerm": {"period": 12, "periodType": "Month", "termType": "TERMED", "startDate": "$(date +%Y-%m-%d)"},
              "renewalTerms": [{"period": 12, "periodType": "Month"}],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": false
            },
            "subscribeToRatePlans": [{"productRatePlanId": "$RATE_PLAN_ID"}]
          }
        }
      ]
    }
  ]
}
EOF
)
run $ZR order create --body "$ORDER_BODY" --json
ORDER_NUM=$(echo "$RUN_OUT" | jq -r '.orderNumber // empty' 2>/dev/null)
SUB_NUM=$(echo "$RUN_OUT" | jq -r '.subscriptions[0].subscriptionNumber // empty' 2>/dev/null)

if [ -n "$ORDER_NUM" ]; then
  pass "order create → $ORDER_NUM (sub: $SUB_NUM)"
else
  fail "order create (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 3: order get"
# ─────────────────────────────────────────
if [ -n "$ORDER_NUM" ]; then
  echo "  Testing: order get $ORDER_NUM"
  run $ZR order get "$ORDER_NUM" --json
  GET_STATUS=$(echo "$RUN_OUT" | jq -r '.order.status // .status // empty' 2>/dev/null)
  if [ -n "$GET_STATUS" ]; then
    pass "order get → status=$GET_STATUS"
  else
    fail "order get (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "order get → no order number"
fi

# ─────────────────────────────────────────
header "Step 4: order list"
# ─────────────────────────────────────────
# order list has no account filter (only --status/--page); assert it returns a
# tenant-wide .orders array. The specific created order is verified in Step 3.
echo "  Testing: order list"
run $ZR order list --json
LIST_COUNT=$(echo "$RUN_OUT" | jq -r '.orders | length' 2>/dev/null)
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
  pass "order list → .orders array (count=$LIST_COUNT)"
else
  fail "order list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 5: order list-pending"
# ─────────────────────────────────────────
# list-pending takes a <subscription-key> argument.
echo "  Testing: order list-pending validation (no arg)"
expect_fail "order list-pending validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-pending

if [ -n "$SUB_NUM" ]; then
  echo "  Testing: order list-pending $SUB_NUM"
  run $ZR order list-pending "$SUB_NUM" --json
  if echo "$RUN_OUT" | jq -e '.' >/dev/null 2>&1; then
    pass "order list-pending → returned JSON"
  elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
    skip "order list-pending → Zuora API error: $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
  else
    fail "order list-pending (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "order list-pending → no subscription number from order create"
fi

# ─────────────────────────────────────────
header "Step 6: Output formats"
# ─────────────────────────────────────────
if [ -n "$ORDER_NUM" ]; then
  echo "  Testing: order get --jq '.order.orderNumber'"
  # --jq emits the JSON value verbatim (a quoted string), so assert containment.
  run $ZR order get "$ORDER_NUM" --jq '.order.orderNumber'
  if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | grep -qF "$ORDER_NUM"; then
    pass "order get --jq → $RUN_OUT"
  else
    fail "order get --jq → expected to contain $ORDER_NUM, got '$RUN_OUT' (rc=$RUN_RC) ${RUN_ERR}"
  fi

  echo "  Testing: order get --template '{{.order.orderNumber}}'"
  # --template renders a Go text/template against the JSON; the rendered value
  # is emitted raw (unquoted), so assert exact equality.
  run $ZR order get "$ORDER_NUM" --template '{{.order.orderNumber}}'
  if [ "$RUN_RC" -eq 0 ] && [ "$RUN_OUT" = "$ORDER_NUM" ]; then
    pass "order get --template → $RUN_OUT"
  else
    fail "order get --template → expected $ORDER_NUM, got '$RUN_OUT' (rc=$RUN_RC) ${RUN_ERR}"
  fi
else
  skip "order get --jq → no order number"
  skip "order get --template → no order number"
fi

# ─────────────────────────────────────────
header "Step 6.5: --body resolution (@file / stdin / literal)"
# ─────────────────────────────────────────
# cmdutil.ResolveBody: "@file" reads a file, "-" reads stdin, else literal JSON.
# Every mutating command shares this path, so exercise all three forms (plus the
# @missing error) via account create, which needs no special tenant setup.
BODY_JSON='{"name":"E2E-BodyResolve","currency":"JPY","billCycleDay":1,"autoPay":false,"billToContact":{"firstName":"B","lastName":"R","country":"Japan","state":"Tokyo"}}'

echo "  Testing: --body @file"
BODY_FILE="$LOG_DIR/.body.$$.json"
printf '%s' "$BODY_JSON" > "$BODY_FILE"
run $ZR account create --body "@$BODY_FILE" --json
rm -f "$BODY_FILE"
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.accountNumber' >/dev/null 2>&1; then
  pass "--body @file → created $(echo "$RUN_OUT" | jq -r '.accountNumber')"
else
  fail "--body @file (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: --body - (stdin)"
RUN_OUT=$(printf '%s' "$BODY_JSON" | $ZR account create --body - --json 2>"$LOG_DIR/.berr.$$"); RUN_RC=$?
RUN_ERR=$(cat "$LOG_DIR/.berr.$$" 2>/dev/null); rm -f "$LOG_DIR/.berr.$$"
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.accountNumber' >/dev/null 2>&1; then
  pass "--body - (stdin) → created $(echo "$RUN_OUT" | jq -r '.accountNumber')"
else
  fail "--body - (stdin) (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: --body @nonexistent (error path)"
expect_fail "--body @missing → reading file error" "reading body file:" -- $ZR account create --body @/nonexistent/body.json

# ─────────────────────────────────────────
header "Step 7: order preview (read-only)"
# ─────────────────────────────────────────
PREVIEW_BODY=$(cat <<EOF
{
  "existingAccountNumber": "$ACCT_NUM",
  "orderDate": "$(date +%Y-%m-%d)",
  "previewOptions": {"previewThroughType": "SpecificDate", "specificPreviewThroughDate": "$(date -v+1m +%Y-%m-%d 2>/dev/null || date -d '+1 month' +%Y-%m-%d)"},
  "subscriptions": [
    {
      "orderActions": [
        {
          "type": "CreateSubscription",
          "triggerDates": [
            {"name": "ServiceActivation", "triggerDate": "$(date +%Y-%m-%d)"},
            {"name": "CustomerAcceptance", "triggerDate": "$(date +%Y-%m-%d)"}
          ],
          "createSubscription": {
            "terms": {
              "initialTerm": {"period": 12, "periodType": "Month", "termType": "TERMED", "startDate": "$(date +%Y-%m-%d)"},
              "renewalTerms": [{"period": 12, "periodType": "Month"}],
              "renewalSetting": "RENEW_WITH_SPECIFIC_TERM",
              "autoRenew": false
            },
            "subscribeToRatePlans": [{"productRatePlanId": "$RATE_PLAN_ID"}]
          }
        }
      ]
    }
  ]
}
EOF
)
echo "  Testing: order preview"
run $ZR order preview --body "$PREVIEW_BODY" --json
if echo "$RUN_OUT" | jq -e '.' >/dev/null 2>&1; then
  pass "order preview → returned JSON"
elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
  # Tenant/account/rate-plan-specific rejection — narrow, status-specific skip
  # (not a blanket 'grep -qi error' that would also swallow a CLI bug).
  skip "order preview → Zuora API error: $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
else
  fail "order preview (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 8: Validation (read-only checks)"
# ─────────────────────────────────────────
# All checks below require non-zero exit AND the exact CLI/cobra message. They
# never mutate anything (bad/missing args abort before any API call).

echo "  Testing: order get without arg"
expect_fail "order get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order get

echo "  Testing: order create without --body"
expect_fail "order create validation → requires --body" "--body is required" -- $ZR order create

echo "  Testing: order activate without arg"
expect_fail "order activate validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order activate

echo "  Testing: order cancel without arg"
expect_fail "order cancel validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order cancel

echo "  Testing: order cancel without --confirm"
expect_fail "order cancel validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR order cancel O-FAKE

echo "  Testing: order update without arg"
expect_fail "order update validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order update

echo "  Testing: order update without --body"
expect_fail "order update validation → requires --body" "--body is required" -- $ZR order update O-FAKE

echo "  Testing: order update-custom-fields without arg"
expect_fail "order update-custom-fields validation → requires arg" \
  "accepts 1 arg(s), received 0" -- $ZR order update-custom-fields

echo "  Testing: order update-custom-fields without --body"
expect_fail "order update-custom-fields validation → requires --body" \
  "--body is required" -- $ZR order update-custom-fields O-FAKE

echo "  Testing: order update-trigger-dates without arg"
expect_fail "order update-trigger-dates validation → requires arg" \
  "accepts 1 arg(s), received 0" -- $ZR order update-trigger-dates

echo "  Testing: order update-trigger-dates without --body"
expect_fail "order update-trigger-dates validation → requires --body" \
  "--body is required" -- $ZR order update-trigger-dates O-FAKE

echo "  Testing: order revert without arg"
expect_fail "order revert validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order revert

echo "  Testing: order create-async without --body"
expect_fail "order create-async validation → requires --body" "--body is required" -- $ZR order create-async

echo "  Testing: order preview-async without --body"
expect_fail "order preview-async validation → requires --body" "--body is required" -- $ZR order preview-async

echo "  Testing: order delete-async without arg"
expect_fail "order delete-async validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order delete-async

echo "  Testing: order delete without arg"
expect_fail "order delete validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order delete

echo "  Testing: order delete without --confirm"
expect_fail "order delete validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR order delete O-FAKE

echo "  Testing: order job-status without arg"
expect_fail "order job-status validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order job-status

echo "  Testing: order-action update without arg"
expect_fail "order-action update validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order-action update

echo "  Testing: order-line-item get without arg"
expect_fail "order-line-item get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order-line-item get

echo "  Testing: order-line-item update without arg"
expect_fail "order-line-item update validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order-line-item update

echo "  Testing: order-line-item bulk-update without --body"
expect_fail "order-line-item bulk-update validation → requires --body" "--body is required" -- $ZR order-line-item bulk-update

# ─────────────────────────────────────────
header "Step 9: order list-by-* (live reads)"
# ─────────────────────────────────────────
echo "  Testing: order list-by-subscription validation (no arg)"
expect_fail "order list-by-subscription validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-by-subscription

echo "  Testing: order list-by-invoice-owner validation (no arg)"
expect_fail "order list-by-invoice-owner validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-by-invoice-owner

echo "  Testing: order list-by-subscription-owner validation (no arg)"
expect_fail "order list-by-subscription-owner validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR order list-by-subscription-owner

if [ -n "$SUB_NUM" ]; then
  echo "  Testing: order list-by-subscription $SUB_NUM"
  run $ZR order list-by-subscription "$SUB_NUM" --json
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
    pass "order list-by-subscription → .orders array"
  else
    fail "order list-by-subscription (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "order list-by-subscription → no subscription number"
fi

echo "  Testing: order list-by-subscription-owner $ACCT_NUM"
run $ZR order list-by-subscription-owner "$ACCT_NUM" --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
  pass "order list-by-subscription-owner → .orders array"
else
  fail "order list-by-subscription-owner (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: order list-by-invoice-owner $ACCT_NUM"
run $ZR order list-by-invoice-owner "$ACCT_NUM" --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.orders | type == "array"' >/dev/null 2>&1; then
  pass "order list-by-invoice-owner → .orders array"
else
  fail "order list-by-invoice-owner (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Test Account: $ACCT_NUM"
echo "  Order: $ORDER_NUM"
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
