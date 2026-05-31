#!/bin/bash
# E2E Test: Local commands (config / alias / auth token / version / completion)
#
# These are deterministic, offline commands with NO tenant side effects. The
# whole suite runs against an ISOLATED config dir via XDG_CONFIG_HOME so it can
# never touch the user's real ~/.config/zr or the shared keyring. It does NOT
# exercise auth login/logout (those would clobber the shared session).

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-local-${TIMESTAMP}.log"

exec > >(tee >(sed 's/\x1b\[[0-9;]*m//g' > "$LOG_FILE")) 2>&1

green()  { printf "\033[32m%s\033[0m\n" "$1"; }
red()    { printf "\033[31m%s\033[0m\n" "$1"; }
yellow() { printf "\033[33m%s\033[0m\n" "$1"; }

pass() { PASS=$((PASS+1)); green "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); red   "  ✗ $1"; }
skip() { SKIP=$((SKIP+1)); yellow "  ⊘ $1 (skipped)"; }

header() { printf "\n\033[1m=== %s ===\033[0m\n" "$1"; }

# Isolated config dir for the whole run; cleaned up on exit.
ISO_DIR=$(mktemp -d)
export XDG_CONFIG_HOME="$ISO_DIR"
cleanup() { rm -rf "$ISO_DIR"; }
trap cleanup EXIT

# zr <args...> — run the CLI against the isolated config dir.
zr() { "$ZR" "$@"; }

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
# Passes when the command exits non-zero AND output contains the expected string.
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
header "Step 0: binary + isolation check"
# ─────────────────────────────────────────
[ -x "$ZR" ] || { red "zr binary not found/executable at $ZR (build it first)"; exit 1; }
# Sanity: confirm we are NOT pointed at the real config dir.
case "$ISO_DIR" in
  "$HOME"/.config*) red "isolation dir resolved into real config — aborting"; exit 1 ;;
esac
pass "isolated config dir: $ISO_DIR"

# ─────────────────────────────────────────
header "Step 1: version"
# ─────────────────────────────────────────
echo "  Testing: version"
expect_ok "version → prints version string" "zr version" -- zr version

# ─────────────────────────────────────────
header "Step 2: config list / get / set round-trip"
# ─────────────────────────────────────────
echo "  Testing: config list"
expect_ok "config list → shows active_environment" "active_environment" -- zr config list

echo "  Testing: config get without arg"
expect_fail "config get validation → requires arg" "accepts 1 arg(s), received 0" -- zr config get

echo "  Testing: config get unknown key"
expect_fail "config get → rejects unknown key" "unknown config key" -- zr config get no-such-key

echo "  Testing: config get active_environment"
expect_ok "config get → returns a value" "" -- zr config get active_environment

echo "  Testing: config set without args"
expect_fail "config set validation → requires 2 args" "accepts 2 arg(s), received 0" -- zr config set

echo "  Testing: config set with 1 arg"
expect_fail "config set validation → requires 2 args (got 1)" "accepts 2 arg(s), received 1" -- zr config set foo

echo "  Testing: config set unknown key"
expect_fail "config set → rejects unknown key" "unknown config key" -- zr config set bogus_key value

echo "  Testing: config set zuora_version (round-trip)"
expect_ok "config set → accepts zuora_version" "Set zuora_version to 2099-01-01" -- zr config set zuora_version 2099-01-01
GOT=$(zr config get zuora_version 2>&1)
if [ "$GOT" = "2099-01-01" ]; then
  pass "config get → reads back the set value (2099-01-01)"
else
  fail "config get → expected 2099-01-01, got '$GOT'"
fi

echo "  Testing: config set default_output (round-trip)"
expect_ok "config set → accepts default_output" "Set default_output to json" -- zr config set default_output json
GOT2=$(zr config get default_output 2>&1)
if [ "$GOT2" = "json" ]; then
  pass "config get → reads back default_output=json"
else
  fail "config get → expected json, got '$GOT2'"
fi

# ─────────────────────────────────────────
header "Step 3: config env"
# ─────────────────────────────────────────
echo "  Testing: config env without arg"
expect_fail "config env validation → requires arg" "accepts 1 arg(s), received 0" -- zr config env

echo "  Testing: config env unknown environment"
expect_fail "config env → rejects unknown environment" "unknown environment" -- zr config env no-such-env

echo "  Testing: config env apac-sandbox (switch)"
expect_ok "config env → switches active environment" "apac-sandbox" -- zr config env apac-sandbox
GOT3=$(zr config get active_environment 2>&1)
if [ "$GOT3" = "apac-sandbox" ]; then
  pass "config env → active_environment now apac-sandbox"
else
  fail "config env → expected apac-sandbox active, got '$GOT3'"
fi

# ─────────────────────────────────────────
header "Step 4: alias set / list / delete round-trip"
# ─────────────────────────────────────────
echo "  Testing: alias list (empty)"
expect_ok "alias list → reports none configured" "No aliases configured" -- zr alias list

echo "  Testing: alias set without args"
expect_fail "alias set validation → requires 2 args" "accepts 2 arg(s), received 0" -- zr alias set

echo "  Testing: alias set with 1 arg"
expect_fail "alias set validation → requires 2 args (got 1)" "accepts 2 arg(s), received 1" -- zr alias set onlyname

echo "  Testing: alias set la 'account list'"
expect_ok "alias set → creates alias" 'Alias "la" set to' -- zr alias set la "account list"

echo "  Testing: alias list (after set)"
expect_ok "alias list → contains the new alias" "la" -- zr alias list

echo "  Testing: alias delete without arg"
expect_fail "alias delete validation → requires arg" "accepts 1 arg(s), received 0" -- zr alias delete

echo "  Testing: alias delete nonexistent"
expect_fail "alias delete → rejects unknown alias" "not found" -- zr alias delete nope

echo "  Testing: alias delete la"
expect_ok "alias delete → removes alias" "deleted" -- zr alias delete la
if zr alias list 2>&1 | grep -qF "la	"; then
  fail "alias delete → 'la' still present after delete"
else
  pass "alias delete → verified removed"
fi

# ─────────────────────────────────────────
header "Step 5: completion"
# ─────────────────────────────────────────
echo "  Testing: completion bash"
expect_ok "completion bash → emits a completion script" "# bash completion" -- zr completion bash

echo "  Testing: completion zsh"
expect_ok "completion zsh → emits a completion script" "compdef" -- zr completion zsh

# ─────────────────────────────────────────
header "Step 6: auth token / status (read-only, no login/logout)"
# ─────────────────────────────────────────
# In this isolated config dir there is no stored session. The commands must run
# without crashing; whether a token is available depends on ambient credentials,
# so accept either a token-shaped output or a clean "not authenticated" error.
echo "  Testing: auth status (isolated)"
AS_OUT=$(zr auth status 2>&1); AS_RC=$?
if echo "$AS_OUT" | grep -qE "Environment:|not authenticated|No (active )?environment"; then
  pass "auth status → produced a coherent status (rc=$AS_RC)"
else
  fail "auth status → unexpected: $(echo "$AS_OUT" | head -1)"
fi

echo "  Testing: auth token (isolated)"
# SECURITY: `auth token` prints the bearer token on stdout. Never store that
# value in a variable (it could later be echoed into the log). Capture only its
# byte length + exit code; on the error path the message (stderr) is not secret.
AT_LEN=$(zr auth token 2>/dev/null | wc -c | tr -d ' '); AT_RC=${PIPESTATUS[0]}
if [ "$AT_RC" -eq 0 ] && [ "${AT_LEN:-0}" -gt 1 ]; then
  pass "auth token → returned a non-empty token (${AT_LEN} bytes, value not logged)"
else
  AT_ERR=$(zr auth token 2>&1 >/dev/null)   # stderr only; stdout (token) discarded
  if [ "$AT_RC" -ne 0 ] && printf '%s' "$AT_ERR" | grep -qiE "auth|login|credential|environment|token"; then
    pass "auth token → clean error without credentials (rc=$AT_RC)"
  else
    fail "auth token → unexpected: rc=$AT_RC (token value intentionally not shown)"
  fi
fi

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
