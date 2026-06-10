package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeStore map[string]string

func (f fakeStore) Get(name string) (string, bool) {
	v, ok := f[name]
	return v, ok
}

// Characterization tests: these freeze the CURRENT expansion behavior
// (including known bugs, each marked below) so the upcoming refactors are
// provably behavior-preserving except where a fix is intentional.
func TestExpandAlias(t *testing.T) {
	store := fakeStore{
		"ls":         "account list",
		"creditmemo": "account list", // shadows a built-in missing from the manual map
		"q1":         `query "SELECT Id FROM Account"`,
		"self":       "self again",
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
		{"value flag skipped", []string{"zr", "--env", "prod", "ls"}, []string{"zr", "--env", "prod", "account", "list"}},
		{"shorthand value flag skipped", []string{"zr", "-e", "prod", "ls"}, []string{"zr", "-e", "prod", "account", "list"}},
		{"--flag=value form", []string{"zr", "--env=prod", "ls"}, []string{"zr", "--env=prod", "account", "list"}},
		{"builtin never expanded", []string{"zr", "account", "list"}, []string{"zr", "account", "list"}},
		{"alias not found", []string{"zr", "nope"}, []string{"zr", "nope"}},
		{"only flags, no command", []string{"zr", "--json"}, []string{"zr", "--json"}},
		// KNOWN BUG (fixed in a later commit): billrun/creditmemo/debitmemo are
		// registered commands but missing from the manual builtinCommands map,
		// so an alias can shadow them.
		{"BUG: missing builtin is shadowed", []string{"zr", "creditmemo"}, []string{"zr", "account", "list"}},
		// KNOWN BUG (fixed in a later commit): strings.Fields destroys quoting,
		// so an alias wrapping a ZOQL query splits mid-string.
		{"BUG: quotes are destroyed", []string{"zr", "q1"},
			[]string{"zr", "query", `"SELECT`, "Id", "FROM", `Account"`}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, expandAlias(tc.in, store), "input %v", tc.in)
		})
	}
}
