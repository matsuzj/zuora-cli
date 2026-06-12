#!/bin/bash
# Emits the README "Destructive operations" sentence from the ground truth:
# every command package that calls cmdutil.RequireConfirm. make lint fails
# when the README block between the destructive-list markers drifts from
# this output (P7-2 — the hand-written list went stale once already).
set -euo pipefail
cd "$(dirname "$0")/.."
cmds=$(grep -rl "RequireConfirm" pkg/cmd --include='*.go' \
  | grep -v _test \
  | sed 's|pkg/cmd/||; s|/[a-z_-]*\.go$||; s|/| |' \
  | sort \
  | sed 's/^/`zr /; s/$/`/' \
  | paste -sd ',' - | sed 's/,/, /g')
printf '**Destructive operations**: irreversible commands require an explicit `--confirm` flag: %s.\n' "$cmds"
