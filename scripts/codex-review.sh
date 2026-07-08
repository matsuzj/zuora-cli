#!/bin/bash
# codex-review.sh — the ONE place the Codex invocation forms live (#529).
#
# AGENTS.md used to forbid freeform `codex exec` outright (a 2026-06-06 hang)
# while the plan-review protocol depends on exactly that form, which has worked
# reliably since 2026-06-28 (re-verified 2026-07-08: no hang, correct answer).
# Two prose layers held contradictory instructions; one script cannot.
#
#   codex-review.sh diff [base]     branch-diff review (default base: main)
#   codex-review.sh doc <file> [lens-prompt]
#                                   plan/document review — freeform codex exec
#                                   in the read-only sandbox; the file is read
#                                   by codex itself (works for large files)
#
# Codex is non-deterministic: follow the multi-pass convergence protocol in
# docs/plans/README.md (vary the lens each pass; two consecutive clean passes).
# De-scope ladder applies (AGENTS.md): breaks twice or unused for a quarter →
# demote to prose and delete. Kept passthrough-thin on purpose — codex's CLI
# churns fast (the hang→works flip took 22 days).
set -euo pipefail

usage() {
  echo "usage: $0 diff [base] | doc <file> [lens-prompt]" >&2
  exit 2
}
[ $# -ge 1 ] || usage
mode="$1"
shift

case "$mode" in
  diff)
    exec codex exec review --base "${1:-main}"
    ;;
  doc)
    [ $# -ge 1 ] || usage
    file="$1"
    # NOTE: assigned via if (not ${2:-default}) — bash 3.2 mis-parses quotes
    # inside ${parameter:-word} within double quotes.
    lens="${2:-}"
    if [ -z "$lens" ]; then
      lens="Adversarially review this plan/document: factual errors, internal contradictions, decisions that conflict with the repository code or settled conventions, and unstated risks. Verify claims against the repository before reporting. List findings by severity; end with CONVERGED if nothing material remains."
    fi
    [ -r "$file" ] || { echo "not readable: $file" >&2; exit 2; }
    exec codex exec --sandbox read-only --skip-git-repo-check \
      -c model_reasoning_effort="high" \
      "$lens

Review the file at: $file (read it from the working directory)."
    ;;
  *)
    usage
    ;;
esac
