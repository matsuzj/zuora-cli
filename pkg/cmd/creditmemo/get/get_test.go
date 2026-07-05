package get

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestCreditMemoGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/creditmemos/cm-001", map[string]interface{}{
		"id":                "cm-001",
		"number":            "CM00001",
		"creditMemoDate":    "2026-01-15",
		"amount":            100.50,
		"appliedAmount":     60.20,
		"refundAmount":      15.05,
		"unappliedAmount":   25.25,
		"status":            "Posted",
		"currency":          "JPY",
		"reasonCode":        "Standard Adjustment",
		"referredInvoiceId": "inv-ref-777",
		"accountId":         "acc-cm-042",
		"accountNumber":     "A00000001",
		"createdDate":       "2026-01-10T09:30:00Z",
		"success":           true,
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "get", "cm-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Number:\s+CM00001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Posted$`, stdout)
	// Balance is sourced from "unappliedAmount" (credit memos have no "balance"
	// key). Bites if the production read reverts to "balance" (#418).
	assert.Regexp(t, `(?m)^Balance:\s+25\.25$`, stdout)
	// Every remaining prod-read key is pinned with a distinctive value so a key
	// typo or nesting mistake renders "" and fails here (fixture-masking, #482).
	assert.Regexp(t, `(?m)^ID:\s+cm-001$`, stdout)
	assert.Regexp(t, `(?m)^Credit Memo Date:\s+2026-01-15$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+100\.50$`, stdout)        // money
	assert.Regexp(t, `(?m)^Applied Amount:\s+60\.20$`, stdout) // money
	assert.Regexp(t, `(?m)^Refund Amount:\s+15\.05$`, stdout)  // money
	assert.Regexp(t, `(?m)^Currency:\s+JPY$`, stdout)
	assert.Regexp(t, `(?m)^Reason Code:\s+Standard Adjustment$`, stdout)
	assert.Regexp(t, `(?m)^Referred Invoice ID:\s+inv-ref-777$`, stdout)
	assert.Regexp(t, `(?m)^Account ID:\s+acc-cm-042$`, stdout)
	assert.Regexp(t, `(?m)^Account Number:\s+A00000001$`, stdout)
	assert.Regexp(t, `(?m)^Created Date:\s+2026-01-10T09:30:00Z$`, stdout)
}

func TestCreditMemoGet_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id":      "cm-001",
		"number":  "CM00001",
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "get", "cm-001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"number"`)
}

func TestCreditMemoGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "creditmemo", newCmd, nil, "creditmemo", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestCreditMemoGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Credit memo not found")

	_, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "get", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Credit memo not found")
}
