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
			"id":      "pay-001",
			"number":  "P-00000001", // real Payments field is "number" (see payment/get); "paymentNumber" never existed
			"amount":  100.00,
			"status":  "Processed",
			"success": true,
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "apply", "pay-001", "--body", `{"invoices":[{"invoiceId":"inv-001","amount":50}]}`)

	require.NoError(t, err)
	// Label-bound (F-08): each value under its OWN label, not merely present in
	// stdout — a wrong key rendering an empty row would pass a bare Contains
	// (the class that hid the refund-number P1 bug). (#432)
	assert.Regexp(t, `(?m)^ID:\s+pay-001$`, stdout)
	assert.Regexp(t, `(?m)^Payment Number:\s+P-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+100\.00$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Processed$`, stdout)
	assert.Contains(t, stderr, "Payment pay-001 applied.")
}

func TestPaymentApply_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "apply", "pay-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
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
