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
// (DoPaginated propagates the underlying API error). Uses a non-retriable 404
// (not 5xx) so the GET retry/backoff loop doesn't sleep through a coverage-only
// path, and pins the surfaced message so the assertion is tied to this path.
func TestAPI_Paginate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"message": "boom"})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/accounts", "--paginate"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
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

	// The page must be aggregated as a single whole element (the object), not
	// flattened into the top level or dropped: the output is a JSON array of
	// length 1 whose element is the original object.
	var agg []map[string]interface{}
	require.NoError(t, json.Unmarshal(out.Bytes(), &agg))
	require.Len(t, agg, 1)
	assert.Equal(t, "acct-1", agg[0]["id"])
	assert.Equal(t, "Acme", agg[0]["name"])
}

// TestAPI_Write_SuccessFalse_Errors covers that a mutating method surfaces an
// HTTP-200 {"success":false} envelope as an error (non-zero exit), matching the
// typed write commands instead of silently exiting 0.
func TestAPI_Write_SuccessFalse_Errors(t *testing.T) {
	for _, method := range []string{"POST", "PUT", "PATCH", "DELETE"} {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, method, r.Method)
				w.WriteHeader(200)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"reasons": []map[string]interface{}{
						{"code": 50000040, "message": "write rejected"},
					},
				})
			}))
			defer server.Close()

			ios, _, _, _ := iostreams.Test()
			f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

			root := newTestRoot(f)
			root.SetArgs([]string{"api", "/v1/things", "-X", method})
			err := root.Execute()
			require.Error(t, err, "%s with 200 success:false must error", method)
			assert.Contains(t, err.Error(), "write rejected")
		})
	}
}

// TestAPI_GET_SuccessFalse_PassesThrough covers that reads are NOT subject to the
// success-flag check: a GET returning {"success":false} is passed through (the
// raw escape hatch), so the body prints and the command exits 0.
func TestAPI_GET_SuccessFalse_PassesThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "note": "raw"})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/things"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "raw")
}

// TestAPI_CSV_Rejected covers that --csv (a global flag) is rejected for raw api
// output rather than silently ignored.
func TestAPI_CSV_Rejected(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"api", "/v1/things", "--csv"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--csv is not supported")
}
