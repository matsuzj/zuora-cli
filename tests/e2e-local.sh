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

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# Isolated config dir for the whole run; cleaned up on exit. This trap REPLACES
# the lib's drain trap, so chain _drain_log explicitly (see lib contract).
ISO_DIR=$(mktemp -d)
export XDG_CONFIG_HOME="$ISO_DIR"
cleanup() { rm -rf "$ISO_DIR"; }
trap 'cleanup; _drain_log' EXIT

# zr <args...> — run the CLI against the isolated config dir.
zr() { "$ZR" "$@"; }

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

# default_output=json must actually SWITCH the output format for piped
# (non-TTY) invocations carrying no explicit format flag — the P4-3 behavior
# change (#214). Command substitution is non-TTY by nature, so the offline
# `version` command observes it directly.
echo "  Testing: default_output=json switches piped output format"
OUT_JSON=$(zr version 2>&1)
if printf '%s' "$OUT_JSON" | jq -e '.version' >/dev/null 2>&1; then
  pass "default_output=json → piped 'zr version' emits JSON"
else
  fail "default_output=json → expected JSON from 'zr version', got: $(printf '%s' "$OUT_JSON" | head -1)"
fi

# Restore the default: default_output is WIRED now (P4-3) — leaving it on
# json would flip every later piped check (and the following suites) to JSON
# output, which is exactly what broke the alias-execution check when the
# wiring first landed.
expect_ok "config set → restores default_output" "Set default_output to table" -- zr config set default_output table
OUT_TABLE=$(zr version 2>&1)
case "$OUT_TABLE" in
  "zr version"*) pass "default_output=table → 'zr version' back to human text" ;;
  *) fail "default_output restore → expected 'zr version ...', got: $(printf '%s' "$OUT_TABLE" | head -1)" ;;
esac

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
header "Step 4b: alias EXECUTION (expansion through main)"
# ─────────────────────────────────────────
# Until P1-3 no E2E actually ran an alias through the expansion path.
echo "  Testing: set + run a simple alias"
zr alias set v version >/dev/null 2>&1
expect_ok "alias execution → 'zr v' runs 'zr version'" "zr version" -- zr v

echo "  Testing: quoted multi-word expansion survives shlex"
zr alias set vq 'config get "active_environment"' >/dev/null 2>&1
expect_ok "alias execution → quoted expansion works" "" -- zr vq

echo "  Testing: alias set rejects a built-in name"
expect_fail "alias set → rejects built-in name" "built-in command" -- zr alias set account "contact list"

echo "  Testing: alias set rejects a self-reference"
expect_fail "alias set → rejects self-reference" "would invoke itself" -- zr alias set myloop "myloop --json"

echo "  Testing: alias set rejects a malformed expansion"
expect_fail "alias set → rejects unbalanced quotes" "malformed expansion" -- zr alias set bad 'query "SELECT unbalanced'

zr alias delete v >/dev/null 2>&1
zr alias delete vq >/dev/null 2>&1

# ─────────────────────────────────────────
header "Step 5: completion"
# ─────────────────────────────────────────
echo "  Testing: completion bash"
expect_ok "completion bash → emits a completion script" "# bash completion" -- zr completion bash

echo "  Testing: completion zsh"
expect_ok "completion zsh → emits a completion script" "compdef" -- zr completion zsh

# Dynamic completions (P5-3b): __complete is cobra's hidden completion
# entry point; these run fully offline.
echo "  Testing: dynamic completions (__complete)"
expect_ok "config get <TAB> → offers config keys" "default_output" -- zr __complete config get ""
expect_ok "subscription cancel --policy <TAB> → offers policies" "EndOfCurrentTerm" -- zr __complete subscription cancel A-1 --policy ""
expect_ok "--env <TAB> → offers environment names" "apac-sandbox" -- zr __complete --env ""

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
header "Step 7: output flag contract (v0.4.0 behavior changes, offline)"
# ─────────────────────────────────────────
# All three checks fail BEFORE any request is built, so they are exercisable
# in the isolated, unauthenticated config dir.

# Bare --csv on a JSON-only command is an explicit error since #197 (it was
# silently ignored before). charge get rejects it pre-request.
echo "  Testing: --csv on a JSON-only command"
expect_fail "--csv on JSON-only command → explicit error" \
  "--csv is not supported for JSON-only output" -- zr charge get --key FAKE --csv

# List commands reject stray positional arguments (cobra.NoArgs) since the
# P3-2 listcmd migration; previously extra args were silently ignored.
echo "  Testing: stray positional arg on a list command"
expect_fail "list rejects stray positional arg (NoArgs)" \
  'unknown command "stray-arg" for "zr invoice list"' -- zr invoice list stray-arg

# --json + --template is the one documented invalid flag pair (README: --jq
# combinations are valid precedence, cf. the PR #54 regression).
echo "  Testing: --json + --template rejection"
expect_fail "--json + --template → rejected" \
  "cannot use --json and --template together" -- zr version --json --template '{{.version}}'

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
print_summary
