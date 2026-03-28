# Mac mini中心 マルチAI自動開発ワークフロー 最終版
## zuora-cli (zr) — サブスクリプション契約・ローカル実行

---

## 設計思想

### なぜMac mini中心か

GitHub ActionsでAIエージェントを動かすと、ランナーの分単位課金と意図しないワークフロー発火リスクが発生します。Mac miniローカル実行なら、ランナー費用ゼロ、手動コマンド起点で誤発火なし、Ctrl+Cで即停止、ターミナルでリアルタイムデバッグが可能です。

### サブスクリプション契約の前提

3つのAIエージェントすべてサブスクリプション契約済みのため、API従量課金は発生しません。

| エージェント | 契約 | 認証方式 | 主な制約 |
|---|---|---|---|
| Claude Code | Claude Max | OAuth → macOS Keychain | 使用量上限 / リセット周期 |
| Codex CLI | ChatGPT Pro/Plus | OAuth（device code対応） | レート制限 |
| Gemini CLI | Gemini Advanced / 無料枠 | Googleログイン or API Key | 60 RPM / 1000 RPD（無料枠） |

**重要**: `ANTHROPIC_API_KEY` が設定されているとサブスクリプションではなくAPI従量課金が優先されます。必ず未設定であること。

### zuora-cli (zr) プロジェクト固有の前提

- CLI名は `zr`、GitHub CLI (`gh`) のパターンに寄せた設計
- Go 1.26.1（go.mod宣言）
- 既存の `AGENTS.md`（Conventional Commits、`make test`/`make lint`、コマンド配置方針）を尊重
- 既存Makefileに `build/test/lint/fmt/check` ターゲットあり
- Zuora APIはOAuth 2.0クライアントクレデンシャルフロー、環境変数プレフィックス `ZUORA_`
- コマンド配置: `pkg/cmd/<resource>/<action>/`
- 出力制御: `--json`/`--jq`/`--template`

---

## エージェント認証とヘッドレス実行

### Claude Code

**サブスクリプションOAuth認証**

`claude auth login` を実行し、ブラウザでClaude.aiアカウントを認証します。`--console` を付けるとAnthropic Console（API従量課金）ログインになるので、サブスクリプション運用では**付けません**。トークンはmacOS Keychainに暗号化保存されます。

```bash
# サブスクリプションでログイン（--console を付けない）
claude auth login

# 認証状態の確認（exit 0: ログイン済み / exit 1: 未ログイン）
claude auth status

# ⚠️ API Keyが設定されていないことを確認
echo $ANTHROPIC_API_KEY  # 空であること
unset ANTHROPIC_API_KEY  # 設定されている場合
```

**ヘッドレス実行**

- `-p` / `--print` — 非インタラクティブに応答を出力
- `--bare` — hooks/skills/plugins/MCP/CLAUDE.md自動検出をスキップし、スクリプト呼び出しを高速化
- `--tools` — 利用可能な組み込みツール自体を制限（例: `"Bash,Edit,Read"`）
- `--allowedTools` — 権限プロンプトなしで実行できるツールパターンを指定（`--tools` とは目的が異なる）
- `--permission-mode plan` — ファイル変更を行わない計画モード

```bash
# 計画ステージ（ファイル変更なし）
claude --bare --permission-mode plan -p "計画を作成してください"

# 実装ステージ（ツールを限定）
claude --bare --tools "Bash,Edit,Read" -p "実装してください"

# 特定コマンドのみ権限プロンプトなしで許可
claude --allowedTools "Bash(git diff *)" "Bash(make check)" -p "..."
```

**SSH経由ヘッドレスMac miniでの初回認証**

Claude Code公式では、ブラウザが「間違ったマシン」で開く場合、**表示されたURLをコピーしてローカルブラウザで開く**ことが推奨されています。

### Codex CLI

**ChatGPTサブスクリプション認証**

Codex CLIはChatGPT Plus/Pro/Business/Edu/Enterpriseに含まれます。ヘッドレス環境での初回認証には `--device-auth` が用意されており、ブラウザを開かずOAuth device code flowで認証できます。

```bash
# インストール
npm i -g @openai/codex

# ヘッドレス向け認証（ブラウザ不要）
codex login --device-auth

# 認証状態の確認（exit 0: ログイン済み）
codex login status
```

**ヘッドレス実行**

`codex exec` が非インタラクティブ実行コマンドです。`--full-auto` は `--ask-for-approval on-request` と `--sandbox workspace-write` のショートカットです。グローバルフラグはサブコマンドの後ろに置きます。

```bash
codex exec --full-auto "テストを生成して実行してください"

# より厳密な制御
codex exec --ask-for-approval never --sandbox workspace-write "..."
```

**サンドボックス（macOS）**

macOS 12+ではApple Seatbelt（`sandbox-exec`）を使用し、書き込み範囲を限定、アウトバウンドネットワークはデフォルトブロックされます。

**認証情報の保存**

`~/.codex/auth.json` に平文保存される可能性があるため、パスワード同等に扱い、コミット・共有は厳禁です。`.codex/config.toml` で `cli_auth_credentials_store = "keyring"` に設定するとOSのキーチェーンに保存できます。

### Gemini CLI

**Googleログイン認証と制約**

Gemini CLIの認証は「Login with Google」を使い、ブラウザがCLIの待ち受ける `localhost` にリダイレクトされます。ヘッドレス環境では最初の1回だけブラウザ到達性が必要です。認証済みであればキャッシュが使われます。未認証のヘッドレス環境では `GEMINI_API_KEY` が必要です。

```bash
# インストール
brew install gemini-cli

# 初回認証（ブラウザ到達性が必要）
gemini  # → Login with Google を選択

# SSH経由の場合: ポートフォワーディングでlocalhostリダイレクトを中継
ssh -L <PORT>:localhost:<PORT> yusuke@mac-mini.local
```

**ヘッドレス実行**

- `--prompt` / `-p` — 非インタラクティブ強制
- `--yolo` / `-y` — **非推奨（Deprecated）**、代わりに `--approval-mode=yolo` を使用

```bash
# diffをstdinで渡してレビュー
git diff --patch | gemini -p "このdiffをレビューしてください"

# 自動承認（推奨形）
gemini -p "..." --approval-mode=yolo
```

**設定ファイル**: `.gemini/settings.json`（プロジェクトスコープ）
**サンドボックス**: `--sandbox` フラグ + `.gemini/sandbox.Dockerfile` でDockerベースの隔離が可能

---

## アーキテクチャ全体像

```
┌─────────────────────────────────────────────────────┐
│  GitHub: matsuzj/zuora-cli                          │
│  ┌──────────┐    ┌──────────┐    ┌──────────────┐   │
│  │  Issues   │    │ Branches │    │Pull Requests │   │
│  │(ai-impl) │    │ ai/issue-│    │  (自動作成)   │   │
│  └────┬─────┘    └────▲─────┘    └──────▲───────┘   │
│       │               │                │            │
└───────┼───────────────┼────────────────┼────────────┘
        │ gh issue view │ git push       │ gh pr create
        │               │                │
┌───────▼───────────────┼────────────────┼────────────┐
│  Mac mini (常時稼働・ヘッドレス)                      │
│                                                      │
│  ┌──────────────────────────────────────────┐        │
│  │  ai-orchestrator.sh                      │        │
│  │                                          │        │
│  │  1. Issue取得 (gh issue view)             │        │
│  │  2. Worktree作成 (.worktrees/issue-N)    │        │
│  │  3. Plan    → Claude Code (--perm plan)  │        │
│  │  4. Implement → Claude Code              │        │
│  │  5. Review  → Gemini CLI (-p)            │        │
│  │  6. Test    → Codex CLI (exec)           │        │
│  │  7. Commit & Push                        │        │
│  │  8. PR作成  → gh pr create               │        │
│  └──────────────────────────────────────────┘        │
│                                                      │
│  認証: Claude=Keychain, Codex=device-auth,           │
│        Gemini=Google login                           │
└──────────────────────────────────────────────────────┘
```

### エージェント役割分担

| ステージ | エージェント | 選定理由 |
|---|---|---|
| 計画 (Plan) | Claude Code | 推論能力が高く、`--permission-mode plan` で安全に分析 |
| 実装 (Implement) | Claude Code | 自律的な実装・ビルド検証の反復ループに強い |
| クロスレビュー | Gemini CLI | 100万トークンのコンテキストで大きなdiffも一括レビュー。実装者と異なるエージェントが担当 |
| テスト生成 | Codex CLI | Seatbeltサンドボックスでの隔離実行。ネットワーク遮断下で安全にテスト |

### トリガー方式

| 方式 | 説明 | 推奨度 |
|---|---|---|
| 手動コマンド | `make ai ISSUE=42` | ★★★ 最初はこれ |
| cronポーリング | `gh issue list --label ai-implement` を定期実行 | ★★ 安定後に移行 |
| Webhook受信 | Cloudflare Tunnel経由でGitHub Webhook直接受信 | ★ リアルタイム性重視 |

---

## ai-orchestrator.sh

```bash
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

    # ANTHROPIC_API_KEY ガード（サブスク優先のため未設定必須）
    if [[ -n "${ANTHROPIC_API_KEY:-}" ]]; then
        log "❌ ANTHROPIC_API_KEY が設定されています。API従量課金が優先されます。"
        log "   → unset ANTHROPIC_API_KEY を実行してください。"
        exit 1
    fi

    if ! have_cmd gh; then
        log "❌ gh CLI が必要です。"
        exit 1
    fi

    # Claude Code
    if have_cmd claude; then
        if claude auth status >/dev/null 2>&1; then
            log "  ✅ Claude Code: 認証OK"
        else
            log "  ⚠️  Claude Code: 未認証 → claude auth login を実行"
        fi
    else
        log "  ⏭️  Claude Code: 未インストール（plan/implementスキップ）"
    fi

    # Codex CLI
    if have_cmd codex; then
        if codex login status >/dev/null 2>&1; then
            log "  ✅ Codex CLI: 認証OK"
        else
            log "  ⚠️  Codex CLI: 未認証 → codex login --device-auth を実行"
        fi
    else
        log "  ⏭️  Codex CLI: 未インストール（testスキップ）"
    fi

    # Gemini CLI
    if have_cmd gemini; then
        log "  ✅ Gemini CLI: インストール済み"
    else
        log "  ⏭️  Gemini CLI: 未インストール（reviewスキップ）"
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
    gh issue list --label "ai-implement" --state open --limit 1 --json number --jq '.[0].number // empty'
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
    BRANCH="ai/issue-${ISSUE_NUMBER}-${slug}"
    WT_DIR="${REPO_ROOT}/.worktrees/issue-${ISSUE_NUMBER}"
    BASE_REF="origin/${DEFAULT_BASE_BRANCH}"

    log "🌿 Worktree準備: ${WT_DIR} (${BRANCH})"

    git fetch origin "${DEFAULT_BASE_BRANCH}"
    mkdir -p "${REPO_ROOT}/.worktrees"

    if [[ -d "${WT_DIR}" ]]; then
        log "  既存worktreeを再利用"
    else
        git worktree add -b "${BRANCH}" "${WT_DIR}" "${BASE_REF}"
    fi

    # ラベル更新
    gh issue edit "${ISSUE_NUMBER}" --add-label "ai-in-progress" >/dev/null 2>&1 || true
}

#-------------------------------------------------------------------
# ステージ1: 計画 (Claude Code — permission-mode plan)
#-------------------------------------------------------------------
stage_plan() {
    if ! have_cmd claude; then
        log "  ⏭️  Claude未インストール — スキップ"
        return 0
    fi

    log "📐 ステージ1: Claude Code で実装計画を作成..."

    (
        cd "${WT_DIR}"
        echo "ISSUE_JSON:
${ISSUE_JSON}

PROMPT:
zuora-cli (zr) プロジェクトの実装計画を作成してください。
まず AGENTS.md を読み、プロジェクト規約に従ってください。

以下を含む計画を出力:
1. 変更・作成するファイル一覧
2. 各ファイルの具体的な変更内容
3. 追加すべきテストケース
4. 受け入れ基準
5. リスクとロールバック方法

重要:
- コマンド配置は pkg/cmd/<resource>/<action>/
- make check (lint+test) が通ること
- Zuora APIはOAuth 2.0クライアントクレデンシャルフロー
- 環境変数プレフィックスは ZUORA_
" | claude --bare --permission-mode plan -p \
            "上記Issueの実装計画を作成してください。まずAGENTS.mdを読んでください。"
    ) > "${LOG_DIR}/plan.md" 2>&1

    log "  ✅ 計画完了 → ${LOG_DIR}/plan.md"
}

#-------------------------------------------------------------------
# ステージ2: 実装 (Claude Code)
#-------------------------------------------------------------------
stage_implement() {
    if ! have_cmd claude; then
        log "  ⏭️  Claude未インストール — スキップ"
        return 0
    fi

    log "🔨 ステージ2: Claude Code で実装..."

    local plan_content=""
    if [[ -f "${LOG_DIR}/plan.md" ]]; then
        plan_content=$(cat "${LOG_DIR}/plan.md")
    fi

    (
        cd "${WT_DIR}"
        claude --bare \
            --tools "Bash,Edit,Read" \
            --dangerously-skip-permissions \
            -p "以下のIssueを実装してください。まずAGENTS.mdを読んでください。

Issue:
${ISSUE_JSON}

計画:
${plan_content}

重要:
- このリポジトリはzr (zuora-cli) をビルドします
- make check (lint+test) を実行して通ることを確認
- 変更はこのIssueのスコープに限定
- シークレットをログやエラーメッセージに出力しない
"
    ) > "${LOG_DIR}/implement.log" 2>&1

    # ビルド確認
    (cd "${WT_DIR}" && make check 2>/dev/null) && log "  ✅ make check 通過" || log "  ⚠️  make check 失敗（手動確認推奨）"

    log "  ✅ 実装完了"
}

#-------------------------------------------------------------------
# ステージ3: クロスレビュー (Gemini CLI)
#-------------------------------------------------------------------
stage_review() {
    if ! have_cmd gemini; then
        log "  ⏭️  Gemini未インストール — スキップ"
        return 0
    fi

    log "🔍 ステージ3: Gemini CLI でクロスレビュー..."

    (
        cd "${WT_DIR}"
        git diff --patch "${BASE_REF}...HEAD" \
            | gemini -p "あなたはGoのシニアレビュアーです。このdiffをレビューしてください。

コンテキスト:
- プロジェクトは matsuzj/zuora-cli (zr)、gh CLIのパターンに準拠
- 小さく、テスト可能な変更を重視

以下の観点でレビュー:
1. 正確性: 意図通りの実装か
2. セキュリティ: シークレット漏洩、入力バリデーション
3. Goベストプラクティス: エラーハンドリング、命名規則
4. テストカバレッジ: 新しい関数にテストがあるか
5. CLI UX: --json/--jq/--template の一貫性

出力形式:
- サマリー
- 必須修正（ファイル/行ヒント付き）
- 改善提案
- テスト提案
"
    ) > "${LOG_DIR}/review.md" 2>&1

    log "  ✅ レビュー完了 → ${LOG_DIR}/review.md"
}

#-------------------------------------------------------------------
# ステージ4: テスト生成 (Codex CLI)
#-------------------------------------------------------------------
stage_test() {
    if ! have_cmd codex; then
        log "  ⏭️  Codex未インストール — スキップ"
        return 0
    fi

    log "🧪 ステージ4: Codex CLI でテスト生成..."

    (
        cd "${WT_DIR}"
        codex exec \
            --ask-for-approval never \
            --sandbox workspace-write \
            "このブランチの変更に対してGoテストを追加・改善してください。
実行コマンド: make test（または go test -race -count=1 ./...）
失敗があれば修正してください。
AGENTS.md のテスト規約に従ってください。"
    ) > "${LOG_DIR}/test.log" 2>&1

    # テスト確認
    (cd "${WT_DIR}" && make test 2>/dev/null) && log "  ✅ 全テスト通過" || log "  ⚠️  テスト失敗あり（手動確認推奨）"

    log "  ✅ テスト生成完了"
}

#-------------------------------------------------------------------
# ステージ5: Commit & Push & PR作成
#-------------------------------------------------------------------
stage_pr() {
    log "🚀 ステージ5: Commit & Push & PR作成..."

    (
        cd "${WT_DIR}"
        git add -A
        if git diff --cached --quiet; then
            log "  変更なし — コミットスキップ"
            return 0
        fi
        git commit -m "feat: implement issue #${ISSUE_NUMBER}"
        git push -u origin "${BRANCH}" --force-with-lease
    )

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
| クロスレビュー | Gemini CLI | ✅ |
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
    ) 2>/dev/null || log "  ⚠️  PR作成スキップ（既存PRあり？）"

    # ラベル更新
    gh issue edit "${ISSUE_NUMBER}" \
        --add-label "ai-pr-created" \
        --remove-label "ai-in-progress" \
        --remove-label "ai-implement" >/dev/null 2>&1 || true

    log "  ✅ PR作成完了"
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
```

---

## 補助スクリプト

### scripts/poll-and-run.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

LIMIT="${AI_POLL_LIMIT:-3}"
LOCK_FILE="/tmp/ai-orchestrator.lock"

# 二重実行防止
if [[ -f "${LOCK_FILE}" ]]; then
    PID=$(cat "${LOCK_FILE}")
    if kill -0 "${PID}" 2>/dev/null; then
        echo "既に実行中 (PID: ${PID})"
        exit 0
    fi
fi
echo $$ > "${LOCK_FILE}"
trap "rm -f ${LOCK_FILE}" EXIT

issues_json="$(gh issue list --label "ai-implement" --state open --limit "${LIMIT}" --json number)"
count="$(echo "${issues_json}" | jq 'length')"

if [[ "${count}" -eq 0 ]]; then
    echo "ai-implement Issueなし"
    exit 0
fi

echo "${issues_json}" | jq -r '.[].number' | while read -r n; do
    echo "$(date): Issue #${n} の処理を開始"
    ./ai-orchestrator.sh --issue "${n}" --stage all
done
```

### scripts/ai-cross-review.sh

```bash
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

# Claude レビュー
if have_cmd claude && [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
    echo "=== Claude Code レビュー ==="
    cat "${OUT_DIR}/diff.patch" \
        | claude --bare --tools "Read" --permission-mode plan -p \
          "このdiffをバグとセキュリティの観点でレビューしてください。diff外の変更は提案しないでください。" \
        | tee "${OUT_DIR}/claude.review.md"
fi

# Gemini レビュー
if have_cmd gemini; then
    echo ""
    echo "=== Gemini レビュー ==="
    cat "${OUT_DIR}/diff.patch" \
        | gemini -p "このdiffをGoベストプラクティスとパフォーマンスの観点でレビューしてください。" \
        | tee "${OUT_DIR}/gemini.review.md"
fi

# Codex レビュー
if have_cmd codex; then
    echo ""
    echo "=== Codex レビュー ==="
    codex exec --ask-for-approval never --sandbox read-only \
        "このdiffをレビューし、問題点を指摘してください。" \
        < "${OUT_DIR}/diff.patch" \
        | tee "${OUT_DIR}/codex.review.md"
fi

echo ""
echo "レビュー結果: ${OUT_DIR}/"
```

---

## Makefile（既存ターゲットへの追記）

```makefile
# --- AI Orchestration ---
.PHONY: ai ai-plan ai-impl ai-review ai-test ai-pr ai-quick-review ai-auth ai-status

ISSUE ?= $(shell gh issue list --label "ai-implement" --state open --limit 1 --json number --jq '.[0].number // empty')

ai:
	./ai-orchestrator.sh --issue $(ISSUE) --stage all

ai-plan:
	./ai-orchestrator.sh --issue $(ISSUE) --stage plan

ai-impl:
	./ai-orchestrator.sh --issue $(ISSUE) --stage implement

ai-review:
	./ai-orchestrator.sh --issue $(ISSUE) --stage review

ai-test:
	./ai-orchestrator.sh --issue $(ISSUE) --stage test

ai-pr:
	./ai-orchestrator.sh --issue $(ISSUE) --stage pr

ai-quick-review:
	./scripts/ai-cross-review.sh

ai-auth:
	@echo "=== ANTHROPIC_API_KEY ===" && \
	if [ -n "$${ANTHROPIC_API_KEY:-}" ]; then echo "⚠️  設定済み（API課金優先）"; else echo "✅ 未設定（サブスク優先）"; fi
	@echo "=== Claude Code ===" && (command -v claude >/dev/null 2>&1 && claude auth status || echo "未インストール")
	@echo "=== Codex CLI ===" && (command -v codex >/dev/null 2>&1 && codex login status || echo "未インストール")
	@echo "=== Gemini CLI ===" && (command -v gemini >/dev/null 2>&1 && gemini --version || echo "未インストール")

ai-status:
	@echo "Branch: $$(git rev-parse --abbrev-ref HEAD)"
	@echo "Worktrees:"; git worktree list
	@echo "最新ログ:"; ls -td logs/ai-orchestrator/*/ 2>/dev/null | head -1 | xargs -I{} cat {}run.log 2>/dev/null || echo "ログなし"
```

使い方:

```bash
make ai ISSUE=42          # 全パイプライン（Issue指定）
make ai                   # ai-implementラベルの最初のIssueを自動選択
make ai-plan ISSUE=42     # 計画のみ
make ai-quick-review      # ステージ済み変更を即座にレビュー
make ai-auth              # 認証状態確認
```

---

## エージェント設定ファイル

### CLAUDE.md

```markdown
# CLAUDE.md - zuora-cli (zr) プロジェクトコンテキスト

## プロジェクト概要
matsuzj/zuora-cli — Zuora Billing API操作用Go CLI `zr`。
GitHub CLI (gh) のパターンに準拠: サブコマンド、一貫したフラグ、機械読み取り可能な出力。

## ビルド・テスト・Lint
- `make build` → `./bin/zr`
- `make test`  → `go test -race -count=1 ./...`
- `make lint`  → `go vet ./... + staticcheck ./...`
- `make check` → lint + test

## Go バージョン
go 1.26.1 (go.mod)

## アーキテクチャ・規約（AGENTS.md準拠）
- Conventional Commits
- gofmt必須、レーステスト必須
- コマンド配置: `pkg/cmd/<resource>/<action>/`
- `cmd/zr/main.go` は最小限
- DI用Factoryパターン使用

## Zuora認証
- OAuth 2.0 クライアントクレデンシャルフロー
- 環境変数プレフィックス: `ZUORA_`
- シークレットをログ・エラーメッセージに出力しないこと

## コーディングルール
- 不要な依存追加を避ける
- Issue単位の小さなPR
- 出力モード: `--json`/`--jq`/`--template` の一貫性を維持
- テスト: `httptest.NewServer` でHTTPモック、外部APIコール禁止
```

### .claude/settings.json

```json
{
  "permissions": {
    "allow": [
      "Bash(git status)",
      "Bash(git diff *)",
      "Bash(git log *)",
      "Bash(make check)",
      "Bash(make test)",
      "Bash(make lint)",
      "Bash(go test *)",
      "Bash(go vet *)",
      "Bash(staticcheck *)",
      "Read(./**)"
    ],
    "deny": [
      "Bash(curl *)",
      "Read(./.env)",
      "Read(./.env.*)",
      "Read(./secrets/**)",
      "Read(~/.ssh/**)",
      "Read(~/.aws/**)"
    ]
  },
  "sandbox": {
    "enabled": true,
    "autoAllowBashIfSandboxed": true,
    "filesystem": {
      "denyRead": ["~/.aws/credentials"]
    },
    "network": {
      "allowedDomains": [],
      "allowLocalBinding": false
    }
  }
}
```

### .codex/config.toml

```toml
# 公式ドキュメントで確認済みのキーのみ使用
model = "gpt-5-codex"
model_provider = "openai"
cli_auth_credentials_store = "keyring"

developer_instructions = """
This is matsuzj/zuora-cli (zr). Follow AGENTS.md and keep changes minimal and tested.
Prefer make check; avoid external network calls in tests.
"""
```

### .gemini/settings.json

```json
{
  "general": {
    "defaultApprovalMode": "default",
    "enableNotifications": false
  },
  "output": {
    "format": "text"
  },
  "model": {
    "name": "gemini-2.5-pro"
  },
  "tools": {
    "sandboxNetworkAccess": false,
    "sandboxAllowedPaths": ["./"]
  }
}
```

---

## クロスレビューパターン

### 「書いたエージェントは自分の出力をレビューしない」原則

マルチモデル協調の目的は「モデルバイアス・認知的盲点・コンテキスト限界」の回避です。同一モデルによるセルフレビューはこれらの問題を検出できません。本ワークフローでは実装者（Claude）とレビュー者（Gemini）を必ず分離します。

### 3つのパターン

| パターン | 説明 | レート制限への影響 | 推奨場面 |
|---|---|---|---|
| シーケンシャルハンドオフ | Claude→Gemini→Codex の順にバトン | 最小（cooldown挿入容易） | **本ワークフローで採用** |
| パラレルレビュー | Gemini+Codexを並列実行し発見を集約 | 中（同時呼出し） | 徹底的なPRレビュー |
| 敵対的ループ | Geminiがレビュー→Claude修正→再レビューの反復 | 大（呼出し回数倍増） | 品質最重視の重要機能 |

---

## セキュリティガードレール

### サンドボックス隔離

| エージェント | 方式 | ネットワーク |
|---|---|---|
| Claude Code | `.claude/settings.json` の `permissions.deny` + `sandbox` | 明示許可のみ |
| Codex CLI | Apple Seatbelt (`sandbox-exec`) | デフォルト遮断 |
| Gemini CLI | `--sandbox` + Docker (`sandbox.Dockerfile`) | 設定で制御 |

### ブランチ保護

- AIは `ai/issue-*` ブランチ + git worktreeで隔離
- mainへの直接プッシュは禁止（既存AGENTS.mdで規定済み）
- PR必須、CI通過必須、人間の承認必須

### 認証トークン管理

- Claude: macOS Keychainに暗号化保存
- Codex: `cli_auth_credentials_store = "keyring"` でKeychainに保存推奨。`~/.codex/auth.json` が平文の場合はパスワード同等に扱い、コミット・共有厳禁
- Gemini: OSの認証キャッシュ
- `ANTHROPIC_API_KEY` は設定しない

---

## GitHub Actions（最小CI・AIは実行しない）

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:

permissions:
  contents: read

jobs:
  test-and-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true
      - name: Test
        run: go test -race -count=1 ./...
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v9
        with:
          version: v2.11
          args: --timeout=5m
```

---

## 既存オーケストレーター比較

| ツール | 特徴 | 本ワークフローとの関係 |
|---|---|---|
| **Zeroshot** | CLIで `--worktree`/`--pr`/`--ship` 指定、Claude/Codex/Gemini対応 | 抽象化レベルが高い。成熟したら置き換え候補 |
| **parallel-code** | 3エージェントを独立worktreeで並行実行 | GUIアプリのためヘッドレスMac miniとは相性が悪い |
| **claude_code_bridge** | マルチモデル協調の分割ペインターミナル | 対話的ツール。自動化パイプラインには不向き |
| **GitHub Agentic Workflows** | GitHub Actions上でエージェント実行 | 「CIでAI不実行」方針と衝突。比較対象として理解する価値あり |

---

## クイックスタート

```bash
# 1. リポジトリに移動
cd /path/to/zuora-cli

# 2. ツールインストール
npm i -g @anthropic-ai/claude-code
npm i -g @openai/codex
brew install gemini-cli

# 3. 認証（各ツール初回のみ）
unset ANTHROPIC_API_KEY
claude auth login          # ブラウザ認証
codex login --device-auth  # デバイスコード認証（ブラウザ不要）
gemini                     # Login with Google（ブラウザ必要）

# 4. 認証確認
make ai-auth

# 5. ワークフロー用ファイルを配置
chmod +x ai-orchestrator.sh scripts/*.sh

# 6. Issueにラベルを付ける
gh label create "ai-implement" --force
gh issue edit 42 --add-label "ai-implement"

# 7. まず計画だけ試す
make ai-plan ISSUE=42
cat logs/ai-orchestrator/*/plan.md

# 8. 問題なければ全パイプライン
make ai ISSUE=42

# 9. 定期実行したい場合（cron）
crontab -e
# */5 * * * * cd /path/to/zuora-cli && ./scripts/poll-and-run.sh >> /var/log/ai-poll.log 2>&1
```

---

## ディレクトリ構成

```
zuora-cli/
├── ai-orchestrator.sh              # メインパイプライン
├── scripts/
│   ├── poll-and-run.sh             # cronポーリング
│   └── ai-cross-review.sh         # ローカルクロスレビュー
├── logs/                           # 実行ログ（.gitignore推奨）
│   ├── ai-orchestrator/
│   └── ai-cross-review/
├── .worktrees/                     # git worktree（.gitignore推奨）
├── .claude/settings.json           # Claude Code権限・サンドボックス
├── .codex/config.toml              # Codex CLI設定
├── .gemini/settings.json           # Gemini CLI設定
├── CLAUDE.md                       # Claude Codeプロジェクトコンテキスト
├── AGENTS.md                       # 全エージェント共通規約（既存）
├── Makefile                        # ショートカット（AI追記）
├── .github/workflows/ci.yml       # CI（テスト・Lintのみ）
├── pkg/cmd/                        # コマンド実装
├── internal/                       # 内部パッケージ
├── cmd/zr/main.go                  # エントリーポイント
└── go.mod                          # go 1.26.1
```
