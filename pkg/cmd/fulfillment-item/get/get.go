// Package get implements the "zr fulfillment-item get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the fulfillment-item get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <fulfillment-item-id>",
		Short: "Get fulfillment item details",
		Long:  `Get detailed information about a Zuora fulfillment item.`,
		Example: `  zr fulfillment-item get 2c92c0f8...
  zr fulfillment-item get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, itemID string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/fulfillment-items/%s", url.PathEscape(itemID)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Fulfillment Key", Value: cmdutil.GetString(raw, "fulfillmentKey")},
				{Key: "Quantity", Value: cmdutil.GetString(raw, "quantity")},
				{Key: "Description", Value: cmdutil.GetString(raw, "description")},
			}
		},
	})
}
