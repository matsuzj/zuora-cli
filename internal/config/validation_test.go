package config

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateBaseURL(t *testing.T) {
	require.NoError(t, ValidateBaseURL("https://rest.zuora.com"))
	require.NoError(t, ValidateBaseURL("http://localhost:8080"))
	assert.Error(t, ValidateBaseURL(""))
	assert.Error(t, ValidateBaseURL("not a url"))
	assert.Error(t, ValidateBaseURL("ftp://x"))
	assert.Error(t, ValidateBaseURL("/relative/path"))
}

func TestAddEnvironment_RejectsInvalidURL(t *testing.T) {
	c, err := Load(t.TempDir())
	require.NoError(t, err)
	assert.Error(t, c.AddEnvironment("bad", &Environment{BaseURL: ""}))
	assert.Error(t, c.AddEnvironment("bad", &Environment{BaseURL: "nonsense"}))
	assert.NoError(t, c.AddEnvironment("good", &Environment{BaseURL: "https://rest.zuora.com"}))
}

func TestSetZuoraVersion_Validation(t *testing.T) {
	c, err := Load(t.TempDir())
	require.NoError(t, err)
	assert.NoError(t, c.SetZuoraVersion("2025-08-12"))
	assert.Error(t, c.SetZuoraVersion("latest"))
	assert.Error(t, c.SetZuoraVersion("2025/08/12"))
}

func TestSave_AtomicAndPermissions(t *testing.T) {
	dir := t.TempDir()
	c, err := Load(dir)
	require.NoError(t, err)
	require.NoError(t, c.SetToken("sandbox", &TokenEntry{AccessToken: "secret"}))
	require.NoError(t, c.Save())

	// tokens.yml must be 0600 (secret at rest).
	info, err := os.Stat(filepath.Join(dir, "tokens.yml"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// No leftover temp files from the atomic write.
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".zr-", "atomic write must not leave temp files behind")
	}
}

func TestSave_ConcurrentNoRace(t *testing.T) {
	dir := t.TempDir()
	c, err := Load(dir)
	require.NoError(t, err)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = c.SetToken("sandbox", &TokenEntry{AccessToken: "t"})
			_, _ = c.Token("sandbox")
			_ = c.Save()
		}(i)
	}
	wg.Wait()
}
