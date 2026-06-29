package globalflags_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// applyReadOnlyClient runs Register+Apply with args against a test server and
// returns the Apply-wrapped HTTP client plus a flag recording whether any
// request reached the server. It drives the safety-critical wiring that — before
// this test (#324 / F-01) — no unit test exercised: --read-only / ZR_READ_ONLY
// → globalflags.Apply → (*api.Client).SetReadOnly on the factory's client.
func applyReadOnlyClient(t *testing.T, args ...string) (*api.Client, *bool) {
	t.Helper()
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "{}")
	}))
	t.Cleanup(srv.Close)

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), srv.URL, "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags(args))
	require.NoError(t, globalflags.Apply(f, cmd))

	client, err := f.HttpClient()
	require.NoError(t, err)
	return client, &hit
}

// TestApply_ReadOnlyFlagBlocksWritesAllowsReads pins the safety-critical wiring:
// with --read-only, Apply must put the factory's client into read-only mode so a
// write is refused BEFORE leaving the process, while reads still go through. A
// regression dropping globalflags.go's `if readOnly { client.SetReadOnly(true) }`
// would let every write silently hit the live tenant — yet pass all command tests.
func TestApply_ReadOnlyFlagBlocksWritesAllowsReads(t *testing.T) {
	client, hit := applyReadOnlyClient(t, "--read-only")

	var roErr *api.ReadOnlyError
	_, err := client.Post("/v1/orders", nil) // non-allowlisted POST
	require.Error(t, err)
	assert.ErrorAs(t, err, &roErr, "POST must be blocked in read-only mode")

	_, err = client.Delete("/v1/orders/O-1") // PUT/DELETE/PATCH are always blocked
	require.Error(t, err)
	assert.ErrorAs(t, err, &roErr, "DELETE must be blocked in read-only mode")

	assert.False(t, *hit, "blocked writes must not reach the server")

	_, err = client.Get("/v1/orders/O-1")
	require.NoError(t, err, "reads must still go through in read-only mode")
	assert.True(t, *hit, "the GET must reach the server")
}

// TestApply_ZRReadOnlyEnvBlocksWrites: with no flag, ZR_READ_ONLY=true must take
// effect through Apply (the EnvReadOnly fallback when --read-only is unchanged).
func TestApply_ZRReadOnlyEnvBlocksWrites(t *testing.T) {
	t.Setenv("ZR_READ_ONLY", "true")
	client, hit := applyReadOnlyClient(t) // no --read-only flag

	var roErr *api.ReadOnlyError
	_, err := client.Post("/v1/orders", nil)
	require.Error(t, err)
	assert.ErrorAs(t, err, &roErr, "ZR_READ_ONLY=true must block writes via Apply")
	assert.False(t, *hit)
}

// TestApply_ZRReadOnlyAllowDataQueryEnvOptsInWrites: ZR_READ_ONLY=true blocks
// writes, but the ZR_READ_ONLY_ALLOW_DATA_QUERY=1 env var must opt data-query
// submits (POST /query/jobs) back in via Apply — the primary CI deployment
// mechanism. A regression dropping globalflags.go's env fallback would leave the
// opt-in silently inert (all data-query writes blocked). Non-data-query writes
// stay blocked. (#434)
func TestApply_ZRReadOnlyAllowDataQueryEnvOptsInWrites(t *testing.T) {
	t.Setenv("ZR_READ_ONLY", "true")
	t.Setenv("ZR_READ_ONLY_ALLOW_DATA_QUERY", "1")
	client, hit := applyReadOnlyClient(t) // no flags — env path only

	var roErr *api.ReadOnlyError
	// A non-data-query write stays blocked.
	_, err := client.Post("/v1/orders", nil)
	require.Error(t, err)
	assert.ErrorAs(t, err, &roErr, "non-data-query writes stay blocked under ZR_READ_ONLY")
	assert.False(t, *hit, "the blocked write must not reach the server")

	// A data-query submit is opted in by ZR_READ_ONLY_ALLOW_DATA_QUERY=1.
	_, err = client.Post("/query/jobs", nil)
	require.NoError(t, err, "ZR_READ_ONLY_ALLOW_DATA_QUERY=1 (env) must allow data-query submits")
	assert.True(t, *hit, "the data-query submit must reach the server")
}

// TestApply_ReadOnlyFlagFalseOverridesEnv: an explicit --read-only=false wins
// over ZR_READ_ONLY=true (flag precedence — the !Changed guard in Apply).
func TestApply_ReadOnlyFlagFalseOverridesEnv(t *testing.T) {
	t.Setenv("ZR_READ_ONLY", "true")
	client, hit := applyReadOnlyClient(t, "--read-only=false")

	_, err := client.Post("/v1/orders", nil)
	require.NoError(t, err, "--read-only=false must override ZR_READ_ONLY=true")
	assert.True(t, *hit, "the write must reach the server when read-only is explicitly off")
}
