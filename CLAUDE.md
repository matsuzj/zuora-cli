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
