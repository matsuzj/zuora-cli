// Package updatecustomfields implements the "zr order update-custom-fields" command.
package updatecustomfields

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdUpdateCustomFields creates the order update-custom-fields command.
func NewCmdUpdateCustomFields(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "update-custom-fields <order-number>",
		Short: "Update custom fields on an order",
		Long:  `Update custom fields on a Zuora order.`,
		Example: `  zr order update-custom-fields O-00000001 --body @fields.json
  zr order update-custom-fields O-00000001 --body '{"cf_MyField__c":"value"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdateCustomFields(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runUpdateCustomFields(cmd *cobra.Command, f *factory.Factory, orderNumber, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/orders/%s/customFields", url.PathEscape(orderNumber)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Custom fields updated for order %s.\n", orderNumber)
		},
	})
}
