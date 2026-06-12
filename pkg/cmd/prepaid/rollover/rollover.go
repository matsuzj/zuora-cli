// Package rollover implements the "zr prepaid rollover" command.
package rollover

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type rolloverOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdRollover creates the prepaid rollover command.
func NewCmdRollover(f *factory.Factory) *cobra.Command {
	opts := &rolloverOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "rollover",
		Short: "Rollover prepaid balance",
		Long:  `Rollover a prepaid balance in Zuora.`,
		Example: `  zr prepaid rollover --body @rollover.json
  zr prepaid rollover --body '{"subscriptionNumber":"A-S001"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runRollover(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runRollover(cmd *cobra.Command, opts *rolloverOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/ppdd/rollover",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return "Prepaid rollover completed.\n"
		},
	})
}
