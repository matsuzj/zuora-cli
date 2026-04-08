#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

# 前提チェック
for cmd in gh jq; do
    if ! command -v "${cmd}" >/dev/null 2>&1; then
        echo "ERROR: ${cmd} が必要です。" >&2
        exit 1
    fi
done

LIMIT="${AI_POLL_LIMIT:-1}"  # Claude Max レート制限対策: デフォルト1件ずつ処理
LOCK_FILE="/tmp/ai-orchestrator-$(printf '%s' "${REPO_ROOT}" | shasum | cut -d' ' -f1).lock"

# 二重実行防止
if [[ -f "${LOCK_FILE}" ]]; then
    PID=$(cat "${LOCK_FILE}")
    if kill -0 "${PID}" 2>/dev/null; then
        echo "既に実行中 (PID: ${PID})"
        exit 0
    fi
fi
echo $$ > "${LOCK_FILE}"
trap 'rm -f "${LOCK_FILE}"' EXIT

# ai-implement ラベルがあり、かつ ai-in-progress ラベルがない Issue を取得
# limit をフィルタ前に十分大きく取り、フィルタ後に制限する
issues_json="$(gh issue list --label "ai-implement" --state open --limit 50 --json number,labels \
    | jq --argjson limit "${LIMIT}" '[.[] | select(.labels | map(.name) | index("ai-in-progress") | not)] | .[:$limit]')"
count="$(echo "${issues_json}" | jq 'length')"

if [[ "${count}" -eq 0 ]]; then
    echo "ai-implement Issueなし"
    exit 0
fi

echo "${issues_json}" | jq -r '.[].number' | while read -r n; do
    echo "$(date): Issue #${n} の処理を開始"
    ./ai-orchestrator.sh --issue "${n}" --stage all || echo "$(date): ⚠️  Issue #${n} 失敗（続行）"
done
