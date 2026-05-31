package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoad_CorruptConfigYML asserts that a malformed config.yml surfaces a
// wrapped "reading config.yml" error instead of silently falling back to
// defaults (which would hide a corrupted file from the user).
func TestLoad_CorruptConfigYML(t *testing.T) {
	dir := t.TempDir()
	// A scalar where a mapping is expected is invalid for configData.
	writeFile(t, configFilePath(dir), "not: [valid: yaml: : :\n  - broken")

	cfg, err := Load(dir)
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "reading config.yml")
}

// TestLoad_CorruptEnvironmentsYML asserts a malformed environments.yml is
// reported rather than masked by the default environment set.
func TestLoad_CorruptEnvironmentsYML(t *testing.T) {
	dir := t.TempDir()
	// environments must be a map; a sequence cannot unmarshal into it.
	writeFile(t, environmentsFilePath(dir), "environments:\n  - this is a list not a map\n")

	cfg, err := Load(dir)
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "reading environments.yml")
}

// TestLoad_CorruptTokensYML asserts a malformed tokens.yml is reported. A
// silent fallback here would discard a real (but unreadable) token cache.
func TestLoad_CorruptTokensYML(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, tokensFilePath(dir), "tokens: {unterminated\n")

	cfg, err := Load(dir)
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "reading tokens.yml")
}

// TestLoad_TokensNull confirms that a tokens.yml with an explicit null map
// loads successfully and yields a usable (empty, non-nil) token store, so
// SetToken/Token keep working without a nil-map panic.
func TestLoad_TokensNull(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, tokensFilePath(dir), "tokens: null\n")

	cfg, err := Load(dir)
	require.NoError(t, err)

	tok, err := cfg.Token("sandbox")
	require.NoError(t, err)
	assert.Nil(t, tok)

	// The map must be initialized: writing then reading back must work.
	require.NoError(t, cfg.SetToken("sandbox", &TokenEntry{AccessToken: "x"}))
	got, err := cfg.Token("sandbox")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "x", got.AccessToken)
}

// TestLoad_EmptyFileLoadsDefaults confirms an existing-but-empty config.yml
// is treated like a missing file: defaults apply rather than an error.
func TestLoad_EmptyFileLoadsDefaults(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, configFilePath(dir), "")

	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, "sandbox", cfg.ActiveEnvironment())
	assert.Equal(t, "2025-08-12", cfg.ZuoraVersion())
	assert.Equal(t, "table", cfg.DefaultOutput())
}

// TestLoad_NotExistLoadsDefaults confirms the os.IsNotExist branch: a config
// directory with no files at all loads cleanly to defaults.
func TestLoad_NotExistLoadsDefaults(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist-yet")
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, "sandbox", cfg.ActiveEnvironment())
	envs := cfg.Environments()
	assert.Contains(t, envs, "sandbox")
}

// TestSave_UnwritableDir guards the atomic-write path: when the target
// directory cannot be written, Save must return an error (CreateTemp fails)
// rather than panic or silently lose data.
func TestSave_UnwritableDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: chmod is ignored, cannot make dir unwritable")
	}
	dir := t.TempDir()
	// The config dir already exists (so MkdirAll succeeds) but is read+exec
	// only, so os.CreateTemp inside writeYAML cannot create the temp file.
	require.NoError(t, os.Chmod(dir, 0500))
	t.Cleanup(func() { _ = os.Chmod(dir, 0700) })

	cfg, err := Load(dir)
	require.NoError(t, err)

	err = cfg.Save()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "writing config.yml")
}
