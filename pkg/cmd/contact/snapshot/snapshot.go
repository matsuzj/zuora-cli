// Package snapshot implements the "zr contact snapshot" command.
package snapshot

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdSnapshot creates the contact snapshot command.
func NewCmdSnapshot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot <snapshot-id>",
		Short: "Get a contact snapshot",
		Long: `Get a contact snapshot by snapshot ID.

Note: This uses the snapshot ID, not the contact ID.

Examples:
  zr contact snapshot 8aca822f12345
  zr contact snapshot 8aca822f12345 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshot(cmd, f, args[0])
		},
	}
	return cmd
}

func runSnapshot(cmd *cobra.Command, f *factory.Factory, id string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/contact-snapshots/%s", url.PathEscape(id)))
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
		{Key: "Country", Value: getString(raw, "country")},
		{Key: "Contact ID", Value: getString(raw, "contactId")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
