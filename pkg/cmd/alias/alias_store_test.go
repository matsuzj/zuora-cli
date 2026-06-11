package alias

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.Load())

	s.Set("ls", "account list")
	cmd, ok := s.Get("ls")
	assert.True(t, ok)
	assert.Equal(t, "account list", cmd)
}

func TestStore_DeleteExisting(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.Load())

	s.Set("ls", "account list")
	err := s.Delete("ls")
	assert.NoError(t, err)

	_, ok := s.Get("ls")
	assert.False(t, ok)
}

func TestStore_DeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.Load())

	err := s.Delete("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestStore_All_Sorted(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.Load())

	s.Set("zulu", "z-command")
	s.Set("alpha", "a-command")
	s.Set("mike", "m-command")

	entries := s.All()
	require.Len(t, entries, 3)
	assert.Equal(t, "alpha", entries[0].Name)
	assert.Equal(t, "mike", entries[1].Name)
	assert.Equal(t, "zulu", entries[2].Name)
}

func TestStore_SaveAndReload(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.Load())

	s.Set("ls", "account list")
	s.Set("sub", "subscription get")
	require.NoError(t, s.Save())

	// Reload
	s2 := NewStore(dir)
	require.NoError(t, s2.Load())

	cmd, ok := s2.Get("ls")
	assert.True(t, ok)
	assert.Equal(t, "account list", cmd)

	cmd, ok = s2.Get("sub")
	assert.True(t, ok)
	assert.Equal(t, "subscription get", cmd)

	assert.Len(t, s2.All(), 2)
}

func TestStore_LoadEmptyDir(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.Load())
	assert.Len(t, s.All(), 0)
}

func TestStore_LoadExistingFile(t *testing.T) {
	dir := t.TempDir()
	content := "ls: account list\nsub: subscription get\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "aliases.yml"), []byte(content), 0600))

	s := NewStore(dir)
	require.NoError(t, s.Load())
	assert.Len(t, s.All(), 2)

	cmd, ok := s.Get("ls")
	assert.True(t, ok)
	assert.Equal(t, "account list", cmd)
}

func TestStore_OverwriteExisting(t *testing.T) {
	dir := t.TempDir()
	s := NewStore(dir)
	require.NoError(t, s.Load())

	s.Set("ls", "account list")
	s.Set("ls", "subscription list")

	cmd, ok := s.Get("ls")
	assert.True(t, ok)
	assert.Equal(t, "subscription list", cmd)
	assert.Len(t, s.All(), 1)
}
