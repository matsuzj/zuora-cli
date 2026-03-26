// Package subscription implements the "zr subscription" command group.
package subscription

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/cancel"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/changelog"
	changelogbyorder "github.com/matsuzj/zuora-cli/pkg/cmd/subscription/changelog-by-order"
	changelogversion "github.com/matsuzj/zuora-cli/pkg/cmd/subscription/changelog-version"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/create"
	deletecmd "github.com/matsuzj/zuora-cli/pkg/cmd/subscription/delete"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/get"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/list"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/metrics"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/preview"
	previewchange "github.com/matsuzj/zuora-cli/pkg/cmd/subscription/preview-change"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/renew"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/resume"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/suspend"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/update"
	updatecustomfields "github.com/matsuzj/zuora-cli/pkg/cmd/subscription/update-custom-fields"
	"github.com/matsuzj/zuora-cli/pkg/cmd/subscription/versions"
	"github.com/spf13/cobra"
)

// NewCmdSubscription creates the subscription parent command.
func NewCmdSubscription(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subscription <command>",
		Aliases: []string{"sub"},
		Short:   "Manage Zuora subscriptions",
		Long:    "List, view, create, update, and manage Zuora subscriptions.",
	}

	// Read commands
	cmd.AddCommand(list.NewCmdList(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(versions.NewCmdVersions(f))
	cmd.AddCommand(metrics.NewCmdMetrics(f))

	// Write commands
	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(cancel.NewCmdCancel(f))
	cmd.AddCommand(suspend.NewCmdSuspend(f))
	cmd.AddCommand(resume.NewCmdResume(f))
	cmd.AddCommand(renew.NewCmdRenew(f))
	cmd.AddCommand(deletecmd.NewCmdDelete(f))
	cmd.AddCommand(preview.NewCmdPreview(f))
	cmd.AddCommand(previewchange.NewCmdPreviewChange(f))
	cmd.AddCommand(updatecustomfields.NewCmdUpdateCustomFields(f))

	// Changelog commands
	cmd.AddCommand(changelog.NewCmdChangelog(f))
	cmd.AddCommand(changelogbyorder.NewCmdChangelogByOrder(f))
	cmd.AddCommand(changelogversion.NewCmdChangelogVersion(f))

	return cmd
}
