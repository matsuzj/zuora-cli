// Package status implements the "zr meter status" command.
package status

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdStatus creates the meter status command.
func NewCmdStatus(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <meterId> <version>",
		Short: "Get meter run status",
		Long:  `Get the run status of a usage meter by meter ID and version.`,
		Example: `  zr meter status 402880e44c... 1
  zr meter status 402880e44c... 1 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runStatus(cmd *cobra.Command, f *factory.Factory, meterID, version string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/meters/%s/%s/runStatus", url.PathEscape(meterID), url.PathEscape(version)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Meter ID", Value: cmdutil.GetString(raw, "meterId")},
				{Key: "Version", Value: cmdutil.GetString(raw, "version")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Run Type", Value: cmdutil.GetString(raw, "runType")},
				{Key: "Start Time", Value: cmdutil.GetString(raw, "startTime")},
				{Key: "End Time", Value: cmdutil.GetString(raw, "endTime")},
			}
		},
	})
}
