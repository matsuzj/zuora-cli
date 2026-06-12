// Package updatecustomfields implements the "zr subscription update-custom-fields" command.
package updatecustomfields

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdUpdateCustomFields creates the subscription update-custom-fields command.
func NewCmdUpdateCustomFields(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "update-custom-fields <subscription-number> <version>",
		Short: "Update custom fields on a subscription version",
		Long: `Update custom fields on a specific version of a Zuora subscription.

This endpoint accepts a subscription NUMBER (e.g. A-S00000001), not a subscription ID.`,
		Example: `  zr subscription update-custom-fields A-S001 1 --body @fields.json
  zr sub update-custom-fields A-S001 1 --body '{"cf_MyField__c":"value"}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdateCustomFields(cmd, f, args[0], args[1], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runUpdateCustomFields(cmd *cobra.Command, f *factory.Factory, num, ver, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path: fmt.Sprintf("/v1/subscriptions/%s/versions/%s/customFields",
			url.PathEscape(num), url.PathEscape(ver)),
		Body: bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return "Custom fields updated.\n"
		},
	})
}
