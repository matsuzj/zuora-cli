// Package dataquery implements the "zr data-query" parent command for Zuora's
// asynchronous Data Query API. Each subcommand lives in its own package
// (mirroring the other command groups) and shares helpers via the dqutil
// package; see pkg/cmd/data-query/dqutil.
package dataquery

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/cancel"
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/run"
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/submit"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdDataQuery creates the "data-query" parent command.
func NewCmdDataQuery(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "data-query <command>",
		Aliases: []string{"dq"},
		Short:   "Run Zuora Data Query (async read-only SQL)",
		Long: `Submit, track, and download Zuora Data Query jobs.

Data Query runs read-only SQL asynchronously: submit a job, poll until it
completes, then download the result file. Although submit (POST) and cancel
(DELETE) use mutating HTTP methods, Data Query never changes tenant data, so in
read-only mode they are blocked by default and allowed only with
--read-only-allow-data-query (or ZR_READ_ONLY_ALLOW_DATA_QUERY=1).`,
	}

	cmd.AddCommand(submit.NewCmdSubmit(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(cancel.NewCmdCancel(f))
	cmd.AddCommand(run.NewCmdRun(f))

	return cmd
}
