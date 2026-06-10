// Package cmdtest is the shared test harness for command packages: one call
// wires a command under a stub root (with all of root.go's persistent flags),
// backs it with an httptest server, executes it, and returns both streams.
// Handler builders for the canonical Zuora envelopes live in handlers.go.
// Command tests migrate onto this harness in P3-4 (docs/refactoring-plan.md);
// until then the package carries only its own tests.
package cmdtest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
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

// buildRoot constructs a stub root carrying the same 8 persistent flags
// pkg/cmd/root/root.go registers — commands read them via output.FromCmd, and
// cobra rejects unknown flags otherwise. Help texts are intentionally empty:
// the stub is never user-facing, only the names must match.
func buildRoot(f *factory.Factory, parent string, newCmd func(*factory.Factory) *cobra.Command) *cobra.Command {
	root := &cobra.Command{Use: "zr"}

	root.PersistentFlags().StringP("env", "e", "", "")
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	root.PersistentFlags().Bool("csv", false, "")
	root.PersistentFlags().String("zuora-version", "", "")
	root.PersistentFlags().Bool("verbose", false, "")
	root.PersistentFlags().Bool("read-only", false, "")

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
