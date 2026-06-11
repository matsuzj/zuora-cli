package config

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestRoot is used by set_test.go and by tests that need direct cfg access.
func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.AddCommand(NewCmdConfig(f))
	return root
}

// newCmd adapts NewCmdConfig for use with cmdtest.Run (parent="").
func newCmd(f *factory.Factory) *cobra.Command { return NewCmdConfig(f) }

func TestConfigSet(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "zuora_version", "2026-01-01"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, "2026-01-01", cfg.ZuoraVersion())
	assert.Contains(t, out.String(), "Set zuora_version to 2026-01-01")
	assert.Equal(t, 1, cfg.SaveCallCount)
}

func TestConfigSet_InvalidKey(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "set", "unknown_key", "value")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestConfigGet(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "get", "active_environment")

	require.NoError(t, err)
	assert.Equal(t, "sandbox\n", stdout)
}

func TestConfigList(t *testing.T) {
	stdout, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "list")

	require.NoError(t, err)
	assert.Contains(t, stdout, "active_environment: sandbox")
	assert.Contains(t, stdout, "zuora_version: 2025-08-12")
	assert.Contains(t, stdout, "sandbox")
	assert.Contains(t, stdout, "rest.apisandbox.zuora.com")
}

func TestConfigEnv(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "env", "us-production"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, "us-production", cfg.ActiveEnvironment())
	assert.Contains(t, out.String(), "Switched to environment us-production")
}

func TestConfigEnv_Invalid(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "config", "env", "nonexistent")

	assert.Error(t, err)
}
