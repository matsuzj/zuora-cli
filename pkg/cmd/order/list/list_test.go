package list

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestOrderList_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders", map[string]interface{}{
		"orders": []map[string]interface{}{
			{
				"orderNumber":           "O-00000001",
				"status":                "Completed",
				"orderDate":             "2026-01-01",
				"existingAccountNumber": "A00000001",
				"createdDate":           "2026-01-01T00:00:00Z",
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stdout, "Completed")
}

func TestOrderList_WithStatusFilter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Completed", r.URL.Query().Get("status"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orders": []map[string]interface{}{},
		})
	})

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list", "--status", "Completed")
	require.NoError(t, err)
}

// TestOrderList_ScopeFlags folds in the former list-by-* commands: each scope
// flag routes to its dedicated endpoint (#454).
func TestOrderList_ScopeFlags(t *testing.T) {
	cases := []struct {
		flag, value, path string
	}{
		{"subscription", "A-S00000001", "/v1/orders/subscription/A-S00000001"},
		{"subscription-owner", "A00000001", "/v1/orders/subscriptionOwner/A00000001"},
		{"invoice-owner", "A00000002", "/v1/orders/invoiceOwner/A00000002"},
	}
	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			handler := cmdtest.OK(t, "GET", tc.path, map[string]interface{}{
				"orders": []map[string]interface{}{{"orderNumber": "O-00000001"}},
			})
			stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list", "--"+tc.flag, tc.value)
			require.NoError(t, err)
			assert.Contains(t, stdout, "O-00000001")
		})
	}
}

// TestOrderList_SubscriptionPending folds in the former list-pending command.
func TestOrderList_SubscriptionPending(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders/subscription/A-S00000001/pending", map[string]interface{}{
		"orders": []map[string]interface{}{{"orderNumber": "O-PENDING"}},
	})
	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list", "--subscription", "A-S00000001", "--pending")
	require.NoError(t, err)
	assert.Contains(t, stdout, "O-PENDING")
}

func TestOrderList_PendingRequiresSubscription(t *testing.T) {
	// --pending without --subscription is rejected before any request.
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "list", "--pending")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--pending requires --subscription")
}

func TestOrderList_ScopesMutuallyExclusive(t *testing.T) {
	// Two scope flags are rejected before any request (nil handler asserts none).
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "list", "--subscription", "A-S1", "--invoice-owner", "A00000002")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "specify at most one of --subscription, --subscription-owner, --invoice-owner")
}

func TestOrderList_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"orders": []map[string]interface{}{
			{"orderNumber": "O-00000001"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
}
