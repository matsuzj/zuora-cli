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
					"id":             "cm-001",
					"number":         "CM00001",
					"creditMemoDate": "2026-01-15",
					"amount":         100.50,
					"balance":        25.25,
					"status":         "Posted",
					"accountNumber":  "A00000001",
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
			{"id": "cm-001", "number": "CM00001", "amount": 100.5, "balance": 25.25, "status": "Posted", "accountNumber": "A1"},
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
