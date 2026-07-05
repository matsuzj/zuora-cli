package factory

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WiresDependencies(t *testing.T) {
	f := New()
	require.NotNil(t, f)
	assert.NotNil(t, f.IOStreams)
	assert.NotNil(t, f.Config)
	assert.NotNil(t, f.HttpClient)
	assert.NotNil(t, f.AuthToken)
}

func TestNew_ConfigCachedOnce(t *testing.T) {
	// Point config at an isolated dir so this does not touch the real config.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := New()

	c1, err1 := f.Config()
	c2, err2 := f.Config()
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, c1)
	// sync.Once must return the same cached instance on every call.
	assert.True(t, c1 == c2, "Config() must cache and return the same instance")
}

func TestNew_HttpClientBuildsForDefaultEnv(t *testing.T) {
	// Isolated config dir: LoadDefault materializes the default (sandbox)
	// environment, which is all HttpClient needs to build a client. No
	// network and no credentials are required to construct it.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := New()

	client, err := f.HttpClient()
	require.NoError(t, err)
	require.NotNil(t, client, "HttpClient() must build a client from the default config/env")

	// Building the client must be idempotent: it derives everything from the
	// lazily-cached config, so a second call succeeds the same way.
	client2, err2 := f.HttpClient()
	require.NoError(t, err2)
	require.NotNil(t, client2)
}

func TestNew_HttpClientUsesCachedConfig(t *testing.T) {
	// HttpClient internally goes through f.Config(), so it must observe the
	// same cached config that Config() returns rather than reloading.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := New()

	cfg, err := f.Config()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	client, err := f.HttpClient()
	require.NoError(t, err)
	require.NotNil(t, client)

	// The config returned after HttpClient ran must still be the same cached
	// instance (sync.Once was not re-triggered by HttpClient).
	cfgAfter, err := f.Config()
	require.NoError(t, err)
	assert.True(t, cfg == cfgAfter, "HttpClient() must reuse the cached config, not reload it")
}

// TestTokenSource_WriterWiring pins the P6-2 plumbing: a non-nil writer
// installs a Logf that writes to it; a nil writer leaves Logf nil (silent).
func TestTokenSource_WriterWiring(t *testing.T) {
	cfg := config.NewMockConfig()

	var buf strings.Builder
	ts := tokenSource(cfg, &buf)
	require.NotNil(t, ts.Logf)
	ts.Logf("hello %s\n", "world")
	assert.Equal(t, "hello world\n", buf.String())

	silent := tokenSource(cfg, nil)
	assert.Nil(t, silent.Logf)
}

// TestNewTestFactory_NoRealBackoffOnRetry pins the WithSleep seam: a handler
// returning a retryable 429 (no Retry-After, so the real sleeper would spend
// 1-1.5s in jittered backoff) must complete near-instantly under the test
// factory. Bites if NewTestFactory stops injecting the no-backoff sleeper —
// this test then takes >1s and fails the elapsed bound.
func TestNewTestFactory_NoRealBackoffOnRetry(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"message":"rate limited"}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	ios, _, _, _ := iostreams.Test()
	f := NewTestFactory(ios, config.NewMockConfig(), srv.URL, "tok")
	client, err := f.HttpClient()
	require.NoError(t, err)

	start := time.Now()
	resp, err := client.Get("/v1/test")
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&calls))
	assert.Less(t, elapsed, 500*time.Millisecond, "test factory must not spend real time in retry backoff")
}
