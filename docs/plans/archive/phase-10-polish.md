> **歴史的文書(2026-06-13 アーカイブ)**: 初期開発(2026-03)のプラン。進捗表・
> チェックリストは更新されないまま全コマンドが出荷済みで、現状を反映していない。
> 現在の正: 構造は [docs/architecture.md](../../architecture.md)、進行管理は
> [docs/refactoring-plan.md](../../refactoring-plan.md)。

# Phase 10: ポリッシュ + 配布

**依存**: Phase 2 以降いつでも可

## 実装チェックリスト

### エイリアス

- [ ] `zr alias set <name> <cmd>`
- [ ] `zr alias delete <name>`
- [ ] `zr alias list`

### 配布

- [ ] Goreleaser 設定 (darwin/linux × amd64/arm64)
- [ ] Homebrew tap (`matsuzj/homebrew-tap`)
- [ ] GitHub Actions リリース自動化

### ドキュメント

- [ ] README (インストール + クイックスタート)
- [ ] Man page 生成 (Cobra built-in)
- [ ] シェル補完インストールガイド

