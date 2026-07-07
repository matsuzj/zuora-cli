#!/bin/bash
# require-sandbox.sh [zr-binary] — fail-closed tenant-safety gate (#524).
#
# The E2E suites gained this gate in #267; ad-hoc live probes — which
# AGENTS.md and the audit-issue boilerplate mandate, including probes that
# WRITE to the tenant — had none. This is the single implementation:
# tests/lib/e2e-common.sh sources it for require_auth, and any session runs
# it standalone before probing:
#
#     scripts/require-sandbox.sh              # checks ./bin/zr's session
#     scripts/require-sandbox.sh ./bin/zr
#
# Exit 0 = token valid, ZR_ENV (when set) matches the active environment, and
# the active tenant looks like a sandbox — or the operator explicitly opted in
# with ZR_E2E_ALLOW_PROD=1. Unrecognized hosts FAIL CLOSED: a production-host
# denylist would fail OPEN on any host it does not know (a new prod region, a
# custom domain), which is the wrong default in front of irreversible writes.

# require_sandbox_check <zr-binary> — prints one diagnostic line per check
# (OK:/WARN:/FAIL: prefixes); returns 1 on the first failed check. No `exit`,
# no color — callers (the E2E harness, a probing session) decide both.
require_sandbox_check() {
  local zr="$1" auth_out
  [ -x "$zr" ] || {
    echo "FAIL: zr binary not found/executable at $zr (build it first)"
    return 1
  }
  # auth status always exits 0 and prints "Token: valid|expired"; the only
  # reliable signal of a usable session is a "Token: ... valid" line.
  auth_out=$("$zr" auth status 2>&1)
  if ! echo "$auth_out" | grep -qE "Token:[[:space:]]+valid"; then
    echo "FAIL: token not valid: $(echo "$auth_out" | grep -i 'token' | head -1)"
    return 1
  fi

  # When ZR_ENV pins the environment, confirm the binary actually targeted it —
  # a mismatch means something unexpected; never write to an unnamed tenant.
  if [ -n "${ZR_ENV:-}" ]; then
    local env_name
    env_name=$(echo "$auth_out" | awk '/^Environment:/ {print $2}')
    if [ "$env_name" != "$ZR_ENV" ]; then
      echo "FAIL: ZR_ENV=$ZR_ENV but the active environment is '${env_name:-<unknown>}' — refusing an unexpected tenant"
      return 1
    fi
  fi

  # Zuora sandbox hosts carry a "sandbox" or ".test." marker
  # (rest.apisandbox.zuora.com, rest.test.ap.zuora.com); production hosts
  # (rest.zuora.com, rest.na/eu/ap.zuora.com) carry neither.
  local base_url
  base_url=$(echo "$auth_out" | awk '/^Base URL:/ {print $NF}')
  case "$base_url" in
    *sandbox*|*.test.*)
      echo "OK: sandbox tenant: $base_url"
      ;;
    *)
      if [ "${ZR_E2E_ALLOW_PROD:-0}" = "1" ]; then
        echo "WARN: ZR_E2E_ALLOW_PROD=1 — proceeding against a NON-sandbox tenant: ${base_url:-<unknown>}"
      else
        echo "FAIL: active tenant does not look like a sandbox: ${base_url:-<unknown>} — writes must NOT hit production. Switch with: zr config env apac-sandbox   (intentional override: ZR_E2E_ALLOW_PROD=1)"
        return 1
      fi
      ;;
  esac
  return 0
}

# Standalone execution (not sourced): check $1, or ./bin/zr at the repo root.
if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  set -uo pipefail
  root="$(cd "$(dirname "$0")/.." && pwd)"
  require_sandbox_check "${1:-$root/bin/zr}"
fi
