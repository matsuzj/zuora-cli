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

	refreshCalled := false
	client := NewClient(
		WithBaseURL(server.URL),
		WithTokenSource(func() (string, error) {
			if refreshCalled {
				return "refreshed-token", nil
			}
			refreshCalled = true
			return "refreshed-token", nil
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
