package list

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	root.PersistentFlags().Bool("csv", false, "")
	dm := &cobra.Command{Use: "debitmemo"}
	dm.AddCommand(NewCmdList(f))
	root.AddCommand(dm)
	return root
}

func TestDebitMemoList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"debitmemo", "list", "--account-number", "A00000001"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "DM00001")
	assert.Contains(t, out.String(), "Posted")
}

func TestDebitMemoList_CSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"debitmemos": []map[string]interface{}{
				{"id": "dm-001", "number": "DM00001", "amount": 110.0, "balance": 110.0, "status": "Posted", "accountNumber": "A1"},
			},
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"debitmemo", "list", "--csv"})
	require.NoError(t, root.Execute())
	output := out.String()
	assert.Contains(t, output, "DM00001")
	assert.Contains(t, output, ",")
}

func TestDebitMemoList_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 50000040, "message": "Invalid account"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"debitmemo", "list", "--account-id", "BAD"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid account")
}
