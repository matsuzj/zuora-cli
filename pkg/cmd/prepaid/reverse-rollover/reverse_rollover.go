// Package reverserollover implements the "zr prepaid reverse-rollover" command.
package reverserollover

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type reverseRolloverOptions struct {
	Factory *factory.Factory
	Body    string
	Confirm bool
}

// NewCmdReverseRollover creates the prepaid reverse-rollover command.
func NewCmdReverseRollover(f *factory.Factory) *cobra.Command {
	opts := &reverseRolloverOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "reverse-rollover",
		Short: "Reverse a prepaid rollover",
		Long:  `Reverse a prepaid balance rollover in Zuora.`,
		Example: `  zr prepaid reverse-rollover --body @reverse.json
  zr prepaid reverse-rollover --body '{"subscriptionNumber":"A-S001"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(opts.Confirm); err != nil {
				return err
			}
			return runReverseRollover(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)
	cmdutil.AddConfirmFlag(cmd, &opts.Confirm, "reversal (this action is irreversible)")

	return cmd
}

func runReverseRollover(cmd *cobra.Command, opts *reverseRolloverOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/ppdd/reverse-rollover",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return "Prepaid reverse rollover completed.\n"
		},
	})
}
