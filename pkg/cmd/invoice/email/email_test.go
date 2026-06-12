package email

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdEmail(f) }

func TestInvoiceEmail_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/invoices/inv-001/emails", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		cmdtest.OK(t, "", "", map[string]interface{}{
			"success": true,
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "email", "inv-001", "--body", `{"emailAddresses":"user@example.com"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "true")
	assert.Contains(t, stderr, "Invoice inv-001 email sent.")
}

func TestInvoiceEmail_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "email", "inv-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestInvoiceEmail_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "invoice", newCmd, nil, "invoice", "email")

	assert.Error(t, err)
}

func TestInvoiceEmail_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Email sending failed")

	_, _, err := cmdtest.Run(t, "invoice", newCmd, handler, "invoice", "email", "inv-001", "--body", `{"emailAddresses":"user@example.com"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Email sending failed")
}
