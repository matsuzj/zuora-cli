package output

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmptyBody_AllModesConsistent covers that an empty / whitespace-only body
// (e.g. an HTTP 204) is a silent success in every output mode. Previously the
// pretty-print and raw paths exited 0 but --jq and --template failed with an
// EOF parse error, an inconsistency across output flags.
func TestEmptyBody_AllModesConsistent(t *testing.T) {
	bodies := map[string][]byte{
		"nil":        nil,
		"empty":      []byte(""),
		"whitespace": []byte("  \n\t "),
	}
	for name, data := range bodies {
		t.Run(name, func(t *testing.T) {
			// Pretty-print (no jq)
			ios, _, out, _ := iostreams.Test()
			require.NoError(t, PrintJSON(ios, data, ""))
			assert.Empty(t, out.String())

			// jq
			ios, _, out, _ = iostreams.Test()
			require.NoError(t, PrintJSON(ios, data, ".foo"), "jq on an empty body must succeed")
			assert.Empty(t, out.String())

			// template
			ios, _, out, _ = iostreams.Test()
			require.NoError(t, PrintTemplate(ios, data, "{{.foo}}"), "template on an empty body must succeed")
			assert.Empty(t, out.String())

			// raw
			ios, _, out, _ = iostreams.Test()
			require.NoError(t, PrintRawOrJSON(ios, data))
			assert.Empty(t, out.String())
		})
	}
}
