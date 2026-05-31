package output

import (
	"io"
	"strings"
	"unicode"

	"github.com/olekukonko/tablewriter"
)

// PrintTable writes a formatted table to w.
func PrintTable(w io.Writer, rows [][]string, columns []Column) error {
	table := tablewriter.NewTable(w)

	headers := make([]any, len(columns))
	for i, col := range columns {
		headers[i] = sanitizeCell(col.Header)
	}
	table.Header(headers...)

	for _, row := range rows {
		vals := make([]any, len(row))
		for i, v := range row {
			vals[i] = sanitizeCell(v)
		}
		if err := table.Append(vals...); err != nil {
			return err
		}
	}

	return table.Render()
}

// sanitizeCell makes an arbitrary API string safe to render in a table cell:
// newlines/tabs/carriage returns collapse to spaces and other control
// characters (including ANSI escape sequences) are dropped, so a multiline or
// hostile field cannot break the table layout or write escape codes to the
// terminal.
func sanitizeCell(s string) string {
	if s == "" {
		return s
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\t', '\r':
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}
