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

func TestNew_HttpClientBuildsForDefaultEnv(t *testing.T) {
	// Isolated config dir: LoadDefault materializes the default (sandbox)
	// environment, which is all HttpClient needs to build a client. No
	// network and no credentials are required to construct it.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := New()

	client, err := f.HttpClient()
	require.NoError(t, err)
	require.NotNil(t, client, "HttpClient() must build a client from the default config/env")

	// Building the client must be idempotent: it derives everything from the
	// lazily-cached config, so a second call succeeds the same way.
	client2, err2 := f.HttpClient()
	require.NoError(t, err2)
	require.NotNil(t, client2)
}

func TestNew_HttpClientUsesCachedConfig(t *testing.T) {
	// HttpClient internally goes through f.Config(), so it must observe the
	// same cached config that Config() returns rather than reloading.
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	f := New()

	cfg, err := f.Config()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	client, err := f.HttpClient()
	require.NoError(t, err)
	require.NotNil(t, client)

	// The config returned after HttpClient ran must still be the same cached
	// instance (sync.Once was not re-triggered by HttpClient).
	cfgAfter, err := f.Config()
	require.NoError(t, err)
	assert.True(t, cfg == cfgAfter, "HttpClient() must reuse the cached config, not reload it")
}
