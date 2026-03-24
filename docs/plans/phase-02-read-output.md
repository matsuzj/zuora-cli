---
title: "Phase 2: 出力フォーマッタ + Account + Subscription (Read)"
status: not-started
depends_on: [phase-01]
---

# Phase 2: 出力フォーマッタ + Account + Subscription (Read)

## 実装チェックリスト

### 出力フォーマッタ (pkg/output/)

- [ ] テーブル出力 (カラム整形 + 色付き)
- [ ] JSON 出力 + `--jq` フィルタ (gojq)
- [ ] Go テンプレート出力 (`--template`)
- [ ] CSV 出力 (ZOQL 用)
- [ ] ページャ連携 (less/more)

### Account — Read 操作

- [ ] `zr account list` — `GET /object-query/accounts` (--page-size, --cursor でページネーション。--filter で filter[] パラメータ指定。例: `--filter "status=Active"`。注意: v1 ではなく object-query API)
- [ ] `zr account get <key>` — `GET /v1/accounts/{account-key}`
- [ ] `zr account summary <key>` — `GET /v1/accounts/{account-key}/summary`
- [ ] `zr account payment-methods <key>` — `GET /v1/accounts/{account-key}/payment-methods`
- [ ] `zr account payment-methods-default <key>` — `GET /v1/accounts/{account-key}/payment-methods/default`
- [ ] `zr account payment-methods-cascading <key>` — `GET /v1/accounts/{account-key}/payment-methods/cascading`

### Subscription — Read 操作

- [ ] `zr subscription list --account <key>` — `GET /v1/subscriptions/accounts/{account-key}` (--account フラグ必須、値を URL パスに埋め込み)
- [ ] `zr subscription get <key>` — `GET /v1/subscriptions/{subscription-key}`
- [ ] `zr subscription versions <key> <ver>` — `GET /v1/subscriptions/{subscription-key}/versions/{version}`
- [ ] `zr subscription metrics --subscription-numbers <nums>` — `GET /v1/subscriptions/subscription-metrics` (--subscription-numbers 必須、カンマ区切り)

### Unit テスト

- [ ] Account コマンド群のテスト (HTTP モック)
- [ ] Subscription コマンド群のテスト (HTTP モック)
- [ ] 出力フォーマッタのテスト

### Sandbox 検証

- [ ] `zr account list --json` 動作確認
- [ ] `zr subscription list --account <key>` 動作確認

## コマンド詳細仕様

### account コマンド (Read)

| サブコマンド | API | フラグ | テーブル出力カラム |
|---|---|---|---|
| `list` | `GET /object-query/accounts` | `--page-size`, `--cursor`, `--filter` | ID, NAME, NUMBER, STATUS, BALANCE, CREATED |
| `get <key>` | `GET /v1/accounts/{key}` | — | 詳細ビュー (key-value 形式) |
| `summary <key>` | `GET /v1/accounts/{key}/summary` | — | 残高・サブスク数・請求サマリー |
| `payment-methods <key>` | `GET /v1/accounts/{key}/payment-methods` | — | TYPE, LAST4, DEFAULT, STATUS |
| `payment-methods-default <key>` | `GET /v1/accounts/{key}/payment-methods/default` | — | 詳細ビュー |
| `payment-methods-cascading <key>` | `GET /v1/accounts/{key}/payment-methods/cascading` | — | 詳細ビュー |

### subscription コマンド (Read)

| サブコマンド | API | フラグ | テーブル出力カラム |
|---|---|---|---|
| `list --account <key>` | `GET /v1/subscriptions/accounts/{key}` | `--account` (必須, URL パスに埋め込み), `--page-size`, `--charge-detail` | ID, NUMBER, NAME, STATUS, TERM_TYPE, START, END |
| `get <key>` | `GET /v1/subscriptions/{key}` | — | 詳細ビュー |
| `versions <key> <ver>` | `GET /v1/subscriptions/{key}/versions/{ver}` | — | 詳細ビュー |
| `metrics` | `GET /v1/subscriptions/subscription-metrics` | `--subscription-numbers` | MRR, TCV, TCB 等 |
