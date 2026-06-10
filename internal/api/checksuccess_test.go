package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The success-flag check is ON BY DEFAULT: HTTP 200 responses carrying
// {"success": false} (v1) or {"Success": false} (Object CRUD) are the only
// signal that an otherwise-2xx call actually failed, so they must be turned
// into errors unless a caller explicitly opts out (WithoutCheckSuccess —
// reserved for the raw zr api GET/HEAD passthrough).
func TestCheckSuccess_LowercaseFalse_IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"success":false,"reasons":[{"code":"X","message":"nope"}]}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.Post("/v1/action", strings.NewReader(`{}`))
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, "X", apiErr.Code)
}

func TestCheckSuccess_UppercaseFalse_IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"Success":false}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.Post("/v1/objects", strings.NewReader(`{}`))
	require.Error(t, err, "uppercase Object-CRUD Success:false must be detected")
}

func TestCheckSuccess_True_IsOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"success":true,"id":"1"}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	resp, err := c.Post("/v1/action", strings.NewReader(`{}`))
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestCheckSuccess_NoSuccessField_PassesThrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"1","name":"x"}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.Post("/v1/action", strings.NewReader(`{}`))
	require.NoError(t, err, "a body without a success field is not a failure")
}

func TestCheckSuccess_MalformedJSON_PassesThrough(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	resp, err := c.Post("/v1/action", strings.NewReader(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "not json", resp.String())
}

func TestAPIError_401_ExitCodeAndHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"message":"token expired"}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL)) // no refresh -> 401 surfaces
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 2, apiErr.ExitCode(), "401 should map to the auth exit code (2)")
	assert.Contains(t, apiErr.Error(), "zr auth login")
}

func TestParseAPIError_TruncatesHugeBody(t *testing.T) {
	huge := strings.Repeat("A", 5000)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		w.Write([]byte(huge))
	}))
	defer srv.Close()

	c := newNoSleepClient(WithBaseURL(srv.URL), WithHTTPClient(srv.Client()))
	_, err := c.Get("/v1/test")
	require.Error(t, err)
	assert.LessOrEqual(t, len(err.Error()), 700, "an oversized non-JSON error body must be truncated")
}

func TestClient_SetsUserAgent(t *testing.T) {
	var ua string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	assert.Contains(t, ua, "zuora-cli/")
}

func TestClient_GET_NoIdempotencyKey(t *testing.T) {
	var hasKey bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, hasKey = r.Header["Idempotency-Key"]
		w.WriteHeader(200)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.Get("/v1/test")
	require.NoError(t, err)
	assert.False(t, hasKey, "GET must not carry an Idempotency-Key")
}

func TestCheckSuccess_DefaultOnForGet(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"success":false,"reasons":[{"code":"Y","message":"hidden failure"}]}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	_, err := c.Get("/v1/thing")
	require.Error(t, err, "the check must be on by default — no option passed")
}

func TestCheckSuccess_WithoutCheckSuccess_OptsOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"success":false,"reasons":[{"code":"Z","message":"raw passthrough"}]}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	resp, err := c.Get("/v1/raw", WithoutCheckSuccess())
	require.NoError(t, err, "opt-out must deliver the body uninterpreted")
	assert.Contains(t, resp.String(), "raw passthrough")
}
