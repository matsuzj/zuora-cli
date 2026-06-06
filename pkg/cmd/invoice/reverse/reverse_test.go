package reverse

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
	inv.AddCommand(NewCmdReverse(f))
	root.AddCommand(inv)
	return root
}

func TestInvoiceReverse_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/invoices/inv-001/reverse", r.URL.Path)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "inv-001",
			"status":  "Reversed",
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "reverse", "inv-001", "--confirm"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Reversed")
}

func TestInvoiceReverse_RequiresConfirm(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "reverse", "inv-001"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestInvoiceReverse_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "reverse"})
	assert.Error(t, root.Execute())
}
