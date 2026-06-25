package globalflags_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// zuoraVersionHeader runs Register+Apply with the given args against a test
// server that records the Zuora-Version header, performs one GET through the
// factory's Apply-wrapped HTTP client, and returns the captured header value.
// The config's version is seeded to cfgVersion before the factory is built
// (NewTestFactory snapshots cfg.ZuoraVersion() into the client at that point).
func zuoraVersionHeader(t *testing.T, cfgVersion string, args ...string) string {
	t.Helper()

	var captured string
	var sawRequest bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawRequest = true
		captured = r.Header.Get("Zuora-Version")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "{}")
	}))
	t.Cleanup(srv.Close)

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetZuoraVersion(cfgVersion))
	f := factory.NewTestFactory(ios, cfg, srv.URL, "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags(args))
	require.NoError(t, globalflags.Apply(f, cmd))

	client, err := f.HttpClient()
	require.NoError(t, err)
	_, err = client.Get("/v1/test")
	require.NoError(t, err)
	require.True(t, sawRequest, "the GET must reach the test server")

	return captured
}

// --zuora-version overrides the Zuora-Version header on every request. This
// exercises the `zv != ""` branch of globalflags.Apply (globalflags.go) and
// (*api.Client).SetZuoraVersion — the override path that, before this test, no
// unit test reached. The flag value differs from the seeded config version so
// the assertion bites: a regression that dropped the override would surface the
// config version "2025-08-12" instead.
func TestApply_ZuoraVersionFlagOverridesHeader(t *testing.T) {
	got := zuoraVersionHeader(t, "2025-08-12", "--zuora-version", "2030-01-01")
	assert.Equal(t, "2030-01-01", got)
}

// Without --zuora-version, the header keeps the config's version: the flag's
// empty default must NOT blank an otherwise-set header (the `zv != ""` guard
// false branch).
func TestApply_ZuoraVersionDefaultsToConfigVersion(t *testing.T) {
	got := zuoraVersionHeader(t, "2025-08-12")
	assert.Equal(t, "2025-08-12", got)
}

// TestApply_ZuoraVersionFlagOnPostRequest pins that the --zuora-version override
// also reaches the WRITE path: a POST additionally carries a body and an
// Idempotency-Key, and the Zuora-Version header must survive alongside them. The
// helper above only exercises a GET (F-35: the override was never verified on a
// mutating request).
func TestApply_ZuoraVersionFlagOnPostRequest(t *testing.T) {
	var captured string
	var sawPost bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			sawPost = true
			captured = r.Header.Get("Zuora-Version")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "{}")
	}))
	t.Cleanup(srv.Close)

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetZuoraVersion("2025-08-12"))
	f := factory.NewTestFactory(ios, cfg, srv.URL, "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags([]string{"--zuora-version", "2030-01-01"}))
	require.NoError(t, globalflags.Apply(f, cmd))

	client, err := f.HttpClient()
	require.NoError(t, err)
	_, err = client.Post("/v1/test", strings.NewReader("{}"))
	require.NoError(t, err)
	require.True(t, sawPost, "the POST must reach the test server")
	assert.Equal(t, "2030-01-01", captured, "Zuora-Version override must be set on POST (write path), not just GET")
}
