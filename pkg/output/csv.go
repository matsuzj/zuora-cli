package output

import (
	"encoding/csv"
	"io"
	"regexp"
)

// PrintCSV writes data as CSV, neutralizing spreadsheet formula injection.
func PrintCSV(w io.Writer, rows [][]string, columns []Column) error {
	writer := csv.NewWriter(w)

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = sanitizeCSVField(sanitizeCSVCell(col.Header))
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		sanitized := make([]string, len(row))
		for i, v := range row {
			sanitized[i] = sanitizeCSVField(sanitizeCSVCell(v))
		}
		if err := writer.Write(sanitized); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// sanitizeCSVCell strips terminal-escape and text-spoofing characters from a CSV
// cell while preserving newlines, which encoding/csv safely quotes — so a value
// piped through a terminal cannot execute escape codes and cannot spoof text
// direction, yet legitimate multi-line cells survive the export. (PrintTable's
// sanitizeCell collapses newlines because they would break a fixed-width table;
// CSV keeps them because a quoted field is structurally fine.)
func sanitizeCSVCell(s string) string {
	return sanitizeRunes(s, true)
}

// sanitizeCSVField neutralizes CSV/spreadsheet formula injection (CWE-1236).
// A field whose first non-whitespace character is one a spreadsheet may
// interpret as a formula (= + - @) is prefixed with a single quote so it is
// treated as text rather than executed (see OWASP "CSV Injection"). Leading
// whitespace is skipped for classification because a spreadsheet may trim it
// and then execute the following character, but the original value is preserved.
//
// A leading + or - on a value that is actually a number (e.g. "-10.50", a
// credit amount) is legitimate data, not a formula, so it is left untouched to
// avoid corrupting numeric columns on export.
func sanitizeCSVField(s string) string {
	if s == "" {
		return s
	}
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\r' || s[i] == '\n') {
		i++
	}
	if i == len(s) {
		return s // all whitespace: nothing a spreadsheet would execute
	}
	switch s[i] {
	case '=', '@':
		return "'" + s
	case '+', '-':
		if isNumeric(s[i:]) {
			return s
		}
		return "'" + s
	}
	return s
}

// isNumeric reports whether s is a plain decimal numeric literal, so a leading
// sign is data rather than the start of a formula. It deliberately rejects the
// Go-only forms strconv.ParseFloat accepts (Inf, NaN, hex/0x..p exponents,
// underscores), since those still begin with a spreadsheet formula trigger.
func isNumeric(s string) bool {
	return decimalNumberRE.MatchString(s)
}

// decimalNumberRE matches an optional sign, integer/decimal digits, and an
// optional base-10 exponent — nothing else.
var decimalNumberRE = regexp.MustCompile(`^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d+)?$`)
