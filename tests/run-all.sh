#!/bin/bash
# Run all E2E suites against the currently-authenticated Zuora tenant and print
# an aggregated pass/fail summary. Requires: `zr auth login` already done, and
# the binary built at ./bin/zr (run `task build` or `make build` first).
#
# Usage:
#   ./tests/run-all.sh                       # run every suite
#   ./tests/run-all.sh order usage-meter     # run only the named suites
#   ZR_ENV=apac-sandbox ./tests/run-all.sh   # pin the environment for this run
#
# Environment selection: the suites use whichever environment `zr` resolves —
# ZR_ENV (exported, inherited by every suite) wins over the persisted
# active_environment. When ZR_ENV is set, require_auth also asserts the active
# environment matches it, so a typo can't route writes at another tenant.
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
RUN_START=$(date +%s)
PARTIAL=false
[ "$#" -gt 0 ] && PARTIAL=true

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

# Machine receipt (#527): every run writes tests/logs/summary-<sha>-<ts>.json
# as a side effect, so an "11/11 green" claim in a PR/release is one file away
# from its machine source. make release-check asserts a clean, non-partial
# receipt exists for the exact HEAD being tagged. HONEST-MISTAKE GUARD ONLY:
# the receipt shares the agent's trust root (any writer can forge a file) —
# it catches partial runs and wrong commits, not fabrication.
# json_str_array <elems...> — emits "a","b" (no trailing comma); nothing for
# an empty list, so an empty bash array renders as [] and not [""].
json_str_array() {
  [ "$#" -eq 0 ] && return 0
  printf '"%s",' "$@" | sed 's/,$//'
}

write_receipt() {
  local repo sha dirty ts dur envname burl auth_out receipt
  repo="$SCRIPT_DIR/.."
  sha=$(git -C "$repo" rev-parse HEAD 2>/dev/null || echo unknown)
  dirty=false
  [ -n "$(git -C "$repo" status --porcelain 2>/dev/null)" ] && dirty=true
  ts=$(date -u '+%Y-%m-%dT%H:%M:%SZ')
  dur=$(( $(date +%s) - RUN_START ))
  auth_out=$("$ZR" auth status 2>&1 || true)
  envname=$(echo "$auth_out" | awk '/^Environment:/ {print $2}')
  burl=$(echo "$auth_out" | awk '/^Base URL:/ {print $NF}')
  receipt="$SCRIPT_DIR/logs/summary-$sha-$(date +%Y%m%d-%H%M%S).json"
  mkdir -p "$SCRIPT_DIR/logs"
  {
    printf '{\n'
    printf '  "git_sha": "%s",\n' "$sha"
    printf '  "dirty": %s,\n' "$dirty"
    printf '  "partial": %s,\n' "$PARTIAL"
    printf '  "timestamp": "%s",\n' "$ts"
    printf '  "duration_seconds": %s,\n' "$dur"
    printf '  "environment": "%s",\n' "$envname"
    printf '  "base_url": "%s",\n' "$burl"
    printf '  "passed": [%s],\n' "$(json_str_array ${OK_SUITES[@]+"${OK_SUITES[@]}"})"
    printf '  "failed": [%s]\n' "$(json_str_array ${FAILED_SUITES[@]+"${FAILED_SUITES[@]}"})"
    printf '}\n'
  } > "$receipt"
  echo "  Receipt: $receipt"
}

bold "════════ E2E roll-up ════════"
green "  Passed suites: ${#OK_SUITES[@]}"
# Guard the expansion: bash 3.2 (macOS /bin/bash) treats an empty array as
# unbound under `set -u` and would crash the roll-up.
if [ "${#OK_SUITES[@]}" -gt 0 ]; then
  for s in "${OK_SUITES[@]}"; do green "    ✓ $s"; done
fi
write_receipt
if [ "${#FAILED_SUITES[@]}" -gt 0 ]; then
  red "  Failed suites: ${#FAILED_SUITES[@]}"
  for s in "${FAILED_SUITES[@]}"; do red "    ✗ $s"; done
  exit 1
fi
green "  ALL E2E SUITES PASSED"
