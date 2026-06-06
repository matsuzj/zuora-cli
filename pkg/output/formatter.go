// Package output provides formatters for CLI output (table, JSON, Go template, CSV).
package output

import (
	"fmt"
	"io"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

// Column defines a table column.
type Column struct {
	Header string
	Field  string
}

// DetailField defines a key-value pair for detail output.
type DetailField struct {
	Key   string
	Value string
}

// FormatOptions holds output format flags.
type FormatOptions struct {
	JSON     bool
	JQ       string
	Template string
	CSV      bool
}

// FromCmd reads FormatOptions from cobra command flags.
func FromCmd(cmd *cobra.Command) FormatOptions {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	jq, _ := cmd.Flags().GetString("jq")
	tmpl, _ := cmd.Flags().GetString("template")
	csvFlag, _ := cmd.Flags().GetBool("csv")
	return FormatOptions{JSON: jsonFlag, JQ: jq, Template: tmpl, CSV: csvFlag}
}

// Render outputs data in the appropriate format for table commands.
func Render(ios *iostreams.IOStreams, rawJSON []byte, opts FormatOptions, rows [][]string, cols []Column) error {
	if opts.JQ != "" {
		return PrintJSON(ios, rawJSON, opts.JQ)
	}
	if opts.JSON {
		return PrintJSON(ios, rawJSON, "")
	}
	if opts.Template != "" {
		return PrintTemplate(ios, rawJSON, opts.Template)
	}
	if opts.CSV {
		return PrintCSV(ios.Out, rows, cols)
	}
	w := io.Writer(ios.Out)
	if pager, err := StartPager(ios); err != nil {
		// Pager failed to start — fall back to direct output but tell the user why.
		fmt.Fprintf(ios.ErrOut, "warning: could not start pager: %v\n", err)
	} else {
		defer pager.Close()
		w = pager
	}
	return PrintTable(w, rows, cols)
}

// RenderDetail outputs data in the appropriate format for detail commands.
func RenderDetail(ios *iostreams.IOStreams, rawJSON []byte, opts FormatOptions, fields []DetailField) error {
	if opts.JQ != "" {
		return PrintJSON(ios, rawJSON, opts.JQ)
	}
	if opts.JSON {
		return PrintJSON(ios, rawJSON, "")
	}
	if opts.Template != "" {
		return PrintTemplate(ios, rawJSON, opts.Template)
	}
	if opts.CSV {
		rows := make([][]string, len(fields))
		for i, f := range fields {
			rows[i] = []string{f.Key, f.Value}
		}
		return PrintCSV(ios.Out, rows, []Column{{Header: "Field"}, {Header: "Value"}})
	}
	return PrintDetail(ios, fields)
}
