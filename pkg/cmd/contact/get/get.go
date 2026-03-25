// Package get implements the "zr contact get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the contact get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <contact-id>",
		Short: "Get contact details",
		Long: `Get detailed information about a Zuora contact.

Examples:
  zr contact get 8aca822f12345
  zr contact get 8aca822f12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, id string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/contacts/%s", url.PathEscape(id)))
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: getString(raw, "id")},
		{Key: "First Name", Value: getString(raw, "firstName")},
		{Key: "Last Name", Value: getString(raw, "lastName")},
		{Key: "Email", Value: getString(raw, "workEmail")},
		{Key: "Phone", Value: getString(raw, "workPhone")},
		{Key: "Country", Value: getString(raw, "country")},
		{Key: "State", Value: getString(raw, "state")},
		{Key: "City", Value: getString(raw, "city")},
		{Key: "Address 1", Value: getString(raw, "address1")},
		{Key: "Postal Code", Value: getString(raw, "postalCode")},
		{Key: "Account ID", Value: getString(raw, "accountId")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
