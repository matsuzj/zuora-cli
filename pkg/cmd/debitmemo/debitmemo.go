// Package debitmemo implements the "zr debitmemo" command group.
package debitmemo

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/debitmemo/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/debitmemo/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdDebitMemo creates the debitmemo parent command.
func NewCmdDebitMemo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "debitmemo <command>",
		Aliases: []string{"debit-memo"},
		Short:   "Manage Zuora debit memos",
		Long:    "List and view Zuora debit memos.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))

	return cmd
}
