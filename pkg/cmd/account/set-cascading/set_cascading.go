// Package setcascading implements the "zr account set-cascading" command.
package setcascading

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

type setCascadingOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdSetCascading creates the account set-cascading command.
func NewCmdSetCascading(f *factory.Factory) *cobra.Command {
	opts := &setCascadingOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "set-cascading <account-key>",
		Short: "Configure cascading payment methods",
		Long: `Configure cascading payment methods for a Zuora billing account.

Examples:
  zr account set-cascading A00000001 --body @cascading.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runSetCascading(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runSetCascading(cmd *cobra.Command, opts *setCascadingOptions, key string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/accounts/%s/payment-methods/cascading", url.PathEscape(key)), bodyReader, api.WithCheckSuccess())
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

	fmt.Fprintf(f.IOStreams.ErrOut, "Cascading payment methods updated for account %s.\n", key)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
