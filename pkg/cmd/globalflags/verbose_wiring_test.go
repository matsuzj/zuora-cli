package globalflags_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApply_ZRDebugAPIEnablesBodyLogging pins the ZR_DEBUG=api → verbose-body
// wiring: Apply must enable level-2 verbose so the client logs request/response
// BODIES to ErrOut. VerboseLevels is unit-tested in isolation, but nothing drove
// the Apply → (*api.Client).SetVerbose/SetVerboseBody path end-to-end — a
// regression dropping that wiring would silently disable -vv/ZR_DEBUG body
// observability. (See #342 / F-19.)
func TestApply_ZRDebugAPIEnablesBodyLogging(t *testing.T) {
	t.Setenv("ZR_DEBUG", "api")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"marker":"BODY-MARKER-9000"}`)
	}))
	t.Cleanup(srv.Close)

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), srv.URL, "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags([]string{}))
	require.NoError(t, globalflags.Apply(f, cmd))

	client, err := f.HttpClient()
	require.NoError(t, err)
	_, err = client.Get("/v1/test")
	require.NoError(t, err)

	assert.Contains(t, errOut.String(), "BODY-MARKER-9000",
		"ZR_DEBUG=api must enable verbose body logging (response body to ErrOut)")
}

// TestApply_ZRDebugUnknownValueWarns pins #456 item 7: a non-empty ZR_DEBUG that
// is not exactly "api" (here the wrong-case "API") is a silent no-op, so Apply
// warns to stderr instead of leaving the user to wonder why bodies never appear.
func TestApply_ZRDebugUnknownValueWarns(t *testing.T) {
	t.Setenv("ZR_DEBUG", "API")

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "", "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags([]string{}))
	require.NoError(t, globalflags.Apply(f, cmd))

	assert.Contains(t, errOut.String(), `ZR_DEBUG="API" is not recognized`)
}

// TestApply_ZRDebugAPIDoesNotWarn confirms the exact supported value stays quiet.
func TestApply_ZRDebugAPIDoesNotWarn(t *testing.T) {
	t.Setenv("ZR_DEBUG", "api")

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "", "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags([]string{}))
	require.NoError(t, globalflags.Apply(f, cmd))

	assert.NotContains(t, errOut.String(), "not recognized")
}
