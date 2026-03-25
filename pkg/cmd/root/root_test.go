package root

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
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

	flags := []string{"env", "json", "jq", "template", "zuora-version", "verbose"}
	for _, name := range flags {
		assert.NotNil(t, cmd.PersistentFlags().Lookup(name), "missing flag: %s", name)
	}
}
