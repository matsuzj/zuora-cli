package output

import (
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PrintDetail must strip ANSI/control characters from API-controlled field
// values, matching the table path, to prevent terminal-escape injection.
func TestPrintDetail_SanitizesControlChars(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	fields := []DetailField{
		{Key: "Name", Value: "ev\x1b[31mil\x1b[0m"},
		{Key: "Note", Value: "line1\nline2\twith\rcontrol"},
	}
	require.NoError(t, PrintDetail(ios, fields))

	s := out.String()
	assert.NotContains(t, s, "\x1b", "ANSI escape (ESC) must be stripped from detail output")
	assert.NotContains(t, s, "\n\n", "embedded newlines must be collapsed, not break layout")
	assert.NotContains(t, s, "\r", "carriage returns must be collapsed")
	assert.Contains(t, s, "Name", "the key/label is still rendered")
}

// A value with no control characters must pass through unchanged.
func TestPrintDetail_PlainValueUnchanged(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	require.NoError(t, PrintDetail(ios, []DetailField{{Key: "Amount", Value: "1000000"}}))
	assert.True(t, strings.Contains(out.String(), "1000000"))
}
