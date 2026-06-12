#!/bin/bash
# E2E Test: Commerce / Catalog (Product, Plan, Charge, RatePlan)
# テナント: apac-sandbox
# 注意: product list-legacy / plan list は --body を必須とする検索 API。テナントの
#       Product Catalog 設定に依存するが、空フィルタ {} なら結果(空配列可)を返す。

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ZR="$SCRIPT_DIR/../bin/zr"
PASS=0
FAIL=0
SKIP=0
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_DIR="$SCRIPT_DIR/logs"
mkdir -p "$LOG_DIR"
LOG_FILE="$LOG_DIR/e2e-commerce-${TIMESTAMP}.log"

source "$SCRIPT_DIR/lib/e2e-common.sh"
setup_log

# ─────────────────────────────────────────
header "Step 0: Auth check"
# ─────────────────────────────────────────
require_auth

# ─────────────────────────────────────────
header "Step 1: Validation (read commands)"
# ─────────────────────────────────────────
echo "  Testing: product get without argument"
expect_fail "product get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR product get

echo "  Testing: product list-legacy without --body"
expect_fail "product list-legacy validation → requires --body" 'required flag(s) "body" not set' -- $ZR product list-legacy

echo "  Testing: plan get without --key"
expect_fail "plan get validation → requires the key arg" "accepts 1 arg(s), received 0" -- $ZR plan get

echo "  Testing: plan list without --body"
expect_fail "plan list validation → requires --body" 'required flag(s) "body" not set' -- $ZR plan list

echo "  Testing: plan purchase-options without --plan"
expect_fail "plan purchase-options validation → requires --plan" 'required flag(s) "plan" not set' -- $ZR plan purchase-options

echo "  Testing: charge get without --key"
expect_fail "charge get validation → requires the key arg" "accepts 1 arg(s), received 0" -- $ZR charge get

echo "  Testing: rateplan get without argument"
expect_fail "rateplan get validation → requires argument" "accepts 1 arg(s), received 0" -- $ZR rateplan get

# ─────────────────────────────────────────
header "Step 2: Validation (mutating commands)"
# ─────────────────────────────────────────
echo "  Testing: product create without --body"
expect_fail "product create validation → requires --body" 'required flag(s) "body" not set' -- $ZR product create

echo "  Testing: product update without --body"
expect_fail "product update validation → requires --body" 'required flag(s) "body" not set' -- $ZR product update

echo "  Testing: plan create without --body"
expect_fail "plan create validation → requires --body" 'required flag(s) "body" not set' -- $ZR plan create

echo "  Testing: plan update without --body"
expect_fail "plan update validation → requires --body" 'required flag(s) "body" not set' -- $ZR plan update

echo "  Testing: charge create without --body"
expect_fail "charge create validation → requires --body" 'required flag(s) "body" not set' -- $ZR charge create

echo "  Testing: charge update without --body"
expect_fail "charge update validation → requires --body" 'required flag(s) "body" not set' -- $ZR charge update

echo "  Testing: charge update-tiers without --body"
expect_fail "charge update-tiers validation → requires --body" 'required flag(s) "body" not set' -- $ZR charge update-tiers

# ─────────────────────────────────────────
header "Step 3: NoArgs rejection (stray positional)"
# ─────────────────────────────────────────
echo "  Testing: product create with stray arg"
expect_fail "product create → rejects stray positional arg" 'unknown command "extraArg"' -- $ZR product create extraArg --body '{}'

# ─────────────────────────────────────────
header "Step 4: Live read commands"
# ─────────────────────────────────────────
# product list-legacy is a --body search returning {"products":[...]} (empty OK).
echo "  Testing: product list-legacy --body '{}'"
read_or_skip_on "product list-legacy → .products array" '.products | type == "array"' "no Route matched" -- $ZR product list-legacy --body '{}' --json

# plan list is a --body search returning {"plans":[...]} (empty OK).
echo "  Testing: plan list --body '{}'"
read_or_skip_on "plan list → .plans array" '.plans | type == "array"' "no Route matched" -- $ZR plan list --body '{}' --json

# rateplan get resolves a *subscription* rate plan id (v1 /v1/rateplans/{id}),
# not a product rate plan id — passing the latter 404s. Derive a real
# subscription rate plan id from the tenant via ZOQL; skip only if the tenant
# genuinely has none.
SUB_RP_ID=$($ZR query "SELECT Id FROM RatePlan" --jq '.records[0].Id // ""' 2>/dev/null | tr -d '"')
if [ -n "$SUB_RP_ID" ]; then
  echo "  Testing: rateplan get $SUB_RP_ID (subscription rate plan)"
  read_or_skip_on "rateplan get → JSON object" 'type == "object" and .success == true' "HTTP 404" -- $ZR rateplan get "$SUB_RP_ID" --json
else
  skip "rateplan get → no subscription rate plan available in tenant"
fi

# ─────────────────────────────────────────
header "Summary"
# ─────────────────────────────────────────
echo ""
print_summary
