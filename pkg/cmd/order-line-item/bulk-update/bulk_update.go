// Package bulkupdate implements the "zr order-line-item bulk-update" command.
package bulkupdate

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdBulkUpdate creates the order-line-item bulk-update command.
func NewCmdBulkUpdate(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "bulk-update",
		Short: "Bulk update order line items",
		Long: `Bulk update Zuora order line items (max 100 items per request).

The --body must contain a JSON object with an "orderLineItems" array.

Examples:
  zr order-line-item bulk-update --body @items.json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runBulkUpdate(cmd, f, body)
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	return cmd
}

func runBulkUpdate(cmd *cobra.Command, f *factory.Factory, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/order-line-items/bulk", bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Order line items bulk updated.\n")
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
