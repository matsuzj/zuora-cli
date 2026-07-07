#!/bin/bash
# Emits the "pending live verification" ledger from the ground truth: every
# LIVE-UNVERIFIED(<what>; since <date>; trigger: <unblock condition>) marker in
# the tree. make lint fails when the block between the pending-live markers in
# docs/e2e-test-skips.md drifts from this output (#521 — deferrals recorded
# only in PR bodies / closed issues evaporated; the marker is the registry).
# Resolve a marker by live-probing the assumption and replacing it with a
# dated "live-verified against <tenant> (<date>)" comment, then regenerate.
# Entries carry file paths only (no line numbers) so unrelated edits to a
# marker-bearing file do not churn the generated block; `make pending-live`
# prints exact line numbers on demand.
set -euo pipefail
cd "$(dirname "$0")/.."
# `|| true`: an empty marker set is a legitimate (target) state — grep's
# exit-1-on-no-match must not kill the script under set -e/pipefail.
{ grep -rn --include='*.go' --include='*.sh' 'LIVE-UNVERIFIED(' pkg internal tests 2>/dev/null || true; } \
  | sed -E 's/^([^:]+):[0-9]+:.*LIVE-UNVERIFIED\((.*)\).*$/- `\1` — \2/' \
  | sort -u
