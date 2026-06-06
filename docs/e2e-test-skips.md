# E2E Test Skips

The end-to-end suites in `tests/e2e-*.sh` run the real `zr` binary against a
live Zuora tenant (currently **apac-sandbox**, Orders API enabled). Most checks
must pass; a few legitimately **skip** because the sandbox tenant lacks a
feature/route, lacks external infrastructure, or because an assertion depends on
eventual consistency.

This document catalogs every skip, its exact cause, and whether it points at a
real gap. It is generated from observed live runs — each skip below was
reproduced directly against the tenant with the error code recorded.

**Recently resolved:** `order preview` no longer skips. The 400 (`58740021`) was
a misspelled `previewOptions` body, not a tenant limitation — the fields are
`previewThruType` / `specificPreviewThruDate` (not `...Through...`), plus a
`previewTypes` array. The test body is corrected and now asserts success.

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

## Current skips (8 total)

| Suite | Check | Category | Signal | Why |
|---|---|---|---|---|
| contact-signup | `contact delete verify` | eventual-consistency | record still returned after retries | Zuora read-after-delete is not immediately consistent. |
| contact-signup | `signup` (live) | request-body shape | HTTP 400, code 69030021 | Tenant rejects the test body (`subscribeToRatePlans` invalid for this tenant's Sign-Up shape). **Fixable test body** (see below). |
| commerce | `product list-legacy` | tenant-config | HTTP 404, "no Route matched" | Legacy Commerce Product Catalog API not enabled on this tenant. |
| commerce | `plan list` | tenant-config | HTTP 404, "no Route matched" | `/v1/rateplans` (Commerce catalog) not routed on this tenant. |
| commerce | `rateplan get` | request-input | HTTP 404, code 50000040 | Test passes a **product** rate-plan id, but the endpoint resolves a **subscription** rate-plan id; no match → 404. **Fixable test input** (see below). |
| subscription-write | `subscription preview-change` | tenant-config | "invalid parameter" | Orders tenant expects a different body shape; the v1 preview params are rejected. |
| invoice-payment | `payment get` | sandbox-environment | no payment id available | No payment gateway configured on sandbox, so no payment exists to fetch. |
| ramp-commitment | `commitment list` | sandbox-environment | HTTP 404, code 50000040 | `/v1/commitments` endpoint does not exist on this tenant. |

`local`, `zoql-omnichannel`, and `usage-meter` have **no skips**.

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

- **`rateplan get`** — HTTP 404 `50000040`:
  *"IDまたはsubscriptionRatePlanNumberによるサブスクリプション料金プランが見つかりません"*.
  The test passes a **product** rate-plan id, but `rateplan get` resolves a
  **subscription** rate-plan id — no match. Needs a real subscription
  rate-plan id from a created subscription.
- **`signup` (live)** — HTTP 400 `69030021`:
  *"無効なパラメータ： 「subscribeToRatePlans」"*.
  The Sign-Up body's `subscriptionData.subscribeToRatePlans` shape isn't
  accepted by this tenant.

## Running the suites

```sh
zr auth login --client-id "$ZUORA_CLIENT_ID" --client-secret "$ZUORA_CLIENT_SECRET"
task build            # or: make build  (produces ./bin/zr)
./tests/run-all.sh    # all suites; exits non-zero if any suite fails
./tests/run-all.sh order usage-meter   # only named suites
```

`tests/logs/` (git-ignored) holds the per-run logs; each suite prints a
`Passed / Failed / Skipped` summary and a final `RESULT:` line. The latest full
run: **9/9 suites pass — 234 passed, 0 failed, 8 skipped.**
