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

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

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
print_summary
