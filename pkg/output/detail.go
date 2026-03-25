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
		fmt.Fprintf(w, "%s:\t%s\n", f.Key, f.Value)
	}
	return w.Flush()
}
