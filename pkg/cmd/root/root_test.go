package root

import (
	"github.com/matsuzj/zuora-cli/internal/testutil"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestRootJsonTemplateExclusion(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)
	cmd.SetArgs([]string{"version", "--json", "--template", "foo"})
	err := cmd.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --json and --template together")
}

func TestRootGlobalFlags(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	cmd := NewCmdRoot(f)

	flags := []string{"env", "json", "jq", "template", "zuora-version", "verbose", "read-only"}
	for _, name := range flags {
		assert.NotNil(t, cmd.PersistentFlags().Lookup(name), "missing flag: %s", name)
	}
}

func TestRootReadOnlyFlag_BlocksWriteCommand(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"acc-123","success":true}`))
	}))

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

func TestRootReadOnlyEnvVar_BlocksWriteCommand(t *testing.T) {
	t.Setenv("ZR_READ_ONLY", "true")

	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`{"success":true}`))
	}))

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
