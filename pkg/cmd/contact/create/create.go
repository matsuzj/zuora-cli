// Package create implements the "zr contact create" command.
package create

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdCreate creates the contact create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a contact",
		Long: `Create a new Zuora contact.

Examples:
  zr contact create --body @contact.json
  zr contact create --body '{"accountId":"...","firstName":"John","lastName":"Doe","country":"US"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCreate(cmd, f, body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runCreate(cmd *cobra.Command, f *factory.Factory, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/contacts", bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Contact ID", Value: cmdutil.GetString(raw, "id")},
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	if id := cmdutil.GetString(raw, "id"); id != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Contact %s created.\n", id)
	}
	return nil
}
