// Package delete implements the "zr usage delete" command.
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

// NewCmdDelete creates the usage delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	opts := &deleteOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a usage record",
		Long: `Delete a usage record by ID via the CRUD API.

This action is irreversible. Use --confirm to proceed.

Examples:
  zr usage delete 2c92a0f96bd... --confirm`,
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

func runDelete(cmd *cobra.Command, opts *deleteOptions, id string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/v1/object/usage/%s", url.PathEscape(id)), api.WithCheckSuccess())
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
		fmt.Fprintf(f.IOStreams.ErrOut, "Usage record %s deleted.\n", id)
		return nil
	}

	// Response has a body
	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Unexpected response from server while deleting usage record %s:\n%s\n", id, string(resp.Body))
		return fmt.Errorf("failed to parse server response for usage delete")
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(raw, "Id")},
		{Key: "Success", Value: getString(raw, "Success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Usage record %s deleted.\n", id)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
