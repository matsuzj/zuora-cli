// Package get implements the "zr ramp get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the ramp get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <ramp-number>",
		Short: "Get ramp details",
		Long: `Get detailed information about a Zuora ramp.

Examples:
  zr ramp get R-00000001
  zr ramp get R-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, rampNumber string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/ramps/%s", url.PathEscape(rampNumber)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// GET /v1/ramps/{id} wraps the ramp under a top-level "ramp" object, and its
	// number field is "number" (not "rampNumber"). Fall back to the top level so
	// an unwrapped response still renders.
	ramp, _ := raw["ramp"].(map[string]interface{})
	if ramp == nil {
		ramp = raw
	}
	fields := []output.DetailField{
		{Key: "Ramp Number", Value: cmdutil.GetString(ramp, "number")},
		{Key: "Name", Value: cmdutil.GetString(ramp, "name")},
		{Key: "Description", Value: cmdutil.GetString(ramp, "description")},
		{Key: "Subscription Number", Value: cmdutil.GetString(ramp, "subscriptionNumber")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
