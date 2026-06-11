// Package create implements the "zr fulfillment create" command.
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

// NewCmdCreate creates the fulfillment create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a fulfillment",
		Long: `Create a new Zuora fulfillment.

Examples:
  zr fulfillment create --body @fulfillment.json
  zr fulfillment create --body '{"orderLineItemId":"..."}'`,
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

	resp, err := client.Post("/v1/fulfillments", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// POST /v1/fulfillments is the BULK create endpoint: the created object(s) are
	// returned under a top-level "fulfillments" array (keyed by id/fulfillmentNumber),
	// not a flat top-level "key". "success" is top-level.
	var fulfillmentID, fulfillmentNumber string
	if arr, ok := raw["fulfillments"].([]interface{}); ok && len(arr) > 0 {
		if first, ok := arr[0].(map[string]interface{}); ok {
			fulfillmentID = cmdutil.GetString(first, "id")
			fulfillmentNumber = cmdutil.GetString(first, "fulfillmentNumber")
		}
	}

	fields := []output.DetailField{
		{Key: "Fulfillment Number", Value: fulfillmentNumber},
		{Key: "ID", Value: fulfillmentID},
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	ref := fulfillmentNumber
	if ref == "" {
		ref = fulfillmentID
	}
	if ref != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Fulfillment %s created.\n", ref)
	}
	return nil
}
