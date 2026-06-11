package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	internalapi "github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAPICmd(f *factory.Factory) *cobra.Command { return NewCmdAPI(f) }

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
	missing := filepath.Join(t.TempDir(), "does-not-exist.json")
	_, _, err := cmdtest.Run(t, "", newAPICmd, nil,
		"api", "/v1/accounts", "-X", "POST", "--body", "@"+missing)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading body file")
}

// TestAPI_Paginate_Error covers the error return when a page request fails
// (DoPaginated propagates the underlying API error). Uses a non-retriable 404
// (not 5xx) so the GET retry/backoff loop doesn't sleep through a coverage-only
// path, and pins the surfaced message so the assertion is tied to this path.
func TestAPI_Paginate_Error(t *testing.T) {
	handler := cmdtest.Status(t, "GET", "/v1/accounts", http.StatusNotFound,
		map[string]interface{}{"message": "boom"})

	_, _, err := cmdtest.Run(t, "", newAPICmd, handler,
		"api", "/v1/accounts", "--paginate")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

// TestAPI_Paginate_NonArrayPage covers the branch where an aggregated page is a
// JSON object rather than an array: it must be appended whole, not flattened.
// A single response with no "data" envelope and no "nextPage" yields exactly
// this shape from DoPaginated.
func TestAPI_Paginate_NonArrayPage(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/accounts/acct-1",
		map[string]interface{}{"id": "acct-1", "name": "Acme"})

	stdout, _, err := cmdtest.Run(t, "", newAPICmd, handler,
		"api", "/v1/accounts/acct-1", "--paginate")
	require.NoError(t, err)

	// The page must be aggregated as a single whole element (the object), not
	// flattened into the top level or dropped: the output is a JSON array of
	// length 1 whose element is the original object.
	var agg []map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(stdout), &agg))
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
			m := method
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, m, r.Method)
				w.WriteHeader(200)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"reasons": []map[string]interface{}{
						{"code": 50000040, "message": "write rejected"},
					},
				})
			})

			_, _, err := cmdtest.Run(t, "", newAPICmd, handler,
				"api", "/v1/things", "-X", method)
			require.Error(t, err, "%s with 200 success:false must error", method)
			assert.Contains(t, err.Error(), "write rejected")
		})
	}
}

// TestAPI_GET_SuccessFalse_PassesThrough covers that reads are NOT subject to the
// success-flag check: a GET returning {"success":false} is passed through (the
// raw escape hatch), so the body prints and the command exits 0.
func TestAPI_GET_SuccessFalse_PassesThrough(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "note": "raw"})
	})

	stdout, _, err := cmdtest.Run(t, "", newAPICmd, handler, "api", "/v1/things")
	require.NoError(t, err)
	assert.Contains(t, stdout, "raw")
}

// TestAPI_CSV_Rejected covers that --csv (a global flag) is rejected for raw api
// output rather than silently ignored.
func TestAPI_CSV_Rejected(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newAPICmd, nil, "api", "/v1/things", "--csv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--csv is not supported")
}
