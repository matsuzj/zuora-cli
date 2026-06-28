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

func TestContactList_RejectsStrayArg(t *testing.T) {
	// contact list filters via flags and takes no positional args; a stray
	// positional must be rejected (cobra.NoArgs), not silently ignored.
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "list", "stray", "--account-id", "acct-123")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown command "stray"`)
}

func TestContactList_Success(t *testing.T) {
	handler := cmdtest.OK(t, "POST", "/v1/action/query", map[string]interface{}{
		"records": []map[string]interface{}{
			{"Id": "c-1", "FirstName": "John", "LastName": "Doe", "WorkEmail": "j@example.com"},
		},
		"size": 1,
	})

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "list", "--account-id", "acct-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "John")
	assert.Contains(t, stdout, "Doe")
	assert.Contains(t, stdout, "j@example.com")
}

func TestContactList_RejectsZOQLInjection(t *testing.T) {
	// A crafted --account-id must be rejected BEFORE any query is sent. The nil
	// handler means reaching the API yields a connection error, not "invalid
	// --account-id" — so the message assert confirms the *validation* rejected it.
	for _, bad := range []string{
		`x' OR '1'='1`,
		`x' OR Id != '`,
		`acct'; DELETE FROM Contact WHERE Id != '`,
		`has space`,
		`quote'inside`,
	} {
		_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "list", "--account-id", bad)
		require.Error(t, err, "account-id %q must be rejected", bad)
		assert.Contains(t, err.Error(), "invalid --account-id", "must fail validation, not reach the API: %q", bad)
	}
}

func TestContactList_Pagination(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "list", "--account-id", "acct-123")
	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	assert.Contains(t, stdout, "Page1")
	assert.Contains(t, stdout, "Page2")
}

func TestContactList_Pagination_JSON(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	})

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "list", "--account-id", "acct-123", "--json")
	require.NoError(t, err)
	// JSON output should contain both records
	assert.Contains(t, stdout, "c-1")
	assert.Contains(t, stdout, "c-2")
}

// TestContactList_SuccessFalse_IsError pins that an action/query returning
// HTTP 200 with {"success":false} (e.g. invalid ZOQL) surfaces as an error
// rather than silently printing zero contacts. Guards the WithCheckSuccess wiring.
func TestContactList_SuccessFalse_IsError(t *testing.T) {
	handler := cmdtest.Reasons(t, "INVALID_FIELD", "invalid query")

	_, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "list", "--account-id", "acct-123")
	require.Error(t, err, "success:false from action/query must surface as an error")
	assert.Contains(t, err.Error(), "invalid query")
}

func TestContactList_RequiresAccountID(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "list")
	assert.Error(t, err)
}
