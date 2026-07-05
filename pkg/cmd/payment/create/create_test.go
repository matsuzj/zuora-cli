package create

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestPaymentCreate_Success(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/payments", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		cmdtest.OK(t, "", "", map[string]interface{}{
			"id":      "pay-001",
			"number":  "P-00000001", // real Payments field is "number" (see payment/get); "paymentNumber" never existed
			"amount":  100.00,
			"status":  "Processed",
			"success": true,
		})(w, r)
	}

	stdout, stderr, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "create", "--body", `{"amount":100,"accountId":"acc-001"}`)

	require.NoError(t, err)
	// Label-bound (F-08): each value under its OWN label — a wrong key rendering
	// an empty row would pass a bare Contains. (#432)
	assert.Regexp(t, `(?m)^ID:\s+pay-001$`, stdout)
	assert.Regexp(t, `(?m)^Payment Number:\s+P-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+100\.00$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Processed$`, stdout)
	assert.Contains(t, stderr, "Payment pay-001 created.")
}

func TestPaymentCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "payment", newCmd, handler, "payment", "create", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestPaymentCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestPaymentCreate_RejectsArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "payment", newCmd, nil, "payment", "create", "extra-arg", "--body", `{}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `unknown command "extra-arg"`)
}
