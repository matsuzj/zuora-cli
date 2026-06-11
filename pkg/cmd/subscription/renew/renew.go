// Package renew implements the "zr subscription renew" command.
package renew

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdRenew creates the subscription renew command.
func NewCmdRenew(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "renew <subscription-key>",
		Short: "Renew a subscription",
		Long: `Renew an existing Zuora subscription.

Renews using existing term settings. Use --body to override billing options.

Examples:
  zr subscription renew SUB-001
  zr subscription renew SUB-001 --body '{"collect":true}'
  zr sub renew SUB-001 --body @renew.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRenew(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, false)
	return cmd
}

func runRenew(cmd *cobra.Command, f *factory.Factory, key, body string) error {
	var bodyReader io.Reader
	if body != "" {
		r, err := cmdutil.ResolveBody(body, f.IOStreams.In)
		if err != nil {
			return err
		}
		bodyReader = r
	} else {
		bodyReader = strings.NewReader("{}")
	}

	path := fmt.Sprintf("/v1/subscriptions/%s/renew", url.PathEscape(key))
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   path,
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Subscription %s renewed.\n", key)
		},
	})
}
