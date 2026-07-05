// Package cmdtest is the shared test harness for command packages: one call
// wires a command under a stub root (with all of root.go's persistent flags),
// backs it with an httptest server, executes it, and returns both streams.
// Handler builders for the canonical Zuora envelopes live in handlers.go.
// This is the standard command-test harness: command packages across pkg/cmd/**
// drive their commands through Run/OK/Reasons/Status (docs/refactoring-plan.md
// records the migration that landed it).
package cmdtest

import (
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

// Run builds a minimal cobra root, wires newCmd under an optional parent
// group, starts a test HTTP server backed by handler, executes args, and
// returns stdout, stderr, and the Execute error.
//
// parent is the Use string of the intermediate group command (e.g. "order");
// pass "" for root-level commands. handler may be nil for tests that make no
// HTTP request (flag validation, --confirm guards) — the factory then points
// at http://localhost so an unexpected request fails loudly instead of
// silently succeeding. args are the tokens as typed, without the program
// name.
//
// Run calls t.Setenv (below) to neutralize ambient ZR_* env, so a test using
// Run MUST NOT call t.Parallel(): the Go testing runtime forbids t.Parallel in
// any test that has called t.Setenv, and will panic.
func Run(t *testing.T, parent string, newCmd func(*factory.Factory) *cobra.Command, handler http.HandlerFunc, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	// The harness root applies the REAL global flags, so an ambient
	// ZR_READ_ONLY exported as a machine-level safety default would block
	// every write-command success test (review finding on the wave-2
	// migration). Neutralize it per test — t.Setenv restores it afterwards;
	// EnvReadOnly treats empty as off. Tests that exercise the env knob
	// itself use their own harness, not Run.
	t.Setenv("ZR_READ_ONLY", "")
	t.Setenv("ZR_READ_ONLY_ALLOW_DATA_QUERY", "")
	t.Setenv("ZR_DEBUG", "")
	t.Setenv("ZR_ENV", "")

	serverURL := "http://localhost"
	if handler != nil {
		server := httptest.NewServer(handler)
		t.Cleanup(server.Close)
		serverURL = server.URL
	}

	ios, _, out, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), serverURL, "test-token")

	root := buildRoot(f, parent, newCmd)
	root.SetArgs(args)
	err = root.Execute()
	return out.String(), errOut.String(), err
}

// RequiresConfirm pins the cmdutil.RequireConfirm guard for a destructive
// command: executing args (as typed, WITHOUT --confirm) must fail with the
// canonical guard error before any HTTP request. The handler is nil, so a
// command that reaches the network fails with a connection error instead of
// the guard message — asserting the message therefore proves the guard fired,
// not merely that something errored. parent/newCmd/args mirror Run. Guard
// tests that assert anything beyond this shared shape keep their own bodies.
func RequiresConfirm(t *testing.T, parent string, newCmd func(*factory.Factory) *cobra.Command, args ...string) {
	t.Helper()
	_, _, err := Run(t, parent, newCmd, nil, args...)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
	assert.Contains(t, err.Error(), "this action is irreversible")
}

// buildRoot constructs a stub root that carries the REAL global-flag
// behavior: globalflags.Register defines the same persistent flags as the
// production root, and globalflags.Apply runs as PersistentPreRunE — so
// --read-only blocking, the --json+--template rejection, --env validation and
// --zuora-version/--verbose wiring all behave exactly as in the shipped CLI
// (a review caught that a name-only stub would let migrated tests drift).
func buildRoot(f *factory.Factory, parent string, newCmd func(*factory.Factory) *cobra.Command) *cobra.Command {
	root := &cobra.Command{
		Use:           "zr",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return globalflags.Apply(f, cmd)
		},
	}
	globalflags.Register(root)

	leaf := newCmd(f)
	if parent == "" {
		root.AddCommand(leaf)
	} else {
		grp := &cobra.Command{Use: parent}
		grp.AddCommand(leaf)
		root.AddCommand(grp)
	}
	return root
}
