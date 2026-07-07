package api

// ReadOnlyDocData is the read-only gate's ground truth, exposed for
// cmd/gen-readonly-doc (which generates the docs/read-only.md allowlist block,
// drift-gated by make lint). Any NEW allowlist variable added to the gate
// (e.g. a future readOnlyPUTPatterns) MUST be added here and to the generator
// too — the generator prints exactly what this returns, so a forgotten field
// silently regenerates a stale doc (#526).
type ReadOnlyDocData struct {
	// POSTAllowList are exact-match POST paths allowed in read-only mode.
	POSTAllowList []string
	// POSTPatterns are regex sources for POST paths with dynamic segments.
	POSTPatterns []string
	// DataQuerySubmitPath / DataQueryCancelPattern are permitted only under
	// the --read-only-allow-data-query opt-in (default: blocked).
	DataQuerySubmitPath    string
	DataQueryCancelPattern string
}

// ReadOnlyDocForDocs returns copies of the live allowlist data.
func ReadOnlyDocForDocs() ReadOnlyDocData {
	patterns := make([]string, 0, len(readOnlyPOSTPatterns))
	for _, re := range readOnlyPOSTPatterns {
		patterns = append(patterns, re.String())
	}
	return ReadOnlyDocData{
		POSTAllowList:          append([]string(nil), readOnlyPOSTAllowList...),
		POSTPatterns:           patterns,
		DataQuerySubmitPath:    "query/jobs",
		DataQueryCancelPattern: dataQueryJobPattern.String(),
	}
}
