package globalflags_test

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func applyWithDefaultOutput(t *testing.T, defaultOutput string, args ...string) *cobra.Command {
	t.Helper()
	ios, _, _, _ := iostreams.Test() // buffer-backed: stdout is NOT a terminal
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetDefaultOutput(defaultOutput))
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags(args))
	require.NoError(t, globalflags.Apply(f, cmd))
	return cmd
}

// default_output=json + no format flag + non-TTY stdout → behaves as --json.
func TestApply_DefaultOutputJSON_AppliesWhenPiped(t *testing.T) {
	cmd := applyWithDefaultOutput(t, "json")
	v, _ := cmd.Flags().GetBool("json")
	assert.True(t, v, "default_output=json must act as the --json default when piped")
}

func TestApply_DefaultOutputTable_NoEffect(t *testing.T) {
	cmd := applyWithDefaultOutput(t, "table")
	v, _ := cmd.Flags().GetBool("json")
	assert.False(t, v)
}

// An explicit format flag always wins over the configured default.
func TestApply_DefaultOutputJSON_ExplicitFlagWins(t *testing.T) {
	cmd := applyWithDefaultOutput(t, "json", "--csv")
	v, _ := cmd.Flags().GetBool("json")
	assert.False(t, v, "an explicit --csv must suppress the json default")

	cmd = applyWithDefaultOutput(t, "json", "--template", "{{.id}}")
	v, _ = cmd.Flags().GetBool("json")
	assert.False(t, v, "an explicit --template must suppress the json default")
}
