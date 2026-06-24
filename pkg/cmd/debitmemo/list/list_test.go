package list

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestDebitMemoList_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/debitmemos", r.URL.Path)
		assert.Equal(t, "A00000001", r.URL.Query().Get("accountNumber"))
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"debitmemos": []map[string]interface{}{
				{
					"id":            "dm-001",
					"number":        "DM00001",
					"debitMemoDate": "2026-01-15",
					"amount":        110.00,
					"balance":       110.00,
					"status":        "Posted",
					"accountNumber": "A00000001",
				},
			},
			"success": true,
		})
	})

	stdout, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "list", "--account-number", "A00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "DM00001")
	assert.Contains(t, stdout, "Posted")
}

func TestDebitMemoList_CSV(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"debitmemos": []map[string]interface{}{
			{"id": "dm-001", "number": "DM00001", "amount": 110.0, "balance": 110.0, "status": "Posted", "accountNumber": "A1"},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "list", "--csv")
	require.NoError(t, err)
	assert.Contains(t, stdout, "DM00001")
	assert.Contains(t, stdout, ",")
}

func TestDebitMemoList_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Invalid account")

	_, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "list", "--account-id", "BAD")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid account")
}

func TestDebitMemoList_AccountIDMapsToAccountIdQuery(t *testing.T) {
	// --account-id must populate the accountId query param (NOT accountNumber) —
	// the flag-vocabulary contract AGENTS.md warns is easy to mis-wire.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "acct-123", r.URL.Query().Get("accountId"))
		assert.Empty(t, r.URL.Query().Get("accountNumber"), "--account-id must not populate accountNumber")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"debitmemos": []map[string]interface{}{}, "success": true})
	})

	_, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "list", "--account-id", "acct-123")
	require.NoError(t, err)
}

func TestDebitMemoList_CSVHeaderAndColumnOrder(t *testing.T) {
	// The CSV header row and column ORDER are a compatibility contract for scripts
	// (cut -d, / fixed columns). Pin them so a Columns reorder/rename is a visible,
	// reviewed change rather than a silent break.
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"debitmemos": []map[string]interface{}{
			{"id": "dm-001", "number": "DM00001", "debitMemoDate": "2026-01-15", "amount": 110.0, "balance": 95.5, "status": "Posted", "accountNumber": "A1"},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "debitmemo", newCmd, handler, "debitmemo", "list", "--csv")
	require.NoError(t, err)

	rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rows), 2, "CSV must have a header row plus the data row")
	assert.Equal(t, []string{"ID", "NUMBER", "DATE", "AMOUNT", "BALANCE", "STATUS", "ACCOUNT"}, rows[0])
	assert.Equal(t, "DM00001", rows[1][1]) // NUMBER
	assert.Equal(t, "110.00", rows[1][3])  // AMOUNT (Money: two decimals)
	assert.Equal(t, "95.50", rows[1][4])   // BALANCE (Money)
	assert.Equal(t, "Posted", rows[1][5])  // STATUS
}
