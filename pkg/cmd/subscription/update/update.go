// Package update implements the "zr subscription update" command.
package update

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdUpdate creates the subscription update command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "update <subscription-key>",
		Short: "Update a subscription",
		Long:  `Update an existing Zuora subscription.`,
		Example: `  zr subscription update SUB-001 --body @update.json
  zr sub update SUB-001 --body '{"notes":"updated"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdate(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runUpdate(cmd *cobra.Command, f *factory.Factory, key, body string) error {
	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/subscriptions/%s", url.PathEscape(key)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Subscription %s updated.\n", key)
		},
	})
}
