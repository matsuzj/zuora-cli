# Phase 8: Ramp + Commitment + Fulfillment + Prepaid

**依存**: Phase 2

## 実装チェックリスト

### Ramp

- [ ] `zr ramp get <num>` — `GET /v1/ramps/{rampNumber}`
- [ ] `zr ramp get-by-subscription <key>` — `GET /v1/subscriptions/{subscriptionKey}/ramps`
- [ ] `zr ramp metrics <num>` — `GET /v1/ramps/{rampNumber}/ramp-metrics`
- [ ] `zr ramp metrics-by-subscription <key>` — `GET /v1/subscriptions/{subscriptionKey}/ramp-metrics`
- [ ] `zr ramp metrics-by-order <num>` — `GET /v1/orders/{orderNumber}/ramp-metrics`

### Commitment

- [ ] `zr commitment list --account <num>` — `GET /v1/commitments` (--account 必須、accountNumber クエリパラメータ)
- [ ] `zr commitment get <key>` — `GET /v1/commitments/{commitmentKey}`
- [ ] `zr commitment periods --commitment <key>` — `GET /v1/commitments/periods?commitmentKey={key}` (--commitment 必須。または --account + --start-date + --end-date の組み合わせ)
- [ ] `zr commitment balance <id>` — `GET /v1/commitments/{commitmentId}/balance`
- [ ] `zr commitment schedules <key>` — `GET /v1/commitments/{commitmentKey}/schedules`

### Fulfillment

- [ ] `zr fulfillment create` — `POST /v1/fulfillments`
- [ ] `zr fulfillment get <key>` — `GET /v1/fulfillments/{key}`
- [ ] `zr fulfillment update <key>` — `PUT /v1/fulfillments/{key}`
- [ ] `zr fulfillment delete <key>` — `DELETE /v1/fulfillments/{key}`
- [ ] `zr fulfillment-item create` — `POST /v1/fulfillment-items`
- [ ] `zr fulfillment-item get <id>` — `GET /v1/fulfillment-items/{id}`
- [ ] `zr fulfillment-item update <id>` — `PUT /v1/fulfillment-items/{id}`
- [ ] `zr fulfillment-item delete <id>` — `DELETE /v1/fulfillment-items/{id}`

### Prepaid with Drawdown

- [ ] `zr prepaid rollover` — `POST /v1/ppdd/rollover`
- [ ] `zr prepaid reverse-rollover` — `POST /v1/ppdd/reverse-rollover`
- [ ] `zr prepaid deplete` — `POST /v1/prepaid-balance-funds/deplete`
