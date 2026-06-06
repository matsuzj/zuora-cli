// Package invoice implements the "zr invoice" command group.
package invoice

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/email"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/files"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/items"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/post"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/reverse"
	usageratedetail "github.com/matsuzj/zuora-cli/pkg/cmd/invoice/usage-rate-detail"
	"github.com/matsuzj/zuora-cli/pkg/cmd/invoice/writeoff"
	"github.com/spf13/cobra"
)

// NewCmdInvoice creates the invoice parent command.
func NewCmdInvoice(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoice <command>",
		Short: "Manage Zuora invoices",
		Long:  "List, view, and manage Zuora invoices.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(items.NewCmdItems(f))
	cmd.AddCommand(files.NewCmdFiles(f))
	cmd.AddCommand(email.NewCmdEmail(f))
	cmd.AddCommand(usageratedetail.NewCmdUsageRateDetail(f))
	cmd.AddCommand(post.NewCmdPost(f))
	cmd.AddCommand(reverse.NewCmdReverse(f))
	cmd.AddCommand(writeoff.NewCmdWriteoff(f))

	return cmd
}
