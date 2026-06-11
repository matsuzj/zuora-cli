package completion

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCompletion(f) }

func TestCompletionBash(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "completion", "bash")

	assert.NoError(t, err)
	assert.NotEmpty(t, stdout)
}

func TestCompletionZsh(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "completion", "zsh")

	assert.NoError(t, err)
	assert.NotEmpty(t, stdout)
}

func TestCompletionFish(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "completion", "fish")

	assert.NoError(t, err)
	assert.NotEmpty(t, stdout)
}

func TestCompletionInvalidShell(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "completion", "invalid")

	assert.Error(t, err)
}
