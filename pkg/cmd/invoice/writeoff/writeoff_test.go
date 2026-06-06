package writeoff

import (
	"encoding/json"
	"io"
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
	inv.AddCommand(NewCmdWriteoff(f))
	root.AddCommand(inv)
	return root
}

func TestInvoiceWriteoff_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/invoices/inv-001/write-off", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "bad debt")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			// Use Zuora's real credit-memo field name ("number"), matching the
			// creditmemo get/list commands, so this test actually guards the key.
			"creditMemo": map[string]interface{}{"id": "cm-001", "number": "CM00001"},
			"success":    true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "writeoff", "inv-001", "--confirm", "--body", `{"comment":"bad debt"}`})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "cm-001")
	assert.Contains(t, out.String(), "CM00001")
}

func TestInvoiceWriteoff_NoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body, "no --body should send an empty request body")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "writeoff", "inv-001", "--confirm"})
	require.NoError(t, root.Execute())
}

func TestInvoiceWriteoff_RequiresConfirm(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "writeoff", "inv-001"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestInvoiceWriteoff_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 58730040, "message": "Invoice balance is zero"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"invoice", "writeoff", "inv-001", "--confirm"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "balance is zero")
}
