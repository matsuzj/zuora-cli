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
| contact-signup | `signup` (live) | sandbox-environment | HTTP 500, code 69000060 | Body shape corrected (`ratePlans` + `terms`); field validation passes, the residual 500 is a tenant configuration limit. A reappearing 69030021 now FAILS the suite. |
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
- **meter family live verification** — Usage mediation is not enabled on
  this tenant (HTTP 400 `70002004`), so `meter audit`/`meter summary`
  response shapes remain doc-unverified (their sparse `Fields` were flagged
  by the 2026-06-13 assessment) until a mediation-enabled tenant is
  available. Validation paths are E2E-covered.
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

### request-input / body shape — RESOLVED (2026-06-13)
The `signup` body was the last test-input skip: the invalid
`subscribeToRatePlans` field is now the correct `ratePlans` + `terms` shape
per the official Sign-Up API. Field validation passes; the remaining skip is
the tenant-side HTTP 500 (`69000060`), catalogued under sandbox-environment
above. A reappearing `69030021` fails the suite (shape-regression guard).

## Pending live verification (generated — do not hand-edit)

Doc-verified assumptions awaiting a live probe. The ground truth is the
`LIVE-UNVERIFIED(<what>; since <date>; trigger: <unblock condition>)` markers in
the code — `make pending-live` lists them with line numbers. This block is
generated by `scripts/gen-pending-live.sh` and drift-gated by `make lint`.
Resolve a marker by live-probing the assumption and replacing it with a dated
`live-verified against <tenant> (<date>)` comment, then regenerate.

<!-- pending-live:begin -->
- `pkg/cmd/account/payment-methods-cascading/payment_methods_cascading.go` — cascading payment-methods shape {consent,priorities[]}; since 2026-07-05; trigger: tenant with cascading payments enabled
- `pkg/cmd/commitment/list/list.go` — commitment list shape {total,page,page_size,commitments[]}; since 2026-07-05; trigger: tenant with /v1/commitments provisioned
- `pkg/cmd/contact/snapshot/snapshot.go` — contact-snapshot postalCode field spelling; since 2026-07-05; trigger: invoice with a non-null contact-snapshot ID
- `pkg/cmd/data-query/dqutil/dqutil.go` — data-query columnSeparator/readDeleted submit options; since 2026-06-29; trigger: any tenant - submit a DSV/readDeleted job
- `pkg/cmd/meter/audit/audit.go` — meter auditTrail data[] entry shape and required queryFromTime/queryToTime params; since 2026-07-05; trigger: tenant with mediation/metering provisioned
- `pkg/cmd/meter/status/status.go` — meter runStatus envelope {success,data:{runStatus int enum,...}}; since 2026-07-05; trigger: tenant with mediation/metering provisioned
- `pkg/cmd/meter/summary/summary.go` — meter summary envelope {success,data:{query,output[]}}; since 2026-07-05; trigger: tenant with mediation/metering provisioned
- `pkg/cmd/omnichannel/create/create.go` — omni-channel create POST response {subscriptionId,subscriptionNumber,...}; since 2026-07-05; trigger: tenant with Omni-Channel provisioned
- `pkg/cmd/omnichannel/get/get.go` — omni-channel subscription GET flat shape incl. state/externalState; since 2026-07-05; trigger: tenant with Omni-Channel provisioned
- `pkg/cmd/product/get/get.go` — POST /commerce/products/{key} documented-read semantics incl. GET-vs-POST; since 2026-07-05; trigger: Commerce-provisioned tenant
<!-- pending-live:end -->

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
run passes **all suites** — `ls tests/e2e-*.sh` is the authoritative count
(11 since PR #411 added the dataquery suite on 2026-06-29; don't hand-copy the
number forward). Check counts grow with coverage — see the latest run logs
for exact numbers. Prune old logs with `make e2e-clean` (deletes
`tests/logs/*.log` older than 30 days).

## Manual cleanup after a broken run

The write suites delete what they create on a clean run, but there is **no
auto-teardown**: a mid-suite failure can leave a sandbox account behind. There
is deliberately no `make` target for this (account deletion is async and active
subscriptions can block it), so prune manually. Each suite names its account in
a comment right after `setup_log`.

| Suite | Account name | Note |
|-------|--------------|------|
| `e2e-order` | `E2E-Order-Test` | also removes its order/subscription |
| `e2e-subscription-write` | `E2E-Sub-Write-Test` | **cancel SUB_A/SUB_B/SUB_C first** — active subscriptions block account deletion |
| `e2e-contact-signup` | `E2E-Contact-Test` | |
| `e2e-invoice-payment` | `E2E-InvoicePay-Test` | |

```sh
zr account list | grep <name>            # find the leftover account key
zr account delete <account-key> --confirm   # async; returns a Job ID
```

(e2e-order's `--body` resolution step also created throwaway `E2E-BodyResolve`
accounts; the suite now deletes those inline — see #257.)
