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

func TestCreditMemoGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/creditmemos/cm-001", map[string]interface{}{
		"id":             "cm-001",
		"number":         "CM00001",
		"creditMemoDate": "2026-01-15",
		"amount":         100.50,
		"balance":        25.25,
		"status":         "Posted",
		"accountNumber":  "A00000001",
		"success":        true,
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "get", "cm-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Number:\s+CM00001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Posted$`, stdout)
}

func TestCreditMemoGet_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id":      "cm-001",
		"number":  "CM00001",
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "get", "cm-001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"number"`)
}

func TestCreditMemoGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "creditmemo", newCmd, nil, "creditmemo", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestCreditMemoGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Credit memo not found")

	_, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "get", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Credit memo not found")
}
