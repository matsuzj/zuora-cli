package refund

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdRefund(f) }

func TestPaymentRefund_Success(t *testing.T) {
	// Expect (not OK) so the --body payload is asserted to reach the server
	// intact (#484): the previous handler ignored r.Body entirely.
	handler := cmdtest.Expect{
		Method:   "POST",
		Path:     "/v1/payments/pay-001/refunds",
		JSONBody: `{"amount":50,"type":"External"}`,
		Respond: map[string]interface{}{
			"id":      "ref-001",
			"number":  "R-00001",
			"amount":  50.00,
			"status":  "Processed",
			"success": true,
		},
	}.Handler(t)

	stdout, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "refund", "pay-001", "--body", `{"amount":50,"type":"External"}`, "--confirm")
	require.NoError(t, err)
	// Refund Number is sourced from "number" (live-verified; "refundNumber"
	// never existed). Bites if production reverts to the wrong key. (#420)
	assert.Regexp(t, `(?m)^Refund Number:\s+R-00001$`, stdout)
	// Label-bound Amount/Status — a wrong key rendering an empty row would pass
	// a bare Contains. (#432)
	assert.Regexp(t, `(?m)^Amount:\s+50\.00$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Processed$`, stdout)
}

func TestPaymentRefund_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "refund", "pay-001")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPaymentRefund_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "payment", newCmd, "payment", "refund", "pay-001", "--body", `{"amount":50,"type":"External"}`)
}
