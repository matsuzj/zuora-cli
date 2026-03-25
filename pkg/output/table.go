package output

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

// PrintTable writes a formatted table to w.
func PrintTable(w io.Writer, rows [][]string, columns []Column) error {
	table := tablewriter.NewTable(w)

	headers := make([]any, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}
	table.Header(headers...)

	for _, row := range rows {
		vals := make([]any, len(row))
		for i, v := range row {
			vals[i] = v
		}
		if err := table.Append(vals...); err != nil {
			return err
		}
	}

	return table.Render()
}
