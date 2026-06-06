package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPI_Paginate covers the --paginate branch: multiple pages are fetched and
// their `data` arrays flattened into a single aggregated JSON array.
func TestAPI_Paginate(t *testing.T) {
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		resp := map[string]interface{}{
			"data": []map[string]string{{"id": fmt.Sprintf("acct-%d", page)}},
		}
		if page < 2 {
			resp["nextPage"] = fmt.Sprintf("/v1/accounts?page=%d", page+1)
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts", "--paginate"})
	require.NoError(t, root.Execute())

	output := out.String()
	assert.Contains(t, output, "acct-1")
	assert.Contains(t, output, "acct-2", "page 2 data must be aggregated into the output")
}

// TestAPI_Paginate_ObjectQueryRejected covers the guard that --paginate is not
// supported for Object Query endpoints (cursor-based, not URL-based pagination).
func TestAPI_Paginate_ObjectQueryRejected(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/object-query/accounts", "--paginate"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Object Query")
}

// TestAPI_InvalidHeader covers the malformed -H value guard.
func TestAPI_InvalidHeader(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/test", "-H", "NoColonHeader"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid header format")
}

// TestAPI_Template covers the --template output branch.
func TestAPI_Template(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "acct-9", "name": "Acme"})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts/acct-9", "--template", "{{.name}}"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "Acme")
}
