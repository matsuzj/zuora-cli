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
		Use:   "status <meter-id> <version>",
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
			// Real shape per the official API reference (doc-verified 2026-07-05,
			// #486; this sandbox cannot probe mediation endpoints): the envelope is
			// {success, data:{runStatus, runStatusDescription}} — runStatus is an
			// integer enum (1=NEVER_RUN … 13=CONSUME_COMPLETED) and the previous
			// flat keys (meterId/version/status/runType/startTime/endTime) do not
			// exist in the response.
			// LIVE-UNVERIFIED(meter runStatus envelope {success,data:{runStatus int enum,...}}; since 2026-07-05; trigger: tenant with mediation/metering provisioned)
			data, _ := raw["data"].(map[string]interface{})
			return []output.DetailField{
				{Key: "Run Status", Value: cmdutil.GetDecimal(data, "runStatus")},
				{Key: "Description", Value: cmdutil.GetString(data, "runStatusDescription")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
	})
}
