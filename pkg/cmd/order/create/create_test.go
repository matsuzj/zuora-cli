package create

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestOrderCreate_Success(t *testing.T) {
	handler := cmdtest.Expect{
		Method:   "POST",
		Path:     "/v1/orders",
		JSONBody: `{"existingAccountNumber":"A001"}`,
		Respond: map[string]interface{}{
			"success":       true,
			"orderNumber":   "O-00000001",
			"accountNumber": "ACCT-CREATE-7777777",
			"status":        "Pending",
		},
	}.Handler(t)

	stdout, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "create", "--body", `{"existingAccountNumber":"A001"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")
	assert.Contains(t, stderr, "Order O-00000001 created.")
	// Label-bound (F-08) asserts for the previously unfixtured keys (#482):
	// accountNumber and status never appeared in any fixture, so a key typo
	// would render "" while the test stayed green.
	assert.Regexp(t, `(?m)^Order Number:\s+O-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Account Number:\s+ACCT-CREATE-7777777$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Pending$`, stdout)
	// success is a JSON boolean read via GetString — renders fmt %v "true".
	assert.Regexp(t, `(?m)^Success:\s+true$`, stdout)
}

func TestOrderCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Missing required field")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "create", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required field")
}

func TestOrderCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
