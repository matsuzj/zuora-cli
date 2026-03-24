---
title: Zuora CLI (zr) 開発プラン
date: 2026-03-24
category: work
tags: [zuora, cli, go, api]
---

# Zuora CLI 開発プラン

## Context

Zuora の API 操作を効率化するため、GitHub CLI (gh) をモデルにした CLI ツールを開発する。
目的: Zuora のアカウント・サブスクリプション・オーダー等のリソース管理をターミナルから直感的に行えるようにする。
余っている MacBook Air で Claude Code を使って自動開発する。

## 決定事項

- **言語**: Go (Cobra フレームワーク)
- **リポジトリ**: `matsuzj/zuora-cli` (新規リポジトリ)
- **初期スコープ**: Account + Subscription の Read 操作のみ → 全 API を段階実装
- **テスト環境**: Zuora Sandbox 利用可能

## セッション引継ぎガイド

新しい Claude Code セッションでこのプランの続きを実装する際の手順:

1. **チェックリストを確認**: 各 Phase ファイルの実装チェックリストで `- [x]` の項目を確認し、次に実装すべき Phase を特定
2. **依存関係を確認**: 下記の Phase 依存関係を確認し、前提 Phase が完了しているか検証
3. **API 仕様を参照**: 各コマンドの Zuora API パスが記載されている。詳細なリクエスト/レスポンス仕様は https://developer.zuora.com/v1-api-reference/api.md を `WebFetch` で取得して確認
4. **既存コードを読む**: 既に実装済みのコマンド (例: `pkg/cmd/account/list/`) のパターンを踏襲して新コマンドを実装
5. **テスト実行**: `task test` で既存テストが壊れていないことを確認してから新コードを追加
6. **Sandbox 検証**: 実装後は `zr auth login` で Sandbox に接続し、実際の API 呼び出しで動作確認

## Phase 依存関係

```
Phase 0 (基盤) ← 全 Phase の前提
Phase 1 (認証+設定+API) ← Phase 0
Phase 2 (Account+Subscription Read + 出力) ← Phase 1
Phase 3 (Account+Subscription Write + Contact) ← Phase 2
Phase 4 (Order) ← Phase 2
Phase 5 (Commerce) ← Phase 2
Phase 6 (Invoice+Payment) ← Phase 2
Phase 7 (Usage+Meter) ← Phase 2
Phase 8 (Ramp+Commitment+Fulfillment) ← Phase 2
Phase 9 (ZOQL+ChangeLog+Omnichannel) ← Phase 2
Phase 10 (ポリッシュ+配布) ← Phase 2 以降いつでも可
```

Phase 3〜9 は Phase 2 完了後に任意順序で実装可能。

## Phase 進捗サマリ

| Phase | 名前 | ステータス | ファイル |
|-------|------|-----------|---------|
| 0 | プロジェクト基盤 | 未着手 | [phase-00-foundation.md](phase-00-foundation.md) |
| 1 | 認証 + 設定 + Raw API | 未着手 | [phase-01-auth-config-api.md](phase-01-auth-config-api.md) |
| 2 | 出力フォーマッタ + Account + Subscription (Read) | 未着手 | [phase-02-read-output.md](phase-02-read-output.md) |
| 3 | Account + Subscription (Write) + Contact | 未着手 | [phase-03-write-contact.md](phase-03-write-contact.md) |
| 4 | Order | 未着手 | [phase-04-order.md](phase-04-order.md) |
| 5 | Commerce (Product / Plan / Charge) | 未着手 | [phase-05-commerce.md](phase-05-commerce.md) |
| 6 | Invoice + Payment | 未着手 | [phase-06-invoice-payment.md](phase-06-invoice-payment.md) |
| 7 | Usage + Meter | 未着手 | [phase-07-usage-meter.md](phase-07-usage-meter.md) |
| 8 | Ramp + Commitment + Fulfillment + Prepaid | 未着手 | [phase-08-ramp-commit.md](phase-08-ramp-commit.md) |
| 9 | ZOQL + Subscription Change Log + Omnichannel | 未着手 | [phase-09-zoql-omni.md](phase-09-zoql-omni.md) |
| 10 | ポリッシュ + 配布 | 未着手 | [phase-10-polish.md](phase-10-polish.md) |

## 技術選定: Go

gh CLI と同じ言語・フレームワーク。小バイナリ (10-15MB)、高速起動 (<10ms)、Goreleaser + Homebrew 配布が容易。

## リポジトリ: matsuzj/zuora-cli

```
zuora-cli/
├── go.mod                        # module github.com/matsuzj/zuora-cli
├── go.sum
├── .goreleaser.yml
├── Taskfile.yml                  # task build, task test
├── .github/workflows/ci.yml
├── README.md
├── LICENSE
├── AGENTS.md
│
├── cmd/zr/
│   └── main.go                   # エントリポイント
│
├── pkg/
│   ├── cmd/
│   │   ├── root/root.go          # ルートコマンド + グローバルフラグ
│   │   ├── factory/factory.go    # DI (IOStreams, Config, HTTPClient)
│   │   ├── auth/                 # auth login/logout/status/token
│   │   ├── api/                  # Raw API コマンド
│   │   ├── config/               # config set/get/list/env
│   │   ├── account/              # account CRUD + summary
│   │   ├── subscription/         # subscription CRUD + lifecycle
│   │   ├── order/                # order CRUD + lifecycle
│   │   ├── product/              # product CRUD
│   │   ├── plan/                 # plan CRUD + charges
│   │   ├── invoice/              # invoice list/get/items/pdf/email
│   │   ├── payment/              # payment CRUD + apply/refund
│   │   ├── usage/                # usage CRUD + import
│   │   ├── contact/              # contact CRUD + transfer/scrub
│   │   ├── ramp/                 # ramp list/get/metrics
│   │   ├── commitment/           # commitment list/get/balance/schedule
│   │   ├── fulfillment/          # fulfillment CRUD
│   │   ├── fulfillment-item/     # fulfillment-item CRUD
│   │   ├── charge/               # charge CRUD (commerce API)
│   │   ├── rateplan/             # rateplan get (v1 API)
│   │   ├── meter/                # meter run/debug/status/summary/audit
│   │   ├── prepaid/              # prepaid rollover/deplete
│   │   ├── omnichannel/          # omnichannel subscription CRUD
│   │   ├── order-action/         # order-action update
│   │   ├── order-line-item/      # order-line-item CRUD + bulk
│   │   ├── query/                # ZOQL クエリ
│   │   ├── signup/               # sign-up
│   │   ├── alias/                # alias set/delete/list
│   │   ├── version/              # version
│   │   └── completion/           # completion
│   │
│   ├── iostreams/iostreams.go
│   └── output/                   # table / json+jq / template / csv
│
├── internal/
│   ├── auth/                     # OAuth 2.0 + keyring + 自動更新
│   ├── config/                   # YAML 設定 + マルチ環境
│   ├── api/                      # HTTP クライアント + ページネーション + エラー
│   ├── zoql/                     # ZOQL パーサ + エクスポート
│   └── build/                    # ビルド情報
│
└── test/                         # fixtures + mock server
```

## 認証設計

### OAuth 2.0 フロー

```
zr auth login
  1. Client ID / Client Secret 入力 (対話 or フラグ)
  2. 環境選択 (environments.yml のキー名から選択、またはカスタム URL 入力)
  3. POST /oauth/token (grant_type=client_credentials)
  4. Client ID/Secret → OS キーチェーン (macOS: Keychain, Linux: Secret Service 等の暗号化ストレージ)
     - デフォルト: キーチェーン等の OS 提供暗号化ストレージ「のみ」に保存を許可
     - キーチェーン非対応 OS では、Client Secret の永続保存機能はデフォルト無効
     - その場合は毎回入力 or 環境変数 `ZR_CLIENT_ID` / `ZR_CLIENT_SECRET` による運用を前提とする
     - どうしても平文ファイル保存を許可したい場合は、明示的な opt-in / 確認プロンプトを表示し、
       `~/.config/zr/credentials` (0600) への保存は「自己責任の危険なオプション」として扱う
  5. Access Token + 有効期限 → 設定ファイル
```

- 期限 60秒前に自動リフレッシュ / 401 応答時に1回リトライ
- 環境変数 `ZR_CLIENT_ID`, `ZR_CLIENT_SECRET`, `ZR_ENV` にも対応

### 設定ファイル (~/.config/zr/)

```yaml
# config.yml
active_environment: sandbox
zuora_version: "2025-08-12"
default_output: table

# tokens.yml — 環境ごとの認証状態 (自動管理、ユーザー編集不要)
tokens:
  sandbox:
    access_token: "Bearer ..."
    expires_at: "2026-03-24T12:00:00Z"
  us-production:
    access_token: "Bearer ..."
    expires_at: "2026-03-24T12:00:00Z"
# Client ID/Secret は OS キーチェーンに環境名をキーとして保存
# キーチェーン不在時は ~/.config/zr/credentials に保存 (後述)

# environments.yml — ユーザーが自由に環境を追加可能
environments:
  sandbox:
    base_url: "https://rest.apisandbox.zuora.com"
  us-production:
    base_url: "https://rest.na.zuora.com"
  us-production-cloud2:
    base_url: "https://rest.zuora.com"
  eu-production:
    base_url: "https://rest.eu.zuora.com"
  apac-production:
    base_url: "https://rest.ap.zuora.com"
```

### グローバルフラグ

```
--env, -e <name>      # 環境指定 (environments.yml のキー名: sandbox, us-production, eu-production 等)
--json                # JSON 出力
--jq <expr>           # jq フィルタ
--template <tmpl>     # Go テンプレート出力
--zuora-version <ver> # API バージョン指定
--verbose             # デバッグ出力
```

**出力フラグルール:**
- `--jq` は暗黙的に `--json` を有効化 (`--json` 単独指定は不要)
- 優先順位: `--jq` > `--json` > `--template` > テーブル (デフォルト)。`query` コマンドのみ追加で `--csv` があり、テーブルの前に入る (`--jq` > `--json` > `--template` > `--csv` > テーブル)
- `--json` と `--template` は排他 (両方指定時はエラー)
- ページャ (less/more) はテーブル出力時のみ有効 (JSON/template 時は無効)

## 主要依存パッケージ

| パッケージ | 用途 |
|-----------|------|
| `github.com/spf13/cobra` | CLI コマンドフレームワーク |
| `github.com/spf13/viper` | 設定管理 (YAML, 環境変数) |
| `github.com/zalando/go-keyring` | OS キーチェーン連携 (macOS: Keychain, Linux: Secret Service。不在時: 環境変数 or 毎回入力。opt-in で平文ファイル保存可) |
| `github.com/itchyny/gojq` | jq フィルタリング |
| `github.com/olekukonenko/tablewriter` | テーブル出力 |
| `golang.org/x/term` | ターミナル検出 |
| `github.com/stretchr/testify` | テストアサーション |

## コマンド実装パターン

新コマンドを追加する際は、以下のパターンに統一する (gh CLI の設計を踏襲)。

### ディレクトリ構成

```
pkg/cmd/<resource>/<action>/
├── <action>.go       # コマンド定義 + 実行ロジック
└── <action>_test.go  # テスト
```

**アクション名の正規化規則:**
- ディレクトリ名: ケバブケースをそのまま使用 (例: `payment-methods-default/`)
- Go package 名: ハイフンをアンダースコアに変換 (例: `package payment_methods_default`)
- ファイル名: ハイフンをアンダースコアに変換 (例: `payment_methods_default.go`)
- Cobra の `Use` フィールド: ケバブケースのまま (例: `Use: "payment-methods-default"`)

### コードテンプレート (例: `account list`)

```go
// pkg/cmd/account/list/list.go
package list

import (
    "github.com/matsuzj/zuora-cli/internal/api"
    "github.com/matsuzj/zuora-cli/pkg/cmd/factory"
    "github.com/matsuzj/zuora-cli/pkg/output"
    "github.com/spf13/cobra"
)

type Options struct {
    Factory  *factory.Factory
    PageSize int
    Cursor   string
    Filter   []string
}

func NewCmdList(f *factory.Factory) *cobra.Command {
    opts := &Options{Factory: f}

    cmd := &cobra.Command{
        Use:   "list",
        Short: "List accounts",
        Long:  "List all accounts in the current Zuora environment.",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runList(cmd, opts)
        },
    }

    cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Number of results per page")
    cmd.Flags().StringVar(&opts.Cursor, "cursor", "", "Pagination cursor for next page")
    cmd.Flags().StringSliceVar(&opts.Filter, "filter", nil, "Filter conditions (e.g. 'status=Active')")
    // --json, --jq, --template は root の PersistentFlags で定義済み
    // cmd.Flags().Lookup("json") 等で参照可能

    return cmd
}

func runList(cmd *cobra.Command, opts *Options) error {
    client, err := opts.Factory.HttpClient()
    if err != nil {
        return err
    }

    // API 呼び出し
    path := "/object-query/accounts"
    resp, err := client.Get(path, api.WithQuery("pageSize", opts.PageSize), api.WithQuery("cursor", opts.Cursor))
    if err != nil {
        return err
    }

    // グローバルフラグを root の PersistentFlags から取得
    jsonFlag, _ := cmd.Flags().GetBool("json")
    jqExpr, _ := cmd.Flags().GetString("jq")
    tmpl, _ := cmd.Flags().GetString("template")

    // 排他チェック: --json と --template の両方指定はエラー (root の PersistentPreRunE で実施)
    // ここでは root で検証済みの前提で分岐のみ行う

    // 出力フォーマット分岐
    if jsonFlag || jqExpr != "" {
        return output.PrintJSON(opts.Factory.IOStreams, resp, jqExpr)
    }
    if tmpl != "" {
        return output.PrintTemplate(opts.Factory.IOStreams, resp, tmpl)
    }
    return output.PrintTable(opts.Factory.IOStreams, resp, accountColumns)
}

// テーブル出力のカラム定義
var accountColumns = []output.Column{
    {Header: "ID", Field: "id"},
    {Header: "NAME", Field: "name"},
    {Header: "NUMBER", Field: "accountNumber"},
    {Header: "STATUS", Field: "status"},
    {Header: "BALANCE", Field: "balance"},
    {Header: "CREATED", Field: "createdDate"},
}
```

### 親コマンド登録パターン

```go
// pkg/cmd/account/account.go
package account

import (
    "github.com/matsuzj/zuora-cli/pkg/cmd/account/get"
    "github.com/matsuzj/zuora-cli/pkg/cmd/account/list"
    "github.com/matsuzj/zuora-cli/pkg/cmd/account/summary"
    "github.com/matsuzj/zuora-cli/pkg/cmd/factory"
    "github.com/spf13/cobra"
)

func NewCmdAccount(f *factory.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "account",
        Short: "Manage Zuora accounts",
    }
    cmd.AddCommand(list.NewCmdList(f))
    cmd.AddCommand(get.NewCmdGet(f))
    cmd.AddCommand(summary.NewCmdSummary(f))
    return cmd
}
```

### テストパターン

```go
// pkg/cmd/account/list/list_test.go
func TestListAccounts(t *testing.T) {
    // 1. テスト用 HTTP モックを設定
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/object-query/accounts", r.URL.Path)
        assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(fixtures.AccountListResponse)
    }))
    defer server.Close()

    // 2. テスト用 Factory 生成
    ios := iostreams.Test()
    f := factory.NewTestFactory(ios, server.URL, "test-token")

    // 3. root コマンド配下にぶら下げて実行 (PersistentFlags を有効化)
    rootCmd := root.NewCmdRoot(f)  // --json 等の PersistentFlags が登録される
    accountCmd := account.NewCmdAccount(f)
    rootCmd.AddCommand(accountCmd)
    rootCmd.SetArgs([]string{"account", "list", "--json"})
    err := rootCmd.Execute()

    // 4. アサーション
    assert.NoError(t, err)
    assert.Contains(t, ios.Out.String(), "A-001")
}
```

### Factory インターフェース

```go
// pkg/cmd/factory/factory.go
type Factory struct {
    IOStreams   *iostreams.IOStreams
    Config     func() (config.Config, error)
    HttpClient func() (*api.Client, error)
    AuthToken  func() (string, error)
}
```

### Write 操作コマンドのパターン (--body フラグ)

Write 操作 (create/update) は `--body` フラグで JSON を受け取る:

```
zr account create --body '{"name": "Test", ...}'
zr account create --body @account.json      # ファイルから読み込み
zr account create                           # 対話モード (TTY 時)
```

- `@` プレフィックスでファイルパス指定
- TTY 接続時は対話プロンプトで必須フィールドを入力
- `--body -` で stdin から読み込み

**Write 系コマンドの入力方式ルール:**
- **フルボディ操作** (create, update): `--body` で JSON 全体を渡す
- **単純操作** (cancel, suspend, resume, activate): 専用フラグ (`--effective-date`, `--policy` 等) で引数を受け取り、CLI がリクエストボディを組み立てる。`--body` も併用可能で、指定時は専用フラグより優先
- **ボディ必須の操作** (renew, revert, email, refund 等): `--body` 必須。API が複雑なボディを要求するため専用フラグでは不十分
- **原則**: 専用フラグがない操作は `--body` 必須

**ファイル出力ルール:**
- バイナリダウンロード (`invoice files --download` 等): `--output <path>` フラグ省略可。未指定時はデフォルトファイル名 (`invoice-{id}.pdf`) でカレントディレクトリに保存。既存ファイルがある場合はエラー、`--force` で上書き
- エクスポート (`query --export` 等): `--export <path>` でファイル出力。既存ファイルがある場合はエラー、`--force` で上書き
- stdout にはバイナリを流さない (TTY 検出時にエラー)

## エラーハンドリング仕様

### Zuora API エラー応答の表示

```
Error: Zuora API error (HTTP 400)
  Code: INVALID_VALUE
  Message: The account key 'XXX' is invalid.
```

### 認証エラー

```
Error: Authentication failed. Run 'zr auth login' to authenticate.
```

### 接続エラー

```
Error: Could not connect to https://rest.apisandbox.zuora.com
  Check your network connection and environment settings.
```

### 終了コード

| コード | 意味 |
|--------|------|
| 0 | 成功 |
| 1 | 一般エラー |
| 2 | 認証エラー |
| 3 | API エラー (4xx) |
| 4 | サーバーエラー (5xx) |

## テスト戦略

- **Unit**: 各コマンドに `*_test.go`、HTTP モックで API 応答を再現
- **Integration**: `//go:build integration` タグ、Zuora Sandbox で実行
- **CI**: `go vet` + `staticcheck` + `go test -race`

## 検証方法

1. `task build` でバイナリビルド → `./bin/zr version`
2. `zr auth login` → Sandbox 認証テスト
3. `zr auth status` → トークン状態確認
4. `zr api /object-query/accounts` → Raw API 動作確認 (GET はデフォルト、`-X POST` で変更)
5. `zr account list --json` → フォーマット出力確認
6. `zr subscription list --account <key>` → サブスクリプション参照
7. `go test ./...` で全テスト通過

## 参照資料一覧

### Zuora API リファレンス

| 資料 | URL | 用途 |
|------|-----|------|
| Zuora v1 API Reference | https://developer.zuora.com/v1-api-reference/api.md | 全エンドポイント定義・リクエスト/レスポンス仕様 |
| Zuora Developer Portal | https://developer.zuora.com/ | API 概要・ガイド・チュートリアル |
| Zuora OpenAPI Spec (OTC) | `zuora-openapi-for-otc.yaml` (developer.zuora.com 内) | OpenAPI 3.0 スキーマ定義 |
| Zuora OpenAPI Spec (Full) | `zuora-openapi-full-compact.yaml` (developer.zuora.com 内) | 全 API の OpenAPI スキーマ |
| Older API (Transactions) | https://developer.zuora.com/v1-api-reference/older-api/transactions/ | Invoice/Payment の account 別一覧 |

### Zuora API Operation ページ (レビュー検証済み)

各コマンド実装時に参照すべき個別 operation ドキュメント:

| リソース | Operation | URL |
|----------|-----------|-----|
| Account | List (object-query, cursor+filter) | https://developer.zuora.com/v1-api-reference/api/operation/queryAccounts/ |
| Subscription | Cancel | https://developer.zuora.com/v1-api-reference/api/operation/PUT_CancelSubscription/ |
| Subscription | Suspend | https://developer.zuora.com/v1-api-reference/api/operation/PUT_SuspendSubscription/ |
| Subscription | Resume | https://developer.zuora.com/v1-api-reference/api/operation/PUT_ResumeSubscription/ |
| Subscription | Renew | https://developer.zuora.com/v1-api-reference/api/operation/PUT_RenewSubscription/ |
| Subscription | By Account | https://developer.zuora.com/v1-api-reference/api/operation/GET_SubscriptionsByAccount/ |
| Subscription | Metrics | https://developer.zuora.com/v1-api-reference/api/operation/GetMetricsBySubscriptionNumbers/ |
| Order | List All | https://developer.zuora.com/v1-api-reference/api/operation/GET_AllOrders/ |
| Order | Update (full payload) | https://developer.zuora.com/v1-api-reference/api/operation/PUT_Order/ |
| Order | Revert | https://developer.zuora.com/v1-api-reference/api/operation/revertOrder/ |
| Order | Trigger Dates | https://developer.zuora.com/v1-api-reference/api/operation/PUT_OrderTriggerDates/ |
| Order Line Item | Bulk Update | https://developer.zuora.com/v1-api-reference/api/operation/Post_OrderLineItems/ |
| Commerce | Product Get | https://developer.zuora.com/v1-api-reference/api/operation/GET_RetrieveProductByKey/ |
| Commerce | Plan Query | https://developer.zuora.com/v1-api-reference/api/operation/queryCommerceProductRatePlans/ |
| Commerce | Plan List | https://developer.zuora.com/v1-api-reference/api/operation/queryCommercePlansList/ |
| Commerce | Charge Query | https://developer.zuora.com/v1-api-reference/api/operation/queryProductRatePlanChargeWithDynamicPricing/ |
| Commerce | Purchase Options | https://developer.zuora.com/v1-api-reference/api/operation/queryPurchaseOptionsbyPRPID/ |
| Commerce | Legacy Products | https://developer.zuora.com/v1-api-reference/api/operation/queryLegacyProducts/ |
| Invoice | Files | https://developer.zuora.com/v1-api-reference/api/operation/GET_InvoiceFiles/ |
| Invoice | Email | https://developer.zuora.com/v1-api-reference/api/operation/POST_EmailInvoice/ |
| Invoice | By Account (Older) | https://developer.zuora.com/v1-api-reference/older-api/transactions/get_transactioninvoice/ |
| Payment | Apply | https://developer.zuora.com/v1-api-reference/api/operation/PUT_ApplyPayment/ |
| Payment | Refund | https://developer.zuora.com/v1-api-reference/api/operation/POST_RefundPayment/ |
| Payment | By Account (Older) | https://developer.zuora.com/v1-api-reference/older-api/transactions/get_transactionpayment/ |
| Usage | Post (CSV multipart) | https://developer.zuora.com/v1-api-reference/api/operation/POST_Usage/ |
| Meter | Summary | https://developer.zuora.com/v1-api-reference/api/operation/retrieveMeterSummaryData/ |
| Meter | Audit Trail | https://developer.zuora.com/v1-api-reference/api/operation/getAuditTrailEntriesForMeter/ |
| Commitment | List | https://developer.zuora.com/v1-api-reference/api/operation/getCommitments/ |
| Commitment | Periods | https://developer.zuora.com/v1-api-reference/api/operation/getCommitmentPeriods/ |
| Contact | Snapshot | https://developer.zuora.com/v1-api-reference/api/operation/GET_ContactSnapshot/ |

### GitHub CLI (設計参考)

| 資料 | URL | 用途 |
|------|-----|------|
| GitHub CLI 公式サイト | https://cli.github.com/ | CLI コマンド体系・設計思想 |
| GitHub CLI マニュアル | https://cli.github.com/manual/ | 全コマンドリファレンス |
| GitHub CLI ソースコード | https://github.com/cli/cli | アーキテクチャ・ディレクトリ構造 |
| go-gh ライブラリ | https://pkg.go.dev/github.com/cli/go-gh/v2 | API クライアント・IOStreams パターン |
| go-gh API パッケージ | https://pkg.go.dev/github.com/cli/go-gh/v2/pkg/api | REST/GraphQL クライアント設計 |
| gh 出力フォーマット | https://cli.github.com/manual/gh_help_formatting | JSON/jq/テンプレート出力仕様 |
| gh 認証コマンド | https://cli.github.com/manual/gh_auth_login | OAuth 認証フロー設計 |

### Go ライブラリ (依存パッケージ)

| 資料 | URL | 用途 |
|------|-----|------|
| Cobra CLI フレームワーク | https://github.com/spf13/cobra | コマンド定義・フラグ管理 |
| Viper 設定管理 | https://github.com/spf13/viper | YAML 設定・環境変数 |
| go-keyring | https://github.com/zalando/go-keyring | OS キーチェーン連携 |
| gojq | https://github.com/itchyny/gojq | jq フィルタリング |
| Goreleaser | https://goreleaser.com/ | バイナリビルド・リリース自動化 |
| testify | https://github.com/stretchr/testify | テストアサーション |

### Zuora API 認証

| 項目 | 詳細 |
|------|------|
| OAuth エンドポイント | `POST /oauth/token` (grant_type=client_credentials) |
| 認証ヘッダ | `Authorization: Bearer <access_token>` |
| API バージョンヘッダ | `Zuora-Version: 2025-08-12` |
| レート制限 | IP アドレス単位 |

### Zuora リージョン別エンドポイント

| リージョン | Base URL |
|-----------|----------|
| US Sandbox | `https://rest.apisandbox.zuora.com` |
| US Production (Cloud 1) | `https://rest.na.zuora.com` |
| US Production (Cloud 2) | `https://rest.zuora.com` |
| EU Production | `https://rest.eu.zuora.com` |
| APAC Production | `https://rest.ap.zuora.com` |
