package main

import (
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRender pins that the generated block is compiled from the live gate
// data: entries the gate matches against must appear verbatim. Bite: removing
// an entry from internal/api's allowlist (or from the accessor) changes this
// output, which the make lint drift gate then catches against docs/read-only.md.
func TestRender(t *testing.T) {
	d := api.ReadOnlyDocForDocs()
	require.NotEmpty(t, d.POSTAllowList)
	require.NotEmpty(t, d.POSTPatterns)

	got := render(d)

	for _, p := range d.POSTAllowList {
		assert.Contains(t, got, "`"+p+"`")
	}
	for _, p := range d.POSTPatterns {
		assert.Contains(t, got, "`"+p+"`")
	}
	assert.Contains(t, got, "PUT / DELETE / PATCH** — always blocked")
	assert.Contains(t, got, "--read-only-allow-data-query")
	assert.Contains(t, got, "POST "+d.DataQuerySubmitPath)
	assert.True(t, strings.HasSuffix(got, "\n"), "block must end with a newline for the marker sandwich")
}
