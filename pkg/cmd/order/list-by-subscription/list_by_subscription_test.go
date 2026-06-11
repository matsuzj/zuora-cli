package listbysubscription

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdListBySubscription(f) }

func TestOrderListBySubscription_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders/subscription/A-S00000001", map[string]interface{}{
		"success": true,
		"orders": []map[string]interface{}{
			{
				"orderNumber":           "O-00000001",
				"status":                "Completed",
				"orderDate":             "2026-01-01",
				"existingAccountNumber": "A001",
				"createdDate":           "2026-01-02",
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-by-subscription", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stdout, "Completed")
}

func TestOrderListBySubscription_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Subscription not found")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-by-subscription", "A-S99999999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription not found")
}

func TestOrderListBySubscription_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "list-by-subscription")
	assert.Error(t, err)
}

func TestOrderListBySubscription_NextPageHint(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"success":  true,
		"orders":   []map[string]interface{}{{"orderNumber": "O-00000001"}},
		"nextPage": "https://rest.example.com/v1/orders/subscription/A-S00000001?page=2",
	})

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-by-subscription", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stderr, "More results available")
}
