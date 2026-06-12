// Package listcmd provides a declarative runner for standard table list
// commands (P3-2). A command declares its cobra surface, query flags, request
// path, response items key, and columns as a Spec; the runner owns the shared
// machinery: flag registration, conditional query assembly, the GET request,
// envelope decoding, cell extraction, output.Render, and the canonical
// nextPage hint.
package listcmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// ColumnSpec maps one JSON key of each list item to a table column.
type ColumnSpec struct {
	Header string
	Key    string
	// Money renders float64 values with %.2f, preserving the current monetary
	// display ("100.00", not "100"). An absent or null key renders "0.00" —
	// the zero value the migrated commands' typed structs produced.
	Money bool
}

// Flag declares one cobra flag of a list command.
type Flag struct {
	// Name is the cobra flag name (e.g. "page-size").
	Name string
	// Query is the request query-parameter name (e.g. "pageSize"). Empty
	// means the flag is not sent as a query parameter (path-only flags such
	// as subscription list's --account, consumed by Spec.Path).
	Query string
	Usage string
	// Required marks the flag required at the cobra level.
	Required bool
	// Repeatable registers a StringArray flag; non-empty values are sent via
	// api.WithQuerySlice (e.g. account list's --filter → "filter[]").
	Repeatable bool
	// Int registers an int flag with IntDefault. Int flags are ALWAYS sent
	// (strconv.Itoa), matching account list's --page-size default 20.
	Int        bool
	IntDefault int
	// DeprecatedName optionally registers the flag's OLD spelling as a
	// hidden, deprecated alias writing into the same destination (string
	// flags only). Renamed flags keep working for one release; the alias is
	// removed in the next minor. When combined with Required, the value (not
	// cobra's per-flag Changed bit) is what run() enforces, so the alias
	// satisfies the requirement.
	DeprecatedName string
}

// NextPage declares how the canonical pagination hint carries the next-page
// value. The zero value disables reconstruction: the runner falls back to the
// generic "Use --json to see nextPage URL." message.
type NextPage struct {
	// Flag is the flag name that carries the next-page value in the
	// reconstructed command (e.g. "page" or "cursor"). It must name a plain
	// string Flag of the Spec (not Int or Repeatable).
	Flag string
	// FromURL, when non-empty, parses the response's nextPage as a URL and
	// takes this query parameter as the value (page-based APIs). When empty,
	// the raw nextPage string is used verbatim (cursor-based APIs).
	FromURL string
}

// Spec declares a standard table list command.
type Spec struct {
	Use   string
	Short string
	Long  string
	// Example feeds cobra's Example field (rendered under "Examples:" in
	// help). Example invocations belong here, not embedded in Long.
	Example string
	Aliases []string
	// Args validates positional arguments (defaults to cobra.NoArgs).
	Args cobra.PositionalArgs
	// Flags are registered in order; query parameters are assembled in the
	// same order.
	Flags []Flag
	// Path builds the request path from positional args and the current
	// string/int flag values (keyed by flag name; ints are formatted with
	// strconv.Itoa, repeatable flags are not included).
	Path func(args []string, flags map[string]string) string
	// ItemsKey is the response envelope key holding the item array.
	ItemsKey string
	Columns  []ColumnSpec
	NextPage NextPage
}

// New builds the cobra command for a Spec.
func New(f *factory.Factory, spec Spec) *cobra.Command {
	strVals := make(map[string]*string)
	intVals := make(map[string]*int)
	arrVals := make(map[string]*[]string)

	args := spec.Args
	if args == nil {
		args = cobra.NoArgs
	}

	cmd := &cobra.Command{
		Use:     spec.Use,
		Short:   spec.Short,
		Long:    spec.Long,
		Example: spec.Example,
		Aliases: spec.Aliases,
		Args:    args,
		RunE: func(cmd *cobra.Command, posArgs []string) error {
			return run(cmd, f, spec, posArgs, strVals, intVals, arrVals)
		},
	}

	for _, fl := range spec.Flags {
		switch {
		case fl.Int:
			v := new(int)
			cmd.Flags().IntVar(v, fl.Name, fl.IntDefault, fl.Usage)
			intVals[fl.Name] = v
		case fl.Repeatable:
			v := new([]string)
			cmd.Flags().StringArrayVar(v, fl.Name, nil, fl.Usage)
			arrVals[fl.Name] = v
		default:
			v := new(string)
			cmd.Flags().StringVar(v, fl.Name, "", fl.Usage)
			strVals[fl.Name] = v
			if fl.DeprecatedName != "" {
				// The old spelling writes into the SAME destination and
				// keeps working through the deprecation window (removed in
				// the next minor); pflag prints the notice on use and hides
				// the flag from help.
				cmd.Flags().StringVar(v, fl.DeprecatedName, "", fl.Usage)
				_ = cmd.Flags().MarkDeprecated(fl.DeprecatedName, "use --"+fl.Name+" instead")
			}
		}
		if fl.Required {
			if fl.DeprecatedName == "" {
				_ = cmd.MarkFlagRequired(fl.Name)
			}
			// With a deprecated alias, cobra's required check would reject
			// invocations that set only the alias (it inspects the canonical
			// flag's Changed bit, not the shared destination). run() enforces
			// the requirement on the VALUE instead, with cobra's wording.
		}
	}

	return cmd
}

func run(cmd *cobra.Command, f *factory.Factory, spec Spec, posArgs []string, strVals map[string]*string, intVals map[string]*int, arrVals map[string]*[]string) error {
	// Required flags that carry a deprecated alias are enforced on the value
	// (cobra's required check would not see the alias as satisfying the
	// canonical flag). The wording matches cobra's exactly.
	for _, fl := range spec.Flags {
		if fl.Required && fl.DeprecatedName != "" {
			if v := strVals[fl.Name]; v != nil && *v == "" {
				return fmt.Errorf("required flag(s) %q not set", fl.Name)
			}
		}
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	flagVals := make(map[string]string, len(strVals)+len(intVals))
	for name, v := range strVals {
		flagVals[name] = *v
	}
	for name, v := range intVals {
		flagVals[name] = strconv.Itoa(*v)
	}

	var reqOpts []api.RequestOption
	for _, fl := range spec.Flags {
		if fl.Query == "" {
			continue
		}
		switch {
		case fl.Int:
			reqOpts = append(reqOpts, api.WithQuery(fl.Query, strconv.Itoa(*intVals[fl.Name])))
		case fl.Repeatable:
			if vals := *arrVals[fl.Name]; len(vals) > 0 {
				reqOpts = append(reqOpts, api.WithQuerySlice(fl.Query, vals))
			}
		default:
			if v := *strVals[fl.Name]; v != "" {
				reqOpts = append(reqOpts, api.WithQuery(fl.Query, v))
			}
		}
	}

	resp, err := client.Get(spec.Path(posArgs, flagVals), reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var envelope map[string]interface{}
	if err := json.Unmarshal(resp.Body, &envelope); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	rawItems, ok := envelope[spec.ItemsKey].([]interface{})
	if !ok && envelope[spec.ItemsKey] != nil {
		// The hand-written commands' typed structs error on a non-array items
		// key; a silent empty table would hide a response-shape change.
		return fmt.Errorf("parsing response: %q is not an array", spec.ItemsKey)
	}
	rows := make([][]string, len(rawItems))
	for i, ri := range rawItems {
		item, _ := ri.(map[string]interface{})
		row := make([]string, len(spec.Columns))
		for j, col := range spec.Columns {
			row[j] = cell(item, col)
		}
		rows[i] = row
	}

	cols := make([]output.Column, len(spec.Columns))
	for i, c := range spec.Columns {
		cols[i] = output.Column{Header: c.Header}
	}

	if err := output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols); err != nil {
		return err
	}

	nextPage, _ := envelope["nextPage"].(string)
	if nextPage != "" && !fmtOpts.JSON && fmtOpts.JQ == "" && fmtOpts.Template == "" {
		printHint(cmd, spec, posArgs, strVals, intVals, arrVals, nextPage, f)
	}

	return nil
}

// cell extracts one display cell from an item. Money columns preserve the
// migrated commands' typed-struct semantics: float64 → %.2f, absent/null →
// "0.00" (the struct zero value), anything else → %v.
func cell(item map[string]interface{}, col ColumnSpec) string {
	if item == nil {
		if col.Money {
			return "0.00"
		}
		return ""
	}
	if col.Money {
		v, ok := item[col.Key]
		if !ok || v == nil {
			return "0.00"
		}
		if f, ok := v.(float64); ok {
			return fmt.Sprintf("%.2f", f)
		}
		return fmt.Sprintf("%v", v)
	}
	return cmdutil.GetString(item, col.Key)
}

// printHint emits the canonical pagination hint: a copy-pasteable next
// command rebuilt from cobra's command path, the positional args, every
// non-default flag value, and the next-page value carried by the declared
// flag. When the Spec declares no NextPage (or the value cannot be
// extracted), it falls back to the generic message.
func printHint(cmd *cobra.Command, spec Spec, posArgs []string, strVals map[string]*string, intVals map[string]*int, arrVals map[string]*[]string, nextPage string, f *factory.Factory) {
	value := nextPage
	if spec.NextPage.FromURL != "" {
		value = ""
		if u, err := url.Parse(nextPage); err == nil {
			value = u.Query().Get(spec.NextPage.FromURL)
		}
	}
	if spec.NextPage.Flag == "" || value == "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "\nMore results available. Use --json to see nextPage URL.\n")
		return
	}

	parts := []string{cmd.CommandPath()}
	for _, a := range posArgs {
		parts = append(parts, quoteIfNeeded(a))
	}
	for _, fl := range spec.Flags {
		if fl.Name == spec.NextPage.Flag {
			continue // re-emitted below with the new value
		}
		switch {
		case fl.Int:
			if v := *intVals[fl.Name]; v != fl.IntDefault {
				parts = append(parts, "--"+fl.Name, strconv.Itoa(v))
			}
		case fl.Repeatable:
			for _, v := range *arrVals[fl.Name] {
				parts = append(parts, "--"+fl.Name, quoteIfNeeded(v))
			}
		default:
			if v := *strVals[fl.Name]; v != "" {
				parts = append(parts, "--"+fl.Name, quoteIfNeeded(v))
			}
		}
	}
	parts = append(parts, "--"+spec.NextPage.Flag, quoteIfNeeded(value))

	fmt.Fprintf(f.IOStreams.ErrOut, "\nMore results available. Next page:\n  %s\n", strings.Join(parts, " "))
}

// quoteIfNeeded wraps a value in shell single quotes when it contains
// characters that would not survive a shell unquoted; plain tokens stay bare
// so the common hints read naturally. Single quotes (not Go/double quoting —
// review finding) so that $VAR, backticks, and backslashes paste verbatim; an
// embedded single quote is escaped by closing the quotes, emitting a
// backslash-quote, and reopening (quote backslash-quote quote). Control and
// Unicode format characters are mapped to spaces first: the hint goes to the
// terminal unfiltered, so the table renderer's escape-sequence sanitization
// (CWE-150) must not be bypassable via nextPage/cursor values.
func quoteIfNeeded(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.In(r, unicode.Cf) {
			return ' '
		}
		return r
	}, s)
	if s == "" {
		return "''"
	}
	for _, r := range s {
		isSafe := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '-' || r == '_' || r == '.' || r == ':' || r == '/' || r == '=' || r == ',' || r == '%' || r == '+'
		if !isSafe {
			return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
		}
	}
	return s
}
