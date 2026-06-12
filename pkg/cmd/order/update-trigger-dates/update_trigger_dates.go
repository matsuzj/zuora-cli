// Package updatetriggerdates implements the "zr order update-trigger-dates" command.
package updatetriggerdates

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdUpdateTriggerDates creates the order update-trigger-dates command.
func NewCmdUpdateTriggerDates(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:     "update-trigger-dates <order-number>",
		Short:   "Update trigger dates on an order",
		Long:    `Update trigger dates on a Zuora order.`,
		Example: `  zr order update-trigger-dates O-00000001 --body @dates.json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateTriggerDates(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runUpdateTriggerDates(cmd *cobra.Command, f *factory.Factory, orderNumber, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/orders/%s/triggerDates", url.PathEscape(orderNumber)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Trigger dates updated for order %s.\n", orderNumber)
		},
	})
}
