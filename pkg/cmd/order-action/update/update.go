// Package update implements the "zr order-action update" command.
package update

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdUpdate creates the order-action update command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:     "update <order-action-id>",
		Short:   "Update an order action",
		Long:    `Update a Zuora order action.`,
		Example: `  zr order-action update 2c92c0f9876... --body @action.json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdate(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runUpdate(cmd *cobra.Command, f *factory.Factory, id, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/orderActions/%s", url.PathEscape(id)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Order action %s updated.\n", id)
		},
	})
}
