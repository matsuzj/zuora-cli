package list

import (
	"encoding/json"
	httptest "github.com/matsuzj/zuora-cli/internal/testutil/httpmock"
	"net/http"
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
	sub := &cobra.Command{Use: "contact"}
	sub.AddCommand(NewCmdList(f))
	root.AddCommand(sub)
	return root
}

func TestContactList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/action/query", r.URL.Path)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": []map[string]interface{}{
				{"Id": "c-1", "FirstName": "John", "LastName": "Doe", "WorkEmail": "j@example.com"},
			},
			"size": 1,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "list", "--account-id", "acct-123"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "John")
	assert.Contains(t, out.String(), "Doe")
	assert.Contains(t, out.String(), "j@example.com")
}

func TestContactList_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		if callCount == 1 {
			assert.Equal(t, "/v1/action/query", r.URL.Path)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"records": []map[string]interface{}{
					{"Id": "c-1", "FirstName": "Page1", "LastName": "User", "WorkEmail": "p1@example.com"},
				},
				"size":         1,
				"done":         false,
				"queryLocator": "loc-abc",
			})
		} else {
			assert.Equal(t, "/v1/action/queryMore", r.URL.Path)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"records": []map[string]interface{}{
					{"Id": "c-2", "FirstName": "Page2", "LastName": "User", "WorkEmail": "p2@example.com"},
				},
				"size": 1,
				"done": true,
			})
		}
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "list", "--account-id", "acct-123"})
	require.NoError(t, root.Execute())
	assert.Equal(t, 2, callCount)
	assert.Contains(t, out.String(), "Page1")
	assert.Contains(t, out.String(), "Page2")
}

func TestContactList_Pagination_JSON(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"records":      []map[string]interface{}{{"Id": "c-1", "FirstName": "A", "LastName": "B", "WorkEmail": "a@b.com"}},
				"size":         1,
				"done":         false,
				"queryLocator": "loc-xyz",
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"records": []map[string]interface{}{{"Id": "c-2", "FirstName": "C", "LastName": "D", "WorkEmail": "c@d.com"}},
				"size":    1,
				"done":    true,
			})
		}
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")

	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "list", "--account-id", "acct-123", "--json"})
	require.NoError(t, root.Execute())
	// JSON output should contain both records
	assert.Contains(t, out.String(), "c-1")
	assert.Contains(t, out.String(), "c-2")
}

func TestContactList_RequiresAccountID(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"contact", "list"})
	assert.Error(t, root.Execute())
}
