package reverse

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdReverse(f) }

func TestInvoiceReverse_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/invoices/inv-001/reverse", map[string]interface{}{
		"id":            "inv-001",
		"invoiceNumber": "INV-REV-9001",
		"status":        "Reversed",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "reverse", "inv-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Reversed")
	// Label-bound (F-08): every prod-read key is pinned with a distinctive
	// value so a key typo renders "" and fails here (fixture-masking, #482).
	assert.Regexp(t, `(?m)^ID:\s+inv-001$`, stdout)
	assert.Regexp(t, `(?m)^Invoice Number:\s+INV-REV-9001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Reversed$`, stdout)
	// "success" is a JSON bool; GetString formats it as "true".
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestInvoiceReverse_RequiresConfirm(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "reverse", "inv-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestInvoiceReverse_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "reverse")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

// TestInvoiceReverse_SendsEmptyJSONBody pins the 415 fix: Zuora's endpoint binds a Map body parameter
// and rejects requests without a Content-Type, which the client sets only
// when a body is present — the command must send an explicit "{}".
func TestInvoiceReverse_SendsEmptyJSONBody(t *testing.T) {
	inner := cmdtest.OK(t, "PUT", "/v1/invoices/inv-001/reverse", map[string]interface{}{
		"id": "inv-001", "status": "Reversed", "success": true,
	})
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "{}", string(b))
		inner(w, r)
	}

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "reverse", "inv-001", "--confirm")
	require.NoError(t, err)
}
