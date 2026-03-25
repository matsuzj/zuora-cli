// Package scrub implements the "zr contact scrub" command.
package scrub

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdScrub creates the contact scrub command.
func NewCmdScrub(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scrub <contact-id>",
		Short: "Scrub personal data from a contact",
		Long: `Scrub (anonymize) personal data from a Zuora contact.

This replaces personal fields with anonymized values for data privacy compliance.

Examples:
  zr contact scrub 8aca822f12345`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runScrub(cmd, f, args[0])
		},
	}
	return cmd
}

func runScrub(cmd *cobra.Command, f *factory.Factory, id string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/contacts/%s/scrub", url.PathEscape(id)), nil, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Contact %s scrubbed.\n", id)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
