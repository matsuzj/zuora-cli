package config

import (
	"fmt"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigSet_ActiveEnvironment(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "active_environment", "us-production"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "us-production", cfg.ActiveEnvironment())
	assert.Contains(t, errOut.String(), "Set active_environment to us-production")
	assert.Empty(t, out.String(), "stdout is reserved for data (#453/#519)")
	assert.Equal(t, 1, cfg.SaveCallCount)
}

func TestConfigSet_ActiveEnvironment_Unknown(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "active_environment", "does-not-exist"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown environment")
	assert.Equal(t, 0, cfg.SaveCallCount, "a failed set must not persist")
}

func TestConfigSet_DefaultOutput(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "default_output", "json"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "json", cfg.DefaultOutput())
	assert.Contains(t, errOut.String(), "Set default_output to json")
	assert.Empty(t, out.String(), "stdout is reserved for data (#453/#519)")
}

func TestConfigSet_DefaultOutput_Invalid(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "default_output", "xml"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output format")
	assert.Equal(t, 0, cfg.SaveCallCount, "a failed set must not persist")
}

// TestConfigSet_SaveError covers the branch where the value is accepted but
// persisting it fails: the error must propagate (Save was still attempted).
func TestConfigSet_SaveError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	cfg.SaveError = fmt.Errorf("disk full")
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "default_output", "json"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
	assert.Equal(t, 1, cfg.SaveCallCount, "Save must have been attempted before the error")
}

// TestConfigSet_ConfigError covers the early return when the factory cannot
// load the configuration at all.
func TestConfigSet_ConfigError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return nil, fmt.Errorf("cannot load config")
		},
	}

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "default_output", "json"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot load config")
}
