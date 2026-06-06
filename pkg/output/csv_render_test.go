package output

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRender_CSV(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cols := []Column{{Header: "ID"}, {Header: "Name"}}
	rows := [][]string{{"a1", "Acme"}, {"a2", "Beta"}}

	require.NoError(t, Render(ios, nil, FormatOptions{CSV: true}, rows, cols))

	output := out.String()
	assert.Contains(t, output, "ID,Name")
	assert.Contains(t, output, "a1,Acme")
	assert.Contains(t, output, "a2,Beta")
}

func TestRenderDetail_CSV(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	fields := []DetailField{{Key: "ID", Value: "p-1"}, {Key: "Amount", Value: "1000000"}}

	require.NoError(t, RenderDetail(ios, nil, FormatOptions{CSV: true}, fields))

	output := out.String()
	assert.Contains(t, output, "Field,Value")
	assert.Contains(t, output, "ID,p-1")
	assert.Contains(t, output, "Amount,1000000")
}

func TestFromCmd_CSV(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().String("jq", "", "")
	cmd.Flags().String("template", "", "")
	cmd.Flags().Bool("csv", false, "")
	require.NoError(t, cmd.Flags().Set("csv", "true"))

	opts := FromCmd(cmd)
	assert.True(t, opts.CSV)
	assert.False(t, opts.JSON)
}

// --json must win over --csv: CSV is the lowest-priority of the table formats.
func TestRender_JSONBeatsCSV(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	cols := []Column{{Header: "ID"}}
	rows := [][]string{{"a1"}}

	require.NoError(t, Render(ios, []byte(`[{"id":"a1"}]`), FormatOptions{JSON: true, CSV: true}, rows, cols))

	output := out.String()
	assert.Contains(t, output, `"id"`)
	assert.NotContains(t, output, "ID,")
}
