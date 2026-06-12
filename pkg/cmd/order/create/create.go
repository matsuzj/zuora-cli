// Package create implements the "zr order create" command.
package create

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the order create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an order",
		Long:  `Create a new Zuora order.`,
		Example: `  zr order create --body @order.json
  zr order create --body '{"existingAccountNumber":"A001","orderDate":"2026-01-01","orderLineItems":[...]}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/orders",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Order Number", Value: cmdutil.GetString(raw, "orderNumber")},
				{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
				{Key: "Status", Value: cmdutil.GetString(raw, "status")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if num := cmdutil.GetString(raw, "orderNumber"); num != "" {
				return fmt.Sprintf("Order %s created.\n", num)
			}
			return ""
		},
	})
}
