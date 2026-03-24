---
title: "Phase 9: ZOQL + Subscription Change Log + Omnichannel"
status: not-started
depends_on: [phase-02]
---

# Phase 9: ZOQL + Subscription Change Log + Omnichannel

## 実装チェックリスト

### ZOQL クエリ

- [ ] `zr query "<ZOQL>"` — `POST /v1/action/query`
- [ ] テーブル / JSON / CSV 出力
- [ ] `--export <file>` ファイルエクスポート
- [ ] ページネーション (大量結果)

### Subscription Change Log

- [ ] `zr subscription changelog <num>` — `GET /v1/subscription-change-logs/{subscriptionNumber}`
- [ ] `zr subscription changelog-by-order <num>` — `GET /v1/subscription-change-logs/orders/{orderNumber}`
- [ ] `zr subscription changelog-version <num> <ver>` — `GET /v1/subscription-change-logs/{subscriptionNumber}/versions/{version}`

### Omnichannel Subscription

- [ ] `zr omnichannel create` — `POST /v1/omni-channel-subscriptions`
- [ ] `zr omnichannel get <key>` — `GET /v1/omni-channel-subscriptions/{subscriptionKey}`
- [ ] `zr omnichannel delete <key>` — `DELETE /v1/omni-channel-subscriptions/{subscriptionKey}`

## コマンド詳細仕様

### query コマンド (ZOQL)

```
zr query "<ZOQL>" [flags]

Examples:
  zr query "SELECT Id, Name, Status FROM Account WHERE Status = 'Active'"
  zr query "SELECT Id, SubscriptionNumber FROM Subscription" --json
  zr query "SELECT Id FROM Invoice" --export invoices.csv --csv

Flags:
      --csv              Output as CSV (query 固有。--json/--jq/--template はグローバルフラグを継承)
      --export string    Export results to file (--csv or --json と組み合わせ)
      --limit int        Maximum number of rows (default: all)
```

**出力フラグ優先順位**: `--jq` > `--json` > `--template` > `--csv` > テーブル (デフォルト)
グローバルフラグ (`--json`, `--jq`, `--template`) は query でもそのまま使用可能。`--csv` は query 固有の追加フラグ。
