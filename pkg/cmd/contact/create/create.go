// Package create implements the "zr contact create" command.
package create

import (
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
		Long:  `Create a new Zuora contact.`,
		Example: `  zr contact create --body @contact.json
  zr contact create --body '{"accountId":"...","firstName":"John","lastName":"Doe","country":"US"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(cmd, f, body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runCreate(cmd *cobra.Command, f *factory.Factory, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/contacts",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Contact ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if id := cmdutil.GetString(raw, "id"); id != "" {
				return fmt.Sprintf("Contact %s created.\n", id)
			}
			return ""
		},
	})
}
