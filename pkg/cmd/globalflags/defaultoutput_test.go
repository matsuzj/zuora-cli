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

// A subcommand with a LOCAL --csv shadowing the root persistent flag must not
// see the json default applied when --csv was given at the ROOT level.
func TestApply_DefaultOutputJSON_RootLevelShadowedFlagWins(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetDefaultOutput("json"))
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "tok")

	root := &cobra.Command{Use: "zr"}
	globalflags.Register(root)
	sub := &cobra.Command{Use: "query", RunE: func(*cobra.Command, []string) error { return nil }}
	sub.Flags().Bool("csv", false, "local shadow")
	root.AddCommand(sub)

	require.NoError(t, root.PersistentFlags().Set("csv", "true")) // zr --csv query ...
	require.NoError(t, globalflags.Apply(f, sub))
	v, _ := sub.Flags().GetBool("json")
	assert.False(t, v, "an explicit root-level --csv must suppress the json default even when shadowed")
}

// TestApply_DefaultOutputJSON_SkippedOnTTY covers the human/TTY branch that was
// previously untestable: on an interactive terminal, default_output=json must
// NOT force --json (humans keep the readable table). Reachable now via
// SetTTYForTest (F-25). The piped counterpart is _AppliesWhenPiped above.
func TestApply_DefaultOutputJSON_SkippedOnTTY(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.SetTTYForTest(true) // simulate an interactive terminal
	cfg := config.NewMockConfig()
	require.NoError(t, cfg.SetDefaultOutput("json"))
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "tok")

	cmd := &cobra.Command{Use: "x"}
	globalflags.Register(cmd)
	require.NoError(t, cmd.ParseFlags(nil))
	require.NoError(t, globalflags.Apply(f, cmd))

	v, _ := cmd.Flags().GetBool("json")
	assert.False(t, v, "default_output=json must NOT apply on a TTY — humans get the table")
}
