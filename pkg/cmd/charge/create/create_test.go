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

func TestChargeCreate_Success(t *testing.T) {
	handler := cmdtest.Expect{
		Method:   "POST",
		Path:     "/commerce/charges",
		Headers:  map[string]string{"Content-Type": "application/json"},
		JSONBody: `{"name":"Monthly Charge","plan_id":"plan-001"}`,
		Respond: map[string]interface{}{
			"id":   "chg-001",
			"name": "Monthly Charge",
		},
	}.Handler(t)

	stdout, stderr, err := cmdtest.Run(t, "charge", newCmd, handler,
		"charge", "create", "--body", `{"name":"Monthly Charge","plan_id":"plan-001"}`)

	require.NoError(t, err)
	assert.Contains(t, stdout, "chg-001")
	assert.Contains(t, stdout, "Monthly Charge")
	assert.Contains(t, stderr, "Charge created.")
}

func TestChargeCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "create")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestChargeCreate_BareCSVRejectedBeforePost(t *testing.T) {
	// --csv on a JSON-only write must be rejected BEFORE any HTTP call — a
	// rejected-then-retried create could otherwise double-create. nil handler =
	// unexpected requests fail loudly; surfacing the CSV error (not a connection
	// error) proves no POST was attempted.
	_, _, err := cmdtest.Run(t, "charge", newCmd, nil, "charge", "create", "--body", `{"name":"X"}`, "--csv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--csv is not supported for JSON-only output")
}
