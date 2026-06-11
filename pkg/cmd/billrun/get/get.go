// Package get implements the "zr billrun get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the billrun get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <bill-run-id>",
		Short: "Get bill run details",
		Long: `Get detailed information about a Zuora bill run.

Examples:
  zr billrun get BR-00000001
  zr billrun get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, billRunID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/bill-runs/%s", url.PathEscape(billRunID)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetDecimal(raw, "id")},
				{Key: "Bill Run Number", Value: cmdutil.GetDecimal(raw, "billRunNumber")},
				{Key: "Name", Value: cmdutil.GetDecimal(raw, "name")},
				{Key: "Status", Value: cmdutil.GetDecimal(raw, "status")},
				{Key: "Invoice Date", Value: cmdutil.GetDecimal(raw, "invoiceDate")},
				{Key: "Target Date", Value: cmdutil.GetDecimal(raw, "targetDate")},
				{Key: "Auto Post", Value: cmdutil.GetDecimal(raw, "autoPost")},
				{Key: "Auto Email", Value: cmdutil.GetDecimal(raw, "autoEmail")},
				{Key: "Bill Cycle Day", Value: cmdutil.GetDecimal(raw, "billCycleDay")},
				{Key: "Scheduled Execution Time", Value: cmdutil.GetDecimal(raw, "scheduledExecutionTime")},
				{Key: "Created Date", Value: cmdutil.GetDecimal(raw, "createdDate")},
			}
		},
	})
}
