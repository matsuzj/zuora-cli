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

func TestAccountGet_Detail(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001", map[string]interface{}{
		"basicInfo": map[string]interface{}{
			"id": "id-1", "name": "Acme Corp", "accountNumber": "A001",
			"status": "Active",
		},
		"billingAndPayment": map[string]interface{}{
			"autoPay": true, "billCycleDay": 1, "currency": "USD",
		},
		"metrics": map[string]interface{}{
			"balance": "250.00",
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Acme Corp")
	assert.Contains(t, stdout, "Active")
	assert.Contains(t, stdout, "250.00")
}

func TestAccountGet_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"id": "id-1", "name": "Acme",
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "A001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"name"`)
}

func TestAccountGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "account", newCmd, nil, "account", "get")
	assert.Error(t, err)
}

func TestAccountGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
