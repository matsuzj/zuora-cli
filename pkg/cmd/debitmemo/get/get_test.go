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
	root.PersistentFlags().Bool("csv", false, "")
	dm := &cobra.Command{Use: "debitmemo"}
	dm.AddCommand(NewCmdGet(f))
	root.AddCommand(dm)
	return root
}

func TestDebitMemoGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/debitmemos/dm-001", r.URL.Path)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "dm-001",
			"number":        "DM00001",
			"debitMemoDate": "2026-01-15",
			"amount":        110.00,
			"balance":       110.00,
			"status":        "Posted",
			"accountNumber": "A00000001",
			"success":       true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"debitmemo", "get", "dm-001"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "DM00001")
	assert.Contains(t, out.String(), "Posted")
}

func TestDebitMemoGet_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "dm-001",
			"number":  "DM00001",
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"debitmemo", "get", "dm-001", "--json"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), `"number"`)
}

func TestDebitMemoGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"debitmemo", "get"})
	assert.Error(t, root.Execute())
}

func TestDebitMemoGet_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 50000040, "message": "Debit memo not found"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"debitmemo", "get", "bad-id"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Debit memo not found")
}
