package version

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdVersion(f) }

func TestVersionOutput(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "version")

	require.NoError(t, err)
	assert.Contains(t, stdout, "zr version")
}

func TestVersionJSON(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "version", "--json")

	require.NoError(t, err)
	assert.Contains(t, stdout, `"version"`)
	assert.Contains(t, stdout, `"commit"`)
}

func TestVersionJQ(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "version", "--jq", ".version")

	require.NoError(t, err)
	assert.Contains(t, stdout, "dev")
}

// TestVersionCSV pins the --csv branch added with the #519 funnel fix
// (version shared config get's gap: --csv fell through to human text).
func TestVersionCSV(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "version", "--csv")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Field,Value")
	assert.Contains(t, stdout, "Version,")
}
