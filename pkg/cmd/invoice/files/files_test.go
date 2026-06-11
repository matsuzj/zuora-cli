package files

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdFiles(f) }

func TestInvoiceFiles_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/invoices/inv-001/files", map[string]interface{}{
		"success": true,
		"files": []map[string]interface{}{
			{"id": "file-001", "pdfFileUrl": "https://example.com/file.pdf"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "files", "inv-001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "file-001")
}

func TestInvoiceFiles_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "files")
	assert.Error(t, err)
}
