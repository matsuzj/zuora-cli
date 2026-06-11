package output

import (
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// PrintRawOrJSON must pass a non-JSON body through verbatim, with NO added
// trailing newline, so a binary/exact-byte `zr api ... > file` is not corrupted.
func TestPrintRawOrJSON_NonJSON_Verbatim(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	body := []byte("\x00\x01rawbytes-no-newline")
	require.NoError(t, PrintRawOrJSON(ios, body))
	assert.Equal(t, body, out.Bytes(), "non-JSON body must be written byte-for-byte with no added newline")
}

// Valid JSON is still pretty-printed.
func TestPrintRawOrJSON_JSON_Pretty(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	require.NoError(t, PrintRawOrJSON(ios, []byte(`{"a":1}`)))
	assert.Contains(t, out.String(), "\"a\": 1")
}

// --jq / --template decoding must reject trailing garbage after the first JSON
// value, matching strict json.Unmarshal behavior.
func TestPrintJSON_JQ_RejectsTrailingGarbage(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := PrintJSON(ios, []byte(`{"ok":true} and then junk`), ".ok")
	require.Error(t, err, "trailing data after the JSON value must be rejected")
}

func TestPrintTemplate_RejectsTrailingGarbage(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	err := PrintTemplate(ios, []byte(`{"ok":true} junk`), "{{.ok}}")
	require.Error(t, err)
}

// isNumeric must accept plain decimals (so signed amounts are not corrupted) and
// reject the Go-only forms (Inf/NaN/hex) that still begin with a formula trigger.
func TestIsNumeric_StrictDecimalOnly(t *testing.T) {
	for _, ok := range []string{"-10.50", "+3", "-5", "1e3", "-1.2e-3", ".5", "0"} {
		assert.True(t, isNumeric(ok), "%q should be numeric", ok)
	}
	for _, bad := range []string{"+Inf", "-Inf", "NaN", "0x1p-2", "-0x1.5p3", "1_000", "+cmd", "=1+1", "-1+1"} {
		assert.False(t, isNumeric(bad), "%q must NOT be treated as numeric", bad)
	}
}

// The dangerous Go-float forms must therefore be quoted in CSV output.
func TestPrintCSV_QuotesInfAndHex(t *testing.T) {
	var buf strings.Builder
	cols := []Column{{Header: "V"}}
	rows := [][]string{{"+Inf"}, {"-0x1p2"}, {"-10.50"}}
	require.NoError(t, PrintCSV(&buf, rows, cols))
	got := buf.String()
	assert.Contains(t, got, "'+Inf")
	assert.Contains(t, got, "'-0x1p2")
	assert.NotContains(t, got, "'-10.50") // legitimate number stays unquoted
}
