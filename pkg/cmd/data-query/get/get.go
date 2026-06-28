// Package get implements the "zr data-query get" command.
package get

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/dqutil"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the data-query get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <job-id>",
		Aliases: []string{"status"},
		Short:   "Get a Data Query job's status and result URL",
		Example: `  zr data-query get 2c92c0f8...
  zr data-query get 2c92c0f8... --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdutil.RunDetail(cmd, f, cmdutil.Action{
				Method: "GET",
				Path:   dqutil.JobPath(args[0]),
				Fields: func(raw map[string]interface{}) []output.DetailField {
					return dqutil.DetailFields(dqutil.UnwrapData(raw))
				},
			})
		},
	}
}
