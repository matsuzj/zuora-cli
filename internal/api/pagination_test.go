package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoPaginated_TruncationReturnsErrorAndPartialData(t *testing.T) {
	// Server always advertises a next page, forcing the maxPages cap.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":     []map[string]string{{"id": "x"}},
			"nextPage": "/v1/test?page=next",
		})
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	data, err := c.DoPaginated("GET", "/v1/test")
	require.Error(t, err, "hitting the page cap must return an error")
	assert.Contains(t, err.Error(), "pagination limit")
	assert.NotEmpty(t, data, "partial data must still be returned alongside the truncation error")
}

func TestDoPaginated_NonPaginatedFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"single","name":"x"}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	data, err := c.DoPaginated("GET", "/v1/test")
	require.NoError(t, err)
	require.Len(t, data, 1, "a non-paginated body should come back as a single element")
}

func TestDoPaginated_TwoPages(t *testing.T) {
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		resp := map[string]interface{}{
			"data": []map[string]string{{"id": fmt.Sprintf("p%d", page)}},
		}
		if page < 2 {
			resp["nextPage"] = fmt.Sprintf("/v1/test?page=%d", page+1)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	data, err := c.DoPaginated("GET", "/v1/test")
	require.NoError(t, err)
	assert.Len(t, data, 2)
}

// TestDoPaginated_ResendsBodyEachPage guards against the body-exhaustion bug:
// the request body is a single-use io.Reader, so without buffering it the first
// page drains it and page 2+ POSTs an empty body. Every page must receive the
// full body.
func TestDoPaginated_ResendsBodyEachPage(t *testing.T) {
	const wantBody = `{"queryString":"select id from account"}`
	var gotBodies []string
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBodies = append(gotBodies, string(b))
		page++
		resp := map[string]interface{}{
			"data": []map[string]string{{"id": fmt.Sprintf("p%d", page)}},
		}
		if page < 2 {
			resp["nextPage"] = fmt.Sprintf("/v1/test?page=%d", page+1)
		}
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	data, err := c.DoPaginated("POST", "/v1/test", WithBody(strings.NewReader(wantBody)))
	require.NoError(t, err)
	assert.Len(t, data, 2)

	require.Len(t, gotBodies, 2, "expected exactly two page requests")
	for i, got := range gotBodies {
		assert.Equal(t, wantBody, got, "page %d must receive the full request body, not an exhausted/empty one", i+1)
	}
}
