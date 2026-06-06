package output

import (
	"fmt"
	"text/tabwriter"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
)

// PrintDetail writes key-value pairs in aligned format.
func PrintDetail(ios *iostreams.IOStreams, fields []DetailField) error {
	w := tabwriter.NewWriter(ios.Out, 0, 0, 2, ' ', 0)
	for _, f := range fields {
		// Sanitize like the table path: a hostile/compromised API field value
		// must not write ANSI/control sequences to the user's terminal.
		fmt.Fprintf(w, "%s:\t%s\n", sanitizeCell(f.Key), sanitizeCell(f.Value))
	}
	return w.Flush()
}
