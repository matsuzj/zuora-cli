package cmdutil

import (
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func deleteFields(raw map[string]interface{}) []output.DetailField {
	return []output.DetailField{{Key: "ID", Value: GetString(raw, "id")}}
}

// The three response shapes delete endpoints produce, judged uniformly.
func TestRenderDeleteResult_204IsSuccess(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	resp := &api.Response{StatusCode: 204, Body: nil}

	err := RenderDeleteResult(ios, resp, output.FormatOptions{}, "Thing T-1 deleted.\n", deleteFields)
	require.NoError(t, err)
	assert.Empty(t, out.String(), "human path keeps stdout clean")
	assert.Contains(t, errOut.String(), "Thing T-1 deleted.")
}

func TestRenderDeleteResult_Empty200IsSuccess(t *testing.T) {
	ios, _, _, errOut := iostreams.Test()
	resp := &api.Response{StatusCode: 200, Body: []byte("  ")}

	err := RenderDeleteResult(ios, resp, output.FormatOptions{}, "Thing T-1 deleted.\n", deleteFields)
	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "deleted")
}

func TestRenderDeleteResult_NonJSON200IsSuccess(t *testing.T) {
	ios, _, _, errOut := iostreams.Test()
	resp := &api.Response{StatusCode: 200, Body: []byte("OK")}

	err := RenderDeleteResult(ios, resp, output.FormatOptions{}, "Thing T-1 deleted.\n", deleteFields)
	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "deleted")
}

func TestRenderDeleteResult_JSONBodyRendersDetail(t *testing.T) {
	ios, _, out, _ := iostreams.Test()
	resp := &api.Response{StatusCode: 200, Body: []byte(`{"success":true,"id":"D-9"}`)}

	err := RenderDeleteResult(ios, resp, output.FormatOptions{}, "ignored\n", deleteFields)
	require.NoError(t, err)
	assert.Contains(t, out.String(), "D-9")
}

func TestRenderDeleteResult_SuccessSynthesizedForJSONFlag(t *testing.T) {
	ios, _, out, errOut := iostreams.Test()
	resp := &api.Response{StatusCode: 204, Body: nil}

	err := RenderDeleteResult(ios, resp, output.FormatOptions{JSON: true}, "Thing T-1 deleted.\n", deleteFields)
	require.NoError(t, err)
	assert.Contains(t, out.String(), `"success": true`)
	assert.Empty(t, errOut.String(), "machine output suppresses the human message")
}
