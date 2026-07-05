package transfer

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdTransfer(f) }

func TestPaymentTransfer_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/payments/pay-001/transfer", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		// The --body payload must reach the server intact (#484): the handler
		// previously ignored r.Body.
		body, rerr := io.ReadAll(r.Body)
		if assert.NoError(t, rerr) {
			assert.JSONEq(t, `{"accountId":"acc-002"}`, string(body))
		}
		cmdtest.OK(t, "", "", map[string]interface{}{
			"id":        "pay-001",
			"number":    "P-00000001", // real Payments field is "number" (see payment/get); "paymentNumber" never existed
			"accountId": "acc-002",
			"amount":    100.00,
			"status":    "Processed",
			"success":   true,
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "transfer", "pay-001", "--body", `{"accountId":"acc-002"}`)

	require.NoError(t, err)
	assert.Regexp(t, `(?m)^ID:\s+pay-001$`, stdout)
	assert.Regexp(t, `(?m)^Payment Number:\s+P-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Account ID:\s+acc-002$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+100\.00$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Processed$`, stdout)
	assert.Contains(t, stderr, "Payment pay-001 transferred.")
}

func TestPaymentTransfer_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "transfer", "pay-001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPaymentTransfer_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "transfer")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestPaymentTransfer_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53830020, "Cannot transfer payment")

	_, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "transfer", "pay-001", "--body", `{"accountId":"acc-002"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot transfer payment")
}
