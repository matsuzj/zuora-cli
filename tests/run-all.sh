#!/bin/bash
# Run all E2E suites against the currently-authenticated Zuora tenant and print
# an aggregated pass/fail summary. Requires: `zr auth login` already done, and
# the binary built at ./bin/zr (run `task build` or `make build` first).
#
# Usage:
#   ./tests/run-all.sh                # run every suite
#   ./tests/run-all.sh order usage-meter   # run only the named suites
#
# Each suite is an independent script that exits non-zero on failure; this
# runner reports per-suite RESULT and a final roll-up, and exits non-zero if any
# suite failed.

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"

green() { printf "\033[32m%s\033[0m\n" "$1"; }
red()   { printf "\033[31m%s\033[0m\n" "$1"; }
bold()  { printf "\033[1m%s\033[0m\n" "$1"; }

if [ ! -x "$ZR" ]; then
  red "Binary not found at $ZR — build it first (task build / make build)."
  exit 1
fi

# Suites to run: all e2e-*.sh by default, or the names passed as arguments.
if [ "$#" -gt 0 ]; then
  SUITES=()
  for name in "$@"; do
    SUITES+=("$SCRIPT_DIR/e2e-${name}.sh")
  done
else
  SUITES=("$SCRIPT_DIR"/e2e-*.sh)
fi

declare -a OK_SUITES=()
declare -a FAILED_SUITES=()

for suite in "${SUITES[@]}"; do
  name="$(basename "$suite")"
  if [ ! -x "$suite" ]; then
    red "  skip $name (not found/executable)"
    continue
  fi
  bold "▶ $name"
  if bash "$suite"; then
    OK_SUITES+=("$name")
  else
    FAILED_SUITES+=("$name")
  fi
  echo
done

bold "════════ E2E roll-up ════════"
green "  Passed suites: ${#OK_SUITES[@]}"
for s in "${OK_SUITES[@]}"; do green "    ✓ $s"; done
if [ "${#FAILED_SUITES[@]}" -gt 0 ]; then
  red "  Failed suites: ${#FAILED_SUITES[@]}"
  for s in "${FAILED_SUITES[@]}"; do red "    ✗ $s"; done
  exit 1
fi
green "  ALL E2E SUITES PASSED"
