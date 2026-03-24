# Phase 5: Commerce (Product / Plan / Charge)

**依存**: Phase 2

## 実装チェックリスト

### Product

- [ ] `zr product create` — `POST /commerce/products`
- [ ] `zr product update` — `PUT /commerce/products`
- [ ] `zr product get <key>` — `GET /commerce/products/{product_key}`
- [ ] `zr product list-legacy` — `POST /commerce/legacy/products/list` (--body 必須、フィルタ条件 JSON)

### Plan (Rate Plan)

- [ ] `zr plan create` — `POST /commerce/plans`
- [ ] `zr plan update` — `PUT /commerce/plans`
- [ ] `zr plan get --key <ratePlanKey>` — `POST /commerce/plans/query` (--key 必須、body の product_rate_plan_key に設定)
- [ ] `zr plan list` — `POST /commerce/plans/list` (--body 必須、フィルタ条件 JSON)
- [ ] `zr plan purchase-options --plan <ratePlanId>` — `POST /commerce/purchase-options/list` (--plan の値を body の filters 配列 `[{"field": "prp_id", "operator": "=", "value": {"string_value": "<ratePlanId>"}}]` に設定)

### Charge (Rate Plan Charge)

- [ ] `zr charge create` — `POST /commerce/charges`
- [ ] `zr charge update` — `PUT /commerce/charges`
- [ ] `zr charge get --key <chargeKey>` — `POST /commerce/charges/query` (--key 必須、body の product_rate_plan_charge_key に設定)
- [ ] `zr charge update-tiers` — `PUT /commerce/tiers`

### Rate Plan (v1)

- [ ] `zr rateplan get <id>` — `GET /v1/rateplans/{ratePlanId}`
