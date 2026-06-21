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

func TestOrderGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders/O-00000001", map[string]interface{}{
		"success": true,
		"order": map[string]interface{}{
			"orderNumber": "O-00000001",
			"status":      "Completed",
			"orderDate":   "2026-01-01",
			// Distinctive value so the assertion below genuinely bites: the
			// "Account Number" column reads "existingAccountNumber" (NOT the
			// flatter "accountNumber"), a drift-prone key. Without asserting it,
			// a swap to the wrong key would render empty yet keep the test green.
			"existingAccountNumber": "ACCT-9000001",
			"createdDate":           "2026-01-01T00:00:00Z",
		},
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "get", "O-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")   // orderNumber (nested)
	assert.Contains(t, stdout, "Completed")    // status (nested)
	assert.Contains(t, stdout, "ACCT-9000001") // existingAccountNumber (nested, drift-prone)
}

func TestOrderGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "get")
	assert.Error(t, err)
}

func TestOrderGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730020, "Order not found")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "get", "O-99999999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Order not found")
}
