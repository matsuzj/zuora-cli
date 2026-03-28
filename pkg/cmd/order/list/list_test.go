package list

import (
	"encoding/json"
	"github.com/matsuzj/zuora-cli/internal/testutil"
	"net/http"
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
	order.AddCommand(NewCmdList(f))
	root.AddCommand(order)
	return root
}

func TestOrderList_Success(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/orders", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
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
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "O-00000001")
	assert.Contains(t, out.String(), "Completed")
}

func TestOrderList_WithStatusFilter(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Completed", r.URL.Query().Get("status"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orders": []map[string]interface{}{},
		})
	}))

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list", "--status", "Completed"})
	err := root.Execute()

	require.NoError(t, err)
}

func TestOrderList_JSON(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"orders": []map[string]interface{}{
				{"orderNumber": "O-00000001"},
			},
		})
	}))

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "O-00000001")
}
