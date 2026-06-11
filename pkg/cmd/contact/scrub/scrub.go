// Package scrub implements the "zr contact scrub" command.
package scrub

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdScrub creates the contact scrub command.
func NewCmdScrub(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "scrub <contact-id>",
		Short: "Scrub personal data from a contact",
		Long: `Scrub (anonymize) personal data from a Zuora contact.

This replaces personal fields with anonymized values for data privacy compliance.
This action is irreversible. Use --confirm to proceed.

Examples:
  zr contact scrub 8aca822f12345 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runScrub(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "scrub")
	return cmd
}

func runScrub(cmd *cobra.Command, f *factory.Factory, id string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/contacts/%s/scrub", url.PathEscape(id)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Contact %s scrubbed.\n", id)
		},
	})
}
