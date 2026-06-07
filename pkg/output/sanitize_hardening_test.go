package output

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSanitizeCell_BiDiAndLineSeparators covers the hardening added for table /
// detail output: ANSI escapes and Unicode format (Cf) controls are dropped, and
// the Unicode line/paragraph separators collapse to spaces.
func TestSanitizeCell_BiDiAndLineSeparators(t *testing.T) {
	got := sanitizeCell("\x1b[31mred\x1b[0m")
	assert.NotContains(t, got, "\x1b", "ESC byte must be stripped")
	assert.Contains(t, got, "red")

	// U+202E RIGHT-TO-LEFT OVERRIDE (filename-spoofing) is category Cf — dropped.
	assert.Equal(t, "userexe", sanitizeCell("user\u202Eexe"))
	// Zero-width space (U+200B, also Cf) — dropped.
	assert.Equal(t, "ab", sanitizeCell("a\u200Bb"))
	// Line / paragraph separators collapse to a space (they are not IsControl).
	assert.Equal(t, "a b", sanitizeCell("a\u2028b"))
	assert.Equal(t, "a b", sanitizeCell("a\u2029b"))
	// Ordinary text is untouched.
	assert.Equal(t, "Acme Inc.", sanitizeCell("Acme Inc."))
}

// TestSanitizeCSVCell covers the CSV-specific sanitizer: it strips terminal
// escapes and Cf/line-separator characters but PRESERVES newlines (encoding/csv
// quotes them), unlike the table sanitizer which collapses them.
func TestSanitizeCSVCell(t *testing.T) {
	assert.Equal(t, "line1\nline2", sanitizeCSVCell("line1\nline2"), "newline preserved")
	assert.NotContains(t, sanitizeCSVCell("a\x1b[31mb"), "\x1b", "ESC stripped")
	assert.NotContains(t, sanitizeCSVCell("user\u202Etxt"), "\u202E", "BiDi override stripped")
	assert.Equal(t, "a b", sanitizeCSVCell("a\rb"), "CR -> space")
	assert.Equal(t, "a b", sanitizeCSVCell("a\u2028b"), "line separator -> space")
}

// TestSanitizeCSVField_LeadingWhitespace covers that formula-injection
// classification skips leading whitespace (the bypass the audit found) while
// leaving genuine numbers and ordinary text untouched.
func TestSanitizeCSVField_LeadingWhitespace(t *testing.T) {
	assert.Equal(t, "'=cmd", sanitizeCSVField("=cmd"))
	assert.Equal(t, "' =cmd", sanitizeCSVField(" =cmd"), "leading space must not hide the formula")
	assert.Equal(t, "'\t@x", sanitizeCSVField("\t@x"), "leading tab must not hide the formula")
	assert.Equal(t, "' -cmd", sanitizeCSVField(" -cmd"), "leading space + non-numeric - is a formula")
	assert.Equal(t, "-10.50", sanitizeCSVField("-10.50"), "a real negative number is data, untouched")
	assert.Equal(t, "+42", sanitizeCSVField("+42"), "a real signed number is data, untouched")
	assert.Equal(t, "normal", sanitizeCSVField("normal"))
}

// TestPrintCSV_HardenedEndToEnd exercises the full PrintCSV path: escapes are
// gone, formulas are neutralized, and a legitimate multi-line cell survives.
func TestPrintCSV_HardenedEndToEnd(t *testing.T) {
	var buf bytes.Buffer
	rows := [][]string{
		{"a\x1b[31mEVIL", "=SUM(A1)"},
		{"multi\nline", "ok"},
	}
	cols := []Column{{Header: "C1"}, {Header: "C2"}}
	require.NoError(t, PrintCSV(&buf, rows, cols))

	out := buf.String()
	assert.NotContains(t, out, "\x1b", "no ESC byte may reach CSV output")
	assert.Contains(t, out, "'=SUM(A1)", "formula must be neutralized")
	assert.Contains(t, out, "multi\nline", "legitimate newline must be preserved (csv-quoted)")
}
