// Package debug implements the "zr meter debug" command.
package debug

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdDebug creates the meter debug command.
func NewCmdDebug(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug <meterId> <version>",
		Short: "Debug a usage meter",
		Long: `Debug a usage meter by meter ID and version.

Examples:
  zr meter debug 402880e44c... 1
  zr meter debug 402880e44c... 1 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDebug(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runDebug(cmd *cobra.Command, f *factory.Factory, meterID, version string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   fmt.Sprintf("/meters/debug/%s/%s", url.PathEscape(meterID), url.PathEscape(version)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
				{Key: "Message", Value: cmdutil.GetString(raw, "message")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return "Meter debug started.\n"
		},
	})
}
