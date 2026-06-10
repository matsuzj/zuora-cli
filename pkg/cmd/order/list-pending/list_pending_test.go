package listpending

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	order := &cobra.Command{Use: "order"}
	order.AddCommand(NewCmdListPending(f))
	root.AddCommand(order)
	return root
}

func TestOrderListPending_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/orders/subscription/A-S00000001/pending", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
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
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-pending", "A-S00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "ORDER NUMBER")
	assert.Contains(t, out.String(), "O-00000001")
	assert.Contains(t, out.String(), "Pending")
}

func TestOrderListPending_WithQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/orders/subscription/A-S00000001/pending", r.URL.Path)
		assert.Equal(t, "2", r.URL.Query().Get("page"))
		assert.Equal(t, "10", r.URL.Query().Get("pageSize"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"orders":  []map[string]interface{}{},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-pending", "A-S00000001", "--page", "2", "--page-size", "10"})
	err := root.Execute()

	require.NoError(t, err)
}

func TestOrderListPending_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-pending"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestOrderListPending_NextPageHint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"orders":   []map[string]interface{}{{"orderNumber": "O-00000001"}},
			"nextPage": "https://rest.example.com/v1/orders/pending?page=2",
		})
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-pending", "A-S00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "More results available")
}
