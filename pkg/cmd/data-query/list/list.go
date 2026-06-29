// Package list implements the "zr data-query list" command.
package list

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/dqutil"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the data-query list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	return listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List recent Data Query jobs",
		Example: `  zr data-query list
  zr data-query list --status completed --page-size 20`,
		Flags: []listcmd.Flag{
			{Name: "status", Query: "queryStatus", Usage: "Filter by job status", Enum: dqutil.JobStatuses},
			{Name: "page-size", Query: "pageSize", Usage: "Maximum jobs to return (<= 1000)", Int: true, OmitZero: true},
		},
		Path: func(args []string, flags map[string]string) string {
			return "/query/jobs"
		},
		ItemsKey: "data",
		Columns: []listcmd.ColumnSpec{
			// id + status only: listcmd renders non-Money columns via GetString,
			// which would scientific-notate a large numeric outputRows. The count
			// is shown (as a plain decimal) in the detail view instead.
			{Header: "ID", Key: "id"},
			{Header: "Status", Key: "queryStatus"},
		},
	})
}
