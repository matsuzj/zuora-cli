package output

import (
	"encoding/csv"
	"io"
	"strconv"
)

// PrintCSV writes data as CSV, neutralizing spreadsheet formula injection.
func PrintCSV(w io.Writer, rows [][]string, columns []Column) error {
	writer := csv.NewWriter(w)

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = sanitizeCSVField(col.Header)
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		sanitized := make([]string, len(row))
		for i, v := range row {
			sanitized[i] = sanitizeCSVField(v)
		}
		if err := writer.Write(sanitized); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

// sanitizeCSVField neutralizes CSV/spreadsheet formula injection (CWE-1236).
// A field whose first character is one a spreadsheet may interpret as a formula
// (= + - @, or a leading tab/CR) is prefixed with a single quote so it is
// treated as text rather than executed (see OWASP "CSV Injection").
//
// A leading + or - on a value that is actually a number (e.g. "-10.50", a
// credit amount) is legitimate data, not a formula, so it is left untouched to
// avoid corrupting numeric columns on export.
func sanitizeCSVField(s string) string {
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '@', '\t', '\r':
		return "'" + s
	case '+', '-':
		if isNumeric(s) {
			return s
		}
		return "'" + s
	}
	return s
}

// isNumeric reports whether s is a plain numeric literal (so a leading sign is
// data, not the start of a formula).
func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}
