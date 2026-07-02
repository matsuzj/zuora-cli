// Package billrun implements the "zr billrun" command group.
package billrun

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/billrun/cancel"
	"github.com/matsuzj/zuora-cli/pkg/cmd/billrun/create"
	"github.com/matsuzj/zuora-cli/pkg/cmd/billrun/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/billrun/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/billrun/post"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdBillRun creates the billrun parent command.
func NewCmdBillRun(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "billrun <command>",
		// Accept the kebab-case spelling too; multi-word resources are otherwise
		// unguessable (some are concatenated, some hyphenated). Additive only —
		// `billrun` stays the canonical name.
		Aliases: []string{"bill-run"},
		Short:   "Manage Zuora bill runs",
		Long:    "Create, view, post, cancel, and delete Zuora bill runs.",
	}

	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(post.NewCmdPost(f))
	cmd.AddCommand(cancel.NewCmdCancel(f))
	cmd.AddCommand(delete.NewCmdDelete(f))

	return cmd
}
