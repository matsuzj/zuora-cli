// Package get implements the "zr contact get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the contact get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <contact-id>",
		Short: "Get contact details",
		Long:  `Get detailed information about a Zuora contact.`,
		Example: `  zr contact get 8aca822f12345
  zr contact get 8aca822f12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, id string) error {
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/contacts/%s", url.PathEscape(id)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "First Name", Value: cmdutil.GetString(raw, "firstName")},
				{Key: "Last Name", Value: cmdutil.GetString(raw, "lastName")},
				{Key: "Email", Value: cmdutil.GetString(raw, "workEmail")},
				{Key: "Phone", Value: cmdutil.GetString(raw, "workPhone")},
				{Key: "Country", Value: cmdutil.GetString(raw, "country")},
				{Key: "State", Value: cmdutil.GetString(raw, "state")},
				{Key: "City", Value: cmdutil.GetString(raw, "city")},
				{Key: "Address 1", Value: cmdutil.GetString(raw, "address1")},
				{Key: "Postal Code", Value: cmdutil.GetString(raw, "zipCode")},
				{Key: "Account ID", Value: cmdutil.GetString(raw, "accountId")},
			}
		},
	})
}
