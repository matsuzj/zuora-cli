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

func TestPaymentGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/payments/pay-001", map[string]interface{}{
		"id":            "pay-001",
		"number":        "P-00000001", // real field is "number" (live-verified); "paymentNumber" never existed
		"effectiveDate": "2026-01-15",
		"amount":        100.00,
		"status":        "Processed",
		"type":          "External",
		"accountId":     "acc-001",
		"gatewayState":  "Settled",
		"createdDate":   "2026-01-10T10:00:00Z",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "get", "pay-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "P-00000001")
	assert.Contains(t, stdout, "100.00", "amount must render with cents")
	assert.Contains(t, stdout, "Processed")
}

func TestPaymentGet_LargeAmountNotScientific(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id":     "pay-002",
		"amount": 1234567.89,
		"status": "Processed",
	})

	stdout, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "get", "pay-002")
	require.NoError(t, err)
	assert.Contains(t, stdout, "1234567.89", "large amount must render as a plain decimal")
	assert.NotContains(t, stdout, "e+", "amount must not use scientific notation")
}

func TestPaymentGet_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id":            "pay-001",
		"paymentNumber": "P-00000001",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "get", "pay-001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"paymentNumber"`)
}

func TestPaymentGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "get")
	assert.Error(t, err)
}

func TestPaymentGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Payment not found")

	_, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "get", "bad-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Payment not found")
}
