// Package get implements the "zr usage get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the usage get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <usage-id>",
		Short: "Get a usage record",
		Long:  `Get a usage record by ID via the CRUD API.`,
		Example: `  zr usage get 2c92a0f96bd...
  zr usage get 2c92a0f96bd... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, id string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/object/usage/%s", url.PathEscape(id)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "Id")},
				{Key: "Account ID", Value: cmdutil.GetString(raw, "AccountId")},
				{Key: "Quantity", Value: cmdutil.GetDecimal(raw, "Quantity")},
				{Key: "Start Date", Value: cmdutil.GetString(raw, "StartDateTime")},
				{Key: "End Date", Value: cmdutil.GetString(raw, "EndDateTime")},
				{Key: "UOM", Value: cmdutil.GetString(raw, "UOM")},
				{Key: "Status", Value: cmdutil.GetString(raw, "Status")},
				{Key: "Subscription ID", Value: cmdutil.GetString(raw, "SubscriptionId")},
				{Key: "Charge ID", Value: cmdutil.GetString(raw, "ChargeId")},
				{Key: "Created Date", Value: cmdutil.GetString(raw, "CreatedDate")},
				{Key: "Updated Date", Value: cmdutil.GetString(raw, "UpdatedDate")},
			}
		},
	})
}
