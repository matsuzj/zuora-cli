// Package snapshot implements the "zr contact snapshot" command.
package snapshot

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
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

	resp, err := client.Get(fmt.Sprintf("/v1/contact-snapshots/%s", url.PathEscape(id)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "ID", Value: cmdutil.GetString(raw, "id")},
		{Key: "First Name", Value: cmdutil.GetString(raw, "firstName")},
		{Key: "Last Name", Value: cmdutil.GetString(raw, "lastName")},
		{Key: "Email", Value: cmdutil.GetString(raw, "workEmail")},
		{Key: "Country", Value: cmdutil.GetString(raw, "country")},
		{Key: "Contact ID", Value: cmdutil.GetString(raw, "contactId")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
