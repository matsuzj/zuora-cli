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

func TestAccountList_Table(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/object-query/accounts", r.URL.Path)
		assert.Equal(t, "10", r.URL.Query().Get("pageSize"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "id-1", "name": "Acme Corp", "accountNumber": "A001", "status": "Active", "balance": 100.50, "createdDate": "2025-01-01"},
			},
		})
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "list", "--page-size", "10")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Acme Corp")
	assert.Contains(t, stdout, "A001")
	assert.Contains(t, stdout, "Active")
}

func TestAccountList_JSON(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": "id-1", "name": "Acme"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "list", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"name"`)
}

func TestAccountList_Filter(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filters := r.URL.Query()["filter[]"]
		assert.Equal(t, []string{"status.EQ:Active"}, filters)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	})

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "list", "--filter", "status.EQ:Active")
	require.NoError(t, err)
}

func TestAccountList_JQ(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"data": []map[string]interface{}{
			{"id": "id-1", "name": "Acme"},
			{"id": "id-2", "name": "Beta"},
		},
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "list", "--jq", ".data[].name")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Acme")
	assert.Contains(t, stdout, "Beta")
}

func TestAccountList_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Account not found")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Account not found")
}
