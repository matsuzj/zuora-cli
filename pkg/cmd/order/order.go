// Package order implements the "zr order" command group.
package order

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/activate"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/cancel"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/create"
	createasync "github.com/matsuzj/zuora-cli/pkg/cmd/order/create-async"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/order/delete"
	deleteasync "github.com/matsuzj/zuora-cli/pkg/cmd/order/delete-async"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/get"
	jobstatus "github.com/matsuzj/zuora-cli/pkg/cmd/order/job-status"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/list"
	listbyinvoiceowner "github.com/matsuzj/zuora-cli/pkg/cmd/order/list-by-invoice-owner"
	listbysubscription "github.com/matsuzj/zuora-cli/pkg/cmd/order/list-by-subscription"
	listbysubscriptionowner "github.com/matsuzj/zuora-cli/pkg/cmd/order/list-by-subscription-owner"
	listpending "github.com/matsuzj/zuora-cli/pkg/cmd/order/list-pending"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/preview"
	previewasync "github.com/matsuzj/zuora-cli/pkg/cmd/order/preview-async"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/revert"
	"github.com/matsuzj/zuora-cli/pkg/cmd/order/update"
	updatecustomfields "github.com/matsuzj/zuora-cli/pkg/cmd/order/update-custom-fields"
	updatetriggerdates "github.com/matsuzj/zuora-cli/pkg/cmd/order/update-trigger-dates"
	"github.com/spf13/cobra"
)

// NewCmdOrder creates the order parent command.
func NewCmdOrder(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order <command>",
		Short: "Manage Zuora orders",
		Long:  "List, view, create, update, delete, and manage Zuora orders.",
	}

	// CRUD commands
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	// Lifecycle commands
	cmd.AddCommand(activate.NewCmdActivate(f))
	cmd.AddCommand(cancel.NewCmdCancel(f))
	cmd.AddCommand(revert.NewCmdRevert(f))
	cmd.AddCommand(preview.NewCmdPreview(f))

	// Query commands
	cmd.AddCommand(listbysubscriptionowner.NewCmdListBySubscriptionOwner(f))
	cmd.AddCommand(listbysubscription.NewCmdListBySubscription(f))
	cmd.AddCommand(listpending.NewCmdListPending(f))
	cmd.AddCommand(listbyinvoiceowner.NewCmdListByInvoiceOwner(f))

	// Custom fields & trigger dates
	cmd.AddCommand(updatecustomfields.NewCmdUpdateCustomFields(f))
	cmd.AddCommand(updatetriggerdates.NewCmdUpdateTriggerDates(f))

	// Async operations
	cmd.AddCommand(createasync.NewCmdCreateAsync(f))
	cmd.AddCommand(previewasync.NewCmdPreviewAsync(f))
	cmd.AddCommand(deleteasync.NewCmdDeleteAsync(f))
	cmd.AddCommand(jobstatus.NewCmdJobStatus(f))

	return cmd
}
