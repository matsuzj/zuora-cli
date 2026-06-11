// Package get implements the "zr ramp get" command.
package get

import (
	"fmt"
	"net/url"

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
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/ramps/%s", url.PathEscape(rampNumber)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// GET /v1/ramps/{id} wraps the ramp under a top-level "ramp" object, and its
			// number field is "number" (not "rampNumber"). Fall back to the top level so
			// an unwrapped response still renders.
			ramp, _ := raw["ramp"].(map[string]interface{})
			if ramp == nil {
				ramp = raw
			}
			return []output.DetailField{
				{Key: "Ramp Number", Value: cmdutil.GetString(ramp, "number")},
				{Key: "Name", Value: cmdutil.GetString(ramp, "name")},
				{Key: "Description", Value: cmdutil.GetString(ramp, "description")},
				{Key: "Subscription Number", Value: cmdutil.GetString(ramp, "subscriptionNumber")},
			}
		},
	})
}
