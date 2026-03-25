package output

import (
	"encoding/csv"
	"io"
)

// PrintCSV writes data as CSV.
func PrintCSV(w io.Writer, rows [][]string, columns []Column) error {
	writer := csv.NewWriter(w)

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Header
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
