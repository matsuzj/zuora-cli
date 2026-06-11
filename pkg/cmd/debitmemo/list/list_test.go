package list

import (
	"encoding/json"
	"net/http"
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
