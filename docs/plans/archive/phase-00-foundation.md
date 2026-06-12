> **歴史的文書(2026-06-13 アーカイブ)**: 初期開発(2026-03)のプラン。進捗表・
> チェックリストは更新されないまま全コマンドが出荷済みで、現状を反映していない。
> 現在の正: 構造は [docs/architecture.md](../../architecture.md)、進行管理は
> [docs/refactoring-plan.md](../../refactoring-plan.md)。

# Phase 0: プロジェクト基盤

**依存**: なし (全 Phase の前提)

## 実装チェックリスト

- [ ] `matsuzj/zuora-cli` GitHub リポジトリ作成
- [ ] `go mod init github.com/matsuzj/zuora-cli`
- [ ] Cobra ルートコマンド (`pkg/cmd/root/`)
- [ ] Factory パターン (`pkg/cmd/factory/`)
- [ ] IOStreams (`pkg/iostreams/`)
- [ ] ビルド情報 (`internal/build/`)
- [ ] `zr version`
- [ ] `zr completion` (bash/zsh/fish)
- [ ] Taskfile.yml (`task build`, `task test`, `task lint`)
- [ ] GitHub Actions CI (`go vet`, `staticcheck`, `go test -race`)
- [ ] AGENTS.md (内容: ブランチ命名 feature/fix/docs/chore、Conventional Commits、main ブランチ保護、Go コード規約: gofmt, staticcheck, golangci-lint、テスト必須)
- [ ] README.md
