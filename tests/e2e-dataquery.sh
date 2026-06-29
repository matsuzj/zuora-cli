#!/bin/bash
# E2E Test: Data Query (zr data-query) + read-only opt-in
# テナント: apac-sandbox (Account テーブルにシードデータあり)
#
# Covers: submit / get / run(+S3 download) / list / cancel / failed job, and the
# --read-only-allow-data-query / ZR_READ_ONLY_ALLOW_DATA_QUERY opt-in (blocked by
# default, allowed with the toggle, never widens other writes).

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-dataquery-${TIMESTAMP}.log"

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

DQ_SQL="SELECT accountnumber FROM account"

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

# ─────────────────────────────────────────
header "Step 1: Validation (offline / cobra-level)"
# ─────────────────────────────────────────
echo "  Testing: submit without SQL and without --file"
expect_fail "submit validation → requires SQL or --file" "provide the SQL" -- $ZR data-query submit

CONFLICT_FILE=$(mktemp); printf 'SELECT 1' > "$CONFLICT_FILE"
echo "  Testing: submit with both positional SQL and --file"
expect_fail "submit validation → SQL XOR --file" "not both" -- $ZR data-query submit "SELECT 1" --file "$CONFLICT_FILE"
rm -f "$CONFLICT_FILE"

echo "  Testing: get without argument"
expect_fail "get validation → requires job-id" "accepts 1 arg(s), received 0" -- $ZR data-query get

echo "  Testing: cancel without argument"
expect_fail "cancel validation → requires job-id" "accepts 1 arg(s), received 0" -- $ZR data-query cancel

echo "  Testing: cancel without --confirm"
expect_fail "cancel validation → requires --confirm" "this action is irreversible" -- $ZR data-query cancel FAKE-JOB-ID

# ─────────────────────────────────────────
header "Step 2: submit (live)"
# ─────────────────────────────────────────
echo "  Testing: data-query submit"
run_retry 3 $ZR data-query submit "$DQ_SQL" --json
JOB_ID=$(printf '%s' "$RUN_OUT" | jq -r '.data.id // empty' 2>/dev/null)
if [ "$RUN_RC" -eq 0 ] && [ -n "$JOB_ID" ] \
   && printf '%s' "$RUN_OUT" | jq -e '.data.queryStatus == "accepted"' >/dev/null 2>&1; then
  pass "submit → job accepted (id: $JOB_ID)"
else
  fail "submit (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 3: get / poll until terminal (live)"
# ─────────────────────────────────────────
if [ -n "${JOB_ID:-}" ]; then
  DQ_STATUS=""
  for _ in 1 2 3 4 5 6 7 8 9 10; do
    run $ZR data-query get "$JOB_ID" --json
    DQ_STATUS=$(printf '%s' "$RUN_OUT" | jq -r '.data.queryStatus // empty' 2>/dev/null)
    case "$DQ_STATUS" in completed|failed|cancelled) break;; esac
    sleep 2
  done
  echo "  Testing: get → completed with numeric outputRows + dataFile"
  if [ "$DQ_STATUS" = "completed" ] \
     && printf '%s' "$RUN_OUT" | jq -e '(.data.outputRows | type == "number") and (.data.dataFile | type == "string")' >/dev/null 2>&1; then
    ROWS=$(printf '%s' "$RUN_OUT" | jq -r '.data.outputRows')
    pass "get → completed, outputRows=$ROWS (number), dataFile present"
  else
    fail "get → status='$DQ_STATUS' (rc=$RUN_RC): $(printf '%s' "${RUN_ERR:-$RUN_OUT}" | head -1)"
  fi
else
  skip "get → no job id from submit"
fi

# ─────────────────────────────────────────
header "Step 4: run = submit → poll → S3 download (live)"
# ─────────────────────────────────────────
DQ_OUT=$(mktemp)
echo "  Testing: data-query run --output (downloads the S3 dataFile)"
run $ZR data-query run "$DQ_SQL LIMIT 5" --output "$DQ_OUT" --interval 2s
if [ "$RUN_RC" -eq 0 ] && [ -s "$DQ_OUT" ] \
   && head -1 "$DQ_OUT" | jq -e '.accountnumber' >/dev/null 2>&1 \
   && printf '%s' "$RUN_ERR" | grep -qF "completed"; then
  pass "run → downloaded $(wc -l < "$DQ_OUT" | tr -d ' ') JSONL rows with accountnumber"
else
  fail "run (rc=$RUN_RC) → file '$(head -1 "$DQ_OUT" 2>/dev/null)' err: $(printf '%s' "$RUN_ERR" | head -1)"
fi
rm -f "$DQ_OUT"

echo "  Testing: data-query run --output - (raw bytes to stdout, no metadata)"
run $ZR data-query run "$DQ_SQL LIMIT 1" --output - --interval 2s
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | jq -e '.accountnumber' >/dev/null 2>&1; then
  pass "run --output - → stdout is raw JSONL (no metadata mixed in)"
else
  fail "run --output - (rc=$RUN_RC) → stdout '$(printf '%s' "$RUN_OUT" | head -1)' err: $(printf '%s' "$RUN_ERR" | head -1)"
fi

# ─────────────────────────────────────────
header "Step 5: list (live)"
# ─────────────────────────────────────────
echo "  Testing: data-query list --json (data is an array)"
run_retry 3 $ZR data-query list --page-size 5 --json
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | jq -e '.data | type == "array"' >/dev/null 2>&1; then
  pass "list → .data is an array ($(printf '%s' "$RUN_OUT" | jq -r '.data | length') jobs)"
else
  fail "list (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: data-query list --status completed (queryStatus filter)"
run_retry 3 $ZR data-query list --status completed --page-size 5 --json
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | jq -e '.data | type == "array"' >/dev/null 2>&1; then
  pass "list --status completed → accepted by the API"
else
  fail "list --status (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 6: cancel (live)"
# ─────────────────────────────────────────
echo "  Testing: submit a fresh job then cancel it"
CANCEL_ID=$($ZR data-query submit "$DQ_SQL" --json 2>/dev/null | jq -r '.data.id // empty')
if [ -n "$CANCEL_ID" ]; then
  run $ZR data-query cancel "$CANCEL_ID" --confirm
  # The job may finish before the cancel lands; either a cancelled status or a
  # clean 2xx (RenderDeleteResult) is an acceptable terminal outcome.
  if [ "$RUN_RC" -eq 0 ] && printf '%s%s' "$RUN_OUT" "$RUN_ERR" | grep -qiE "cancelled|completed"; then
    pass "cancel → DELETE accepted (status: $(printf '%s%s' "$RUN_OUT" "$RUN_ERR" | grep -oiE 'cancelled|completed' | head -1))"
  else
    fail "cancel (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
  fi
else
  skip "cancel → could not submit a job to cancel"
fi

# ─────────────────────────────────────────
header "Step 7: failed job surfaces its error (live)"
# ─────────────────────────────────────────
echo "  Testing: run with an invalid table → non-zero exit + error surfaced"
run $ZR data-query run "SELECT * FROM no_such_table_e2e_xyz" --interval 2s
if [ "$RUN_RC" -ne 0 ] && printf '%s%s' "$RUN_ERR" "$RUN_OUT" | grep -qiE "failed|does not exist|Zuora API error"; then
  pass "failed job → non-zero exit, error surfaced: $(printf '%s' "${RUN_ERR:-$RUN_OUT}" | grep -oiE 'failed:.*' | head -1)"
else
  fail "failed job → expected rc!=0 + error, got rc=$RUN_RC: $(printf '%s' "${RUN_ERR:-$RUN_OUT}" | head -1)"
fi

# ─────────────────────────────────────────
header "Step 8: read-only opt-in (the core feature)"
# ─────────────────────────────────────────
# 8a: blocked by default, WITH the Data Query hint.
echo "  Testing: ZR_READ_ONLY=1 blocks submit (default), with the opt-in hint"
RO_OUT=$(ZR_READ_ONLY=1 $ZR data-query submit "$DQ_SQL" 2>&1); RO_RC=$?
if [ "$RO_RC" -eq 5 ] && printf '%s' "$RO_OUT" | grep -qF "not allowed in read-only mode" \
   && printf '%s' "$RO_OUT" | grep -qF "read-only-allow-data-query"; then
  pass "read-only default → submit blocked (exit 5) with --read-only-allow-data-query hint"
else
  fail "read-only default → rc=$RO_RC: $(printf '%s' "$RO_OUT" | head -1)"
fi

# 8b: allowed with the env opt-in.
echo "  Testing: ZR_READ_ONLY=1 + ZR_READ_ONLY_ALLOW_DATA_QUERY=1 allows submit"
run env ZR_READ_ONLY=1 ZR_READ_ONLY_ALLOW_DATA_QUERY=1 $ZR data-query submit "$DQ_SQL" --json
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | jq -e '.data.id' >/dev/null 2>&1; then
  pass "read-only + env opt-in → submit allowed"
else
  fail "read-only + env opt-in → rc=$RUN_RC: ${RUN_ERR:-$RUN_OUT}"
fi

# 8c: allowed with the FLAG opt-in (flag form, alongside --read-only).
echo "  Testing: --read-only --read-only-allow-data-query allows submit"
run $ZR --read-only --read-only-allow-data-query data-query submit "$DQ_SQL" --json
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | jq -e '.data.id' >/dev/null 2>&1; then
  pass "read-only + flag opt-in → submit allowed"
else
  fail "read-only + flag opt-in → rc=$RUN_RC: ${RUN_ERR:-$RUN_OUT}"
fi

# 8d: the toggle must NOT widen ordinary writes (regression guard).
echo "  Testing: opt-in does NOT allow a normal write (account create stays blocked)"
RO_ACCT=$(ZR_READ_ONLY=1 ZR_READ_ONLY_ALLOW_DATA_QUERY=1 $ZR account create --body '{"name":"e2e-dq"}' 2>&1); RO_ACCT_RC=$?
if [ "$RO_ACCT_RC" -ne 0 ] && printf '%s' "$RO_ACCT" | grep -qF "not allowed in read-only mode" \
   && ! printf '%s' "$RO_ACCT" | grep -qF "read-only-allow-data-query"; then
  pass "opt-in → account create still blocked (no DQ hint; widens only Data Query)"
else
  fail "opt-in widened a non-Data-Query write? rc=$RO_ACCT_RC: $(printf '%s' "$RO_ACCT" | head -1)"
fi

# 8e: env fail-safe is conservative — an unrecognized value must NOT enable it.
echo "  Testing: ZR_READ_ONLY_ALLOW_DATA_QUERY=maybe (unrecognized) does NOT enable"
RO_MAYBE=$(ZR_READ_ONLY=1 ZR_READ_ONLY_ALLOW_DATA_QUERY=maybe $ZR data-query submit "$DQ_SQL" 2>&1); RO_MAYBE_RC=$?
if [ "$RO_MAYBE_RC" -eq 5 ] && printf '%s' "$RO_MAYBE" | grep -qF "not allowed in read-only mode"; then
  pass "opt-in env unrecognized value → fails safe (stays blocked)"
else
  fail "opt-in env 'maybe' → expected blocked, rc=$RO_MAYBE_RC: $(printf '%s' "$RO_MAYBE" | head -1)"
fi

# 8f: GET (list) is always allowed under read-only.
echo "  Testing: ZR_READ_ONLY=1 allows list (GET)"
run env ZR_READ_ONLY=1 $ZR data-query list --page-size 3 --json
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | jq -e '.data | type == "array"' >/dev/null 2>&1; then
  pass "read-only → list (GET) allowed"
else
  fail "read-only → list blocked? rc=$RUN_RC: ${RUN_ERR:-$RUN_OUT}"
fi

# 8g: cancel (DELETE) blocked by default under read-only, with the hint (gate
# fires before the request, so a fake id is fine for the BLOCK assertion).
echo "  Testing: ZR_READ_ONLY=1 blocks cancel (DELETE) with the opt-in hint"
RO_CANCEL=$(ZR_READ_ONLY=1 $ZR data-query cancel FAKE-JOB-ID --confirm 2>&1); RO_CANCEL_RC=$?
if [ "$RO_CANCEL_RC" -eq 5 ] && printf '%s' "$RO_CANCEL" | grep -qF "read-only-allow-data-query"; then
  pass "read-only default → cancel (DELETE) blocked with hint"
else
  fail "read-only default → cancel rc=$RO_CANCEL_RC: $(printf '%s' "$RO_CANCEL" | head -1)"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
print_summary
