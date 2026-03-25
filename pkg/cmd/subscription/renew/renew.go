// Package renew implements the "zr subscription renew" command.
package renew

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/internal/api"
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

	cmd.Flags().StringVarP(&body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	return cmd
}

func runRenew(cmd *cobra.Command, f *factory.Factory, key, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != "" {
		bodyReader, err = cmdutil.ResolveBody(body, f.IOStreams.In)
		if err != nil {
			return err
		}
	} else {
		bodyReader = strings.NewReader("{}")
	}

	path := fmt.Sprintf("/v1/subscriptions/%s/renew", url.PathEscape(key))
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

	fmt.Fprintf(f.IOStreams.ErrOut, "Subscription %s renewed.\n", key)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
