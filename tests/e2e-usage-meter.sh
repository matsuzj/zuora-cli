#!/bin/bash
# E2E Test: Usage & Meter Commands (Sub-phase 3d)
# テナント: apac-sandbox
# 注意: usage/meter は専用のテナント設定が必要なため、バリデーション中心のテスト。
#       各ケースは「終了コードが非ゼロ」かつ「想定したエラーメッセージを含む」ことを
#       固定文字列で検証する。help にフォールバックしたり exit 0 になる退行を検出できる。

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
LOG_FILE="$LOG_DIR/e2e-usage-meter-${TIMESTAMP}.log"

# Tee all output to log file
exec > >(tee >(sed 's/\x1b\[[0-9;]*m//g' > "$LOG_FILE")) 2>&1

green()  { printf "\033[32m%s\033[0m\n" "$1"; }
red()    { printf "\033[31m%s\033[0m\n" "$1"; }
yellow() { printf "\033[33m%s\033[0m\n" "$1"; }

pass() { PASS=$((PASS+1)); green "  ✓ $1"; }
fail() { FAIL=$((FAIL+1)); red   "  ✗ $1"; }
skip() { SKIP=$((SKIP+1)); yellow "  ⊘ $1 (skipped)"; }

header() { printf "\n\033[1m=== %s ===\033[0m\n" "$1"; }

# expect_fail <description> <expected-substring> -- <command...>
# Passes only when the command exits non-zero AND its combined (stdout+stderr)
# output contains the exact expected substring (fixed-string match). This makes a
# regression that drops the validation, prints help, or exits 0 a real FAIL —
# unlike a loose 'grep -qi arg|required' which any usage banner would satisfy.
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
header "Step 1: Usage Validation"
# ─────────────────────────────────────────
echo "  Testing: usage post without --file"
expect_fail "usage post validation → requires --file" "--file is required" -- $ZR usage post

echo "  Testing: usage post with nonexistent file (local IO error, not API)"
expect_fail "usage post validation → file not found (local)" \
  "no such file or directory" -- $ZR usage post --file /nonexistent/file.csv

echo "  Testing: usage get without argument"
expect_fail "usage get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR usage get

echo "  Testing: usage create without --body"
expect_fail "usage create validation → requires --body" "--body is required" -- $ZR usage create

echo "  Testing: usage update without argument"
expect_fail "usage update validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR usage update

echo "  Testing: usage delete without argument"
expect_fail "usage delete validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR usage delete

# ─────────────────────────────────────────
header "Step 2: Meter Validation"
# ─────────────────────────────────────────
echo "  Testing: meter run without arguments"
expect_fail "meter run validation → requires 2 arguments" "accepts 2 arg(s), received 0" -- $ZR meter run

echo "  Testing: meter run with only 1 argument"
expect_fail "meter run validation → requires 2 arguments (got 1)" "accepts 2 arg(s), received 1" -- $ZR meter run ONLY-ONE

echo "  Testing: meter debug without arguments"
expect_fail "meter debug validation → requires 2 arguments" "accepts 2 arg(s), received 0" -- $ZR meter debug

echo "  Testing: meter status without arguments"
expect_fail "meter status validation → requires 2 arguments" "accepts 2 arg(s), received 0" -- $ZR meter status

echo "  Testing: meter summary without argument"
expect_fail "meter summary validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR meter summary

# With the positional present, the missing required --run-type flag is reported.
echo "  Testing: meter summary without --run-type"
expect_fail "meter summary validation → requires --run-type" \
  "--run-type is required" -- $ZR meter summary FAKE-ID

echo "  Testing: meter audit without required flags"
# meter audit validates required flags only after the positional arg is supplied,
# so the first missing flag reported is --export-type. Pin to that exact text
# rather than a loose alternation that included the substring 'to' (matches almost
# any output) — the old predicate could not detect the flag check regressing.
expect_fail "meter audit validation → requires --export-type" \
  "--export-type is required" -- $ZR meter audit FAKE-ID

echo "  Testing: meter audit without argument"
expect_fail "meter audit validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR meter audit

# ─────────────────────────────────────────
header "Step 3: Summary"
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
