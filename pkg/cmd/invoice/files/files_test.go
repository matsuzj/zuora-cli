package files

import (
	"encoding/json"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdFiles(f) }

// realFilesResponse mirrors the live-verified GET /v1/invoices/{id}/files shape
// (2026-07-02): the array key is "invoiceFiles" (NOT "files"), each entry has
// id/versionNumber/pdfFileUrl. The old fixture used "files", which the raw-JSON
// dump masked; the table build reads "invoiceFiles", so a wrong key renders an
// empty table and fails the row assertions below.
func realFilesResponse() map[string]interface{} {
	return map[string]interface{}{
		"invoiceFiles": []map[string]interface{}{
			{
				"id":            "file-001",
				"versionNumber": 1770622200033,
				"pdfFileUrl":    "/v1/files/abc123",
			},
		},
		"success": true,
	}
}

func TestInvoiceFiles_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/invoices/inv-001/files", realFilesResponse())

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "files", "inv-001")
	require.NoError(t, err)
	// Table columns sourced from the real invoiceFiles entry.
	assert.Contains(t, stdout, "file-001")
	assert.Contains(t, stdout, "1770622200033", "versionNumber must render as a plain integer, not 1.77e+12")
	assert.Contains(t, stdout, "/v1/files/abc123")
}

// TestInvoiceFiles_JSON pins that --json still emits the raw response body (the
// machine contract is unchanged by the table conversion).
func TestInvoiceFiles_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/invoices/inv-001/files", realFilesResponse())

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "files", "inv-001", "--json")
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &got), "--json must emit valid JSON")
	arr, ok := got["invoiceFiles"].([]interface{})
	require.True(t, ok, "invoiceFiles array must be present in --json output")
	require.Len(t, arr, 1)
}

func TestInvoiceFiles_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "files")
	assert.Error(t, err)
}
