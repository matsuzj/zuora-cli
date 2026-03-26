# Read-Only Mode for zuora-cli

## Context

zuora-cli は Claude Code のスキルから自動実行されることを想定している。誤って書き込み API (POST/PUT/DELETE/PATCH) を呼ぶリスクを防ぐため、読み取り専用モードを追加する。

## レビュー指摘と設計変更

Codex レビューで P1 が検出された: Commerce API (`charge get`, `plan list`, `plan get`, `plan purchase-options`, `product list-legacy`) や ZOQL query は HTTP POST を使う読み取りコマンドであり、HTTP メソッドだけで判定すると正当な読み取りコマンドがブロックされる。

**対策**: HTTP メソッドベースの allowlist を拡充し、POST を使う全ての読み取りエンドポイントを含める。パスは `strings.ToLower` で正規化してマッチング。

## 方針

- `--read-only` グローバルフラグ **または** `ZR_READ_ONLY=true` 環境変数で有効化
- API クライアントの `Do()` メソッドで HTTP メソッド + パスをチェックし、書き込み操作をブロック
- GET/HEAD/OPTIONS は常に許可
- POST は allowlist に含まれるパスのみ許可（下記参照）
- PUT/DELETE/PATCH は常にブロック
- 終了コード 5 で read-only 違反を報告 (1=general, 2=auth, 3=4xx, 4=5xx, 5=read-only)

### POST Allowlist (読み取り POST エンドポイント)

```go
var readOnlyPOSTAllowList = []string{
    // ZOQL query
    "v1/action/query",
    "v1/action/querymore",
    // Commerce API query/list (POST だが読み取り専用)
    "commerce/charges/query",
    "commerce/plans/query",
    "commerce/plans/list",
    "commerce/purchase-options/list",
    "commerce/legacy/products/list",
    // Preview (データ変更なし、シミュレーション)
    "v1/orders/preview",
    "v1/async/orders/preview",
    "v1/subscriptions/preview",
    // subscription preview-change: パスに動的セグメントあり → regex match (下記参照)
}

// regex match (動的パスセグメント対応)
var readOnlyPOSTPatterns = []*regexp.Regexp{
    regexp.MustCompile(`^v1/subscriptions/[^/]+/preview$`),  // preview-change
    regexp.MustCompile(`^meters/[^/]+/summary$`),            // meter summary (読み取り)
}
```

**パス正規化** (`extractPath` ヘルパー):
1. 絶対 URL (`https://...`) の場合: `url.Parse` でパスを抽出
2. 相対パスの場合: そのまま使用
3. 先頭 `/` を除去、小文字化
4. クエリパラメータを除去

```go
func extractPath(rawPath string) string {
    p := rawPath
    if strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
        if u, err := url.Parse(p); err == nil {
            p = u.Path
        }
    }
    if idx := strings.Index(p, "?"); idx >= 0 {
        p = p[:idx]
    }
    return strings.ToLower(strings.TrimLeft(p, "/"))
}
```

allowlist マッチング:
- 完全一致: `readOnlyPOSTAllowList` に含まれるか
- regex 一致: `readOnlyPOSTPatterns` のいずれかにマッチするか（動的パスセグメント対応）

meter debug はブロック対象（メーター実行操作のため）。meter summary は読み取り（集計データ取得）のため regex で許可。

## 変更ファイル

### 1. `internal/api/errors.go` — ReadOnlyError 追加

```go
type ReadOnlyError struct{}
func (e *ReadOnlyError) Error() string {
    return "blocked: write operation not allowed in read-only mode. Remove --read-only flag or unset ZR_READ_ONLY to enable write operations"
}
func (e *ReadOnlyError) ExitCode() int { return 5 }  // 1=general, 2=auth, 3=4xx, 4=5xx, 5=read-only
```

### 2. `internal/api/client.go` — コア実装

- `Client` 構造体に `readOnly bool` フィールド追加
- `WithReadOnly() ClientOption` 追加
- `SetReadOnly(bool)` セッター追加
- `Do()` の先頭にガード追加: `if c.readOnly && !isReadOnlyAllowed(method, path)`
- `isReadOnlyAllowed()` ヘルパー: GET/HEAD/OPTIONS は許可、POST は allowlist (完全一致 + regex 一致) のみ許可、PUT/DELETE/PATCH はブロック
- `extractPath()` ヘルパー: 絶対 URL・クエリパラメータ・大文字小文字を正規化してからマッチング

### 3. `pkg/cmd/root/root.go` — フラグ + 環境変数

- `--read-only` PersistentFlag 追加
- PersistentPreRunE で `--read-only` フラグまたは `ZR_READ_ONLY=true` を検出
- `--verbose` と同じパターンで `f.HttpClient` をラップし `client.SetReadOnly(true)`
- `"os"` を import に追加

### 4. `pkg/cmd/factory/testing.go` — テスト用ヘルパー

- `NewTestFactoryReadOnly()` 追加

### 5. `internal/api/client_test.go` — ユニットテスト

- POST/PUT/DELETE/PATCH がブロックされること
- GET が許可されること
- ZOQL query/queryMore POST が許可されること
- Commerce API 読み取り POST が許可されること (charge query, plan list 等)
- subscription preview-change (regex match `v1/subscriptions/[^/]+/preview`) が許可されること
- meter summary (regex match `meters/[^/]+/summary`) が許可されること
- 絶対 URL (https://...) が正しく正規化されること
- クエリパラメータ付きパスが正しく正規化されること
- `SetReadOnly()` が動作すること
- ReadOnlyError の ExitCode が 5 であること

### 6. `pkg/cmd/root/root_test.go` — コマンドレベルテスト

- `--read-only` フラグが登録されていること
- `ZR_READ_ONLY=true` で HttpClient に readOnly が設定されること
- コマンドレベル統合テスト: `--read-only` 付きで write コマンド (account create) を実行し ReadOnlyError が返ること

## 変更しないファイル

- `cmd/zr/main.go` — `--read-only` はブール型フラグなので alias 展開のスキップリストに追加不要
- auth 関連 — OAuth トークン取得は独自の `http.Client` を使うため影響なし

## 検証方法

```bash
# ビルド
go build -o bin/zr ./cmd/zr/

# ユニットテスト
go test -race ./internal/api/...

# 全テスト
go test -race ./...

# E2E: 読み取りが通ること
./bin/zr --read-only account list
./bin/zr --read-only plan list --body '{}'
./bin/zr --read-only charge get --key "test"

# E2E: 書き込みがブロックされること
./bin/zr --read-only account create --body '{}'
# → "blocked: write operation not allowed in read-only mode..."

# E2E: ZOQL が通ること
./bin/zr --read-only query "SELECT Id FROM Account LIMIT 1"

# E2E: 環境変数
ZR_READ_ONLY=true ./bin/zr account create --body '{}'
# → blocked

# E2E: zr api
./bin/zr --read-only api /v1/accounts  # → OK (GET)
./bin/zr --read-only api -X POST /v1/orders --body '{}'  # → blocked
./bin/zr --read-only api -X POST /v1/action/query --body '{"queryString":"SELECT Id FROM Account"}' # → OK (allowlisted)
```
