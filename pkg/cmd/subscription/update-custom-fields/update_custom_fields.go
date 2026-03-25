// Package updatecustomfields implements the "zr subscription update-custom-fields" command.
package updatecustomfields

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

// NewCmdUpdateCustomFields creates the subscription update-custom-fields command.
func NewCmdUpdateCustomFields(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "update-custom-fields <subscription-number> <version>",
		Short: "Update custom fields on a subscription version",
		Long: `Update custom fields on a specific version of a Zuora subscription.

Examples:
  zr subscription update-custom-fields A-S001 1 --body @fields.json
  zr sub update-custom-fields A-S001 1 --body '{"cf_MyField__c":"value"}'`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdateCustomFields(cmd, f, args[0], args[1], body)
		},
	}

	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	return cmd
}

func runUpdateCustomFields(cmd *cobra.Command, f *factory.Factory, num, ver, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/v1/subscriptions/%s/versions/%s/customFields",
		url.PathEscape(num), url.PathEscape(ver))
	resp, err := client.Put(path, bodyReader, api.WithCheckSuccess())
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

	fmt.Fprintf(f.IOStreams.ErrOut, "Custom fields updated.\n")
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
