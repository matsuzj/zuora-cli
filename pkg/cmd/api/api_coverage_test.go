package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	internalapi "github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPI_HttpClientError covers the early return when the factory cannot build
// an HTTP client (e.g. no active environment / missing credentials).
func TestAPI_HttpClientError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		HttpClient: func() (*internalapi.Client, error) {
			return nil, fmt.Errorf("no active environment configured")
		},
	}

	err := runAPI(&apiOptions{Factory: f, Method: "GET"}, "/v1/accounts")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no active environment configured")
}

// TestAPI_BodyResolveError covers the --body resolution error branch: a @file
// pointing at a nonexistent path must surface a clear error before any request.
func TestAPI_BodyResolveError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")

	missing := filepath.Join(t.TempDir(), "does-not-exist.json")
	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts", "-X", "POST", "--body", "@" + missing})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading body file")
}

// TestAPI_Paginate_Error covers the error return when a page request fails
// (DoPaginated propagates the underlying API error).
func TestAPI_Paginate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": "boom"})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts", "--paginate"})
	err := root.Execute()
	require.Error(t, err)
}

// TestAPI_Paginate_NonArrayPage covers the branch where an aggregated page is a
// JSON object rather than an array: it must be appended whole, not flattened.
// A single response with no "data" envelope and no "nextPage" yields exactly
// this shape from DoPaginated.
func TestAPI_Paginate_NonArrayPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "acct-1", "name": "Acme"})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts/acct-1", "--paginate"})
	require.NoError(t, root.Execute())

	output := out.String()
	assert.Contains(t, output, "acct-1")
	assert.Contains(t, output, "Acme")
}
