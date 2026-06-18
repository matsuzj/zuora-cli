#!/bin/bash
# Run all E2E suites against the currently-authenticated Zuora tenant and print
# an aggregated pass/fail summary. Requires: `zr auth login` already done, and
# the binary built at ./bin/zr (run `task build` or `make build` first).
#
# Usage:
#   ./tests/run-all.sh                # run every suite
#   ./tests/run-all.sh order usage-meter   # run only the named suites
#
# Tenant safety: the live suites refuse to run unless the active environment is a
# sandbox (require_auth in tests/lib/e2e-common.sh fails closed). To run write
# suites against a non-sandbox tenant on purpose, set ZR_E2E_ALLOW_PROD=1.
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

# Suites with no tenant writes may run concurrently: e2e-local is fully
# isolated (own XDG_CONFIG_HOME, offline), e2e-usage-meter is pure cobra
# validation, e2e-commerce adds only read-only live calls. Their outputs are
# buffered and replayed in order so logs never interleave. Everything else
# creates tenant state and stays strictly serial. The split applies only to
# full runs — explicit suite arguments keep today's serial behavior.
PARALLEL_SAFE="e2e-local.sh e2e-commerce.sh e2e-usage-meter.sh"

is_parallel_safe() {
  case " $PARALLEL_SAFE " in *" $1 "*) return 0;; esac
  return 1
}

if [ $# -eq 0 ]; then
  declare -a PAR_SUITES=()
  declare -a SER_SUITES=()
  for suite in "${SUITES[@]}"; do
    if is_parallel_safe "$(basename "$suite")"; then
      PAR_SUITES+=("$suite")
    else
      SER_SUITES+=("$suite")
    fi
  done

  if [ "${#PAR_SUITES[@]}" -gt 0 ]; then
    PARLOG_DIR=$(mktemp -d)
    i=0
    for suite in "${PAR_SUITES[@]}"; do
      (
        if bash "$suite" > "$PARLOG_DIR/$i.log" 2>&1; then
          : > "$PARLOG_DIR/$i.ok"
        fi
      ) &
      i=$((i + 1))
    done
    wait
    i=0
    for suite in "${PAR_SUITES[@]}"; do
      name="$(basename "$suite")"
      bold "▶ $name (parallel)"
      cat "$PARLOG_DIR/$i.log"
      if [ -f "$PARLOG_DIR/$i.ok" ]; then
        OK_SUITES+=("$name")
      else
        FAILED_SUITES+=("$name")
      fi
      echo
      i=$((i + 1))
    done
    rm -rf "$PARLOG_DIR"
  fi

  SUITES=()
  if [ "${#SER_SUITES[@]}" -gt 0 ]; then
    SUITES=("${SER_SUITES[@]}")
  fi
fi

for suite in "${SUITES[@]}"; do
  name="$(basename "$suite")"
  if [ ! -x "$suite" ]; then
    # A requested-but-missing suite must FAIL the run, not skip silently —
    # otherwise a typoed name produces a green E2E run that tested nothing.
    red "  ✗ $name (not found/executable)"
    FAILED_SUITES+=("$name (not found)")
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
# Guard the expansion: bash 3.2 (macOS /bin/bash) treats an empty array as
# unbound under `set -u` and would crash the roll-up.
if [ "${#OK_SUITES[@]}" -gt 0 ]; then
  for s in "${OK_SUITES[@]}"; do green "    ✓ $s"; done
fi
if [ "${#FAILED_SUITES[@]}" -gt 0 ]; then
  red "  Failed suites: ${#FAILED_SUITES[@]}"
  for s in "${FAILED_SUITES[@]}"; do red "    ✗ $s"; done
  exit 1
fi
green "  ALL E2E SUITES PASSED"
