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

func TestCreditMemoList_StatusCompletionUsesCanonicalSpelling(t *testing.T) {
	// Zuora returns the US spelling "Canceled" for credit-memo status
	// (live-verified). Offering "Cancelled" (British) in completion sends a
	// value that matches no records and returns silently empty. (#422)
	cmd := NewCmdList(&factory.Factory{})
	fn, ok := cmd.GetFlagCompletionFunc("status")
	require.True(t, ok, "status flag must register a completion func")
	vals, _ := fn(cmd, nil, "")
	assert.Contains(t, vals, "Canceled")
	assert.NotContains(t, vals, "Cancelled")
}

func TestCreditMemoList_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/creditmemos", r.URL.Path)
		assert.Equal(t, "A00000001", r.URL.Query().Get("accountNumber"))
		assert.Equal(t, "Posted", r.URL.Query().Get("status"))
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"creditmemos": []map[string]interface{}{
				{
					"id":              "cm-001",
					"number":          "CM00001",
					"creditMemoDate":  "2026-01-15",
					"amount":          100.50,
					"unappliedAmount": 25.25,
					"status":          "Posted",
					"accountNumber":   "A00000001",
				},
			},
			"success": true,
		})
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "list", "--account-number", "A00000001", "--status", "Posted")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CM00001")
	assert.Contains(t, stdout, "Posted")
}

func TestCreditMemoList_NoFilter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No filters → no query params besides none required.
		assert.Empty(t, r.URL.Query().Get("accountId"))
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"creditmemos": []map[string]interface{}{
				{"id": "cm-001", "number": "CM00001", "amount": 10.0, "status": "Draft"},
			},
			"success": true,
		})
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "list")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CM00001")
}

func TestCreditMemoList_CSV(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/creditmemos", map[string]interface{}{
		"creditmemos": []map[string]interface{}{
			{"id": "cm-001", "number": "CM00001", "amount": 100.5, "unappliedAmount": 25.25, "status": "Posted", "accountNumber": "A1"},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "list", "--csv")
	require.NoError(t, err)
	assert.Contains(t, stdout, "CM00001")
	assert.Contains(t, stdout, ",")
}

func TestCreditMemoList_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Invalid account")

	_, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "list", "--account-id", "BAD")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid account")
}

func TestCreditMemoList_AccountIDMapsToAccountIdQuery(t *testing.T) {
	// --account-id must populate the accountId query param (NOT accountNumber) —
	// the flag-vocabulary contract AGENTS.md warns is easy to mis-wire.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "acct-123", r.URL.Query().Get("accountId"))
		assert.Empty(t, r.URL.Query().Get("accountNumber"), "--account-id must not populate accountNumber")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"creditmemos": []map[string]interface{}{}, "success": true})
	})

	_, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "list", "--account-id", "acct-123")
	require.NoError(t, err)
}

func TestCreditMemoList_CSVHeaderAndColumnOrder(t *testing.T) {
	// The CSV header row and column ORDER are a compatibility contract for scripts
	// (cut -d, / fixed columns). Pin them so a Columns reorder/rename is a visible,
	// reviewed change rather than a silent break.
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"creditmemos": []map[string]interface{}{
			{"id": "cm-001", "number": "CM00001", "creditMemoDate": "2026-01-15", "amount": 100.5, "unappliedAmount": 25.25, "status": "Posted", "accountNumber": "A1"},
		},
		"success": true,
	})

	stdout, _, err := cmdtest.Run(t, "creditmemo", newCmd, handler, "creditmemo", "list", "--csv")
	require.NoError(t, err)

	rows, err := csv.NewReader(strings.NewReader(stdout)).ReadAll()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(rows), 2, "CSV must have a header row plus the data row")
	assert.Equal(t, []string{"ID", "NUMBER", "DATE", "AMOUNT", "BALANCE", "STATUS", "ACCOUNT"}, rows[0])
	assert.Equal(t, "CM00001", rows[1][1]) // NUMBER
	assert.Equal(t, "100.50", rows[1][3])  // AMOUNT (Money)
	assert.Equal(t, "25.25", rows[1][4])   // BALANCE (Money)
	assert.Equal(t, "Posted", rows[1][5])  // STATUS
}
