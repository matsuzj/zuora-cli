package post

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
	inv := &cobra.Command{Use: "invoice"}
	inv.AddCommand(NewCmdPost(f))
	root.AddCommand(inv)
	return root
}

func TestInvoicePost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/invoices/inv-001/post", r.URL.Path)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "inv-001",
			"invoiceNumber": "INV00001",
			"status":        "Posted",
			"success":       true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "post", "inv-001"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Posted")
}

func TestInvoicePost_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "post"})
	assert.Error(t, root.Execute())
}

func TestInvoicePost_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 58730020, "message": "Invoice is not in draft status"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "post", "inv-001"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in draft status")
}
