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
			// Numeric so GetMoney's float -> %.2f contract is actually exercised:
			// a string would pass straight through and never bite a
			// GetMoney -> GetDecimal regression.
			"balance": 1234.5,
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "A001")
	require.NoError(t, err)
	// Bind each value to its label (not a bare substring) and source it from the
	// nested key, so a wrong key or wrong render helper renders empty / mis-formats
	// and fails here.
	assert.Regexp(t, `(?m)^Name:\s+Acme Corp$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Active$`, stdout)
	assert.Regexp(t, `(?m)^Balance:\s+1234\.50$`, stdout) // GetMoney: float -> two decimals
	assert.Regexp(t, `(?m)^Auto Pay:\s+true$`, stdout)    // GetBool
	assert.Regexp(t, `(?m)^Bill Cycle Day:\s+1$`, stdout) // GetInt
	assert.Regexp(t, `(?m)^Currency:\s+USD$`, stdout)
}

func TestAccountGet_CurrencyFallsBackToMetrics(t *testing.T) {
	// F-18: currency placement varies. When billingAndPayment carries no currency,
	// fall back to metrics so the Currency row isn't blank (a real account has it
	// in both billingAndPayment and metrics; verified by live probe).
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001", map[string]interface{}{
		"basicInfo":         map[string]interface{}{"name": "Acme", "status": "Active"},
		"billingAndPayment": map[string]interface{}{"autoPay": true}, // no currency here
		"metrics":           map[string]interface{}{"balance": 0.0, "currency": "JPY"},
		"success":           true,
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "A001")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Currency:\s+JPY$`, stdout, "currency must fall back to metrics when billingAndPayment lacks it")
}

func TestAccountGet_TypeMismatchRendersEmpty(t *testing.T) {
	// GetInt silently drops a type-mismatched value (the helper contract): a
	// string billCycleDay must render an EMPTY Bill Cycle Day, not the raw string.
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001", map[string]interface{}{
		"basicInfo":         map[string]interface{}{"name": "Acme"},
		"billingAndPayment": map[string]interface{}{"billCycleDay": "7"}, // string, not int
		"success":           true,
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "A001")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Bill Cycle Day:\s*$`, stdout, "type-mismatched billCycleDay must render empty")
	assert.NotContains(t, stdout, "7")
}

func TestAccountGet_JSON(t *testing.T) {
	// --json is a raw passthrough, so feed the REAL nested envelope (the same
	// shape the detail view unwraps) and assert the nested keys survive verbatim.
	// A flat {id,name} fixture never reflects what `account get --json` emits.
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001", map[string]interface{}{
		"basicInfo": map[string]interface{}{
			"id": "id-1", "name": "Acme Corp", "accountNumber": "A001",
		},
		"billingAndPayment": map[string]interface{}{"currency": "USD"},
		"metrics":           map[string]interface{}{"balance": 1234.5},
		"success":           true,
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "A001", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"basicInfo"`)
	assert.Contains(t, stdout, `"billingAndPayment"`)
	assert.Contains(t, stdout, `"Acme Corp"`)
}

func TestAccountGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "account", newCmd, nil, "account", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestAccountGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "get", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
