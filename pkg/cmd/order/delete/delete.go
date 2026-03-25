// Package delete implements the "zr order delete" command.
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

// NewCmdDelete creates the order delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <order-number>",
		Short: "Delete an order",
		Long: `Delete a Zuora order.

This action is irreversible. Use --confirm to proceed.

Examples:
  zr order delete O-00000001 --confirm`,
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

func runDelete(cmd *cobra.Command, opts *deleteOptions, orderNumber string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/orders/%s", url.PathEscape(orderNumber)), api.WithCheckSuccess())
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
		fmt.Fprintf(f.IOStreams.ErrOut, "Order %s deleted.\n", orderNumber)
		return nil
	}

	// Response has a body (e.g. async delete returns job info)
	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Unexpected response from server while deleting order %s:\n%s\n", orderNumber, string(resp.Body))
		return fmt.Errorf("failed to parse server response for order delete")
	}

	fields := []output.DetailField{
		{Key: "Success", Value: fmt.Sprintf("%v", raw["success"])},
	}
	if jobID, ok := raw["jobId"].(string); ok {
		fields = append(fields, output.DetailField{Key: "Job ID", Value: jobID})
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
