package alias

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/globalflags"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
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

// TestListCommand_JSON pins the output-consistency fix (#453): `alias list
// --json` must emit a structured JSON array, not the silently-ignored
// tab-separated text. Removing the format-flag branch makes json.Unmarshal fail.
func TestListCommand_JSON(t *testing.T) {
	dir := t.TempDir()
	content := "alpha: a-command\nzulu: z-command\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "aliases.yml"), []byte(content), 0600))

	ios, _, out, _ := iostreams.Test()
	f := &factory.Factory{
		IOStreams: ios,
		Config:    func() (config.Config, error) { return config.Load(dir) },
	}

	root := newTestRootWithAlias(f, ios)
	globalflags.Register(root)
	root.SetArgs([]string{"alias", "list", "--json"})
	require.NoError(t, root.Execute())

	var got []map[string]interface{}
	require.NoError(t, json.Unmarshal(out.Bytes(), &got),
		"alias list --json must emit valid JSON, not tab-separated text")
	require.Len(t, got, 2)
	assert.Equal(t, "alpha", got[0]["name"])
	assert.Equal(t, "a-command", got[0]["command"])
}

func TestSetCommand_ExactArgs(t *testing.T) {
	f, ios, _ := newTestFactory(t)

	cmd := NewCmdAlias(f)
	cmd.SetOut(ios.Out)
	cmd.SetErr(ios.ErrOut)
	cmd.SetArgs([]string{"set", "onlyname"})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 2 arg(s), received 1")
}

// newTestRootWithAlias builds a zr-shaped root: the alias group plus a dummy
// "account" command, so runSet's cmd.Root()-derived reserved set includes a
// realistic builtin.
func newTestRootWithAlias(f *factory.Factory, ios *iostreams.IOStreams) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.AddCommand(NewCmdAlias(f))
	root.AddCommand(&cobra.Command{Use: "account", Run: func(*cobra.Command, []string) {}})
	root.SetOut(ios.Out)
	root.SetErr(ios.ErrOut)
	return root
}

func TestSetCommand_RejectsReservedName(t *testing.T) {
	f, ios, dir := newTestFactory(t)

	root := newTestRootWithAlias(f, ios)
	root.SetArgs([]string{"alias", "set", "account", "contact list"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"account" is a built-in command`)

	// Nothing was written.
	s := NewStore(dir)
	require.NoError(t, s.Load())
	_, ok := s.Get("account")
	assert.False(t, ok)
}

func TestSetCommand_RejectsHelp(t *testing.T) {
	f, ios, _ := newTestFactory(t)

	root := newTestRootWithAlias(f, ios)
	root.SetArgs([]string{"alias", "set", "help", "account list"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "built-in command")
}

func TestSetCommand_RejectsSelfReference(t *testing.T) {
	f, ios, dir := newTestFactory(t)

	root := newTestRootWithAlias(f, ios)
	root.SetArgs([]string{"alias", "set", "loop", "loop --json"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `alias "loop" would invoke itself`)

	s := NewStore(dir)
	require.NoError(t, s.Load())
	_, ok := s.Get("loop")
	assert.False(t, ok)
}

func TestSetCommand_RejectsMalformedExpansion(t *testing.T) {
	f, ios, _ := newTestFactory(t)

	root := newTestRootWithAlias(f, ios)
	root.SetArgs([]string{"alias", "set", "bad", `query "SELECT unbalanced`})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "malformed expansion")
}

func TestSetCommand_RejectsEmptyExpansion(t *testing.T) {
	f, ios, _ := newTestFactory(t)

	root := newTestRootWithAlias(f, ios)
	root.SetArgs([]string{"alias", "set", "blank", "   "})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func TestDeleteCommand_AllowsReservedName(t *testing.T) {
	// delete stays permissive so pre-existing polluted entries (created
	// before the set-side guard) can be cleaned up.
	f, ios, dir := newTestFactory(t)

	s := NewStore(dir)
	require.NoError(t, s.Load())
	s.Set("account", "contact list")
	require.NoError(t, s.Save())

	root := newTestRootWithAlias(f, ios)
	root.SetArgs([]string{"alias", "delete", "account"})
	require.NoError(t, root.Execute())

	s2 := NewStore(dir)
	require.NoError(t, s2.Load())
	_, ok := s2.Get("account")
	assert.False(t, ok)
}
