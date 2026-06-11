package apply

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdApply(f) }

func TestPaymentApply_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/payments/pay-001/apply", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		cmdtest.OK(t, "", "", map[string]interface{}{
			"id":            "pay-001",
			"paymentNumber": "P-00000001",
			"amount":        100.00,
			"status":        "Processed",
			"success":       true,
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "apply", "pay-001", "--body", `{"invoices":[{"invoiceId":"inv-001","amount":50}]}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "pay-001")
	assert.Contains(t, stderr, "Payment pay-001 applied.")
}

func TestPaymentApply_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "apply", "pay-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestPaymentApply_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "apply")

	assert.Error(t, err)
}

func TestPaymentApply_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Cannot apply payment")

	_, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "apply", "pay-001", "--body", `{"invoices":[]}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot apply payment")
}
