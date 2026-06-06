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
	cm := &cobra.Command{Use: "creditmemo"}
	cm.AddCommand(NewCmdList(f))
	root.AddCommand(cm)
	return root
}

func TestCreditMemoList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"creditmemo", "list", "--account-number", "A00000001", "--status", "Posted"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "CM00001")
	assert.Contains(t, out.String(), "Posted")
}

func TestCreditMemoList_NoFilter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No filters → no query params besides none required.
		assert.Empty(t, r.URL.Query().Get("accountId"))
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"creditmemos": []map[string]interface{}{
				{"id": "cm-001", "number": "CM00001", "amount": 10.0, "status": "Draft"},
			},
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"creditmemo", "list"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "CM00001")
}

func TestCreditMemoList_CSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"creditmemos": []map[string]interface{}{
				{"id": "cm-001", "number": "CM00001", "amount": 100.5, "balance": 25.25, "status": "Posted", "accountNumber": "A1"},
			},
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"creditmemo", "list", "--csv"})
	require.NoError(t, root.Execute())
	output := out.String()
	assert.Contains(t, output, "CM00001")
	assert.Contains(t, output, ",")
}

func TestCreditMemoList_SuccessFalse(t *testing.T) {
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
	root.SetArgs([]string{"creditmemo", "list", "--account-id", "BAD"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid account")
}
