// Package delete implements the "zr fulfillment delete" command.
package delete

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	Factory *factory.Factory
	Confirm bool
}

// NewCmdDelete creates the fulfillment delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <fulfillment-key>",
		Short: "Delete a fulfillment",
		Long: `Delete a Zuora fulfillment.

This action is irreversible. Use --confirm to proceed.

Examples:
  zr fulfillment delete F-00000001 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !opts.Confirm {
				return fmt.Errorf("this action is irreversible. Use --confirm to proceed")
			}
			return runDelete(cmd, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Confirm the deletion")

	return cmd
}

func runDelete(cmd *cobra.Command, opts *deleteOptions, fulfillmentKey string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/fulfillments/%s", url.PathEscape(fulfillmentKey)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	// DELETE returns 204 (no body) on success
	if resp.StatusCode == 204 {
		synth := []byte(`{"success": true}`)
		if fmtOpts.JQ != "" {
			return output.PrintJSON(f.IOStreams, synth, fmtOpts.JQ)
		}
		if fmtOpts.JSON {
			return output.PrintJSON(f.IOStreams, synth, "")
		}
		if fmtOpts.Template != "" {
			return output.PrintTemplate(f.IOStreams, synth, fmtOpts.Template)
		}
		fmt.Fprintf(f.IOStreams.ErrOut, "Fulfillment %s deleted.\n", fulfillmentKey)
		return nil
	}

	// Response has a body
	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Unexpected response from server while deleting fulfillment %s:\n%s\n", fulfillmentKey, string(resp.Body))
		return fmt.Errorf("failed to parse server response for fulfillment delete")
	}

	fields := []output.DetailField{
		{Key: "Success", Value: fmt.Sprintf("%v", raw["success"])},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
