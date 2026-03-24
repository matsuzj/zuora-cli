package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefault_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, "sandbox", cfg.ActiveEnvironment())
	assert.Equal(t, "2025-08-12", cfg.ZuoraVersion())
	assert.Equal(t, "table", cfg.DefaultOutput())

	envs := cfg.Environments()
	assert.Contains(t, envs, "sandbox")
	assert.Equal(t, "https://rest.apisandbox.zuora.com", envs["sandbox"].BaseURL)
}

func TestLoadExistingConfig(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "config.yml"), `
active_environment: us-production
zuora_version: "2026-01-01"
default_output: json
`)
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.Equal(t, "us-production", cfg.ActiveEnvironment())
	assert.Equal(t, "2026-01-01", cfg.ZuoraVersion())
	assert.Equal(t, "json", cfg.DefaultOutput())
}

func TestSetActiveEnvironment(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	require.NoError(t, err)

	err = cfg.SetActiveEnvironment("us-production")
	assert.NoError(t, err)
	assert.Equal(t, "us-production", cfg.ActiveEnvironment())

	err = cfg.SetActiveEnvironment("nonexistent")
	assert.Error(t, err)
}

func TestSetDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	require.NoError(t, err)

	assert.NoError(t, cfg.SetDefaultOutput("json"))
	assert.Equal(t, "json", cfg.DefaultOutput())

	assert.Error(t, cfg.SetDefaultOutput("csv"))
}

func TestTokenOperations(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	require.NoError(t, err)

	token, err := cfg.Token("sandbox")
	assert.NoError(t, err)
	assert.Nil(t, token)

	entry := &TokenEntry{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	err = cfg.SetToken("sandbox", entry)
	assert.NoError(t, err)

	token, err = cfg.Token("sandbox")
	assert.NoError(t, err)
	assert.Equal(t, "test-token", token.AccessToken)
	assert.True(t, token.IsValid())

	err = cfg.RemoveToken("sandbox")
	assert.NoError(t, err)

	token, err = cfg.Token("sandbox")
	assert.NoError(t, err)
	assert.Nil(t, token)
}

func TestTokenIsValid(t *testing.T) {
	valid := &TokenEntry{
		AccessToken: "token",
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}
	assert.True(t, valid.IsValid())

	expiringSoon := &TokenEntry{
		AccessToken: "token",
		ExpiresAt:   time.Now().Add(30 * time.Second),
	}
	assert.False(t, expiringSoon.IsValid())

	expired := &TokenEntry{
		AccessToken: "token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	assert.False(t, expired.IsValid())

	assert.False(t, (*TokenEntry)(nil).IsValid())
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	require.NoError(t, err)

	require.NoError(t, cfg.SetActiveEnvironment("us-production"))
	require.NoError(t, cfg.SetToken("sandbox", &TokenEntry{
		AccessToken: "saved-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}))
	require.NoError(t, cfg.Save())

	cfg2, err := Load(dir)
	require.NoError(t, err)
	assert.Equal(t, "us-production", cfg2.ActiveEnvironment())

	token, err := cfg2.Token("sandbox")
	assert.NoError(t, err)
	assert.Equal(t, "saved-token", token.AccessToken)
}

func TestAddAndRemoveEnvironment(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	require.NoError(t, err)

	err = cfg.AddEnvironment("custom", &Environment{BaseURL: "https://custom.zuora.com"})
	assert.NoError(t, err)

	env, err := cfg.Environment("custom")
	assert.NoError(t, err)
	assert.Equal(t, "https://custom.zuora.com", env.BaseURL)

	err = cfg.RemoveEnvironment("custom")
	assert.NoError(t, err)

	_, err = cfg.Environment("custom")
	assert.Error(t, err)
}

func TestXDGConfigHome(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	assert.Equal(t, "/custom/config/zr", configDir())
}

func TestXDGConfigHomeDefault(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	dir := configDir()
	home, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(home, ".config", "zr"), dir)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0600))
}
