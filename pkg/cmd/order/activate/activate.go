// Package activate implements the "zr order activate" command.
package activate

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdActivate creates the order activate command.
func NewCmdActivate(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "activate <order-number>",
		Short:   "Activate an order",
		Long:    `Activate a Zuora order.`,
		Example: `  zr order activate O-00000001`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runActivate(cmd, f, args[0])
		},
	}
	return cmd
}

func runActivate(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/orders/%s/activate", url.PathEscape(orderNumber)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Order Number", Value: cmdutil.GetString(raw, "orderNumber")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Order %s activated.\n", orderNumber)
		},
	})
}
