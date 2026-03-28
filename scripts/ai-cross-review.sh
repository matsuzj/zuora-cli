#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

OUT_DIR="${REPO_ROOT}/logs/ai-cross-review/$(date -u '+%Y%m%dT%H%M%SZ')"
mkdir -p "${OUT_DIR}"

have_cmd() { command -v "$1" >/dev/null 2>&1; }

# staged優先、なければworking tree diff
if git diff --cached --quiet; then
    DIFF_CMD=(git diff --patch)
    DIFF_LABEL="working-tree"
else
    DIFF_CMD=(git diff --cached --patch)
    DIFF_LABEL="staged"
fi

echo "Diff source: ${DIFF_LABEL}"
"${DIFF_CMD[@]}" > "${OUT_DIR}/diff.patch"

if [[ ! -s "${OUT_DIR}/diff.patch" ]]; then
    echo "差分がありません"
    exit 0
fi

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

# Codex レビュー
if have_cmd codex; then
    echo ""
    echo "=== Codex レビュー ==="
    codex exec --ask-for-approval never --sandbox read-only \
        "このdiffをレビューし、問題点を指摘してください。" \
        < "${OUT_DIR}/diff.patch" \
        | tee "${OUT_DIR}/codex.review.md" \
        || { echo "⚠️  Codex レビュー失敗（続行）"; review_exit=1; }
fi

echo ""
echo "レビュー結果: ${OUT_DIR}/"
exit "${review_exit}"
