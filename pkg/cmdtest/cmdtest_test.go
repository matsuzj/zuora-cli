package cmdtest

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newProbeCmd is a minimal but realistic command: GET /v1/probe/<id>, render a
// detail view — enough to exercise the server wiring, persistent flags, and
// both streams through the harness.
func newProbeCmd(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:  "probe <id>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HttpClient()
			if err != nil {
				return err
			}
			// NOTE: explicit WithCheckSuccess matches main today; PR #71 flips
			// the default and deletes the option — drop it on update-branch.
			resp, err := client.Get("/v1/probe/"+args[0], api.WithCheckSuccess())
			if err != nil {
				return err
			}
			var raw map[string]interface{}
			if err := json.Unmarshal(resp.Body, &raw); err != nil {
				return fmt.Errorf("parsing response: %w", err)
			}
			fields := []output.DetailField{{Key: "Name", Value: cmdutil.GetString(raw, "name")}}
			return output.RenderDetail(f.IOStreams, resp.Body, output.FromCmd(cmd), fields)
		},
	}
}

func TestRun_RootLevelCommand(t *testing.T) {
	stdout, stderr, err := Run(t, "", newProbeCmd,
		OK(t, "GET", "/v1/probe/P-1", map[string]interface{}{"success": true, "name": "Widget"}),
		"probe", "P-1")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Widget")
	assert.Empty(t, stderr)
}

func TestRun_UnderParentGroup(t *testing.T) {
	stdout, _, err := Run(t, "thing", newProbeCmd,
		OK(t, "GET", "/v1/probe/P-2", map[string]interface{}{"success": true, "name": "Gadget"}),
		"thing", "probe", "P-2")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Gadget")
}

func TestRun_PersistentFlagsWired(t *testing.T) {
	stdout, _, err := Run(t, "", newProbeCmd,
		OK(t, "", "", map[string]interface{}{"success": true, "name": "Widget"}),
		"probe", "P-1", "--json")
	require.NoError(t, err)
	assert.Contains(t, stdout, `"name": "Widget"`, "--json must reach output.FromCmd through the stub root")
}

func TestRun_ReasonsEnvelopeErrors(t *testing.T) {
	// The success-flag check is on by default, so the Reasons envelope must
	// surface as a non-zero error carrying the message.
	_, _, err := Run(t, "", newProbeCmd, Reasons(t, 53100020, "Missing field"), "probe", "P-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing field")
}

func TestRun_NilHandlerForValidationTests(t *testing.T) {
	_, _, err := Run(t, "", newProbeCmd, nil, "probe")
	require.Error(t, err, "arg validation fails before any HTTP call")
	assert.Contains(t, err.Error(), "accepts 1 arg(s)")
}

func TestStatus_ExplicitCode(t *testing.T) {
	_, _, err := Run(t, "", newProbeCmd,
		Status(t, "GET", "/v1/probe/P-9", 404, map[string]interface{}{"message": "not found"}),
		"probe", "P-9")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}
