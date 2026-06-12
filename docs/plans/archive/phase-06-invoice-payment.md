> **歴史的文書(2026-06-13 アーカイブ)**: 初期開発(2026-03)のプラン。進捗表・
> チェックリストは更新されないまま全コマンドが出荷済みで、現状を反映していない。
> 現在の正: 構造は [docs/architecture.md](../../architecture.md)、進行管理は
> [docs/refactoring-plan.md](../../refactoring-plan.md)。

# Phase 6: Invoice + Payment

**依存**: Phase 2

## 実装チェックリスト

### Invoice

- [ ] `zr invoice list --account <key>` — `GET /v1/transactions/invoices/accounts/{account-key}` (--account フラグ必須。注意: Older API。参照: https://developer.zuora.com/v1-api-reference/older-api/transactions/)
- [ ] `zr invoice get <id>` — `GET /v1/invoices/{invoice-id}`
- [ ] `zr invoice items <id>` — `GET /v1/invoices/{invoice-id}/items`
- [ ] `zr invoice files <id>` — `GET /v1/invoices/{invoice-id}/files` (ファイル一覧取得、--download で PDF ダウンロード、--output <path> で保存先指定)
- [ ] `zr invoice email <id>` — `POST /v1/invoices/{invoice-id}/emails` (--body 必須、宛先メール等)
- [ ] `zr invoice usage-rate-detail <item-id>` — `GET /v1/invoices/invoice-item/{invoice-item-id}/usage-rate-detail`

### Payment

- [ ] `zr payment list --account <key>` — `GET /v1/transactions/payments/accounts/{account-key}` (--account フラグ必須。注意: Older API。参照: https://developer.zuora.com/v1-api-reference/older-api/transactions/)
- [ ] `zr payment get <id>` — `GET /v1/payments/{payment-id}`
- [ ] `zr payment create` — `POST /v1/payments` (--body 必須)
- [ ] `zr payment apply <id>` — `PUT /v1/payments/{payment-id}/apply` (--body 必須、invoice/debit-memo 適用先 JSON)
- [ ] `zr payment refund <id>` — `POST /v1/payments/{payment-id}/refunds` (--body 必須、返金額等)
