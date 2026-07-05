# Test fixture provenance

Golden fixtures in this directory hold **real Zuora response shapes** (see
AGENTS.md, "Build fixtures from REAL response shapes") so command tests render
against the true envelope — nesting and all — instead of a hand-written guess
that can mask a wrong-key bug.

Each fixture below records the endpoint it models and how to re-capture it, so
the shape can be re-verified against the live API when Zuora changes a response.

> **Values are anonymised.** IDs, account numbers, and dates are replaced with
> obviously-fake placeholders (e.g. `O-00000001`, `ACCT-9000001`, `2026-01-01`).
> Only the **shape** (keys, nesting, types) is load-bearing — never paste a real
> tenant's data here.

## Re-capturing a fixture

Run the command against a **sandbox** tenant (never production), then
re-anonymise and confirm the test still passes:

    zr <command> <args> --json | jq . > pkg/cmdtest/fixtures/<name>.json

`make lint` fails if a fixture is added without a row in the table below.

## Fixtures

| Fixture | Endpoint (real-shape source) | Used by | Shape notes |
|---|---|---|---|
| `order_get.json` | `GET /v1/orders/{orderNumber}` — [Zuora: Retrieve an order](https://www.zuora.com/developer/api-references/api/operation/GET_Order) | `pkg/cmd/order/get` | Order fields nest under an `order` key alongside a top-level `success`. The test asserts `existingAccountNumber` (NOT the flatter `accountNumber`) — a drift-prone nested key the unwrap fallback depends on. `description`/`createdBy`/`updatedDate`/`updatedBy` were added synthetically (#482) — standard order-envelope siblings per the Zuora API reference, shape-consistent with the captured response but not themselves live-captured. |
