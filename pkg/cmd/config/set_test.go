package config

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigSet_ActiveEnvironment(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "active_environment", "us-production"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "us-production", cfg.ActiveEnvironment())
	assert.Contains(t, out.String(), "Set active_environment to us-production")
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
	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "", "")

	root := newTestRoot(f)
	root.SetArgs([]string{"config", "set", "default_output", "json"})
	require.NoError(t, root.Execute())
	assert.Equal(t, "json", cfg.DefaultOutput())
	assert.Contains(t, out.String(), "Set default_output to json")
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
