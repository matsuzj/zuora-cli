# Phase 1: 認証 + 設定 + Raw API

**依存**: Phase 0

## 実装チェックリスト

### 認証基盤 (internal/auth/)

- [ ] OAuth 2.0 クライアント認証 (`POST /oauth/token`)
- [ ] OS キーチェーン保存/取得 (go-keyring)
- [ ] トークン自動更新 (期限前 + 401 リトライ)
- [ ] 環境変数認証 (`ZR_CLIENT_ID`, `ZR_CLIENT_SECRET`)

### 設定基盤 (internal/config/)

- [ ] YAML 設定ファイル管理 (`~/.config/zr/`)
- [ ] マルチ環境管理 (sandbox/production/custom)
- [ ] リージョン別エンドポイント定義

### HTTP クライアント (internal/api/)

- [ ] Bearer トークン自動注入
- [ ] Zuora-Version ヘッダ管理
- [ ] リトライ (指数バックオフ)
- [ ] ページネーション処理
- [ ] エラー応答パース
- [ ] レート制限ハンドリング
- [ ] `--verbose` デバッグログ

### 認証コマンド

- [ ] `zr auth login` — OAuth 対話認証 + フラグ認証
- [ ] `zr auth logout` — 認証情報削除
- [ ] `zr auth status` — 認証状態・トークン期限表示
- [ ] `zr auth token` — アクセストークン出力 (スクリプト用)

### 設定コマンド

- [ ] `zr config set <key> <value>`
- [ ] `zr config get <key>`
- [ ] `zr config list`
- [ ] `zr config env <name>` — アクティブ環境切替

### Raw API コマンド

- [ ] `zr api <path>` — 任意の API 呼び出し
- [ ] `-X, --method` — HTTP メソッド指定
- [ ] `-b, --body` — リクエストボディ (JSON 文字列 or @ファイル)
- [ ] `-H, --header` — 追加ヘッダ (複数可)
- [ ] `--paginate` — 自動ページネーション
- [ ] `--jq` — jq フィルタ

### Sandbox 検証

- [ ] 認証フロー動作確認
- [ ] `zr api /object-query/accounts` 動作確認 (GET はデフォルト)

## コマンド詳細仕様

### api コマンド (Raw API)

```
zr api [path] [flags]

Examples:
  zr api /v1/accounts                          # GET (デフォルト)
  zr api -X POST /v1/orders --body @order.json # POST with file body
  zr api /v1/accounts --jq '.accounts[].name'  # jq filter
  zr api /v1/accounts --paginate               # 全ページ取得

Flags:
  -X, --method string     HTTP method (GET, POST, PUT, DELETE, PATCH) (default "GET")
  -b, --body string       Request body (JSON string, @file, or - for stdin)
  -H, --header strings    Additional headers (key:value), repeatable
      --paginate           Fetch all pages automatically
      --jq string          Filter JSON output with jq expression
      --zuora-version string  Override Zuora-Version header
```
