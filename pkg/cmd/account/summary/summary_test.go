package summary

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSummary(f) }

func TestAccountSummary_Detail(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/accounts/A001/summary", map[string]interface{}{
		"basicInfo": map[string]interface{}{
			"id": "id-1", "name": "Acme Corp", "accountNumber": "A001",
			"status": "Active", "balance": 100.0, "currency": "USD",
		},
		"subscriptions": []interface{}{map[string]string{"id": "sub-1"}},
		"invoices":      []interface{}{map[string]string{"id": "inv-1"}, map[string]string{"id": "inv-2"}},
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "summary", "A001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Acme Corp")
	assert.Contains(t, stdout, "Subscriptions")
	assert.Contains(t, stdout, "1")
	assert.Contains(t, stdout, "Invoices")
	assert.Contains(t, stdout, "2")
	// Balance is numeric (100.0) in the fixture, so this exercises GetMoney's
	// float -> %.2f contract (a GetMoney -> GetDecimal swap would render "100").
	assert.Regexp(t, `(?m)^Balance:\s+100\.00$`, stdout)
}

func TestAccountSummary_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "summary", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
