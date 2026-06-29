// Package snapshot implements the "zr contact snapshot" command.
package snapshot

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdSnapshot creates the contact snapshot command.
func NewCmdSnapshot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot <snapshot-id>",
		Short: "Get a contact snapshot",
		Long: `Get a contact snapshot by snapshot ID.

Note: This uses the snapshot ID, not the contact ID.`,
		Example: `  zr contact snapshot 8aca822f12345
  zr contact snapshot 8aca822f12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshot(cmd, f, args[0])
		},
	}
	return cmd
}

func runSnapshot(cmd *cobra.Command, f *factory.Factory, id string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/contact-snapshots/%s", url.PathEscape(id)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "First Name", Value: cmdutil.GetString(raw, "firstName")},
				{Key: "Last Name", Value: cmdutil.GetString(raw, "lastName")},
				{Key: "Email", Value: cmdutil.GetString(raw, "workEmail")},
				{Key: "Country", Value: cmdutil.GetString(raw, "country")},
				// A contact snapshot is a point-in-time copy of a contact, so it
				// mirrors the contact field names (zipCode, like contact get — not
				// "postalCode"). The snapshot endpoint is not live-probeable on this
				// sandbox; this matches the verified contact shape. (#427)
				{Key: "Postal Code", Value: cmdutil.GetString(raw, "zipCode")},
				{Key: "Contact ID", Value: cmdutil.GetString(raw, "contactId")},
			}
		},
	})
}
