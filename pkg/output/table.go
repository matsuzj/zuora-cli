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
// newlines/tabs/carriage returns and the Unicode line/paragraph separators
// (U+2028/U+2029) collapse to spaces, and other control characters (ANSI escape
// sequences) plus Unicode format characters (U+202A-U+202E and the other BiDi /
// zero-width controls in category Cf) are dropped — so a multiline or hostile
// field cannot break the table layout, write escape codes to the terminal, or
// spoof text direction.
func sanitizeCell(s string) string {
	return sanitizeRunes(s, false)
}

// sanitizeRunes is the shared body of sanitizeCell and sanitizeCSVCell: tabs,
// CRs and the Unicode line/paragraph separators collapse to spaces; other
// control characters (ANSI escapes) and Unicode format characters (BiDi /
// zero-width, category Cf) are dropped. preserveNewlines keeps '\n' for CSV
// (encoding/csv quotes the field) and collapses it for tables (a newline
// would break the fixed-width layout).
func sanitizeRunes(s string, preserveNewlines bool) string {
	if s == "" {
		return s
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n':
			if preserveNewlines {
				return r
			}
			return ' '
		case '\t', '\r', '\u2028', '\u2029':
			return ' '
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}
