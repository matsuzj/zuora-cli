// Package auth implements the "zr auth" command group.
package auth

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdAuth creates the auth parent command.
func NewCmdAuth(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Manage authentication",
	}

	cmd.AddCommand(newCmdLogin(f))
	cmd.AddCommand(newCmdLogout(f))
	cmd.AddCommand(newCmdStatus(f))
	cmd.AddCommand(newCmdToken(f))

	return cmd
}
