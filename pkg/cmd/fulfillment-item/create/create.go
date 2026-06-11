// Package create implements the "zr fulfillment-item create" command.
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

// NewCmdCreate creates the fulfillment-item create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a fulfillment item",
		Long: `Create a new Zuora fulfillment item.

Examples:
  zr fulfillment-item create --body @item.json
  zr fulfillment-item create --body '{"fulfillmentKey":"F-001","quantity":5}'`,
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
		Path:   "/v1/fulfillment-items",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: firstItemID(raw)},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if itemID := firstItemID(raw); itemID != "" {
				return fmt.Sprintf("Fulfillment item %s created.\n", itemID)
			}
			return ""
		},
	})
}

// firstItemID extracts the first created item id from the bulk response.
// POST /v1/fulfillment-items is the BULK create endpoint: created item ids are
// returned under a top-level "fulfillmentItems" array, not a flat top-level
// "id". "success" is top-level.
func firstItemID(raw map[string]interface{}) string {
	if arr, ok := raw["fulfillmentItems"].([]interface{}); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]interface{}); ok {
			return cmdutil.GetString(first, "id")
		}
	}
	return ""
}
