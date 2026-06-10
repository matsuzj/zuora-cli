package output

import (
	"fmt"
	"text/template"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// PrintTemplate formats data using a Go template.
func PrintTemplate(ios *iostreams.IOStreams, data []byte, tmpl string) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// An empty body (e.g. HTTP 204) has nothing to render — succeed silently,
	// matching the JSON pretty-print and raw paths so the exit code is consistent
	// across output modes.
	if emptyBody(data) {
		return nil
	}

	v, err := decodeJSONPreservingNumbers(data)
	if err != nil {
		return fmt.Errorf("parsing JSON for template: %w", err)
	}

	if err := t.Execute(ios.Out, v); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}
	fmt.Fprintln(ios.Out)
	return nil
}
