package listbysubscription

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
	order.AddCommand(NewCmdListBySubscription(f))
	root.AddCommand(order)
	return root
}

func TestOrderListBySubscription_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/orders/subscription/A-S00000001", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
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
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-by-subscription", "A-S00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "O-00000001")
	assert.Contains(t, out.String(), "Completed")
}

func TestOrderListBySubscription_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000040, "message": "Subscription not found"},
			},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-by-subscription", "A-S99999999"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Subscription not found")
}

func TestOrderListBySubscription_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-by-subscription"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestOrderListBySubscription_NextPageHint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":  true,
			"orders":   []map[string]interface{}{{"orderNumber": "O-00000001"}},
			"nextPage": "https://rest.example.com/v1/orders/subscription/A-S00000001?page=2",
		})
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"order", "list-by-subscription", "A-S00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "More results available")
}
