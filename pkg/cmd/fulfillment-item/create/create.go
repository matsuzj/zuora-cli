// Package create implements the "zr fulfillment-item create" command.
package create

import (
	"encoding/json"
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
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/fulfillment-items", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// POST /v1/fulfillment-items is the BULK create endpoint: created item ids are
	// returned under a top-level "fulfillmentItems" array, not a flat top-level
	// "id". "success" is top-level.
	var itemID string
	if arr, ok := raw["fulfillmentItems"].([]interface{}); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]interface{}); ok {
			itemID = cmdutil.GetString(first, "id")
		}
	}

	fields := []output.DetailField{
		{Key: "ID", Value: itemID},
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	if itemID != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Fulfillment item %s created.\n", itemID)
	}
	return nil
}
