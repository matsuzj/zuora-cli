// Package subscription implements the "zr subscription" command group.
package subscription

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/metrics"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/versions"
	"github.com/spf13/cobra"
)

// NewCmdSubscription creates the subscription parent command.
func NewCmdSubscription(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription <command>",
		Aliases: []string{"sub"},
		Short:   "Manage Zuora subscriptions",
		Long:    "List, view, and inspect Zuora subscriptions.",
	}

	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(versions.NewCmdVersions(f))
	cmd.AddCommand(metrics.NewCmdMetrics(f))

	return cmd
}
