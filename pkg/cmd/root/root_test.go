package root

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// walkCommands visits every command in the tree below root (depth-first).
func walkCommands(root *cobra.Command, fn func(*cobra.Command)) {
	for _, c := range root.Commands() {
		fn(c)
		walkCommands(c, fn)
	}
}

func TestRootHasSubcommands(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)
	subcommands := cmd.Commands()

	names := make([]string, len(subcommands))
	for i, c := range subcommands {
		names[i] = c.Name()
	}
	assert.Contains(t, names, "version")
	assert.Contains(t, names, "completion")
	assert.Contains(t, names, "account")
	assert.Contains(t, names, "subscription")
}

// TestRootKebabCaseAliases pins #455: the concatenated multi-word resource
// groups also answer to their kebab-case spelling (additive — the canonical
// name still works). cobra's Find resolves an alias to its command.
func TestRootKebabCaseAliases(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}
	cmd := NewCmdRoot(f)

	for alias, canonical := range map[string]string{
		"bill-run":    "billrun",
		"credit-memo": "creditmemo",
		"debit-memo":  "debitmemo",
		"rate-plan":   "rateplan",
	} {
		found, _, err := cmd.Find([]string{alias})
		require.NoError(t, err, "alias %q must resolve", alias)
		assert.Equal(t, canonical, found.Name(), "alias %q must route to the %q group", alias, canonical)
	}
}

func TestRootJsonTemplateExclusion(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)
	cmd.SetArgs([]string{"version", "--json", "--template", "foo"})
	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --json and --template together")
}

// TestRootJqCombinationsAllowed guards the documented precedence: --jq implies
// JSON and wins when combined, and --json/--jq/--template all win over --csv, so
// these combinations must NOT be rejected (the renderer picks one). version reads
// no network, so a nil error means the combination was accepted by the guard.
func TestRootJqCombinationsAllowed(t *testing.T) {
	for _, args := range [][]string{
		{"version", "--json", "--jq", ".version"},
		{"version", "--csv", "--jq", ".version"},
	} {
		t.Run(args[1]+"+"+args[2], func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()
			f := &factory.Factory{IOStreams: ios}
			cmd := NewCmdRoot(f)
			cmd.SetArgs(args)
			assert.NoError(t, cmd.Execute())
		})
	}
}

func TestRootGlobalFlags(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)

	flags := []string{"env", "json", "jq", "template", "zuora-version", "verbose", "read-only", "read-only-allow-data-query"}
	for _, name := range flags {
		assert.NotNil(t, cmd.PersistentFlags().Lookup(name), "missing flag: %s", name)
	}
}

func TestRootReadOnlyFlag_BlocksWriteCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"acc-123","success":true}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		HttpClient: func() (*api.Client, error) {
			return api.NewClient(api.WithBaseURL(server.URL)), nil
		},
	}

	cmd := NewCmdRoot(f)
	// account create issues a POST to /v1/accounts — should be blocked
	cmd.SetArgs([]string{"--read-only", "account", "create", "--body", `{}`})
	err := cmd.Execute()
	require.Error(t, err)
	var roErr *api.ReadOnlyError
	assert.ErrorAs(t, err, &roErr, "expected ReadOnlyError, got: %v", err)
}

// TestRootReadOnlyAllowDataQuery_AllowsDataQueryButNotOtherWrites pins that the
// opt-in toggle widens ONLY Data Query: with --read-only --read-only-allow-data-query
// a POST /query/jobs is permitted while an ordinary POST /v1/accounts stays blocked.
func TestRootReadOnlyAllowDataQuery_AllowsDataQueryButNotOtherWrites(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"id":"job-1"}}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		HttpClient: func() (*api.Client, error) {
			return api.NewClient(api.WithBaseURL(server.URL)), nil
		},
	}

	cmd := NewCmdRoot(f)
	cmd.SetArgs([]string{"--read-only", "--read-only-allow-data-query", "version"})
	require.NoError(t, cmd.Execute())

	client, err := f.HttpClient()
	require.NoError(t, err)

	// Data Query submit is allowed by the opt-in.
	_, err = client.Post("/query/jobs", nil)
	require.NoError(t, err, "POST /query/jobs should be allowed with the opt-in")

	// An ordinary write is still blocked.
	_, err = client.Post("/v1/accounts", nil)
	var roErr *api.ReadOnlyError
	require.ErrorAs(t, err, &roErr, "ordinary writes must stay blocked even with the Data Query opt-in")
}

// TestRootReadOnlyAllowDataQuery_ResetOnReapply pins that the unconditional
// client setters make a reused factory idempotent: after a first Apply enables
// the opt-in, a second Apply WITHOUT it (still read-only) must reset the toggle
// so POST /query/jobs is blocked again — guarding the sticky-wrapper edge.
func TestRootReadOnlyAllowDataQuery_ResetOnReapply(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"data":{"id":"job-1"}}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		HttpClient: func() (*api.Client, error) {
			return api.NewClient(api.WithBaseURL(server.URL)), nil
		},
	}

	// First Apply: read-only + opt-in → Data Query submit allowed.
	cmd1 := NewCmdRoot(f)
	cmd1.SetArgs([]string{"--read-only", "--read-only-allow-data-query", "version"})
	require.NoError(t, cmd1.Execute())
	c1, err := f.HttpClient()
	require.NoError(t, err)
	_, err = c1.Post("/query/jobs", nil)
	require.NoError(t, err)

	// Second Apply on the SAME factory: read-only, NO opt-in → must block again.
	cmd2 := NewCmdRoot(f)
	cmd2.SetArgs([]string{"--read-only", "version"})
	require.NoError(t, cmd2.Execute())
	c2, err := f.HttpClient()
	require.NoError(t, err)
	_, err = c2.Post("/query/jobs", nil)
	var roErr *api.ReadOnlyError
	require.ErrorAs(t, err, &roErr, "re-Apply without the opt-in must reset the toggle and block Data Query again")
}

func TestRootReadOnlyEnvVar_BlocksWriteCommand(t *testing.T) {
	t.Setenv("ZR_READ_ONLY", "true")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		HttpClient: func() (*api.Client, error) {
			return api.NewClient(api.WithBaseURL(server.URL)), nil
		},
	}

	cmd := NewCmdRoot(f)
	cmd.SetArgs([]string{"account", "create", "--body", `{}`})
	err := cmd.Execute()
	require.Error(t, err)
	var roErr2 *api.ReadOnlyError
	assert.ErrorAs(t, err, &roErr2)
}

func TestRootReadOnlyFlag_SetsReadOnlyOnClient(t *testing.T) {
	ios, _, _, _ := iostreams.Test()

	f := &factory.Factory{
		IOStreams: ios,
		HttpClient: func() (*api.Client, error) {
			return api.NewClient(api.WithBaseURL("https://example.com")), nil
		},
	}

	cmd := NewCmdRoot(f)
	cmd.SetArgs([]string{"--read-only", "version"})
	require.NoError(t, cmd.Execute())

	// After PersistentPreRunE, the factory's HttpClient wrapper should set readOnly.
	client, err := f.HttpClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	// Verify the client is in read-only mode by attempting a write
	_, writeErr := client.Post("/v1/accounts", nil)
	require.Error(t, writeErr)
	var roErr *api.ReadOnlyError
	assert.ErrorAs(t, writeErr, &roErr)
}

func TestAllCommandsHaveShortHelp(t *testing.T) {
	// Every command surfaced by `zr help` needs a one-line Short, or the help
	// index shows a blank row. A leaf added without one should fail here.
	ios, _, _, _ := iostreams.Test()
	root := NewCmdRoot(&factory.Factory{IOStreams: ios})

	var missing []string
	walkCommands(root, func(c *cobra.Command) {
		if c.Hidden || c.Name() == "help" {
			return // hidden or cobra's auto-generated help shim
		}
		if strings.TrimSpace(c.Short) == "" {
			missing = append(missing, c.CommandPath())
		}
	})
	assert.Empty(t, missing, "these commands need a non-empty Short for `zr help`")
}

func TestRootSuggestsOnUnknownCommand(t *testing.T) {
	// A typo'd subcommand must be rejected with cobra's "did you mean" hint, not
	// silently treated as a positional arg.
	ios, _, _, _ := iostreams.Test()
	root := NewCmdRoot(&factory.Factory{IOStreams: ios})
	root.SetArgs([]string{"accoutn"}) // transposition of "account"

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown command "accoutn"`)
	assert.Contains(t, err.Error(), "Did you mean", "cobra should suggest the closest command")
}
