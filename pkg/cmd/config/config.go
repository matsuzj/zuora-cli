// Package config implements the "zr config" command group.
package config

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdConfig creates the config parent command.
func NewCmdConfig(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <command>",
		Short: "Manage CLI configuration",
	}

	cmd.AddCommand(newCmdSet(f))
	cmd.AddCommand(newCmdGet(f))
	cmd.AddCommand(newCmdList(f))
	cmd.AddCommand(newCmdEnv(f))

	return cmd
}
