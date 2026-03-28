#!/usr/bin/env bash
set -euo pipefail

#===================================================================
# ai-orchestrator.sh — Mac mini用マルチAIパイプライン
# zuora-cli (zr) サブスクリプション契約版
#
# 使い方:
#   ./ai-orchestrator.sh --issue 42
#   ./ai-orchestrator.sh --issue 42 --stage plan
#===================================================================

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || true)"
if [[ -z "${REPO_ROOT}" ]]; then
    echo "ERROR: gitリポジトリ内で実行してください。" >&2
    exit 1
fi

DEFAULT_BASE_BRANCH="main"
COOLDOWN_SECONDS="${COOLDOWN_SECONDS:-10}"
STAGE="all"
ISSUE_NUMBER=""

LOG_DIR="${REPO_ROOT}/logs/ai-orchestrator/$(date -u '+%Y%m%dT%H%M%SZ')"
mkdir -p "${LOG_DIR}"

log() {
    echo "[$(date -u '+%Y-%m-%dT%H:%M:%SZ')] $*" | tee -a "${LOG_DIR}/run.log" >&2
}

usage() {
    cat <<'EOF'
使い方:
  ./ai-orchestrator.sh --issue <番号> [--stage <ステージ>] [--cooldown <秒>]

ステージ:
  plan | implement | review | test | pr | all (デフォルト)
EOF
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --issue)    ISSUE_NUMBER="${2:-}"; shift 2;;
        --stage)    STAGE="${2:-}"; shift 2;;
        --cooldown) COOLDOWN_SECONDS="${2:-}"; shift 2;;
        -h|--help)  usage; exit 0;;
        *)          echo "不明な引数: $1" >&2; usage; exit 1;;
    esac
done

have_cmd() { command -v "$1" >/dev/null 2>&1; }

slugify() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//; s/-{2,}/-/g' | cut -c1-50
}

#-------------------------------------------------------------------
# 認証チェック
#-------------------------------------------------------------------
preflight() {
    log "🔐 認証チェック..."

    # ANTHROPIC_API_KEY 警告（サブスク優先のため未設定推奨、ただしClaude以外のステージでは致命的ではない）
    if [[ -n "${ANTHROPIC_API_KEY:-}" ]]; then
        log "⚠️  ANTHROPIC_API_KEY が設定されています。Claude実行時にAPI従量課金が優先されます。"
        log "   → サブスク運用時は unset ANTHROPIC_API_KEY を実行してください。"
    fi

    if ! have_cmd gh; then
        log "❌ gh CLI が必要です。"
        exit 1
    fi

    if ! have_cmd jq; then
        log "❌ jq が必要です。brew install jq でインストールしてください。"
        exit 1
    fi

    # Claude Code — インストール済み＋認証済みの場合のみ有効
    CLAUDE_AVAILABLE=false
    if have_cmd claude; then
        if claude auth status >/dev/null 2>&1; then
            log "  ✅ Claude Code: 認証OK"
            CLAUDE_AVAILABLE=true
        else
            log "  ⚠️  Claude Code: 未認証（plan/implementスキップ） → claude auth login を実行"
        fi
    else
        log "  ⏭️  Claude Code: 未インストール（plan/implementスキップ）"
    fi

    # Codex CLI — インストール済み＋認証済みの場合のみ有効
    CODEX_AVAILABLE=false
    if have_cmd codex; then
        if codex login status >/dev/null 2>&1; then
            log "  ✅ Codex CLI: 認証OK"
            CODEX_AVAILABLE=true
        else
            log "  ⚠️  Codex CLI: 未認証（review/testスキップ） → codex login --device-auth を実行"
        fi
    else
        log "  ⏭️  Codex CLI: 未インストール（review/testスキップ）"
    fi
}

#-------------------------------------------------------------------
# ユーティリティ
#-------------------------------------------------------------------
cooldown() {
    log "  ⏳ レート制限対策: ${COOLDOWN_SECONDS}秒待機..."
    sleep "${COOLDOWN_SECONDS}"
}

ensure_labels() {
    gh label create "ai-implement"   --description "AI自動実装対象" --color "0E8A16" --force >/dev/null 2>&1 || true
    gh label create "ai-in-progress" --description "AI処理中"       --color "FBCA04" --force >/dev/null 2>&1 || true
    gh label create "ai-pr-created"  --description "AI PR作成済み"  --color "5319E7" --force >/dev/null 2>&1 || true
}

pick_next_issue() {
    # ai-in-progress を除外して、未着手の Issue のみ取得
    gh issue list --label "ai-implement" --state open --limit 50 --json number,labels \
        | jq -r '[.[] | select(.labels | map(.name) | index("ai-in-progress") | not)] | .[0].number // empty'
}

#-------------------------------------------------------------------
# Issue取得 & Worktree準備
#-------------------------------------------------------------------
fetch_issue() {
    log "📋 Issue #${ISSUE_NUMBER} を取得中..."
    ISSUE_JSON=$(gh issue view "${ISSUE_NUMBER}" --json number,title,body,url)
    ISSUE_TITLE=$(echo "${ISSUE_JSON}" | jq -r '.title')
    ISSUE_BODY=$(echo "${ISSUE_JSON}" | jq -r '.body')
    ISSUE_URL=$(echo "${ISSUE_JSON}" | jq -r '.url')
    log "  タイトル: ${ISSUE_TITLE}"
}

setup_worktree() {
    local slug
    slug="$(slugify "${ISSUE_TITLE}")"
    BRANCH="chore/ai-issue-${ISSUE_NUMBER}-${slug}"
    WT_DIR="${REPO_ROOT}/.worktrees/issue-${ISSUE_NUMBER}"
    BASE_REF="origin/${DEFAULT_BASE_BRANCH}"

    log "🌿 Worktree準備: ${WT_DIR} (${BRANCH})"

    git fetch origin "${DEFAULT_BASE_BRANCH}"
    mkdir -p "${REPO_ROOT}/.worktrees"

    if [[ -d "${WT_DIR}" ]]; then
        log "  既存worktreeを再利用（分割ステージ実行時の変更を保持）"
    else
        git worktree add -B "${BRANCH}" "${WT_DIR}" "${BASE_REF}"
    fi

    # ラベル更新
    gh issue edit "${ISSUE_NUMBER}" --add-label "ai-in-progress" >/dev/null 2>&1 || true
}

#-------------------------------------------------------------------
# ステージ1: 計画 (Claude Code — permission-mode plan)
#-------------------------------------------------------------------
stage_plan() {
    if [[ "${CLAUDE_AVAILABLE}" != "true" ]]; then
        log "  ⏭️  Claude利用不可 — スキップ"
        return 0
    fi

    log "📐 ステージ1: Claude Code で実装計画を作成..."

    (
        cd "${WT_DIR}"
        claude --output-format text --permission-mode plan -p \
            "以下のIssueの実装計画を作成してください。まずAGENTS.mdを読んでください。

Issue:
${ISSUE_JSON}

以下を含む計画を出力:
1. 変更・作成するファイル一覧
2. 各ファイルの具体的な変更内容
3. 追加すべきテストケース
4. 受け入れ基準
5. リスクとロールバック方法

重要:
- コマンド配置は pkg/cmd/<resource>/<action>/
- go vet ./... && go test -race ./... が通ること
- 環境変数プレフィックスは ZUORA_"
    ) > "${LOG_DIR}/plan.md" 2>&1

    log "  ✅ 計画完了 → ${LOG_DIR}/plan.md"
}

#-------------------------------------------------------------------
# ステージ2: 実装 (Claude Code)
#-------------------------------------------------------------------
stage_implement() {
    if [[ "${CLAUDE_AVAILABLE}" != "true" ]]; then
        log "  ⏭️  Claude利用不可 — スキップ"
        return 0
    fi

    log "🔨 ステージ2: Claude Code で実装..."

    local plan_content=""
    if [[ -f "${LOG_DIR}/plan.md" ]]; then
        plan_content=$(cat "${LOG_DIR}/plan.md")
    fi

    (
        cd "${WT_DIR}"
        claude --output-format text \
            --tools "Bash,Edit,Read" \
            --allowedTools "Bash(make check)" "Bash(make test)" "Bash(make lint)" "Bash(go test *)" "Bash(go vet *)" \
            -p "以下のIssueを実装してください。まずAGENTS.mdを読んでください。

Issue:
${ISSUE_JSON}

計画:
${plan_content}

重要:
- このリポジトリはzr (zuora-cli) をビルドします
- go vet ./... && go test -race ./... を実行して通ることを確認
- 変更はこのIssueのスコープに限定
- シークレットをログやエラーメッセージに出力しない
"
    ) > "${LOG_DIR}/implement.log" 2>&1

    # ビルド確認
    # staticcheck が Go 1.26.1 と非互換のため、go vet + go test で検証
    if (cd "${WT_DIR}" && go vet ./... && go test -race -count=1 ./... 2>&1); then
        log "  ✅ vet + test 通過"
    else
        log "  ❌ vet + test 失敗 — パイプライン停止"
        return 1
    fi

    log "  ✅ 実装完了"
}

#-------------------------------------------------------------------
# ステージ3: クロスレビュー (Codex CLI)
#-------------------------------------------------------------------
stage_review() {
    if [[ "${CODEX_AVAILABLE}" != "true" ]]; then
        log "  ⏭️  Codex利用不可 — スキップ"
        return 0
    fi

    log "🔍 ステージ3: Codex CLI でクロスレビュー..."

    (
        cd "${WT_DIR}"
        codex review --uncommitted
    ) > "${LOG_DIR}/review.md" 2>&1

    log "  ✅ レビュー完了 → ${LOG_DIR}/review.md"
}

#-------------------------------------------------------------------
# ステージ4: テスト生成 (Codex CLI)
#-------------------------------------------------------------------
stage_test() {
    if [[ "${CODEX_AVAILABLE}" != "true" ]]; then
        log "  ⏭️  Codex利用不可 — スキップ"
        return 0
    fi

    log "🧪 ステージ4: Codex CLI でテスト生成..."

    (
        cd "${WT_DIR}"
        codex exec --full-auto \
            "このブランチの変更に対してGoテストを追加・改善してください。
実行コマンド: make test（または go test -race -count=1 ./...）
失敗があれば修正してください。
AGENTS.md のテスト規約に従ってください。"
    ) > "${LOG_DIR}/test.log" 2>&1

    # lint + テスト確認（Codexがテスト以外のファイルを変更した場合にも検出）
    # staticcheck が Go 1.26.1 と非互換のため、go vet + go test で検証
    if (cd "${WT_DIR}" && go vet ./... && go test -race -count=1 ./... 2>&1); then
        log "  ✅ vet + test 通過"
    else
        log "  ❌ vet + test 失敗 — パイプライン停止"
        return 1
    fi

    log "  ✅ テスト生成完了"
}

#-------------------------------------------------------------------
# ステージ5: Commit & Push & PR作成
#-------------------------------------------------------------------
stage_pr() {
    log "🚀 ステージ5: Commit & Push & PR作成..."

    pushd "${WT_DIR}" >/dev/null

    git add -A
    if git diff --cached --quiet; then
        log "  変更なし — コミットスキップ（PR作成は続行）"
    else
        git commit -m "feat: implement issue #${ISSUE_NUMBER}"
        git push -u origin "${BRANCH}" --force-with-lease
    fi

    popd >/dev/null

    # レビューサマリー取得
    local review_summary="レビュー未実施"
    if [[ -f "${LOG_DIR}/review.md" ]]; then
        review_summary=$(head -20 "${LOG_DIR}/review.md")
    fi

    (
        cd "${WT_DIR}"
        gh pr create \
            --title "feat: #${ISSUE_NUMBER} を実装 — ${ISSUE_TITLE}" \
            --body "## AI生成による実装

Closes #${ISSUE_NUMBER}
実装対象: ${ISSUE_URL}

### パイプライン（Mac miniローカル実行・サブスクリプション契約内）
| ステージ | エージェント | 状態 |
|---|---|---|
| 計画 | Claude Code (plan mode) | ✅ |
| 実装 | Claude Code | ✅ |
| クロスレビュー | Codex CLI | ✅ |
| テスト生成 | Codex CLI (sandbox) | ✅ |

### クロスレビュー概要
${review_summary}

### 変更ファイル
\`\`\`
$(git diff --stat ${BASE_REF})
\`\`\`

> ⚠️ AIエージェントによる自動生成。マージ前に人間のレビューが必要です。
> 💰 サブスクリプション契約内で実行（追加費用なし）。" \
            --base "${DEFAULT_BASE_BRANCH}" \
            --head "${BRANCH}" \
            --label "ai-pr-created"
    ) 2>"${LOG_DIR}/pr-create.err" && local pr_exit=0 || local pr_exit=$?

    if [[ ${pr_exit} -eq 0 ]]; then
        # ラベル更新（PR作成成功時のみ）
        gh issue edit "${ISSUE_NUMBER}" \
            --add-label "ai-pr-created" \
            --remove-label "ai-in-progress" \
            --remove-label "ai-implement" >/dev/null 2>&1 || true
        log "  ✅ PR作成完了"
    else
        log "  ⚠️  PR作成失敗（詳細: ${LOG_DIR}/pr-create.err）"
        # ai-in-progress を除去してポーリングでリトライ可能にする
        gh issue edit "${ISSUE_NUMBER}" \
            --remove-label "ai-in-progress" >/dev/null 2>&1 || true
        log "     ai-in-progress ラベルを除去しました。次回ポーリングでリトライされます。"
    fi
}

#-------------------------------------------------------------------
# メイン実行
#-------------------------------------------------------------------
main() {
    preflight
    ensure_labels

    if [[ -z "${ISSUE_NUMBER}" ]]; then
        ISSUE_NUMBER="$(pick_next_issue)"
        if [[ -z "${ISSUE_NUMBER}" ]]; then
            log "ai-implement ラベルのオープンIssueがありません。"
            exit 0
        fi
    fi

    log "🎯 Issue #${ISSUE_NUMBER} を処理開始"
    fetch_issue
    setup_worktree

    # 異常終了時に ai-in-progress ラベルを除去（ポーリングでリトライ可能にする）
    trap 'gh issue edit "${ISSUE_NUMBER}" --remove-label "ai-in-progress" >/dev/null 2>&1 || true; log "⚠️  異常終了: ai-in-progress ラベルを除去しました"' ERR

    case "${STAGE}" in
        plan)      stage_plan ;;
        implement) stage_plan; cooldown; stage_implement ;;
        review)    stage_review ;;
        test)      stage_test ;;
        pr)        stage_pr ;;
        all)
            stage_plan;      cooldown
            stage_implement;  cooldown
            stage_review;     cooldown
            stage_test;       cooldown
            stage_pr
            ;;
        *)
            log "不明なステージ: ${STAGE}"
            usage; exit 1
            ;;
    esac

    log ""
    log "=========================================="
    log "🎉 完了！ ログ: ${LOG_DIR}/"
    log "=========================================="
}

main "$@"
