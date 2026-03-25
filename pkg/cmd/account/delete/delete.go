// Package delete implements the "zr account delete" command.
package delete

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	Factory *factory.Factory
	Confirm bool
}

// NewCmdDelete creates the account delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <account-key>",
		Short: "Delete a billing account",
		Long: `Delete a Zuora billing account. This is an async operation.

This action is irreversible. Use --confirm to proceed.

Examples:
  zr account delete A00000001 --confirm`,
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

func runDelete(cmd *cobra.Command, opts *deleteOptions, key string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/accounts/%s", url.PathEscape(key)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	// DELETE returns 204 (no body) on success
	if resp.StatusCode == 204 || len(resp.Body) == 0 {
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
		fmt.Fprintf(f.IOStreams.ErrOut, "Account %s deleted.\n", key)
		return nil
	}

	// If response has a body (unexpected for DELETE), render it
	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err == nil {
		return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, []output.DetailField{
			{Key: "Success", Value: fmt.Sprintf("%v", raw["success"])},
		})
	}
	return nil
}
