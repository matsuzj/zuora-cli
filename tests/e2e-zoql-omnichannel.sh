#!/bin/bash
# E2E Test: ZOQL Query + Omnichannel + Changelog + api + Read-only mode
# テナント: apac-sandbox (Account テーブルにシードデータあり)

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-zoql-omnichannel-${TIMESTAMP}.log"

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

# ─────────────────────────────────────────
header "Step 1: ZOQL Query Validation"
# ─────────────────────────────────────────
echo "  Testing: query without argument"
expect_fail "query validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR query

# ─────────────────────────────────────────
header "Step 2: ZOQL Query Execution"
# ─────────────────────────────────────────
# Account is seeded on this tenant, so a successful query MUST return >=1 record.
echo "  Testing: query 'SELECT ... FROM Account'"
run_retry 3 $ZR query "SELECT Id, Name, AccountNumber FROM Account WHERE Status = 'Active'" --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.records' >/dev/null 2>&1; then
  Q_COUNT=$(echo "$RUN_OUT" | jq -r '.records | length')
  if [ "${Q_COUNT:-0}" -ge 1 ]; then
    pass "query → returned $Q_COUNT records"
  else
    fail "query → 0 records (expected >=1 from seeded Account)"
  fi
elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qiE "HTTP 429|rate limit|HTTP 503"; then
  skip "query → transient: $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
else
  fail "query (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: query with --jq '.records | length'"
run_retry 3 $ZR query "SELECT Id FROM Account" --jq '.records | length'
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | grep -qE '^[0-9]+$'; then
  pass "query --jq → numeric count: $RUN_OUT"
else
  fail "query --jq → non-numeric '$RUN_OUT' (rc=$RUN_RC) ${RUN_ERR}"
fi

echo "  Testing: query --csv"
run_retry 3 $ZR query "SELECT Id, Name FROM Account" --csv
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | head -1 | grep -qiE 'id|name'; then
  pass "query CSV → has header row"
else
  fail "query CSV → no header (rc=$RUN_RC) ${RUN_ERR}"
fi

# ─────────────────────────────────────────
header "Step 3: Subscription Changelog Validation"
# ─────────────────────────────────────────
echo "  Testing: subscription changelog without argument"
expect_fail "changelog validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR subscription changelog

echo "  Testing: subscription changelog-by-order without argument"
expect_fail "changelog-by-order validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR subscription changelog-by-order

echo "  Testing: subscription changelog-version without arguments"
expect_fail "changelog-version validation → requires arguments" \
  "accepts 2 arg(s), received 0" -- $ZR subscription changelog-version

# ─────────────────────────────────────────
header "Step 4: Omnichannel Validation"
# ─────────────────────────────────────────
echo "  Testing: omnichannel get without argument"
expect_fail "omnichannel get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR omnichannel get

echo "  Testing: omnichannel create without --body"
expect_fail "omnichannel create validation → requires --body" "--body is required" -- $ZR omnichannel create

echo "  Testing: omnichannel delete without argument"
expect_fail "omnichannel delete validation → requires argument" \
  "accepts 1 arg(s), received 0" -- $ZR omnichannel delete

echo "  Testing: omnichannel delete without --confirm"
expect_fail "omnichannel delete validation → requires --confirm" \
  "this action is irreversible. Use --confirm to proceed" -- $ZR omnichannel delete OC-FAKE

# ─────────────────────────────────────────
header "Step 5: api passthrough"
# ─────────────────────────────────────────
# api takes a single <path> arg; the method defaults to GET (override with -X).
echo "  Testing: api without arguments"
expect_fail "api validation → requires path" "accepts 1 arg(s), received 0" -- $ZR api

echo "  Testing: api /v1/catalog/products (live GET)"
run_retry 3 $ZR api /v1/catalog/products --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.' >/dev/null 2>&1; then
  pass "api GET → returned JSON"
else
  fail "api GET (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: api with --jq scalar"
run_retry 3 $ZR api /v1/catalog/products --jq '.products | type'
if [ "$RUN_RC" -eq 0 ] && printf '%s' "$RUN_OUT" | grep -qF "array"; then
  pass "api GET --jq → .products is array"
else
  fail "api GET --jq → got '$RUN_OUT' (rc=$RUN_RC) ${RUN_ERR}"
fi

echo "  Testing: api on missing resource (error passthrough)"
run $ZR api /v1/accounts/NOPE-DOES-NOT-EXIST
if [ "$RUN_RC" -ne 0 ] && echo "${RUN_ERR:-$RUN_OUT}" | grep -qF "Zuora API error"; then
  pass "api GET 404 → surfaced Zuora API error with non-zero exit"
else
  fail "api GET 404 → expected Zuora API error + rc!=0, got rc=$RUN_RC: ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: api --paginate (auto-follow pages)"
run_retry 3 $ZR api /v1/catalog/products --paginate --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.' >/dev/null 2>&1; then
  pass "api --paginate → returned aggregated JSON"
else
  fail "api --paginate (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

# ─────────────────────────────────────────
header "Step 5.5: global --env flag"
# ─────────────────────────────────────────
# --env overrides the active environment for a single invocation.
echo "  Testing: query --env apac-sandbox (explicit env)"
run_retry 3 $ZR query "SELECT Id FROM Account" --env apac-sandbox --json
if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e '.records' >/dev/null 2>&1; then
  pass "query --env apac-sandbox → returned records"
else
  fail "query --env apac-sandbox (rc=$RUN_RC) → ${RUN_ERR:-$RUN_OUT}"
fi

echo "  Testing: --env unknown-environment (error path)"
expect_fail "--env unknown → rejected" "unknown environment" -- $ZR query "SELECT Id FROM Account" --env no-such-env

# ─────────────────────────────────────────
header "Step 6: Read-only mode"
# ─────────────────────────────────────────
# 6a: ZR_READ_ONLY=1 blocks a write
echo "  Testing: ZR_READ_ONLY=1 blocks account create"
RO_OUT=$(ZR_READ_ONLY=1 $ZR account create --body '{"name":"test"}' 2>&1) || true
if echo "$RO_OUT" | grep -qF "not allowed in read-only mode"; then
  pass "read-only mode → blocks account create"
else
  fail "read-only mode → did not block: $(echo "$RO_OUT" | head -1)"
fi

# 6b: an UNRECOGNIZED value must fail closed (default-true branch)
echo "  Testing: ZR_READ_ONLY=maybe (unrecognized) fails closed"
RO_OUT2=$(ZR_READ_ONLY=maybe $ZR order create --body '{}' 2>&1) || true
if echo "$RO_OUT2" | grep -qF "not allowed in read-only mode"; then
  pass "read-only mode (unrecognized value) → fails closed, blocks write"
else
  fail "read-only mode (maybe) → did not block: $(echo "$RO_OUT2" | head -1)"
fi

# 6c: a recognized falsy value must NOT block; the write proceeds to a real
# API/validation error instead of the read-only guard.
echo "  Testing: ZR_READ_ONLY=0 does not block (off path)"
RO_OFF=$(ZR_READ_ONLY=0 $ZR account create --body '{"name":"test"}' 2>&1) || true
if echo "$RO_OFF" | grep -qF "not allowed in read-only mode"; then
  fail "read-only mode (0) → wrongly blocked the write"
else
  pass "read-only mode (0) → write not blocked (proceeded to API: $(echo "$RO_OFF" | head -1))"
fi

# 6d: read commands are allowed under read-only — assert a real result shape,
# not mere valid JSON (an error rendered as JSON would also be valid JSON).
echo "  Testing: ZR_READ_ONLY=1 allows query (read)"
RO_READ=$(ZR_READ_ONLY=1 $ZR query "SELECT Id FROM Account" --json 2>&1) || true
if echo "$RO_READ" | jq -e '.records' >/dev/null 2>&1 && ! echo "$RO_READ" | grep -qi "read-only"; then
  pass "read-only mode → allows query (has .records, not blocked)"
else
  fail "read-only mode → query not allowed as expected: $(echo "$RO_READ" | head -1)"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
print_summary
