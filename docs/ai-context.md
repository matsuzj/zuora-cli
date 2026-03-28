# AI Context — zuora-cli (zr)

> このファイルは AI エージェント（Claude Code / Codex CLI）が zuora-cli の実装パターンを即座に理解するためのコンテキストです。
> オーケストレーターが Plan / Implement ステージのプロンプトに埋め込みます。

## プロジェクト概要

- **CLI 名**: `zr` (Zuora CLI)
- **設計**: GitHub CLI (`gh`) のパターンに準拠
- **Go**: 1.26.1 (`go.mod`)
- **認証**: Zuora OAuth 2.0 クライアントクレデンシャルフロー
- **環境変数プレフィックス**: `ZUORA_`

## ディレクトリ構造

```
cmd/zr/main.go                  ← エントリポイント（最小限）
pkg/cmd/<resource>/<action>/    ← コマンド実装
pkg/cmd/<resource>/<resource>.go ← 親コマンド（AddCommand）
pkg/cmd/root/root.go            ← ルートコマンド（グローバルフラグ、全コマンド登録）
pkg/cmd/factory/factory.go      ← DI コンテナ（Factory）
pkg/output/                     ← 出力フォーマット（JSON/table/template/jq）
pkg/iostreams/                  ← I/O 抽象化
pkg/cmdutil/                    ← コマンドユーティリティ
internal/api/client.go          ← HTTP クライアント
internal/auth/                  ← 認証
internal/config/                ← 設定管理
internal/build/build.go         ← ビルドメタデータ（Version, Commit, Date）
```

## コマンド実装パターン

### 単一リソース取得 (GET)

```go
// pkg/cmd/<resource>/get/get.go
package get

import (
    "encoding/json"
    "fmt"
    "net/url"

    "github.com/matsuzj/zuora-cli/internal/api"
    "github.com/matsuzj/zuora-cli/pkg/cmd/factory"
    "github.com/matsuzj/zuora-cli/pkg/output"
    "github.com/spf13/cobra"
)

// NewCmdGet は get サブコマンドを生成する。
func NewCmdGet(f *factory.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "get <key>",
        Short: "Get resource details",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return runGet(cmd, f, args[0])
        },
    }
    return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, key string) error {
    // 1. HTTP クライアント取得
    client, err := f.HttpClient()
    if err != nil {
        return err
    }

    // 2. API 呼び出し
    resp, err := client.Get(fmt.Sprintf("/v1/resource/%s", url.PathEscape(key)))
    if err != nil {
        return err
    }

    // 3. レスポンスパース
    fmtOpts := output.FromCmd(cmd)
    var raw map[string]interface{}
    if err := json.Unmarshal(resp.Body, &raw); err != nil {
        return fmt.Errorf("parsing response: %w", err)
    }

    // 4. フィールド抽出 → DetailField スライス
    fields := []output.DetailField{
        {Key: "ID", Value: getString(raw, "id")},
        {Key: "Name", Value: getString(raw, "name")},
        {Key: "Status", Value: getString(raw, "status")},
    }

    // 5. 出力（--json / --jq / --template / テーブル を自動切替）
    return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

// --- 型安全ヘルパー ---
func getString(m map[string]interface{}, key string) string {
    if v, ok := m[key]; ok && v != nil {
        return fmt.Sprintf("%v", v)
    }
    return ""
}
```

### リスト取得 (LIST)

```go
// pkg/cmd/<resource>/list/list.go
package list

import (
    "encoding/json"
    "fmt"
    "strconv"

    "github.com/matsuzj/zuora-cli/internal/api"
    "github.com/matsuzj/zuora-cli/pkg/cmd/factory"
    "github.com/matsuzj/zuora-cli/pkg/output"
    "github.com/spf13/cobra"
)

type listOptions struct {
    Factory  *factory.Factory
    PageSize int
    Cursor   string
}

func NewCmdList(f *factory.Factory) *cobra.Command {
    opts := &listOptions{Factory: f}
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List resources",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runList(cmd, opts)
        },
    }
    cmd.Flags().IntVar(&opts.PageSize, "page-size", 20, "Number of results per page")
    cmd.Flags().StringVar(&opts.Cursor, "cursor", "", "Pagination cursor")
    return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
    client, err := opts.Factory.HttpClient()
    if err != nil {
        return err
    }

    // クエリパラメータ構築
    var reqOpts []api.RequestOption
    reqOpts = append(reqOpts, api.WithQuery("pageSize", strconv.Itoa(opts.PageSize)))
    if opts.Cursor != "" {
        reqOpts = append(reqOpts, api.WithQuery("cursor", opts.Cursor))
    }

    resp, err := client.Get("/object-query/resources", reqOpts...)
    if err != nil {
        return err
    }

    fmtOpts := output.FromCmd(cmd)

    // 型付き構造体でパース
    var body struct {
        Data []struct {
            ID     string `json:"id"`
            Name   string `json:"name"`
            Status string `json:"status"`
        } `json:"data"`
        NextPage string `json:"nextPage"`
    }
    if err := json.Unmarshal(resp.Body, &body); err != nil {
        return fmt.Errorf("parsing response: %w", err)
    }

    // テーブルカラム定義
    cols := []output.Column{
        {Header: "ID", Field: "id"},
        {Header: "NAME", Field: "name"},
        {Header: "STATUS", Field: "status"},
    }

    rows := make([][]string, len(body.Data))
    for i, item := range body.Data {
        rows[i] = []string{item.ID, item.Name, item.Status}
    }

    if err := output.Render(opts.Factory.IOStreams, resp.Body, fmtOpts, rows, cols); err != nil {
        return err
    }

    // ページネーションヒント
    if body.NextPage != "" && !fmtOpts.JSON && fmtOpts.JQ == "" && fmtOpts.Template == "" {
        fmt.Fprintf(opts.Factory.IOStreams.ErrOut, "\nMore results: zr resource list --cursor %q\n", body.NextPage)
    }

    return nil
}
```

### POST / 作成

```go
// POST の場合: --body フラグでリクエストボディを受け取る
func runCreate(cmd *cobra.Command, f *factory.Factory, bodyFlag string) error {
    client, err := f.HttpClient()
    if err != nil {
        return err
    }

    bodyReader, err := cmdutil.ResolveBody(bodyFlag, f.IOStreams.In)
    if err != nil {
        return err
    }

    resp, err := client.Post("/v1/resources", bodyReader, api.WithCheckSuccess())
    if err != nil {
        return err
    }

    // ... 出力
}
```

## テストパターン

```go
// pkg/cmd/<resource>/get/get_test.go
package get

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/matsuzj/zuora-cli/internal/config"
    "github.com/matsuzj/zuora-cli/pkg/cmd/factory"
    "github.com/matsuzj/zuora-cli/pkg/iostreams"
    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// テスト用ルートコマンド（グローバルフラグ含む）
func newTestRoot(f *factory.Factory) *cobra.Command {
    root := &cobra.Command{Use: "zr"}
    root.PersistentFlags().Bool("json", false, "")
    root.PersistentFlags().String("jq", "", "")
    root.PersistentFlags().String("template", "", "")

    parent := &cobra.Command{Use: "resource"}
    parent.AddCommand(NewCmdGet(f))
    root.AddCommand(parent)
    return root
}

func TestGet_Detail(t *testing.T) {
    // 1. HTTPモックサーバー
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "/v1/resource/R001", r.URL.Path)
        assert.Equal(t, http.MethodGet, r.Method)
        w.WriteHeader(200)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id":     "abc123",
            "name":   "Test Resource",
            "status": "Active",
        })
    }))
    defer server.Close()

    // 2. テスト用 Factory
    ios, _, out, _ := iostreams.Test()
    cfg := config.NewMockConfig()
    f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

    // 3. コマンド実行
    root := newTestRoot(f)
    root.SetArgs([]string{"resource", "get", "R001"})
    err := root.Execute()

    // 4. アサーション
    require.NoError(t, err)
    output := out.String()
    assert.Contains(t, output, "Test Resource")
    assert.Contains(t, output, "Active")
}

func TestGet_JSON(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id": "abc123", "name": "Test",
        })
    }))
    defer server.Close()

    ios, _, out, _ := iostreams.Test()
    cfg := config.NewMockConfig()
    f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

    root := newTestRoot(f)
    root.SetArgs([]string{"resource", "get", "R001", "--json"})
    err := root.Execute()

    require.NoError(t, err)
    assert.Contains(t, out.String(), `"id"`)
}

func TestGet_Error(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(404)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "reasons": []map[string]interface{}{
                {"code": 50000020, "message": "not found"},
            },
        })
    }))
    defer server.Close()

    ios, _, _, _ := iostreams.Test()
    cfg := config.NewMockConfig()
    f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

    root := newTestRoot(f)
    root.SetArgs([]string{"resource", "get", "NOTFOUND"})
    err := root.Execute()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not found")
}
```

## 親コマンドパターン

```go
// pkg/cmd/<resource>/<resource>.go
package resource

import (
    "github.com/matsuzj/zuora-cli/pkg/cmd/<resource>/get"
    "github.com/matsuzj/zuora-cli/pkg/cmd/<resource>/list"
    "github.com/matsuzj/zuora-cli/pkg/cmd/factory"
    "github.com/spf13/cobra"
)

func NewCmdResource(f *factory.Factory) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "resource <command>",
        Short: "Manage resources",
    }
    cmd.AddCommand(get.NewCmdGet(f))
    cmd.AddCommand(list.NewCmdList(f))
    return cmd
}
```

## ルートコマンド登録

新しいリソースコマンドを追加する場合、`pkg/cmd/root/root.go` に以下を追加:

```go
import resourcecmd "github.com/matsuzj/zuora-cli/pkg/cmd/resource"

// NewCmdRoot 内:
cmd.AddCommand(resourcecmd.NewCmdResource(f))
```

## 出力フォーマットルール

| フラグ | 動作 | 関数 |
|-------|------|------|
| (なし) | テーブル表示 | `output.Render` (list) / `output.RenderDetail` (get) |
| `--json` | JSON pretty-print | `output.PrintJSON` |
| `--jq <expr>` | jq フィルタ (JSON暗黙有効) | `output.PrintJSON` with jq |
| `--template <tmpl>` | Go テンプレート | `output.PrintTemplate` |

- `--json` と `--template` は排他（root の PersistentPreRunE でチェック）
- `--jq` は `--json` を暗黙的に有効化
- `fmtOpts := output.FromCmd(cmd)` で取得

## Factory パターン

```go
type Factory struct {
    IOStreams   *iostreams.IOStreams
    Config     func() (config.Config, error)    // 遅延評価 + sync.Once キャッシュ
    HttpClient func() (*api.Client, error)       // 遅延評価
    AuthToken  func() (string, error)            // 遅延評価
}
```

- **なぜ関数フィールドか**: root の PersistentPreRunE でミドルウェアとしてラップできる（`--env`, `--verbose`, `--read-only` 等）
- **テスト**: `factory.NewTestFactory(ios, cfg, serverURL, token)` で全依存を注入

## HTTP クライアント

```go
// GET
resp, err := client.Get("/v1/accounts/A001")

// GET with query params
resp, err := client.Get("/object-query/accounts",
    api.WithQuery("pageSize", "20"),
    api.WithQuery("cursor", nextCursor),
)

// POST with body
resp, err := client.Post("/v1/accounts", bodyReader, api.WithCheckSuccess())

// PUT
resp, err := client.Put("/v1/accounts/A001", bodyReader, api.WithCheckSuccess())

// DELETE
resp, err := client.Delete("/v1/accounts/A001")
```

- `api.WithCheckSuccess()`: Zuora の `{"success": false}` レスポンスをエラーとして扱う（POST/PUT で使用）
- `api.WithQuery(key, value)`: クエリパラメータ追加
- `api.WithHeader(key, value)`: カスタムヘッダー追加
- レスポンス: `resp.Body` ([]byte), `resp.StatusCode` (int)
- 400+ ステータスは自動的にエラー変換される

## よくある間違いと正しいやり方

| 間違い | 正しい |
|--------|--------|
| `os.Stdout` に直接書く | `f.IOStreams.Out` を使う |
| エラーを `log.Fatal` で処理 | `return fmt.Errorf("context: %w", err)` で返す |
| JSON を `fmt.Println` で出力 | `output.RenderDetail` / `output.Render` を使う |
| テストで実 API を呼ぶ | `httptest.NewServer` でモック |
| フラグをグローバル変数に保持 | `listOptions` 構造体に格納 |
| `--json` を自前実装 | `output.FromCmd(cmd)` + `output.Render*` で自動対応 |
| import を `_` で未使用にする | 不要な import は削除する |

## 規約（AGENTS.md 準拠）

- **コミット**: Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `test:`, `refactor:`)
- **ブランチ**: `feature/`, `fix/`, `docs/`, `chore/`
- **フォーマット**: `gofmt` 必須
- **テスト**: `-race` フラグ必須、`testify/assert` 使用
- **エクスポート関数**: doc コメント必須
