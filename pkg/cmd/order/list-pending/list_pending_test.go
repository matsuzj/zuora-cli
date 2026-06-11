package listpending

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdListPending(f) }

func TestOrderListPending_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/orders/subscription/A-S00000001/pending", map[string]interface{}{
		"success": true,
		"orders": []map[string]interface{}{
			{
				"orderNumber":           "O-00000001",
				"status":                "Pending",
				"orderDate":             "2026-05-01",
				"existingAccountNumber": "A001",
				"createdDate":           "2026-05-01T10:00:00",
			},
		},
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-pending", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "ORDER NUMBER")
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stdout, "Pending")
}

func TestOrderListPending_WithQuery(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/orders/subscription/A-S00000001/pending", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("pageSize"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"orders":  []map[string]interface{}{},
		})
	})

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-pending", "A-S00000001", "--page", "2", "--page-size", "10")
	require.NoError(t, err)
}

func TestOrderListPending_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "list-pending")
	assert.Error(t, err)
}

func TestOrderListPending_NextPageHint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"orders":   []map[string]interface{}{{"orderNumber": "O-00000001"}},
			"nextPage": "https://rest.example.com/v1/orders/pending?page=2",
		})
	})

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "list-pending", "A-S00000001")
	require.NoError(t, err)
	assert.Contains(t, stderr, "More results available")
}
