package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_WiresDependencies(t *testing.T) {
	f := New()
	require.NotNil(t, f)
	assert.NotNil(t, f.IOStreams)
	assert.NotNil(t, f.Config)
	assert.NotNil(t, f.HttpClient)
	assert.NotNil(t, f.AuthToken)
}

func TestNew_ConfigCachedOnce(t *testing.T) {
	// Point config at an isolated dir so this does not touch the real config.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := New()

	c1, err1 := f.Config()
	c2, err2 := f.Config()
	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, c1)
	// sync.Once must return the same cached instance on every call.
	assert.True(t, c1 == c2, "Config() must cache and return the same instance")
}
