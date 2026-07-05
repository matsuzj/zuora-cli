package unapply

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUnapply(f) }

func TestPaymentUnapply_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/payments/pay-001/unapply", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		cmdtest.OK(t, "", "", map[string]interface{}{
			"id":              "pay-001",
			"number":          "P-00000001", // real Payments field is "number" (see payment/get); "paymentNumber" never existed
			"amount":          100.00,
			"unappliedAmount": 50.00,
			"status":          "Processed",
			"success":         true,
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "unapply", "pay-001", "--body", `{"invoices":[{"invoiceId":"inv-001","amount":50}]}`)

	require.NoError(t, err)
	assert.Regexp(t, `(?m)^ID:\s+pay-001$`, stdout)
	assert.Regexp(t, `(?m)^Payment Number:\s+P-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+100\.00$`, stdout)
	assert.Regexp(t, `(?m)^Unapplied Amount:\s+50\.00$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Processed$`, stdout)
	assert.Contains(t, stderr, "Payment pay-001 unapplied.")
}

func TestPaymentUnapply_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "unapply", "pay-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPaymentUnapply_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "unapply")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestPaymentUnapply_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53840020, "Cannot unapply payment")

	_, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "unapply", "pay-001", "--body", `{"invoices":[]}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot unapply payment")
}
