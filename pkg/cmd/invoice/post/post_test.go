package post

import (
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPost(f) }

func TestInvoicePost_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/invoices/inv-001/post", map[string]interface{}{
		"id":            "inv-001",
		"invoiceNumber": "INV00001",
		"status":        "Posted",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "post", "inv-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Posted")
}

func TestInvoicePost_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "post")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

// TestInvoicePost_RequiresConfirm pins the irreversible-write guard: without
// --confirm the command must refuse before issuing any request (nil handler
// asserts no HTTP call is made).
func TestInvoicePost_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "post", "inv-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestInvoicePost_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730020, "Invoice is not in draft status")

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "post", "inv-001", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in draft status")
}

// TestInvoicePost_SendsEmptyJSONBody pins the 415 fix: Zuora's endpoint binds a Map body parameter
// and rejects requests without a Content-Type, which the client sets only
// when a body is present — the command must send an explicit "{}".
func TestInvoicePost_SendsEmptyJSONBody(t *testing.T) {
	inner := cmdtest.OK(t, "PUT", "/v1/invoices/inv-001/post", map[string]interface{}{
		"id": "inv-001", "status": "Posted", "success": true,
	})
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "{}", string(b))
		inner(w, r)
	}

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "post", "inv-001", "--confirm")
	require.NoError(t, err)
}
