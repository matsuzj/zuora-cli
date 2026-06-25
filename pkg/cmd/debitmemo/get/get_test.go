package get

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestDebitMemoGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/debitmemos/dm-001", map[string]interface{}{
		"id":            "dm-001",
		"number":        "DM00001",
		"debitMemoDate": "2026-01-15",
		"amount":        110.00,
		"balance":       110.00,
		"status":        "Posted",
		"accountNumber": "A00000001",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "get", "dm-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Number:\s+DM00001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Posted$`, stdout)
}

func TestDebitMemoGet_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id":      "dm-001",
		"number":  "DM00001",
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "get", "dm-001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"number"`)
}

func TestDebitMemoGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "debitmemo", newCmd, nil, "debitmemo", "get")
	assert.Error(t, err)
}

func TestDebitMemoGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Debit memo not found")

	_, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "get", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Debit memo not found")
}
