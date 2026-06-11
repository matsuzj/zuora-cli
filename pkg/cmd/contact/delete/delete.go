// Package delete implements the "zr contact delete" command.
package delete

import (
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

	return cmdutil.RenderDeleteResult(f.IOStreams, resp, fmtOpts,
		fmt.Sprintf("Contact %s deleted.\n", id),
		func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		})
}
