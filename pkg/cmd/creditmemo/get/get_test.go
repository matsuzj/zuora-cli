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
	cm := &cobra.Command{Use: "creditmemo"}
	cm.AddCommand(NewCmdGet(f))
	root.AddCommand(cm)
	return root
}

func TestCreditMemoGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/creditmemos/cm-001", r.URL.Path)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":             "cm-001",
			"number":         "CM00001",
			"creditMemoDate": "2026-01-15",
			"amount":         100.50,
			"balance":        25.25,
			"status":         "Posted",
			"accountNumber":  "A00000001",
			"success":        true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"creditmemo", "get", "cm-001"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "CM00001")
	assert.Contains(t, out.String(), "Posted")
}

func TestCreditMemoGet_JSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "cm-001",
			"number":  "CM00001",
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"creditmemo", "get", "cm-001", "--json"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), `"number"`)
}

func TestCreditMemoGet_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"creditmemo", "get"})
	assert.Error(t, root.Execute())
}

func TestCreditMemoGet_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 50000040, "message": "Credit memo not found"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"creditmemo", "get", "bad-id"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Credit memo not found")
}
