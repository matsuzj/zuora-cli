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
