package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		WithTokenSource(func() (string, error) { return "test-token", nil }),
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
		WithTokenSource(func() (string, error) {
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
		WithVerbose(&buf),
		WithTokenSource(func() (string, error) { return "secret-token", nil }),
	)

	_, err := client.Get("/v1/test")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "Bearer ***")
	assert.NotContains(t, output, "secret-token")
	assert.Contains(t, output, "HTTP 200")
}

// --- Read-Only Mode Tests ---

func TestClient_ReadOnly_POSTBlocked(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
	_, err := client.Patch("/v1/accounts/123", strings.NewReader(`{}`))
	require.Error(t, err)
	var roErr *ReadOnlyError
	require.ErrorAs(t, err, &roErr)
}

func TestClient_ReadOnly_GETAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())

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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
	resp, err := client.Post("/v1/subscriptions/SUB-00001234/preview", strings.NewReader(`{}`))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestClient_ReadOnly_MeterSummaryRegexAllowed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())
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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())

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

	client := NewClient(WithBaseURL(server.URL), WithReadOnly())

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
