// Package delete implements the "zr contact delete" command.
package delete

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdDelete creates the contact delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <contact-id>",
		Short: "Delete a contact",
		Long: `Delete a Zuora contact. This action is irreversible.

Examples:
  zr contact delete 8aca822f12345 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runDelete(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "deletion")
	return cmd
}

func runDelete(cmd *cobra.Command, f *factory.Factory, id string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// Zuora DELETE /v1/contacts returns HTTP 200 with {"success": true/false}
	resp, err := client.Delete(fmt.Sprintf("/v1/contacts/%s", url.PathEscape(id)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	// DELETE may return 200 with JSON body or 204 with no body
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
		fmt.Fprintf(f.IOStreams.ErrOut, "Contact %s deleted.\n", id)
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}
	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Contact %s deleted.\n", id)
	return nil
}
