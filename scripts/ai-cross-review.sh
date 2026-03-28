#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

OUT_DIR="${REPO_ROOT}/logs/ai-cross-review/$(date -u '+%Y%m%dT%H%M%SZ')"
mkdir -p "${OUT_DIR}"

have_cmd() { command -v "$1" >/dev/null 2>&1; }

# staged + unstaged + untracked の全変更を単一 patch に統合
# git diff だけでは untracked が漏れるため intent-to-add を一時的に使用
# untracked ファイルを一時的に intent-to-add で追跡し、diff に含める
_intent_files=()
while IFS= read -r f; do
    _intent_files+=("${f}")
done < <(git ls-files --others --exclude-standard)

if [[ ${#_intent_files[@]} -gt 0 ]]; then
    git add -N "${_intent_files[@]}" 2>/dev/null || true
fi

# 全変更（staged + unstaged）を1つの patch に統合
git diff HEAD --patch > "${OUT_DIR}/diff.patch" 2>/dev/null || \
    git diff --patch > "${OUT_DIR}/diff.patch"

# intent-to-add で追加したファイルのみ元に戻す（ユーザーの既存 stage を壊さない）
if [[ ${#_intent_files[@]} -gt 0 ]]; then
    git reset HEAD -- "${_intent_files[@]}" >/dev/null 2>&1 || true
fi

if [[ ! -s "${OUT_DIR}/diff.patch" ]]; then
    echo "差分がありません"
    exit 0
fi

echo "Diff source: staged + unstaged + untracked"

# 各レビュアーの失敗が後続をブロックしないようにする
review_exit=0

# Claude レビュー
if have_cmd claude && [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
    echo "=== Claude Code レビュー ==="
    cat "${OUT_DIR}/diff.patch" \
        | claude --bare --tools "Read" --permission-mode plan -p \
          "このdiffをバグとセキュリティの観点でレビューしてください。diff外の変更は提案しないでください。" \
        | tee "${OUT_DIR}/claude.review.md" \
        || { echo "⚠️  Claude レビュー失敗（続行）"; review_exit=1; }
fi

# Codex レビュー（--uncommitted で untracked も含む）
if have_cmd codex; then
    echo ""
    echo "=== Codex レビュー ==="
    codex review --uncommitted \
        | tee "${OUT_DIR}/codex.review.md" \
        || { echo "⚠️  Codex レビュー失敗（続行）"; review_exit=1; }
fi

echo ""
echo "レビュー結果: ${OUT_DIR}/"
exit "${review_exit}"
