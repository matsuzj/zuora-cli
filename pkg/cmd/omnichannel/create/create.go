// Package create implements the "zr omnichannel create" command.
package create

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the omnichannel create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an omni-channel subscription",
		Long: `Create a new Zuora omni-channel subscription.

Examples:
  zr omnichannel create --body @omnichannel.json
  zr omnichannel create --body '{"subscriptionKey":"S-001",...}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCreate(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/v1/omni-channel-subscriptions", bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Subscription Key", Value: getString(raw, "subscriptionKey")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	if key := getString(raw, "subscriptionKey"); key != "" {
		fmt.Fprintf(f.IOStreams.ErrOut, "Omni-channel subscription %s created.\n", key)
	}
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
