// Package rateplan implements the "zr rateplan" command group.
package rateplan

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/rateplan/get"
	"github.com/spf13/cobra"
)

// NewCmdRatePlan creates the rateplan parent command.
func NewCmdRatePlan(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rateplan <command>",
		Short: "Manage Zuora rate plans",
		Long:  "View Zuora rate plans (v1 API).",
	}

	cmd.AddCommand(get.NewCmdGet(f))

	return cmd
}
