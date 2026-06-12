// Package setcascading implements the "zr account set-cascading" command.
package setcascading

import (
	"fmt"
	"net/url"

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
		Use:     "set-cascading <account-key>",
		Short:   "Configure cascading payment methods",
		Long:    `Configure cascading payment methods for a Zuora billing account.`,
		Example: `  zr account set-cascading A00000001 --body @cascading.json`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runSetCascading(cmd, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runSetCascading(cmd *cobra.Command, opts *setCascadingOptions, key string) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/accounts/%s/payment-methods/cascading", url.PathEscape(key)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Cascading payment methods updated for account %s.\n", key)
		},
	})
}
