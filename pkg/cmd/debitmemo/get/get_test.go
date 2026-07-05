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
		"dueDate":       "2026-02-20",
		"amount":        110.00,
		"balance":       110.00,
		"taxAmount":     10.00,
		"status":        "Posted",
		"currency":      "EUR",
		"reasonCode":    "Charge Dispute",
		"accountId":     "acc-dm-042",
		"accountNumber": "A00000001",
		"createdDate":   "2026-01-11T08:15:00Z",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "get", "dm-001")
	require.NoError(t, err)
	// Label-bound (F-08): values under their own labels.
	assert.Regexp(t, `(?m)^Number:\s+DM00001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Posted$`, stdout)
	// Tax Amount must format with 2 decimals via GetMoney (10.00, not "10").
	// Bites if production reverts to GetDecimal, which renders "10". (#423)
	assert.Regexp(t, `(?m)^Tax Amount:\s+10\.00$`, stdout)
	// Every remaining prod-read key is pinned with a distinctive value so a key
	// typo or nesting mistake renders "" and fails here (fixture-masking, #482).
	assert.Regexp(t, `(?m)^ID:\s+dm-001$`, stdout)
	assert.Regexp(t, `(?m)^Debit Memo Date:\s+2026-01-15$`, stdout)
	assert.Regexp(t, `(?m)^Due Date:\s+2026-02-20$`, stdout)
	assert.Regexp(t, `(?m)^Amount:\s+110\.00$`, stdout)  // money
	assert.Regexp(t, `(?m)^Balance:\s+110\.00$`, stdout) // money
	assert.Regexp(t, `(?m)^Currency:\s+EUR$`, stdout)
	assert.Regexp(t, `(?m)^Reason Code:\s+Charge Dispute$`, stdout)
	assert.Regexp(t, `(?m)^Account ID:\s+acc-dm-042$`, stdout)
	assert.Regexp(t, `(?m)^Account Number:\s+A00000001$`, stdout)
	assert.Regexp(t, `(?m)^Created Date:\s+2026-01-11T08:15:00Z$`, stdout)
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
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestDebitMemoGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Debit memo not found")

	_, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "get", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Debit memo not found")
}
