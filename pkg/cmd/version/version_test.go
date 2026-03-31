package version

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	root.AddCommand(NewCmdVersion(f))
	return root
}

func TestVersionOutput(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	root := newTestRoot(f)
	root.SetArgs([]string{"version"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "zr version")
}

func TestVersionJSON(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	root := newTestRoot(f)
	root.SetArgs([]string{"version", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, `"version"`)
	assert.Contains(t, output, `"commit"`)
}
