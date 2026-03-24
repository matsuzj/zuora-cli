package completion

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.AddCommand(NewCmdCompletion(f))
	return root
}

func TestCompletionBash(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	root := newTestRoot(f)
	root.SetArgs([]string{"completion", "bash"})
	err := root.Execute()

	assert.NoError(t, err)
	assert.NotEmpty(t, out.String())
}

func TestCompletionZsh(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	root := newTestRoot(f)
	root.SetArgs([]string{"completion", "zsh"})
	err := root.Execute()

	assert.NoError(t, err)
	assert.NotEmpty(t, out.String())
}

func TestCompletionFish(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	root := newTestRoot(f)
	root.SetArgs([]string{"completion", "fish"})
	err := root.Execute()

	assert.NoError(t, err)
	assert.NotEmpty(t, out.String())
}

func TestCompletionInvalidShell(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{IOStreams: ios}

	root := newTestRoot(f)
	root.SetArgs([]string{"completion", "invalid"})
	err := root.Execute()

	assert.Error(t, err)
}
