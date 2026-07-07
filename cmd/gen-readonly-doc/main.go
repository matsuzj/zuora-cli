// Command gen-readonly-doc emits the read-only allowlist section of
// docs/read-only.md from the gate's ground truth in internal/api — the same
// generate-plus-drift-gate shape as scripts/gen-destructive-list.sh. The
// hand-copied allowlist in the original plan document went 15 weeks stale and
// misdocumented a SAFETY invariant (#526); this is compiled from the exact
// slices the gate matches against, so it cannot drift silently.
package main

import (
	"fmt"
	"strings"

	"github.com/matsuzj/zuora-cli/internal/api"
)

func render(d api.ReadOnlyDocData) string {
	var b strings.Builder
	b.WriteString("Under `--read-only` / `ZR_READ_ONLY` the API client allows:\n\n")
	b.WriteString("- **GET / HEAD / OPTIONS** — always allowed.\n")
	b.WriteString("- **POST** — allowed only for these read-only endpoints (exact match, after path normalization):\n")
	for _, p := range d.POSTAllowList {
		fmt.Fprintf(&b, "  - `%s`\n", p)
	}
	b.WriteString("- **POST** — allowed for these dynamic-path patterns (regex match):\n")
	for _, p := range d.POSTPatterns {
		fmt.Fprintf(&b, "  - `%s`\n", p)
	}
	b.WriteString("- **PUT / DELETE / PATCH** — always blocked.\n")
	fmt.Fprintf(&b, "- **Data Query opt-in** (`--read-only-allow-data-query` / `ZR_READ_ONLY_ALLOW_DATA_QUERY`, default OFF — fails restrictive): additionally allows `POST %s` (submit) and `DELETE %s` (cancel). Nothing else widens.\n",
		d.DataQuerySubmitPath, d.DataQueryCancelPattern)
	return b.String()
}

func main() {
	fmt.Print(render(api.ReadOnlyDocForDocs()))
}
