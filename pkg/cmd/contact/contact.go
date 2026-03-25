// Package contact implements the "zr contact" command group.
package contact

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/contact/create"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/contact/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/contact/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/contact/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/contact/scrub"
	"github.com/matsuzj/zuora-cli/pkg/cmd/contact/snapshot"
	"github.com/matsuzj/zuora-cli/pkg/cmd/contact/transfer"
	"github.com/matsuzj/zuora-cli/pkg/cmd/contact/update"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdContact creates the contact parent command.
func NewCmdContact(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contact <command>",
		Short: "Manage Zuora contacts",
		Long:  "List, view, create, update, and manage Zuora contacts.",
	}

	// Read commands
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(snapshot.NewCmdSnapshot(f))

	// Write commands
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))
	cmd.AddCommand(transfer.NewCmdTransfer(f))
	cmd.AddCommand(scrub.NewCmdScrub(f))

	return cmd
}
