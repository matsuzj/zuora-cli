// Package output provides formatters for CLI output (table, JSON, Go template, CSV).
package output

import (
	"errors"
	"fmt"
	"io"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

// Column defines a table column.
type Column struct {
	Header string
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

// RenderJSON dispatches rawJSON through the machine-readable format flags in
// the canonical priority order JQ > JSON > Template. It returns handled=true
// when one of those paths produced the output; handled=false means no format
// flag was set and the caller should fall through to its table/detail/CSV
// path. This is the single entry point those branches live in — commands and
// Render/RenderDetail must not re-implement the dispatch (28 hand-rolled
// copies disagreed on the order and all silently ignored --csv; their
// replacement lands with P3-3).
func RenderJSON(ios *iostreams.IOStreams, rawJSON []byte, opts FormatOptions) (bool, error) {
	if opts.JQ != "" {
		return true, PrintJSON(ios, rawJSON, opts.JQ)
	}
	if opts.JSON {
		return true, PrintJSON(ios, rawJSON, "")
	}
	if opts.Template != "" {
		return true, PrintTemplate(ios, rawJSON, opts.Template)
	}
	return false, nil
}

// ErrCSVUnsupportedJSONOnly rejects --csv on commands whose output is JSON
// only. Silently ignoring the flag was a bug (the user asked for CSV and got
// JSON); an explicit error is diagnostic. Mirrors zr api's raw-output
// message. Decided in docs/refactoring-plan.md P3-3; behavior change to be
// noted in the next release tag.
var ErrCSVUnsupportedJSONOnly = errors.New("--csv is not supported for JSON-only output; use --jq or --template to shape the response")

// RenderJSONOnly renders a JSON-only command's response: --jq/--json/
// --template dispatch through RenderJSON FIRST (the documented precedence —
// README: the JSON-family flags win over --csv, cf. the PR #54 regression),
// then a bare --csv is rejected, and the default is pretty-printed JSON. JSON-only read commands should end with exactly this
// call; write commands with a trailing stderr message keep their guard form
// and reject opts.CSV before it.
func RenderJSONOnly(ios *iostreams.IOStreams, rawJSON []byte, opts FormatOptions) error {
	if handled, err := RenderJSON(ios, rawJSON, opts); handled || err != nil {
		return err
	}
	if opts.CSV {
		return ErrCSVUnsupportedJSONOnly
	}
	return PrintJSON(ios, rawJSON, "")
}

// RenderSuccess renders the result of an operation whose response carries no
// usable body (HTTP 204, an empty 200, or a non-JSON 200 — treated as success
// per the delete-policy decision in docs/refactoring-plan.md). Machine-readable
// flags receive a synthesized {"success": true}; otherwise humanMsg (a complete
// sentence with trailing newline) goes to stderr, keeping stdout clean.
func RenderSuccess(ios *iostreams.IOStreams, opts FormatOptions, humanMsg string) error {
	if handled, err := RenderJSON(ios, []byte(`{"success": true}`), opts); handled || err != nil {
		return err
	}
	fmt.Fprint(ios.ErrOut, humanMsg)
	return nil
}

// Render outputs data in the appropriate format for table commands.
func Render(ios *iostreams.IOStreams, rawJSON []byte, opts FormatOptions, rows [][]string, cols []Column) error {
	if handled, err := RenderJSON(ios, rawJSON, opts); handled || err != nil {
		return err
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
	if handled, err := RenderJSON(ios, rawJSON, opts); handled || err != nil {
		return err
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
