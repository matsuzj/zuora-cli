#!/bin/bash
# tests/lib/e2e-common.sh — shared helpers for all E2E suites.
#
# Source contract — every suite does, in this order:
#
#   set -uo pipefail
#   SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
#   ZR="$SCRIPT_DIR/../bin/zr"
#   PASS=0; FAIL=0; SKIP=0
#   TIMESTAMP=$(date +%Y%m%d-%H%M%S)
#   LOG_DIR="$SCRIPT_DIR/logs"; mkdir -p "$LOG_DIR"
#   LOG_FILE="$LOG_DIR/e2e-<name>-${TIMESTAMP}.log"
#   source "$SCRIPT_DIR/lib/e2e-common.sh"
#   setup_log        # bare call — never from a subshell
#   ...suite body...
#   print_summary    # last line; exits 1 when FAIL > 0
#
# The counters (PASS/FAIL/SKIP) and LOG_DIR/LOG_FILE MUST be set before
# sourcing/setup_log — the helpers reference them under `set -u`.
#
# A suite that needs its own EXIT trap (e.g. e2e-local.sh's cleanup) must
# CHAIN the log drain explicitly AFTER setup_log:
#   trap 'cleanup; _drain_log' EXIT
# (a bare `trap cleanup EXIT` would silently replace the drain trap and
# truncate the log tail.)

green()  { printf "\033[32m%s\033[0m\n" "$1"; }
red()    { printf "\033[31m%s\033[0m\n" "$1"; }
yellow() { printf "\033[33m%s\033[0m\n" "$1"; }

pass() { PASS=$((PASS+1)); green "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); red   "  ✗ $1"; }
skip() { SKIP=$((SKIP+1)); yellow "  ⊘ $1 (skipped)"; }

header() { printf "\n\033[1m=== %s ===\033[0m\n" "$1"; }

# Shared rate plan for order/signup bodies (Backlog starter plan, monthly).
# Override with ZR_E2E_RATE_PLAN_ID when the catalog changes; a stale ID makes
# the first order create fail, which is distinguishable from a CLI regression.
# Always defined (set -u): suites that never build an order body just ignore it.
RATE_PLAN_ID="${ZR_E2E_RATE_PLAN_ID:-4c6059a8d8899f453ffa0637451d0003}"

# Drain the tee/sed log pipeline on exit (sed block-buffers to a file; without
# a clean EOF + wait the tail of the log is silently truncated).
_drain_log() { exec 1>&- 2>&-; wait "$LOG_TEE_PID" 2>/dev/null || true; }

# setup_log — tee all output (ANSI-stripped) into $LOG_FILE and arm the drain
# trap. Requires LOG_FILE set and LOG_DIR created before the call.
setup_log() {
  exec > >(tee >(sed 's/\x1b\[[0-9;]*m//g' > "$LOG_FILE")) 2>&1
  LOG_TEE_PID=$!
  trap _drain_log EXIT
}

# run <command...> — capture stdout→RUN_OUT (clean, for jq), stderr→RUN_ERR
# (shown only on failure), exit code→RUN_RC. Keeps JSON parsing reliable while
# making every failure diagnosable (a bare 2>/dev/null discards the reason).
RUN_OUT=""; RUN_ERR=""; RUN_RC=0
run() {
  local ef="$LOG_DIR/.run.$$.err"
  RUN_OUT=$("$@" 2>"$ef"); RUN_RC=$?
  RUN_ERR=$(cat "$ef" 2>/dev/null); rm -f "$ef"
}

# _transient_error <text> — true when an error is worth retrying: Zuora
# throttling/5xx, OR a transport-level network blip. The api client retries
# transport errors only for IDEMPOTENT methods, so a POST like `zr query` (ZOQL
# /v1/action/query) surfaces a raw "connection reset by peer" that this layer
# must absorb — otherwise a network blip reads as a CLI regression. Kept targeted
# (not "retry any error"), per the precise-skip philosophy in docs/e2e-test-skips.md.
_transient_error() {
  printf '%s' "$1" | grep -qiE \
    "HTTP 429|HTTP 5[0-9][0-9]|rate limit|connection reset|connection refused|broken pipe|i/o timeout|TLS handshake timeout|unexpected EOF"
}

# run_retry <attempts> <command...> — run(), retrying transient errors (Zuora
# 429/5xx/rate-limit + transport blips like connection reset) with a 2s pause.
run_retry() {
  local attempts="$1"; shift
  local i
  for ((i=1; i<=attempts; i++)); do
    run "$@"
    [ "$RUN_RC" -eq 0 ] && return 0
    _transient_error "$RUN_ERR" || return "$RUN_RC"
    sleep 2
  done
  return "$RUN_RC"
}

# run_retry_nonempty <attempts> <command...> — like run_retry, but ALSO
# retries when the command exits 0 with EMPTY stdout (defense-in-depth for
# read checks whose empty success output is never legitimate). Sleeps
# escalate (2,4,8,...s). NOTE: the 2026-06-12 query-CSV "flake" this was
# first written for turned out to be a pipefail+EPIPE bug in the CHECK
# pipeline (see e2e-zoql-omnichannel.sh), not an empty API response.
run_retry_nonempty() {
  local attempts="$1"; shift
  local i delay=2
  for ((i=1; i<=attempts; i++)); do
    run "$@"
    if [ "$RUN_RC" -eq 0 ] && [ -n "$RUN_OUT" ]; then
      return 0
    fi
    if [ "$RUN_RC" -ne 0 ]; then
      _transient_error "$RUN_ERR" || return "$RUN_RC"
    fi
    sleep "$delay"; delay=$((delay * 2))
  done
  return "$RUN_RC"
}

# expect_ok <description> <expected-substring> -- <command...>
# Passes when the command exits 0 AND output contains the expected fixed-string.
expect_ok() {
  local desc="$1" want="$2"; shift 2
  [ "${1:-}" = "--" ] && shift
  local out rc
  out=$("$@" 2>&1); rc=$?
  if [ "$rc" -eq 0 ] && printf '%s' "$out" | grep -qF -- "$want"; then
    pass "$desc"
  else
    fail "$desc → rc=$rc, expected '$want', got: $(printf '%s' "$out" | head -1)"
  fi
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

# read_or_skip_on <description> <jq-success-filter> <expected-error-substring> -- <command...>
# Stricter read_or_skip: skip ONLY when the error contains the expected
# fixed-string (the tenant limitation this check is known to hit); any OTHER
# Zuora API error fails. Prevents a blanket skip from masking a new failure
# mode behind a known one.
read_or_skip_on() {
  local desc="$1" filter="$2" expect="$3"; shift 3
  [ "${1:-}" = "--" ] && shift
  run "$@"
  if [ "$RUN_RC" -eq 0 ] && echo "$RUN_OUT" | jq -e "$filter" >/dev/null 2>&1; then
    pass "$desc"
  elif echo "${RUN_ERR:-$RUN_OUT}" | grep -qF -- "$expect"; then
    skip "$desc → $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
  else
    fail "$desc → rc=$RUN_RC: $(echo "${RUN_ERR:-$RUN_OUT}" | head -1)"
  fi
}

# require_auth — Step 0 gate for live suites: binary present, token valid, AND
# the active tenant is a sandbox (fail-closed, ZR_E2E_ALLOW_PROD=1 to override).
# The single implementation lives in scripts/require-sandbox.sh (#524) so
# ad-hoc live probes get the exact same gate the suites use (the #267
# rationale is documented there); this wrapper only maps its diagnostics onto
# the suite's pass/fail counters and hard-exits on failure, as before.
_E2E_LIB_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=../../scripts/require-sandbox.sh
source "$_E2E_LIB_DIR/../../scripts/require-sandbox.sh"

require_auth() {
  local out rc line
  out="$(require_sandbox_check "$ZR")"; rc=$?
  while IFS= read -r line; do
    case "$line" in
      OK:*)   pass "Auth OK (${line#OK: })" ;;
      WARN:*) yellow "  ⚠ ${line#WARN: }"; pass "Auth OK (production override)" ;;
      FAIL:*) fail "Auth gate: ${line#FAIL: }" ;;
    esac
  done <<< "$out"
  [ "$rc" -eq 0 ] || exit 1
}

# print_summary — counts + log path + RESULT line; exits 1 when FAIL > 0.
# Suites print their own "Summary" header / artifact lines before calling this.
print_summary() {
  local total=$((PASS + FAIL + SKIP))
  echo "  Passed:  $PASS / $total"
  echo "  Failed:  $FAIL / $total"
  echo "  Skipped: $SKIP / $total"
  echo ""
  echo "  Log: $LOG_FILE"
  echo ""
  if [ "$FAIL" -gt 0 ]; then
    echo "  RESULT: FAIL"
    exit 1
  fi
  echo "  RESULT: PASS"
}

# ── Deprecated-alias tripwires (#454/#455) ──────────────────────────────────
# Several suites deliberately exercised DEPRECATED alias commands so an
# accidental removal would break E2E loudly. The #512/#513 wave (v0.9.0)
# removed the -by-X commands and flag aliases; those sites now PIN that the
# old names STAY gone ("REMOVED in #512"). `usage post` remains a deliberate
# live alias. When removing an alias for real, update every site in ONE sweep:
#     grep -rn 'ALIAS-TRIPWIRE' tests/
# Each site carries that marker on the line above its check. The past rename
# waves broke E2E precisely because these checks were scattered and unmarked.
