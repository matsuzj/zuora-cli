# Phase 4: Order

**依存**: Phase 2

## 実装チェックリスト

### Order — CRUD

- [ ] `zr order list` — `GET /v1/orders` (--status, --page, --page-size)
- [ ] `zr order get <num>` — `GET /v1/orders/{orderNumber}`
- [ ] `zr order create` — `POST /v1/orders`
- [ ] `zr order update <num>` — `PUT /v1/orders/{orderNumber}` (draft/scheduled。警告: full payload 必須。不足した order actions は削除される。--body で完全な JSON を渡すこと)
- [ ] `zr order delete <num>` — `DELETE /v1/orders/{orderNumber}`

### Order — ライフサイクル

- [ ] `zr order activate <num>` — `PUT /v1/orders/{orderNumber}/activate`
- [ ] `zr order cancel <num>` — `PUT /v1/orders/{orderNumber}/cancel`
- [ ] `zr order revert <num>` — `POST /v1/orders/{orderNumber}/revert` (--body 必須、orderDate 含む JSON)
- [ ] `zr order preview` — `POST /v1/orders/preview`

### Order — クエリ

- [ ] `zr order list-by-subscription-owner <acct>` — `GET /v1/orders/subscriptionOwner/{accountNumber}`
- [ ] `zr order list-by-subscription <key>` — `GET /v1/orders/subscription/{subscription-key}` (subscription number or key)
- [ ] `zr order list-pending <key>` — `GET /v1/orders/subscription/{subscription-key}/pending` (subscription number or key)
- [ ] `zr order list-by-invoice-owner <acct>` — `GET /v1/orders/invoiceOwner/{accountNumber}`

### Order — カスタムフィールド・日付

- [ ] `zr order update-custom-fields <num>` — `PUT /v1/orders/{orderNumber}/customFields` (--body 必須)
- [ ] `zr order update-trigger-dates <num>` — `PUT /v1/orders/{orderNumber}/triggerDates` (--body 必須)

### Order — 非同期操作

- [ ] `zr order create-async` — `POST /v1/async/orders`
- [ ] `zr order preview-async` — `POST /v1/async/orders/preview`
- [ ] `zr order delete-async <num>` — `DELETE /v1/async/orders/{orderNumber}`
- [ ] `zr order job-status <jobId>` — `GET /v1/async-jobs/{jobId}`

### Order Actions

- [ ] `zr order-action update <id>` — `PUT /v1/orderActions/{id}` (--body 必須)

### Order Line Items

- [ ] `zr order-line-item get <id>` — `GET /v1/order-line-items/{itemId}`
- [ ] `zr order-line-item update <id>` — `PUT /v1/order-line-items/{itemId}`
- [ ] `zr order-line-item bulk-update` — `POST /v1/order-line-items/bulk` (--body 必須、orderLineItems 配列 JSON、max 100)

## コマンド詳細仕様

### order コマンド

| サブコマンド | API | フラグ | テーブル出力カラム |
|---|---|---|---|
| `list` | `GET /v1/orders` | `--status`, `--page`, `--page-size` | ORDER_NUMBER, STATUS, CREATED, ACCOUNT |
| `get <num>` | `GET /v1/orders/{num}` | — | 詳細ビュー |
| `create` | `POST /v1/orders` | `--body` | 作成結果 |
| `update <num>` | `PUT /v1/orders/{num}` | `--body` | 更新結果 |
| `delete <num>` | `DELETE /v1/orders/{num}` | `--confirm` | 削除結果 |
| `activate <num>` | `PUT /v1/orders/{num}/activate` | — | 結果 |
| `cancel <num>` | `PUT /v1/orders/{num}/cancel` | — | 結果 |
| `revert <num>` | `POST /v1/orders/{num}/revert` | `--body` (必須, orderDate) | 結果 |
| `preview` | `POST /v1/orders/preview` | `--body` | プレビュー結果 |
| `create-async` | `POST /v1/async/orders` | `--body` | jobId |
| `preview-async` | `POST /v1/async/orders/preview` | `--body` | jobId |
| `delete-async <num>` | `DELETE /v1/async/orders/{num}` | — | jobId |
| `job-status <jobId>` | `GET /v1/async-jobs/{jobId}` | `--watch` | ジョブ状態 |
