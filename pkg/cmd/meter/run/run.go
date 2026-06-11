// Package run implements the "zr meter run" command.
package run

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdRun creates the meter run command.
func NewCmdRun(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <meterId> <version>",
		Short: "Run a usage meter",
		Long: `Run a usage meter by meter ID and version.

Examples:
  zr meter run 402880e44c...  1
  zr meter run 402880e44c...  1 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMeter(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runMeter(cmd *cobra.Command, f *factory.Factory, meterID, version string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   fmt.Sprintf("/meters/run/%s/%s", url.PathEscape(meterID), url.PathEscape(version)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
				{Key: "Message", Value: cmdutil.GetString(raw, "message")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return "Meter run started.\n"
		},
	})
}
