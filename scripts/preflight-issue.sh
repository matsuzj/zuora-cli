#!/bin/bash
# preflight-issue.sh <issue#> [--claim] — duplicate-work preflight for the
# autonomous pipeline (#523). Open issues may already be implemented on main
# (it happened: #403 requested the Expect matcher that PR #319 had merged
# three days earlier), and stale topic branches may pre-exist. Run this
# BEFORE implementing any open issue:
#   1. fetch origin and show the issue's current state/labels
#   2. list remote branches mentioning the issue number
#   3. grep origin/main for the issue body's `backtick` tokens — hits are
#      already-implemented hints, verify them before writing code
#   4. print the isolated-worktree command (never work in the user's checkout)
# --claim adds the ai-in-progress label (the pipeline's live intermediate
# state); refused unless the issue is OPEN.
set -euo pipefail
if [ $# -lt 1 ]; then
  echo "usage: $0 <issue-number> [--claim]" >&2
  exit 2
fi
n="$1"
claim=false
[ "${2:-}" = "--claim" ] && claim=true

cd "$(git rev-parse --show-toplevel)"
git fetch origin --quiet

echo "== issue #$n"
state="$(gh issue view "$n" --json state --jq .state)"
gh issue view "$n" --json state,title,labels,updatedAt \
  --jq '"state: \(.state)\ntitle: \(.title)\nlabels: \([.labels[].name] | join(", "))\nupdated: \(.updatedAt)"'
if [ "$state" != "OPEN" ]; then
  echo "WARNING: issue is $state — the work may already be merged; do not implement from its text alone."
fi

echo
echo "== pre-existing topic branches mentioning #$n (origin)"
# refname column only — issue numbers also appear inside commit SHAs (Codex P3)
git ls-remote --heads origin | awk '{print $2}' | grep -E "(^|[^0-9])$n([^0-9]|\$)" || echo "  (none)"

echo
echo "== issue-body backtick tokens found on origin/main (already-implemented hints)"
tokens="$(gh issue view "$n" --json body --jq .body | grep -oE '\`[^\`]{4,60}\`' | tr -d '\`' | sort -u | head -12 || true)"
if [ -z "$tokens" ]; then
  echo "  (no backtick tokens in the issue body)"
else
  found=false
  while IFS= read -r tok; do
    hits="$(git grep -c -F -- "$tok" origin/main -- '*.go' '*.sh' 2>/dev/null | head -3 || true)"
    if [ -n "$hits" ]; then
      found=true
      echo "  FOUND '$tok':"
      printf '%s\n' "$hits" | sed 's/^/    /'
    fi
  done <<< "$tokens"
  $found || echo "  (none of the tokens exist on origin/main)"
fi

echo
echo "== next steps"
echo "  re-verify each claimed defect against origin/main before writing code, then:"
echo "    git worktree add .claude/worktrees/issue-$n -b <type>/$n-<slug> origin/main"
if $claim; then
  if [ "$state" = "OPEN" ]; then
    gh issue edit "$n" --add-label ai-in-progress >/dev/null
    echo "  labeled ai-in-progress"
  else
    echo "  --claim refused: issue is $state" >&2
    exit 1
  fi
fi
