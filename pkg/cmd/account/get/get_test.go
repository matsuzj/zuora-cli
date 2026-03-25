package get

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
	acct.AddCommand(NewCmdGet(f))
	root.AddCommand(acct)
	return root
}

func TestAccountGet_Detail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/accounts/A001", r.URL.Path)
		w.WriteHeader(200)
		// Zuora v1 nested response: basicInfo, billingAndPayment, metrics
		json.NewEncoder(w).Encode(map[string]interface{}{
			"basicInfo": map[string]interface{}{
				"id": "id-1", "name": "Acme Corp", "accountNumber": "A001",
				"status": "Active",
			},
			"billingAndPayment": map[string]interface{}{
				"autoPay": true, "billCycleDay": 1, "currency": "USD",
			},
			"metrics": map[string]interface{}{
				"balance": "250.00",
			},
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "get", "A001"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "Acme Corp")
	assert.Contains(t, output, "Active")
	assert.Contains(t, output, "250.00")
}

func TestAccountGet_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "id-1", "name": "Acme",
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "get", "A001", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), `"name"`)
}

func TestAccountGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"account", "get"})
	err := root.Execute()

	assert.Error(t, err)
}
