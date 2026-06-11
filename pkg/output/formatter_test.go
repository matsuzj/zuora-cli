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
