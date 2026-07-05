package cancel

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCancel(f) }

func TestPaymentCancel_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/payments/pay-001/cancel", r.URL.Path)
		cmdtest.OK(t, "", "", map[string]interface{}{
			"id":      "pay-001",
			"number":  "P-00000001", // real Payments field is "number" (see payment/get); "paymentNumber" never existed
			"amount":  100.00,
			"status":  "Canceled",
			"success": true,
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "cancel", "pay-001", "--confirm")

	require.NoError(t, err)
	assert.Regexp(t, `(?m)^ID:\s+pay-001$`, stdout)
	assert.Regexp(t, `(?m)^Payment Number:\s+P-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+100\.00$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Canceled$`, stdout)
	assert.Contains(t, stderr, "Payment pay-001 cancelled.")
}

// TestPaymentCancel_RequiresConfirm pins the irreversible-write guard: without
// --confirm the command must refuse before issuing any request.
func TestPaymentCancel_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "payment", newCmd, "payment", "cancel", "pay-001")
}

func TestPaymentCancel_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "cancel")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

// TestPaymentCancel_SendsEmptyJSONBody pins the 415 contract: the bodyless
// lifecycle PUT must still carry Content-Type + "{}" (cf. invoice/billrun post).
func TestPaymentCancel_SendsEmptyJSONBody(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "{}", string(b))
		cmdtest.OK(t, "", "", map[string]interface{}{"id": "pay-001", "status": "Canceled", "success": true})(w, r)
	}

	_, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "cancel", "pay-001", "--confirm")
	require.NoError(t, err)
}

func TestPaymentCancel_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Cannot cancel payment")

	_, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "cancel", "pay-001", "--confirm")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot cancel payment")
}
