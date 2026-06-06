package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
