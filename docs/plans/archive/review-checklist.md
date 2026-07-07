> **歴史的文書(2026-07-08 アーカイブ)**: 「ship パイプライン Phase 2/6」は
> 廃止済みのユーザープロファイル側スキルで、現存しない。E2E 品質の 4 観点は
> AGENTS.md の Testing 節に併合済み。レビュー観点の現在の正:
> AGENTS.md(規約)+ [docs/plans/README.md](../README.md)(計画テンプレート
> とレビュー・プロトコル)。

# レビューチェックリスト

ship パイプラインの Phase 2 / Phase 6 レビューラウンドで確認すべき観点。

## リファレンス

### Zuora REST API

各コマンドの実装が API 仕様と一致しているか確認する。

| 確認項目 | 参照先 |
|----------|--------|
| エンドポイント URL | `docs/zuora-api-reference.md` (ローカル) |
| リクエスト/レスポンス仕様 | https://developer.zuora.com/v1-api-reference/api.md |
| 個別 operation ドキュメント | `docs/plans/archive/README.md`(歴史的文書)の「Zuora API Operation ページ」セクション |

**チェック項目:**
- HTTP メソッド・パスが API 仕様と一致しているか
- レスポンスの JSON 構造（ネストキー含む）を正しく解析しているか
- The Zuora success-flag check is on by default in the API client; flag any `WithoutCheckSuccess()` outside the raw `zr api` GET/HEAD passthrough.
- 必須パラメータ・フラグがプランの仕様表と一致しているか

### GitHub CLI パターン

gh CLI のコード規約に準拠しているか確認する。

| 確認項目 | 参照先 |
|----------|--------|
| ディレクトリ構造 | https://github.com/cli/cli の `pkg/cmd/` |
| コマンド設計 | `docs/architecture.md` の「コマンドの書き方(正準)」 |

**チェック項目:**
- `cobra.ExactArgs(N)` / `cobra.NoArgs` による引数バリデーション
- `--body` フラグは `cmdutil.ResolveBody()` 経由（`-`, `@file`, リテラル JSON）
- `--confirm` フラグによる破壊的操作の安全ガード
- `url.PathEscape()` によるパスパラメータのエスケープ
- `output.FromCmd(cmd)` → `output.Render()` / `output.RenderDetail()` の統一パターン
- 成功メッセージは `f.IOStreams.ErrOut` に出力（stdout はデータ用）

## コード品質

- 各パッケージにユニットテスト (`_test.go`) があるか
- `go build ./...` && `go vet ./...` && `go test -race ./...` が通るか
- 既存テストに回帰がないか

## E2E テスト品質（Phase 6 のみ）

- テストデータはテスト内で作成し、既存データに依存しないか
- テナント固有の制約（Orders 有効/無効等）を考慮して `pass` / `skip` を使い分けているか
- バリデーションテスト（必須フラグ/引数の欠落）を含んでいるか
- `--json` / `--jq` の出力フォーマットテストを含んでいるか
