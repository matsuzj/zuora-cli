package cmdutil

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveBody_Literal(t *testing.T) {
	r, err := ResolveBody(`{"a":1}`, nil)
	require.NoError(t, err)
	b, _ := io.ReadAll(r)
	assert.Equal(t, `{"a":1}`, string(b))
}

func TestResolveBody_Stdin(t *testing.T) {
	r, err := ResolveBody("-", strings.NewReader(`{"from":"stdin"}`))
	require.NoError(t, err)
	b, _ := io.ReadAll(r)
	assert.Equal(t, `{"from":"stdin"}`, string(b))
}

func TestResolveBody_File(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "body.json")
	require.NoError(t, os.WriteFile(p, []byte(`{"file":true}`), 0600))

	r, err := ResolveBody("@"+p, nil)
	require.NoError(t, err)
	b, _ := io.ReadAll(r)
	assert.Equal(t, `{"file":true}`, string(b))
}

func TestResolveBody_MissingFile(t *testing.T) {
	_, err := ResolveBody("@/nonexistent/zzz.json", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading body file")
}

func TestResolveBody_Empty(t *testing.T) {
	r, err := ResolveBody("", nil)
	require.NoError(t, err)
	b, _ := io.ReadAll(r)
	assert.Empty(t, string(b))
}

func TestResolveBody_MidStringAtIsLiteral(t *testing.T) {
	// Only a leading '@' means a file; "user@example.com" is literal text.
	r, err := ResolveBody("user@example.com", nil)
	require.NoError(t, err)
	b, _ := io.ReadAll(r)
	assert.Equal(t, "user@example.com", string(b))
}

func TestRequireConfirm(t *testing.T) {
	assert.NoError(t, RequireConfirm(true))
	err := RequireConfirm(false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
