// Package usage implements the "zr usage" command group.
package usage

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/usage/create"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/usage/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/usage/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/usage/post"
	"github.com/matsuzj/zuora-cli/pkg/cmd/usage/update"
	"github.com/spf13/cobra"
)

// NewCmdUsage creates the usage parent command.
func NewCmdUsage(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "usage <command>",
		Short: "Manage Zuora usage records",
		Long:  "Post, create, view, update, and delete Zuora usage records.",
	}

	cmd.AddCommand(post.NewCmdPost(f))
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))

	return cmd
}
