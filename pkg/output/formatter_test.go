package output

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderJSON_CanonicalOrder(t *testing.T) {
	raw := []byte(`{"a":1}`)

	// JQ wins over JSON.
	ios, _, out, _ := iostreams.Test()
	handled, err := RenderJSON(ios, raw, FormatOptions{JQ: ".a", JSON: true})
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "1\n", out.String())

	// JSON when no JQ.
	ios2, _, out2, _ := iostreams.Test()
	handled, err = RenderJSON(ios2, raw, FormatOptions{JSON: true})
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Contains(t, out2.String(), `"a": 1`)

	// No machine flag → not handled, nothing written.
	ios3, _, out3, _ := iostreams.Test()
	handled, err = RenderJSON(ios3, raw, FormatOptions{CSV: true})
	require.NoError(t, err)
	assert.False(t, handled, "CSV is the caller's branch, not RenderJSON's")
	assert.Empty(t, out3.String())
}

func TestRenderJSON_InvalidJQPropagatesError(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	handled, err := RenderJSON(ios, []byte(`{"a":1}`), FormatOptions{JQ: ".["})
	assert.True(t, handled, "a JQ path was taken")
	require.Error(t, err, "an invalid jq filter must surface an error")
	assert.Empty(t, out.String(), "no partial output on a jq error")
}

func TestRenderJSON_InvalidTemplatePropagatesError(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	handled, err := RenderJSON(ios, []byte(`{"a":1}`), FormatOptions{Template: "{{.x"})
	assert.True(t, handled, "a Template path was taken")
	require.Error(t, err, "an invalid template must surface an error")
	assert.Empty(t, out.String(), "no partial output on a template error")
}

// TestRender_JQErrorDoesNotFallThroughToTable pins the (handled, err) contract:
// when --jq errors, Render must propagate the error and NOT also print the table
// (no double output).
func TestRender_JQErrorDoesNotFallThroughToTable(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	rows := [][]string{{"r1c1", "r1c2"}}
	cols := []Column{{Header: "A"}, {Header: "B"}}
	err := Render(ios, []byte(`{"a":1}`), FormatOptions{JQ: ".["}, rows, cols)
	require.Error(t, err)
	assert.Empty(t, out.String(), "the table must not print when jq errored")
}

func TestRenderDetail_JQErrorDoesNotFallThroughToFields(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	fields := []DetailField{{Key: "K", Value: "V"}}
	err := RenderDetail(ios, []byte(`{"a":1}`), FormatOptions{JQ: ".["}, fields)
	require.Error(t, err)
	assert.Empty(t, out.String(), "the detail view must not print when jq errored")
}

func TestRenderSuccess_HumanAndMachine(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	require.NoError(t, RenderSuccess(ios, FormatOptions{}, "Done.\n"))
	assert.Empty(t, out.String())
	assert.Equal(t, "Done.\n", errOut.String())

	ios2, _, out2, errOut2 := iostreams.Test()
	require.NoError(t, RenderSuccess(ios2, FormatOptions{JSON: true}, "Done.\n"))
	assert.Contains(t, out2.String(), `"success": true`)
	assert.Empty(t, errOut2.String())
}

// TestRenderJSON_JQBeatsJSON pins the precedence inside the central renderer:
// --jq is checked before --json, so when both are set the jq filter wins (F-20).
// (--json+--template is rejected upstream, so the reachable multi-flag cases are
// jq-vs-json and jq-vs-template.)
func TestRenderJSON_JQBeatsJSON(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	handled, err := RenderJSON(ios, []byte(`{"a":1,"b":2}`), FormatOptions{JQ: ".a", JSON: true})
	require.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "1\n", out.String(), "--jq must win over --json")
}
