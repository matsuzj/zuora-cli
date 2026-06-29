package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_BearerTokenInjection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithTokenSource(func(context.Context) (string, error) { return "test-token", nil }),
	)

	resp, err := client.Get("/v1/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_ZuoraVersionHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2025-08-12", r.Header.Get("Zuora-Version"))
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithZuoraVersion("2025-08-12"),
	)

	_, err := client.Get("/v1/test")
	require.NoError(t, err)
}

func TestClient_PostWithBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	body := strings.NewReader(`{"name":"test"}`)
	resp, err := client.Post("/v1/accounts", body)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom"))
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.Get("/v1/test", WithHeader("X-Custom", "custom-value"))
	require.NoError(t, err)
}

func TestClient_QueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "20", r.URL.Query().Get("pageSize"))
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.Get("/v1/test", WithQuery("pageSize", "20"))
	require.NoError(t, err)
}

func TestClient_QuerySlice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()["filter[]"]
		assert.Equal(t, []string{"status.EQ:Active", "balance.GT:0"}, values)
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.Get("/v1/test", WithQuerySlice("filter[]", []string{"status.EQ:Active", "balance.GT:0"}))
	require.NoError(t, err)
}

func TestClient_APIError_V1Format(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 53100020, "message": "The account key 'XXX' is invalid."},
			},
		})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.Get("/v1/accounts/XXX")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 400, apiErr.StatusCode)
	assert.Contains(t, apiErr.Message, "invalid")
	assert.Equal(t, 3, apiErr.ExitCode())
}

func TestClient_APIError_V2Format(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{
				"code":    "NOT_FOUND",
				"message": "Resource not found",
			},
		})
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	_, err := client.Get("/v1/missing")
	require.Error(t, err)

	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "NOT_FOUND", apiErr.Code)
	assert.Equal(t, 3, apiErr.ExitCode())
}

func TestClient_ServerError_ExitCode(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(500)
		w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer server.Close()

	client := NewClient(
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	// Run the retry loop without real backoff sleeps (this test alone was
	// ~7.7s of the package's wall time).
	client.sleep = func(context.Context, time.Duration) error { return nil }
	_, err := client.Get("/v1/test")
	require.Error(t, err)
	assert.True(t, callCount > 1, "should have retried")
}

func TestClient_401_TokenRefresh(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(401)
			w.Write([]byte(`{"message":"unauthorized"}`))
			return
		}
		assert.Equal(t, "Bearer refreshed-token", r.Header.Get("Authorization"))
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	tokenCallCount := 0
	client := NewClient(
		WithBaseURL(server.URL),
		WithTokenSource(func(context.Context) (string, error) {
			tokenCallCount++
			if tokenCallCount > 1 {
				return "refreshed-token", nil
			}
			return "expired-token", nil
		}),
	)

	resp, err := client.Get("/v1/test")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_Verbose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	var buf bytes.Buffer
	client := NewClient(
		WithBaseURL(server.URL),
		WithTokenSource(func(context.Context) (string, error) { return "secret-token", nil }),
	)
	client.SetVerbose(&buf)

	_, err := client.Get("/v1/test")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "Bearer ***")
	assert.NotContains(t, output, "secret-token")
	assert.Contains(t, output, "HTTP 200")
}

// TestClient_Verbose_MasksBearerAcrossMethods pins the credential-masking
// invariant for EVERY method, not just GET: the Authorization header is logged
// as "Bearer ***" and the raw token never appears in verbose output. The
// masking lives in the shared Do() path (client.go), so this is correct today;
// the test guards against a future per-method refactor of the verbose block
// silently unmasking writes (POST/PUT/PATCH/DELETE).
func TestClient_Verbose_MasksBearerAcrossMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	const token = "secret-write-token"

	cases := []struct {
		name string
		do   func(c *Client) (*Response, error)
	}{
		{"GET", func(c *Client) (*Response, error) { return c.Get("/v1/test") }},
		{"POST", func(c *Client) (*Response, error) { return c.Post("/v1/test", strings.NewReader(`{}`)) }},
		{"PUT", func(c *Client) (*Response, error) { return c.Put("/v1/test", strings.NewReader(`{}`)) }},
		// No Client.Patch convenience method exists; drive PATCH through Do.
		{"PATCH", func(c *Client) (*Response, error) {
			return c.Do(http.MethodPatch, "/v1/test", WithBody(strings.NewReader(`{}`)))
		}},
		{"DELETE", func(c *Client) (*Response, error) { return c.Delete("/v1/test") }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			c := NewClient(
				WithBaseURL(server.URL),
				WithTokenSource(func(context.Context) (string, error) { return token, nil }),
			)
			c.SetVerbose(&buf)

			_, err := tc.do(c)
			require.NoError(t, err)

			out := buf.String()
			assert.Contains(t, out, tc.name, "the request method must be logged in verbose output")
			assert.Contains(t, out, "Bearer ***", "Authorization header must be masked")
			assert.NotContains(t, out, token, "the raw bearer token must never appear in verbose output")
		})
	}
}

// TestRedactHeaderValue pins the shared verbose redaction helper: credential-
// and session-bearing headers are masked (Authorization keeps its scheme),
// everything else renders verbatim, and the key match is case-insensitive.
func TestRedactHeaderValue(t *testing.T) {
	cases := []struct {
		key, in, want string
	}{
		{"Authorization", "Bearer abc.def", "Bearer ***"},
		{"Authorization", "opaque-no-scheme", "***"},
		{"Proxy-Authorization", "Basic dXNlcjpwYXNz", "Basic ***"},
		{"Cookie", "ZSession=secret", "***"},
		{"Set-Cookie", "ZSession=secret; HttpOnly; Secure", "***"},
		{"set-cookie", "ZSession=secret", "***"}, // non-canonical key still matches
		{"Content-Type", "application/json", "application/json"},
		{"X-Request-Id", "req-1", "req-1"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, redactHeaderValue(tc.key, tc.in), "%s: %q", tc.key, tc.in)
	}
}

// TestClient_Verbose_RedactsResponseHeaders pins that the -v RESPONSE header
// dump masks a session-bearing Set-Cookie (equivalent to a bearer token)
// symmetrically with the request-side Authorization masking, while leaving
// non-secret response headers verbatim. Guards the asymmetry where the request
// dump masked Authorization but the response dump printed every header raw.
func TestClient_Verbose_RedactsResponseHeaders(t *testing.T) {
	const sessionCookie = "ZSession=super-secret-session-value"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Set-Cookie", sessionCookie)
		w.Header().Set("X-Request-Id", "req-123")
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	var buf bytes.Buffer
	client := NewClient(
		WithBaseURL(server.URL),
		WithTokenSource(func(context.Context) (string, error) { return "secret-token", nil }),
	)
	client.SetVerbose(&buf)

	_, err := client.Get("/v1/test")
	require.NoError(t, err)

	out := buf.String()
	assert.NotContains(t, out, "super-secret-session-value", "Set-Cookie value must not leak in verbose output")
	assert.Contains(t, out, "Set-Cookie: ***", "Set-Cookie must be redacted in the response dump")
	assert.Contains(t, out, "X-Request-Id: req-123", "non-secret response headers still render verbatim")
}

// --- Read-Only Mode Tests ---

func TestClient_ReadOnly_POSTBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	_, err := client.Post("/v1/accounts", strings.NewReader(`{}`))
	require.Error(t, err)
	var roErr *ReadOnlyError
	require.ErrorAs(t, err, &roErr)
}

func TestClient_ReadOnly_PUTBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	_, err := client.Put("/v1/accounts/123", strings.NewReader(`{}`))
	require.Error(t, err)
	var roErr *ReadOnlyError
	require.ErrorAs(t, err, &roErr)
}

func TestClient_ReadOnly_DELETEBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	_, err := client.Delete("/v1/accounts/123")
	require.Error(t, err)
	var roErr *ReadOnlyError
	require.ErrorAs(t, err, &roErr)
}

func TestClient_ReadOnly_PATCHBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	_, err := client.Do(http.MethodPatch, "/v1/accounts/123", WithBody(strings.NewReader(`{}`)))
	require.Error(t, err)
	var roErr *ReadOnlyError
	require.ErrorAs(t, err, &roErr)
}

// TestClient_ReadOnly_DataQuery pins the Data Query opt-in: submit (POST
// query/jobs) and cancel (DELETE query/jobs/{id}) are blocked by default in
// read-only mode and allowed only when SetReadOnlyAllowDataQuery(true) is set.
// The opt-in must widen ONLY those two endpoints (near-misses and ordinary
// writes stay blocked), and a blocked Data Query write must carry the opt-in
// hint while ordinary blocks must not.
func TestClient_ReadOnly_DataQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"id":"job-1"}}`))
	}))
	defer server.Close()

	cases := []struct {
		name        string
		method      string
		path        string
		allowDQ     bool
		wantBlocked bool
		wantDQHint  bool
	}{
		// Opt-in OFF (default): Data Query writes are blocked, with the hint.
		{"submit blocked by default", http.MethodPost, "/query/jobs", false, true, true},
		{"cancel blocked by default", http.MethodDelete, "/query/jobs/job-123", false, true, true},
		// Opt-in ON: submit + cancel allowed.
		{"submit allowed with opt-in", http.MethodPost, "/query/jobs", true, false, false},
		{"cancel allowed with opt-in", http.MethodDelete, "/query/jobs/job-123", true, false, false},
		// Opt-in ON still blocks near-misses, and never sets the DQ hint for them.
		{"id-suffixed POST still blocked", http.MethodPost, "/query/jobs/abc", true, true, false},
		{"collection DELETE still blocked", http.MethodDelete, "/query/jobs", true, true, false},
		{"PUT on a job still blocked", http.MethodPut, "/query/jobs/job-123", true, true, false},
		{"multi-segment DELETE still blocked", http.MethodDelete, "/query/jobs/abc/def", true, true, false},
		{"trailing-slash DELETE still blocked", http.MethodDelete, "/query/jobs/abc/", true, true, false},
		// The opt-in must NOT widen ordinary writes.
		{"normal POST still blocked with opt-in", http.MethodPost, "/v1/accounts", true, true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(WithBaseURL(server.URL))
			client.SetReadOnly(true)
			client.SetReadOnlyAllowDataQuery(tc.allowDQ)

			var err error
			switch tc.method {
			case http.MethodPost:
				_, err = client.Post(tc.path, strings.NewReader(`{}`))
			case http.MethodPut:
				_, err = client.Put(tc.path, strings.NewReader(`{}`))
			case http.MethodDelete:
				_, err = client.Delete(tc.path)
			}

			if !tc.wantBlocked {
				require.NoError(t, err)
				return
			}
			var roErr *ReadOnlyError
			require.ErrorAs(t, err, &roErr)
			if tc.wantDQHint {
				assert.NotEmpty(t, roErr.Hint, "blocked Data Query write must carry the opt-in hint")
				assert.Contains(t, roErr.Error(), "--read-only-allow-data-query")
			} else {
				assert.Empty(t, roErr.Hint, "a non-Data-Query block must not carry the DQ hint")
			}
		})
	}
}

func TestClient_ReadOnly_GETAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	resp, err := client.Get("/v1/accounts")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_ReadOnly_ZOQLQueryAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"done":true,"records":[]}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	resp, err := client.Post("/v1/action/query", strings.NewReader(`{"queryString":"SELECT Id FROM Account"}`))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_ReadOnly_ZOQLQueryMoreAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"done":true,"records":[]}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	resp, err := client.Post("/v1/action/querymore", strings.NewReader(`{"queryLocator":"abc"}`))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_ReadOnly_CommercePOSTAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)

	endpoints := []string{
		"/commerce/charges/query",
		"/commerce/plans/query",
		"/commerce/plans/list",
		"/commerce/purchase-options/list",
		"/commerce/legacy/products/list",
	}
	for _, ep := range endpoints {
		resp, err := client.Post(ep, strings.NewReader(`{}`))
		require.NoError(t, err, "should allow POST to %s", ep)
		assert.Equal(t, 200, resp.StatusCode)
	}
}

func TestClient_ReadOnly_SubscriptionPreviewChangeRegexAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	resp, err := client.Post("/v1/subscriptions/SUB-00001234/preview", strings.NewReader(`{}`))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

// TestIsReadOnlyAllowed_NearMissAndNormalization pins that the read-only POST
// allowlist fails CLOSED on near-miss / normalization tricks: a path must match
// an allowlisted read-only endpoint EXACTLY (after lowercasing + query strip) to
// be permitted. Trailing slashes, dot-segments, and extra path segments must NOT
// sneak a write past the gate, and PUT/DELETE/PATCH are blocked even on an
// allowlisted path.
func TestIsReadOnlyAllowed_NearMissAndNormalization(t *testing.T) {
	cases := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		// allowed baselines
		{"GET always allowed", "GET", "v1/orders/O-1", true},
		{"allowlisted ZOQL query", "POST", "v1/action/query", true},
		{"case-insensitive allowlist", "POST", "V1/ACTION/QUERY", true},
		{"query string stripped before matching", "POST", "v1/orders/preview?async=true", true},
		{"regexp preview-change match", "POST", "v1/subscriptions/SUB-1/preview", true},
		// near-misses: must fail closed (blocked)
		{"trailing slash is not the allowlisted path", "POST", "v1/action/query/", false},
		{"dot-segments do not resolve into an allowlisted path", "POST", "v1/action/query/../../v1/orders", false},
		{"extra segment after preview is not allowlisted", "POST", "v1/subscriptions/SUB-1/preview/extra", false},
		{"non-allowlisted write POST is blocked", "POST", "v1/orders", false},
		{"PUT to an allowlisted path is still blocked", "PUT", "v1/action/query", false},
		{"DELETE is always blocked", "DELETE", "v1/orders/O-1", false},
		{"PATCH is always blocked", "PATCH", "v1/orders/O-1", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, isReadOnlyAllowed(tc.method, tc.path), "%s %s", tc.method, tc.path)
		})
	}
}

func TestClient_ReadOnly_MeterSummaryRegexAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)
	resp, err := client.Post("/meters/meter-abc-123/summary", strings.NewReader(`{}`))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_ReadOnly_AbsoluteURLNormalized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)

	// Absolute URL with allowlisted path should be allowed
	resp, err := client.Do("POST", server.URL+"/v1/action/query")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Absolute URL with non-allowlisted path should be blocked
	_, err = client.Do("POST", server.URL+"/v1/accounts")
	require.Error(t, err)
	var roErr *ReadOnlyError
	require.ErrorAs(t, err, &roErr)
}

func TestClient_ReadOnly_QueryParamNormalized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	client.SetReadOnly(true)

	// Allowlisted path with query params should still be allowed
	resp, err := client.Do("POST", "/v1/action/query?foo=bar")
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_ReadOnly_SetReadOnlyWorks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	// Initially not read-only — POST should work
	_, err := client.Post("/v1/accounts", strings.NewReader(`{}`))
	require.NoError(t, err)

	// Enable read-only
	client.SetReadOnly(true)
	_, err = client.Post("/v1/accounts", strings.NewReader(`{}`))
	require.Error(t, err)
	var roErr *ReadOnlyError
	require.ErrorAs(t, err, &roErr)

	// Disable read-only
	client.SetReadOnly(false)
	_, err = client.Post("/v1/accounts", strings.NewReader(`{}`))
	require.NoError(t, err)
}

func TestReadOnlyError_ExitCode(t *testing.T) {
	err := &ReadOnlyError{}
	assert.Equal(t, 5, err.ExitCode())
	assert.Contains(t, err.Error(), "read-only mode")
}

func TestClient_Pagination(t *testing.T) {
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		resp := map[string]interface{}{
			"data": []map[string]string{{"id": fmt.Sprintf("item-%d", page)}},
		}
		if page < 3 {
			resp["nextPage"] = fmt.Sprintf("/v1/test?page=%d", page+1)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))
	pages, err := client.DoPaginated("GET", "/v1/test")
	require.NoError(t, err)
	assert.Len(t, pages, 3)
}

// ——— moved verbatim from hardening_test.go (P4-2 test consolidation) ———

// A same-host http target must be refused when the configured environment is
// https, so the bearer token is never sent in cleartext (scheme downgrade).
func TestClient_HTTPSBase_SameHostHTTP_Refused(t *testing.T) {
	c := NewClient(
		WithBaseURL("https://rest.zuora.com"),
		WithTokenSource(func(context.Context) (string, error) { return "secret-token", nil }),
	)
	_, err := c.Get("http://rest.zuora.com/v1/accounts")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cleartext")
}

// A cross-host redirect from the configured host must NOT be followed, so the
// request (body, Idempotency-Key, entity ids) never reaches the other host. The
// refusal must also fail fast — a deterministic policy rejection must not be
// retried as if it were a transient transport error.
func TestClient_CrossHostRedirect_Refused(t *testing.T) {
	var attackerHit bool
	attacker := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attackerHit = true
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer attacker.Close()

	var originHits int32
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&originHits, 1)
		http.Redirect(w, r, attacker.URL+"/v1/accounts", http.StatusTemporaryRedirect)
	}))
	defer origin.Close()

	c := NewClient(
		WithBaseURL(origin.URL),
		WithTokenSource(func(context.Context) (string, error) { return "secret-token", nil }),
	)
	_, err := c.Get("/v1/accounts")
	require.Error(t, err)
	assert.False(t, attackerHit, "a cross-host redirect must not be followed to the other host")
	assert.ErrorIs(t, err, errRedirectRefused, "the refusal must carry the no-retry sentinel")
	assert.Equal(t, int32(1), atomic.LoadInt32(&originHits),
		"a blocked redirect must fail fast (origin hit once), not retried as a transient error")
}

// A single http.Client shared across NewClient instances with different base
// URLs must not be mutated, and each client's redirect policy must be bound to
// ITS OWN base host (not whichever NewClient ran first).
func TestClient_SharedInjectedClient_PerBaseRedirectPolicy(t *testing.T) {
	shared := &http.Client{}
	cA := NewClient(WithBaseURL("https://host-a.example"), WithHTTPClient(shared))
	cB := NewClient(WithBaseURL("https://host-b.example"), WithHTTPClient(shared))

	assert.Nil(t, shared.CheckRedirect, "the shared injected client must not be mutated")
	require.NotNil(t, cA.httpClient.CheckRedirect)
	require.NotNil(t, cB.httpClient.CheckRedirect)

	reqB, _ := http.NewRequest(http.MethodGet, "https://host-b.example/v1/x", nil)
	// cA is configured for host-a, so a redirect to host-b is cross-host → refused.
	assert.ErrorIs(t, cA.httpClient.CheckRedirect(reqB, nil), errRedirectRefused,
		"client A (base host-a) must refuse a redirect to host-b")
	// cB is configured for host-b, so the same target is same-host → allowed.
	assert.NoError(t, cB.httpClient.CheckRedirect(reqB, nil),
		"client B (base host-b) must allow a same-host redirect — policy must use B's own base")
}

// PUT must NOT carry an Idempotency-Key: Zuora rejects PUT requests that include
// one ("HTTP 400: Request method 'PUT' not supported with Idempotency-Key
// header"). POST/PATCH still carry it; this guards against reintroducing the key
// on PUT, which would break every PUT action command. POST is checked too as a
// positive control.
func TestClient_PUT_OmitsIdempotencyKey(t *testing.T) {
	var putKey, postKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			putKey = r.Header.Get("Idempotency-Key")
		} else {
			postKey = r.Header.Get("Idempotency-Key")
		}
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Put("/v1/test", strings.NewReader(`{}`))
	require.NoError(t, err)
	_, err = c.Post("/v1/test", strings.NewReader(`{}`))
	require.NoError(t, err)

	assert.Empty(t, putKey, "PUT must NOT carry an Idempotency-Key (Zuora rejects it)")
	assert.True(t, strings.HasPrefix(postKey, "zr-"), "POST still carries the key as a positive control")
}

// Ctrl-C (context cancellation) while honoring a 429 Retry-After must surface as
// cancellation, not a stale HTTP 429 API error.
func TestRetry_429_CancelDuringRetryAfter_ReturnsCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message":"rate limited"}`))
	}))
	defer srv.Close()

	// Inject a sleep that simulates Ctrl-C firing during the Retry-After wait.
	c := NewClient(
		WithBaseURL(srv.URL),
		WithHTTPClient(srv.Client()),
		func(c *Client) {
			c.sleep = func(context.Context, time.Duration) error { return context.Canceled }
		},
	)
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled, "cancellation during Retry-After must surface as cancellation")
	var apiErr *APIError
	assert.False(t, errors.As(err, &apiErr), "must not be classified as a Zuora API error")
}
