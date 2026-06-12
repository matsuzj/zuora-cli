// Package cancel implements the "zr order cancel" command.
package cancel

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdCancel creates the order cancel command.
func NewCmdCancel(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "cancel <order-number>",
		Short: "Cancel an order",
		Long: `Cancel a Zuora order.

This action is irreversible. Use --confirm to proceed.`,
		Example: `  zr order cancel O-00000001 --confirm`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runCancel(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "cancellation")
	return cmd
}

func runCancel(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/orders/%s/cancel", url.PathEscape(orderNumber)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Order Number", Value: cmdutil.GetString(raw, "orderNumber")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Order %s cancelled.\n", orderNumber)
		},
	})
}
