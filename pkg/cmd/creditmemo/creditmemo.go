// Package creditmemo implements the "zr creditmemo" command group.
package creditmemo

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/creditmemo/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/creditmemo/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdCreditMemo creates the creditmemo parent command.
func NewCmdCreditMemo(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "creditmemo <command>",
		Short: "Manage Zuora credit memos",
		Long:  "List and view Zuora credit memos.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))

	return cmd
}
