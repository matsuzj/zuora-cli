package output

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRenderJSONOnly covers the single entry point JSON-only commands end
// with: --csv is an explicit error (previously silently ignored — the user
// asked for CSV and got JSON), the JSON-family flags dispatch through
// RenderJSON, and the default is pretty-printed JSON.
func TestRenderJSONOnly(t *testing.T) {
	body := []byte(`{"id": "x-1", "amount": 100}`)

	t.Run("csv rejected", func(t *testing.T) {
		ios, _, out, _ := iostreams.Test()
		err := RenderJSONOnly(ios, body, FormatOptions{CSV: true})
		require.ErrorIs(t, err, ErrCSVUnsupportedJSONOnly)
		assert.Contains(t, err.Error(), "--csv is not supported for JSON-only output")
		assert.Empty(t, out.String(), "no output may precede the rejection")
	})

	t.Run("documented precedence: jq wins over csv", func(t *testing.T) {
		// README: --jq/--json/--template take precedence when combined with
		// --csv (the PR #54 regression taught us not to reject these combos).
		ios, _, out, _ := iostreams.Test()
		require.NoError(t, RenderJSONOnly(ios, body, FormatOptions{CSV: true, JQ: ".id"}))
		assert.Contains(t, out.String(), "x-1")
	})

	t.Run("documented precedence: json wins over csv", func(t *testing.T) {
		ios, _, out, _ := iostreams.Test()
		require.NoError(t, RenderJSONOnly(ios, body, FormatOptions{CSV: true, JSON: true}))
		assert.Contains(t, out.String(), `"amount"`)
	})

	t.Run("jq", func(t *testing.T) {
		ios, _, out, _ := iostreams.Test()
		require.NoError(t, RenderJSONOnly(ios, body, FormatOptions{JQ: ".id"}))
		assert.Contains(t, out.String(), "x-1")
		assert.NotContains(t, out.String(), "amount")
	})

	t.Run("json", func(t *testing.T) {
		ios, _, out, _ := iostreams.Test()
		require.NoError(t, RenderJSONOnly(ios, body, FormatOptions{JSON: true}))
		assert.Contains(t, out.String(), `"amount"`)
	})

	t.Run("template", func(t *testing.T) {
		ios, _, out, _ := iostreams.Test()
		require.NoError(t, RenderJSONOnly(ios, body, FormatOptions{Template: "{{.id}}"}))
		assert.Contains(t, out.String(), "x-1")
	})

	t.Run("default pretty JSON", func(t *testing.T) {
		ios, _, out, _ := iostreams.Test()
		require.NoError(t, RenderJSONOnly(ios, body, FormatOptions{}))
		assert.Contains(t, out.String(), `"id"`)
		assert.Contains(t, out.String(), `"amount"`)
	})
}

// TestRejectBareCSV pins the pre-request gate write commands use: only a
// BARE --csv is rejected; any JSON-family flag wins by documented precedence.
func TestRejectBareCSV(t *testing.T) {
	assert.ErrorIs(t, RejectBareCSV(FormatOptions{CSV: true}), ErrCSVUnsupportedJSONOnly)
	assert.NoError(t, RejectBareCSV(FormatOptions{CSV: true, JQ: ".id"}))
	assert.NoError(t, RejectBareCSV(FormatOptions{CSV: true, JSON: true}))
	assert.NoError(t, RejectBareCSV(FormatOptions{CSV: true, Template: "{{.id}}"}))
	assert.NoError(t, RejectBareCSV(FormatOptions{}))
}
