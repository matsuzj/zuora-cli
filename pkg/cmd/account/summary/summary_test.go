package summary

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
	acct := &cobra.Command{Use: "account"}
	acct.AddCommand(NewCmdSummary(f))
	root.AddCommand(acct)
	return root
}

func TestAccountSummary_Detail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/accounts/A001/summary", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"basicInfo": map[string]interface{}{
				"id": "id-1", "name": "Acme Corp", "accountNumber": "A001",
				"status": "Active", "balance": 100.0, "currency": "USD",
			},
			"subscriptions": []interface{}{map[string]string{"id": "sub-1"}},
			"invoices":      []interface{}{map[string]string{"id": "inv-1"}, map[string]string{"id": "inv-2"}},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "summary", "A001"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "Acme Corp")
	assert.Contains(t, output, "Subscriptions")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "Invoices")
	assert.Contains(t, output, "2")
}
