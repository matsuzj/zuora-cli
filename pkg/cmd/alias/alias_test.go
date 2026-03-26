package alias

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFactory(t *testing.T) (*factory.Factory, *iostreams.IOStreams, string) {
	t.Helper()
	dir := t.TempDir()

	ios, _, _, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		Config: func() (config.Config, error) {
			return config.Load(dir)
		},
	}
	return f, ios, dir
}

func TestSetCommand(t *testing.T) {
	f, ios, dir := newTestFactory(t)

	cmd := NewCmdAlias(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"set", "ls", "account list"})
	err := cmd.Execute()
	require.NoError(t, err)

	// Verify alias was saved
	s := NewStore(dir)
	require.NoError(t, s.Load())
	cmd2, ok := s.Get("ls")
	assert.True(t, ok)
	assert.Equal(t, "account list", cmd2)
}

func TestDeleteCommand(t *testing.T) {
	f, ios, dir := newTestFactory(t)

	// Pre-create an alias
	s := NewStore(dir)
	require.NoError(t, s.Load())
	s.Set("ls", "account list")
	require.NoError(t, s.Save())

	cmd := NewCmdAlias(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"delete", "ls"})
	err := cmd.Execute()
	require.NoError(t, err)

	// Verify alias was deleted
	s2 := NewStore(dir)
	require.NoError(t, s2.Load())
	_, ok := s2.Get("ls")
	assert.False(t, ok)
}

func TestDeleteCommand_NotFound(t *testing.T) {
	f, ios, _ := newTestFactory(t)

	cmd := NewCmdAlias(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"delete", "nonexistent"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListCommand_Empty(t *testing.T) {
	f, ios, _ := newTestFactory(t)

	cmd := NewCmdAlias(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestListCommand_WithAliases(t *testing.T) {
	f, ios, dir := newTestFactory(t)

	// Pre-create aliases
	content := "alpha: a-command\nzulu: z-command\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "aliases.yml"), []byte(content), 0600))

	cmd := NewCmdAlias(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	require.NoError(t, err)
}

func TestSetCommand_ExactArgs(t *testing.T) {
	f, ios, _ := newTestFactory(t)

	cmd := NewCmdAlias(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"set", "onlyname"})
	err := cmd.Execute()
	assert.Error(t, err)
}
