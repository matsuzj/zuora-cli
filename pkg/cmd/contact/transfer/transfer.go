// Package transfer implements the "zr contact transfer" command.
package transfer

import (
	"fmt"
	"net/url"

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
		Long:  `Transfer a Zuora contact to a different account.`,
		Example: `  zr contact transfer 8aca822f12345 --body '{"destinationAccountId":"8aca999f67890"}'
  zr contact transfer 8aca822f12345 --body @transfer.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTransfer(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runTransfer(cmd *cobra.Command, f *factory.Factory, id, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/contacts/%s/transfer", url.PathEscape(id)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Contact %s transferred.\n", id)
		},
	})
}
