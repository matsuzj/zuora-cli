// Package cmdtest is the shared test harness for command packages: one call
// wires a command under a stub root (with all of root.go's persistent flags),
// backs it with an httptest server, executes it, and returns both streams.
// Handler builders for the canonical Zuora envelopes live in handlers.go.
// Command tests across pkg/cmd use this harness to keep root flag behavior,
// HTTP setup, and output capture consistent.
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
func Run(t *testing.T, parent string, newCmd func(*factory.Factory) *cobra.Command, handler http.HandlerFunc, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	// The harness root applies the REAL global flags, so an ambient
	// ZR_READ_ONLY exported as a machine-level safety default would block
	// every write-command success test (review finding on the wave-2
	// migration). Neutralize it per test — t.Setenv restores it afterwards;
	// EnvReadOnly treats empty as off. Tests that exercise the env knob
	// itself use their own harness, not Run.
	t.Setenv("ZR_READ_ONLY", "")
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
