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
	// CONTRACT CHANGE (P5-2): an empty body used to resolve to an empty
	// reader; with cobra-required flags an explicitly-empty --body ("", or
	// an unset shell variable) satisfies the required check, so ResolveBody
	// now fails fast instead of letting an empty body reach Zuora.
	_, err := ResolveBody("", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request body is empty")
}

func TestResolveBody_MidStringAtIsLiteral(t *testing.T) {
	// Only a LEADING '@' means a file; an '@' inside the (valid JSON) body is
	// literal text, not a file reference.
	r, err := ResolveBody(`{"email":"user@example.com"}`, nil)
	require.NoError(t, err)
	b, _ := io.ReadAll(r)
	assert.Equal(t, `{"email":"user@example.com"}`, string(b))
}

func TestResolveBody_InvalidJSONRejected(t *testing.T) {
	// Malformed JSON is caught locally (F-22) — from a literal, a file, and
	// stdin — instead of being sent to Zuora for a confusing server-side 4xx.
	t.Run("literal", func(t *testing.T) {
		_, err := ResolveBody(`{"a":`, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not valid JSON")
	})
	t.Run("stdin", func(t *testing.T) {
		_, err := ResolveBody("-", strings.NewReader("not json at all"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not valid JSON")
	})
	t.Run("file", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "bad.json")
		require.NoError(t, os.WriteFile(p, []byte("{oops}"), 0600))
		_, err := ResolveBody("@"+p, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not valid JSON")
	})
}

func TestRequireConfirm(t *testing.T) {
	assert.NoError(t, RequireConfirm(true))
	err := RequireConfirm(false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "irreversible")
}
