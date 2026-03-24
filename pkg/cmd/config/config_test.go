package config

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.AddCommand(NewCmdConfig(f))
	return root
}

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
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "unknown_key", "value"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown config key")
}

func TestConfigGet(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "get", "active_environment"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, "sandbox\n", out.String())
}

func TestConfigList(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "list"})
	err := root.Execute()

	require.NoError(t, err)
	output := out.String()
	assert.Contains(t, output, "active_environment: sandbox")
	assert.Contains(t, output, "zuora_version: 2025-08-12")
	assert.Contains(t, output, "sandbox")
	assert.Contains(t, output, "rest.apisandbox.zuora.com")
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
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "env", "nonexistent"})
	err := root.Execute()

	assert.Error(t, err)
}
