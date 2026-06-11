// Package revert implements the "zr order revert" command.
package revert

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdRevert creates the order revert command.
func NewCmdRevert(f *factory.Factory) *cobra.Command {
	var body string
	var confirm bool

	cmd := &cobra.Command{
		Use:   "revert <order-number>",
		Short: "Revert an order",
		Long: `Revert a Zuora order.

Requires --body with a JSON payload containing the orderDate.
This action is irreversible. Use --confirm to proceed.

Examples:
  zr order revert O-00000001 --body '{"orderDate":"2026-01-01"}' --confirm
  zr order revert O-00000001 --body @revert.json --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runRevert(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	cmdutil.AddConfirmFlag(cmd, &confirm, "revert")
	return cmd
}

func runRevert(cmd *cobra.Command, f *factory.Factory, orderNumber, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   fmt.Sprintf("/v1/orders/%s/revert", url.PathEscape(orderNumber)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Order Number", Value: cmdutil.GetString(raw, "orderNumber")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Order %s reverted.\n", orderNumber)
		},
	})
}
