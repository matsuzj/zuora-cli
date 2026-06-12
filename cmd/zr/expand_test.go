package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/root"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore map[string]string

func (f fakeStore) Get(name string) (string, bool) {
	v, ok := f[name]
	return v, ok
}

// Characterization tests over the REAL root command (construction is
// side-effect free): these freeze the expansion behavior — flag-skip rules,
// builtin no-expand, trailing-arg preservation — so refactors are provably
// behavior-preserving except where a fix is intentional and marked.
func TestExpandAlias(t *testing.T) {
	rootCmd := root.NewCmdRoot(factory.New())
	store := fakeStore{
		"ls":         "account list",
		"creditmemo": "account list", // attempts to shadow a registered command
		"billrun":    "account list", // ditto (was shadowable via the old manual map)
		"q1":         `query "SELECT Id FROM Account"`,
		"broken":     `query "SELECT unbalanced`,
	}

	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"no args", []string{"zr"}, []string{"zr"}},
		{"basic expansion", []string{"zr", "ls"}, []string{"zr", "account", "list"}},
		{"trailing args preserved", []string{"zr", "ls", "--page", "2"}, []string{"zr", "account", "list", "--page", "2"}},
		{"leading bool flag", []string{"zr", "--json", "ls"}, []string{"zr", "--json", "account", "list"}},
		{"leading bool flag (csv)", []string{"zr", "--csv", "ls"}, []string{"zr", "--csv", "account", "list"}},
		{"value flag skipped", []string{"zr", "--env", "prod", "ls"}, []string{"zr", "--env", "prod", "account", "list"}},
		{"shorthand value flag skipped", []string{"zr", "-e", "prod", "ls"}, []string{"zr", "-e", "prod", "account", "list"}},
		{"value flag skipped (zuora-version)", []string{"zr", "--zuora-version", "2024-01-01", "ls"},
			[]string{"zr", "--zuora-version", "2024-01-01", "account", "list"}},
		{"--flag=value form", []string{"zr", "--env=prod", "ls"}, []string{"zr", "--env=prod", "account", "list"}},
		{"builtin never expanded", []string{"zr", "account", "list"}, []string{"zr", "account", "list"}},
		{"help never expanded", []string{"zr", "help"}, []string{"zr", "help"}},
		{"alias not found", []string{"zr", "nope"}, []string{"zr", "nope"}},
		{"only flags, no command", []string{"zr", "--json"}, []string{"zr", "--json"}},
		// FIXED (was a bug): billrun/creditmemo/debitmemo were missing from the
		// old manual builtinCommands map, so aliases could shadow them. The
		// builtin set is now derived from the registered commands.
		{"registered command cannot be shadowed (creditmemo)", []string{"zr", "creditmemo"}, []string{"zr", "creditmemo"}},
		{"registered command cannot be shadowed (billrun)", []string{"zr", "billrun"}, []string{"zr", "billrun"}},
		// FIXED (was a bug): strings.Fields destroyed quoting, so an alias
		// wrapping a ZOQL query split mid-string. shlex keeps quoted segments
		// as single arguments (gh behavior).
		{"quoted expansion survives intact", []string{"zr", "q1"},
			[]string{"zr", "query", "SELECT Id FROM Account"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := expandAlias(rootCmd, tc.in, store)
			assert.NoError(t, err, "input %v", tc.in)
			assert.Equal(t, tc.want, got, "input %v", tc.in)
		})
	}
}

// A malformed expansion (unbalanced quoting) must not be half-applied: the
// original args come back with a non-nil error so main can warn and dispatch
// unexpanded.
func TestExpandAlias_MalformedExpansion(t *testing.T) {
	rootCmd := root.NewCmdRoot(factory.New())
	store := fakeStore{"broken": `query "SELECT unbalanced`}

	got, err := expandAlias(rootCmd, []string{"zr", "broken"}, store)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "malformed expansion")
	assert.Equal(t, []string{"zr", "broken"}, got)
}

func TestResolveAliasArgs(t *testing.T) {
	rootCmd := root.NewCmdRoot(factory.New())

	t.Run("expands from a valid aliases.yml", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "aliases.yml"),
			[]byte("ls: account list\n"), 0600))
		var errOut bytes.Buffer
		got := resolveAliasArgs(rootCmd, dir, []string{"zr", "ls"}, &errOut)
		assert.Equal(t, []string{"zr", "account", "list"}, got)
		assert.Empty(t, errOut.String())
	})

	t.Run("missing aliases.yml is fine", func(t *testing.T) {
		var errOut bytes.Buffer
		got := resolveAliasArgs(rootCmd, t.TempDir(), []string{"zr", "ls"}, &errOut)
		assert.Equal(t, []string{"zr", "ls"}, got)
		assert.Empty(t, errOut.String())
	})

	t.Run("broken aliases.yml warns and falls through", func(t *testing.T) {
		// The old behavior silently disabled every alias on a load error,
		// hiding the corruption from the user entirely.
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "aliases.yml"),
			[]byte(":\n\t- not yaml"), 0600))
		var errOut bytes.Buffer
		got := resolveAliasArgs(rootCmd, dir, []string{"zr", "ls"}, &errOut)
		assert.Equal(t, []string{"zr", "ls"}, got)
		assert.Contains(t, errOut.String(), "Warning: ignoring aliases")
		assert.Contains(t, errOut.String(), "aliases.yml")
	})

	t.Run("malformed expansion warns and falls through", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "aliases.yml"),
			[]byte(`broken: 'query "SELECT unbalanced'`+"\n"), 0600))
		var errOut bytes.Buffer
		got := resolveAliasArgs(rootCmd, dir, []string{"zr", "broken"}, &errOut)
		assert.Equal(t, []string{"zr", "broken"}, got)
		assert.Contains(t, errOut.String(), "malformed expansion")
	})
}

// TestExitCode_ContextCanceled pins the 128+SIGINT convention (P6-4): a
// cancellation — even wrapped — exits 130, not an API-error code.
func TestExitCode_ContextCanceled(t *testing.T) {
	assert.Equal(t, 130, exitCode(context.Canceled))
	assert.Equal(t, 130, exitCode(fmt.Errorf("wrapped: %w", context.Canceled)))
	assert.Equal(t, 1, exitCode(fmt.Errorf("plain failure")))
}
