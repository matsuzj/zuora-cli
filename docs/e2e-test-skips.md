# E2E Test Skips

The end-to-end suites in `tests/e2e-*.sh` run the real `zr` binary against a
live Zuora tenant (currently **apac-sandbox**, Orders API enabled). Most checks
must pass; a few legitimately **skip** because the sandbox tenant lacks a
feature/route, lacks external infrastructure, or because an assertion depends on
eventual consistency.

This document catalogs every skip, its exact cause, and whether it points at a
real gap. It is generated from observed live runs — each skip below was
reproduced directly against the tenant with the error code recorded.

**Recently resolved (test-input fixes, not tenant gaps):**
- `order preview` — the 400 (`58740021`) was a misspelled `previewOptions` body:
  the fields are `previewThruType` / `specificPreviewThruDate` (not
  `...Through...`), plus a `previewTypes` array. Body corrected; asserts success.
- `rateplan get` — the 404 (`50000040`) was passing a *product* rate-plan id; the
  endpoint resolves a *subscription* rate-plan id. The test now derives a real one
  via ZOQL (`SELECT Id FROM RatePlan`) and asserts success.

## How skips work in these suites

- **Skips are status-specific, never blanket.** A check only skips on a precise,
  expected signal (a specific HTTP status / Zuora error code, or a documented
  eventual-consistency window). Any *other* failure is a hard `FAIL`. This is
  deliberate: a broad "skip on any error" would let real CLI regressions hide.
- A skipped check means **the CLI built and sent the request correctly** and the
  tenant/environment — not `zr` — is why the call can't be asserted green here.
- The **auth gate is not a skip**: if the stored token is expired, every live
  suite hard-fails at Step 0 (`zr auth status` must show `Token: valid`). Run
  `zr auth login` first. Only `e2e-local.sh` is offline and needs no auth.
- Some checks **pass on an expected error**: where a tenant limitation is
  deterministic, the suite asserts the exact error code as a green check
  instead of skipping (e.g. `payment-methods-default` → 50000040,
  `payment-methods-cascading` → 50000010, `subscription changelog` → 50000010).
  These lock the error-rendering path live and flip loudly if the tenant ever
  gains the feature — update the assertion to a data assertion then.
- A few **dormant skip guards** exist for portability and never fire on this
  runner: the zoql partial-env check skips when the runner has no OS-keyring
  credentials, and `subscription changelog` skips when ZOQL finds no
  subscription.

## Current skips (8 total)

| Suite | Check | Category | Signal | Why |
|---|---|---|---|---|
| contact-signup | `contact delete verify` | eventual-consistency | record still returned after retries | Zuora read-after-delete is not immediately consistent. |
| contact-signup | `signup` (live) | request-body shape | HTTP 400, code 69030021 | Tenant rejects the test body (`subscribeToRatePlans` invalid for this tenant's Sign-Up shape). **Fixable test body** (see below). |
| commerce | `product list-legacy` | tenant-config | HTTP 404, "no Route matched" | Legacy Commerce Product Catalog API not enabled on this tenant. |
| commerce | `plan list` | tenant-config | HTTP 404, "no Route matched" | `/v1/rateplans` (Commerce catalog) not routed on this tenant. |
| subscription-write | `subscription preview-change` | tenant-config | "invalid parameter" | Orders tenant expects a different body shape; the v1 preview params are rejected. |
| invoice-payment | `payment get` | sandbox-environment | no payment id available | No payment gateway configured on sandbox, so no payment exists to fetch. |
| invoice-payment | `invoice post` (live) | tenant-config | invoice already `Posted` | Order-driven billing auto-posts the invoice on this tenant, so there is no Draft to post. The bodyless-PUT 415 contract (#220) stays live-guarded by the billrun suite. |
| ramp-commitment | `commitment list` | sandbox-environment | HTTP 404, code 50000040 | `/v1/commitments` endpoint does not exist on this tenant. |

`local`, `zoql-omnichannel`, `billrun`, and `usage-meter` have **no observed
skips** (zoql carries dormant portability guards, see above).

## Details

### tenant-config-limitation
These pass on a tenant where the corresponding feature/route is provisioned.

- **`product list-legacy`**, **`plan list`** — HTTP 404 "no Route matched with
  those values": the legacy Commerce Product Catalog API (and `/v1/rateplans`)
  is not enabled on apac-sandbox. Needs a tenant with the Commerce catalog
  entitlement.
- **`subscription preview-change`** — the Orders tenant rejects the v1-style
  preview body ("invalid parameter"); a correctly-shaped Orders preview body
  (with real `chargeUpdates`) would be needed to assert success.

### sandbox-environment
These need external infrastructure / endpoints that the sandbox doesn't have.

- **`payment get`** — apac-sandbox has no payment gateway, so no payment is ever
  created; `payment list` returns an empty array and there is no id to `get`.
  Passes on a gateway-configured tenant.
- **`commitment list`** — HTTP 404 `50000040`:
  *"The endpoint /v1/commitments does not exist."* — the Commitments feature is
  not provisioned on this tenant.
- **`invoice post` (live)** — the invoice created by the order's
  `runBilling:true` arrives already `Posted` (tenant auto-post), so the
  Draft→Posted transition can't be exercised here. Validation and the
  `--confirm` guards still assert green, and the billrun suite live-proves the
  bodyless-PUT Content-Type contract every run.

### eventual-consistency
- **`contact delete verify`** — the delete itself succeeds; only the immediate
  read-back can still return the contact. The suite retries (≈5×2s) and skips if
  it is still returned, since this is a Zuora propagation window, not a CLI
  defect.

### request-input / body shape (candidates for a real fix, not permanent skips)
These skip because the test's request input doesn't match what the tenant
expects — a **test-input issue**, not a tenant entitlement gap. The
corresponding argument validation (missing `--body`/arg) is still asserted
green; only the live happy-path is skipped and could pass once the input is
corrected.

- **`signup` (live)** — HTTP 400 `69030021`:
  *"無効なパラメータ： 「subscribeToRatePlans」"*.
  The Sign-Up body's `subscriptionData.subscribeToRatePlans` shape isn't
  accepted by this tenant.

## Commerce live verification — pending on a Commerce-enabled tenant

(docs/plans/phase-05-pending.md から統合、2026-06-13。これは生きた TODO —
apac-sandbox は Commerce API 未提供のため、以下の live happy-path は
Commerce 有効テナントでの手動検証待ち。バリデーション・ユニットテスト・
`rateplan get` の live は検証済み。)

- `product create/update/get/list-legacy` — `/commerce/products` 系の実呼び出し
- `plan create/update/get/list/purchase-options` — `/commerce/plans` 系の実呼び出し
- `charge create/update/get/update-tiers` — `/commerce/charges`・`/commerce/tiers` の実呼び出し

検証方法: Commerce 有効テナントに `zr auth login` → `bash tests/e2e-commerce.sh` →
product list-legacy / plan list のスキップが PASS に変わることを確認。

## Running the suites

```sh
zr auth login --client-id "$ZUORA_CLIENT_ID" --client-secret "$ZUORA_CLIENT_SECRET"
task build            # or: make build  (produces ./bin/zr)
./tests/run-all.sh    # all suites; exits non-zero if any suite fails
./tests/run-all.sh order usage-meter   # only named suites
```

`tests/logs/` (git-ignored) holds the per-run logs; each suite prints a
`Passed / Failed / Skipped` summary and a final `RESULT:` line. The latest full
run: **10/10 suites pass** (the 2026-06-12 expansion added the billrun suite
plus behavior-change, flag-matrix, and lifecycle coverage; check counts grow
with coverage — see the latest run logs for exact numbers).
