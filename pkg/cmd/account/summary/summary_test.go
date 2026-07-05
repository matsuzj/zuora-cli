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
			"defaultPaymentMethod": map[string]interface{}{
				"paymentMethodType": "CreditCardFixture", "id": "pm-777",
			},
		},
		"subscriptions": []interface{}{map[string]string{"id": "sub-1"}},
		"invoices":      []interface{}{map[string]string{"id": "inv-1"}, map[string]string{"id": "inv-2"}},
		"payments":      []interface{}{map[string]string{"id": "p-1"}, map[string]string{"id": "p-2"}, map[string]string{"id": "p-3"}},
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
	// Fixture-masking backfill (#482): pin every prod-read key under its label.
	assert.Regexp(t, `(?m)^ID:\s+id-1$`, stdout)
	assert.Regexp(t, `(?m)^Name:\s+Acme Corp$`, stdout)
	assert.Regexp(t, `(?m)^Account Number:\s+A001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Active$`, stdout)
	assert.Regexp(t, `(?m)^Currency:\s+USD$`, stdout)
	// getPaymentMethodSummary's type+id branch: "Type (id)" from the nested
	// basicInfo.defaultPaymentMethod object.
	assert.Regexp(t, `(?m)^Default Payment Method:\s+CreditCardFixture \(pm-777\)$`, stdout)
	// Array counts, label-bound (bare Contains "1"/"2" matches almost anything).
	assert.Regexp(t, `(?m)^Subscriptions:\s+1$`, stdout)
	assert.Regexp(t, `(?m)^Invoices:\s+2$`, stdout)
	assert.Regexp(t, `(?m)^Payments:\s+3$`, stdout)
}

func TestAccountSummary_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "summary", "bad-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
