package output

import (
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Large integers (Zuora's 19-digit IDs) must survive --jq and --template
// without float64 rounding.
func TestPrintJSON_JQ_PreservesLargeInteger(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{"Id":1234567890123456789}`)
	require.NoError(t, PrintJSON(ios, data, ".Id"))
	assert.Contains(t, out.String(), "1234567890123456789")
	assert.NotContains(t, out.String(), "1234567890123456800")
}

func TestPrintTemplate_PreservesLargeInteger(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	data := []byte(`{"Id":1234567890123456789}`)
	require.NoError(t, PrintTemplate(ios, data, "{{.Id}}"))
	assert.Contains(t, out.String(), "1234567890123456789")
}

func TestPrintJSON_InvalidJSON_ErrorsToStderr(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	err := PrintJSON(ios, []byte(`not json`), "")
	require.Error(t, err, "invalid JSON must surface a non-nil error for scripts")
	assert.Empty(t, out.String(), "invalid JSON must not be written to stdout")
	assert.Contains(t, errOut.String(), "not json")
}

func TestPrintJSON_EmptyBody_NoError(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	assert.NoError(t, PrintJSON(ios, []byte(""), ""))
}

func TestPrintJSON_JQ_EmptyResult_NoBlankLine(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	// .missing[] yields nothing.
	require.NoError(t, PrintJSON(ios, []byte(`{"items":[]}`), ".items[]"))
	assert.Empty(t, out.String(), "an empty jq stream must print nothing, not a blank line")
}

func TestPrintCSV_NeutralizesFormulaInjection(t *testing.T) {
	var buf strings.Builder
	cols := []Column{{Header: "Name"}}
	rows := [][]string{
		{"=1+1"},
		{"+cmd|'/C calc'"},
		{"-1+1+cmd"},
		{"@SUM(A1)"},
		{"safe"},
	}
	require.NoError(t, PrintCSV(&buf, rows, cols))
	got := buf.String()
	assert.Contains(t, got, "'=1+1")
	assert.Contains(t, got, "'+cmd")
	assert.Contains(t, got, "'-1+1+cmd")
	assert.Contains(t, got, "'@SUM(A1)")
	assert.Contains(t, got, "safe")
	assert.NotContains(t, got, "'safe")
}

func TestPrintCSV_PreservesNegativeNumbers(t *testing.T) {
	var buf strings.Builder
	cols := []Column{{Header: "Amount"}}
	rows := [][]string{{"-10.50"}, {"-5"}, {"+3"}, {"-1e3"}}
	require.NoError(t, PrintCSV(&buf, rows, cols))
	got := buf.String()
	// Legitimate signed numbers must NOT be quoted (no data corruption).
	assert.NotContains(t, got, "'-10.50")
	assert.NotContains(t, got, "'-5")
	assert.NotContains(t, got, "'+3")
	assert.Contains(t, got, "-10.50")
}

func TestPrintTable_SanitizesControlChars(t *testing.T) {
	var buf strings.Builder
	cols := []Column{{Header: "V"}}
	rows := [][]string{{"line1\nline2\x1b[31mred"}}
	require.NoError(t, PrintTable(&buf, rows, cols))
	got := buf.String()
	assert.NotContains(t, got, "\x1b", "ANSI/control sequences must be stripped from cells")
}
