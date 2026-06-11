package writeoff

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdWriteoff(f) }

func TestInvoiceWriteoff_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "writeoff", "inv-001", "--confirm", "--body", `{"comment":"bad debt"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "cm-001")
	assert.Contains(t, stdout, "CM00001")
}

func TestInvoiceWriteoff_NoBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body, "no --body should send an empty request body")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	})

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "writeoff", "inv-001", "--confirm")
	require.NoError(t, err)
}

func TestInvoiceWriteoff_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "writeoff", "inv-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestInvoiceWriteoff_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730040, "Invoice balance is zero")

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "writeoff", "inv-001", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "balance is zero")
}
