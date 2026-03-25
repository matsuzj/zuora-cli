// Package transfer implements the "zr contact transfer" command.
package transfer

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

// NewCmdTransfer creates the contact transfer command.
func NewCmdTransfer(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "transfer <contact-id>",
		Short: "Transfer a contact to another account",
		Long: `Transfer a Zuora contact to a different account.

Examples:
  zr contact transfer 8aca822f12345 --body '{"destinationAccountId":"8aca999f67890"}'
  zr contact transfer 8aca822f12345 --body @transfer.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runTransfer(cmd, f, args[0], body)
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	return cmd
}

func runTransfer(cmd *cobra.Command, f *factory.Factory, id, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/contacts/%s/transfer", url.PathEscape(id)), bodyReader, api.WithCheckSuccess())
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

	fmt.Fprintf(f.IOStreams.ErrOut, "Contact %s transferred.\n", id)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
