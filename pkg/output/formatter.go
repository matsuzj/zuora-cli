// Package output provides formatters for CLI output (table, JSON, Go template, CSV).
package output

import (
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
}

// FromCmd reads FormatOptions from cobra command flags.
func FromCmd(cmd *cobra.Command) FormatOptions {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	jq, _ := cmd.Flags().GetString("jq")
	tmpl, _ := cmd.Flags().GetString("template")
	return FormatOptions{JSON: jsonFlag, JQ: jq, Template: tmpl}
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
	w := io.Writer(ios.Out)
	if pager, err := StartPager(ios); err == nil {
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
	return PrintDetail(ios, fields)
}
