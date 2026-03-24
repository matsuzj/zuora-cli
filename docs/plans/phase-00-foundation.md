---
title: "Phase 0: プロジェクト基盤"
status: not-started
depends_on: []
---

# Phase 0: プロジェクト基盤

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
