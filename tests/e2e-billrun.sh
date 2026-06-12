#!/bin/bash
# E2E Test: Bill Run lifecycle
# テナント: apac-sandbox
#
# create → (Processing → Completed) → cancel → (Canceled) → delete を実走する。
# targetDate を過去日 (2000-01-01) にすることで請求書を 1 枚も生成しない安全な
# 実行になる(全アカウントを走査するが対象チャージが存在しない)。post は
# 請求書を確定させるため実走しない(validation のみ)。
#
# このスイートは #220(ボディなし PUT に {} を送る 415 修正)の恒久ゲート:
# billrun cancel は修正前は全ての live 呼び出しが HTTP 415 で失敗していた。
# 注意: cancel のレスポンスは成功を返すが Canceled への遷移は数秒遅れる
# (delete は Canceled 状態のみ受け付ける)ため、間にポーリングを挟む。

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-billrun-${TIMESTAMP}.log"

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

# ─────────────────────────────────────────
header "Step 1: Validation (no API calls)"
# ─────────────────────────────────────────
echo "  Testing: billrun create without --body"
expect_fail "create validation → requires --body" "--body is required" -- $ZR billrun create

echo "  Testing: billrun get without arg"
expect_fail "get validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR billrun get

echo "  Testing: billrun post without arg"
expect_fail "post validation → requires arg" "accepts 1 arg(s), received 0" -- $ZR billrun post

echo "  Testing: billrun cancel without --confirm"
expect_fail "cancel → requires --confirm" "Use --confirm to proceed" -- $ZR billrun cancel FAKE-BR-ID

echo "  Testing: billrun delete without --confirm"
expect_fail "delete → requires --confirm" "Use --confirm to proceed" -- $ZR billrun delete FAKE-BR-ID

# ─────────────────────────────────────────
header "Step 2: billrun create (zero-invoice run, targetDate=2000-01-01)"
# ─────────────────────────────────────────
run $ZR billrun create --body '{"batches":["AllBatches"],"targetDate":"2000-01-01"}' --json
BR_ID=$(echo "$RUN_OUT" | jq -r '.id // empty' 2>/dev/null)
if [ "$RUN_RC" -eq 0 ] && [ -n "$BR_ID" ]; then
  pass "billrun create → $BR_ID"
else
  fail "billrun create → rc=$RUN_RC: $(printf '%s' "${RUN_ERR:-$RUN_OUT}" | head -1)"
  printf '\n'
  red "Cannot proceed without a bill run. Aborting."
  print_summary
  exit 1
fi

# ─────────────────────────────────────────
header "Step 3: billrun get (poll to Completed)"
# ─────────────────────────────────────────
echo "  Polling status (a zero-match run still scans every account; allow ~150s)"
BR_STATUS=""
for _ in $(seq 1 30); do
  BR_STATUS=$($ZR billrun get "$BR_ID" --jq '.status' 2>/dev/null | tr -d '"')
  case "$BR_STATUS" in Completed|Error) break;; esac
  sleep 5
done
if [ "$BR_STATUS" = "Completed" ]; then
  pass "billrun get → status reached Completed"
else
  fail "billrun get → status '$BR_STATUS' (expected Completed within 150s)"
fi

# ─────────────────────────────────────────
header "Step 4: billrun cancel (#220 415-fix gate)"
# ─────────────────────────────────────────
echo "  Testing: billrun cancel $BR_ID --confirm"
CANCEL_RC=0
CANCEL_OUT=$($ZR billrun cancel "$BR_ID" --confirm 2>&1) || CANCEL_RC=$?
if [ "$CANCEL_RC" -eq 0 ] && echo "$CANCEL_OUT" | grep -qF "cancelled."; then
  pass "billrun cancel → succeeded (bodyless PUT carries Content-Type + {})"
elif echo "$CANCEL_OUT" | grep -q "HTTP 415\|50000045"; then
  fail "billrun cancel → 415 REGRESSION (empty-JSON body lost, cf. #220): $(echo "$CANCEL_OUT" | head -2)"
else
  fail "billrun cancel → rc=$CANCEL_RC: $(echo "$CANCEL_OUT" | head -2)"
fi

# The Canceled transition lags the cancel response by a few seconds, and
# delete only accepts Canceled bill runs — poll before deleting.
for _ in $(seq 1 12); do
  BR_STATUS=$($ZR billrun get "$BR_ID" --jq '.status' 2>/dev/null | tr -d '"')
  [ "$BR_STATUS" = "Canceled" ] && break
  sleep 5
done
if [ "$BR_STATUS" = "Canceled" ]; then
  pass "billrun get → status reached Canceled after cancel"
else
  fail "billrun get → status '$BR_STATUS' (expected Canceled within 60s of cancel)"
fi

# ─────────────────────────────────────────
header "Step 5: billrun delete (unified RenderDeleteResult output)"
# ─────────────────────────────────────────
echo "  Testing: billrun delete $BR_ID --confirm"
DEL_RC=0
DEL_OUT=$($ZR billrun delete "$BR_ID" --confirm 2>&1) || DEL_RC=$?
if [ "$DEL_RC" -eq 0 ] && echo "$DEL_OUT" | grep -qF "deleted."; then
  pass "billrun delete → unified delete message (RenderDeleteResult)"
else
  fail "billrun delete → rc=$DEL_RC: $(echo "$DEL_OUT" | head -2)"
  yellow "  ! leftover bill run may remain: $BR_ID (delete manually)"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
echo "  Bill run: $BR_ID"
echo ""
print_summary
