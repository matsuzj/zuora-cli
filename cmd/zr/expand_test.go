package main

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/root"
	"github.com/stretchr/testify/assert"
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
		// KNOWN BUG (fixed in a later commit): strings.Fields destroys quoting,
		// so an alias wrapping a ZOQL query splits mid-string.
		{"BUG: quotes are destroyed", []string{"zr", "q1"},
			[]string{"zr", "query", `"SELECT`, "Id", "FROM", `Account"`}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, expandAlias(rootCmd, tc.in, store), "input %v", tc.in)
		})
	}
}
