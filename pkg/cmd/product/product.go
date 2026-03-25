// Package product implements the "zr product" command group.
package product

import (
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/product/create"
	"github.com/matsuzj/zuora-cli/pkg/cmd/product/get"
	listlegacy "github.com/matsuzj/zuora-cli/pkg/cmd/product/list-legacy"
	"github.com/matsuzj/zuora-cli/pkg/cmd/product/update"
	"github.com/spf13/cobra"
)

// NewCmdProduct creates the product parent command.
func NewCmdProduct(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "product <command>",
		Short: "Manage Zuora commerce products",
		Long:  "Create, update, get, and list Zuora commerce products.",
	}

	cmd.AddCommand(create.NewCmdCreate(f))
	cmd.AddCommand(update.NewCmdUpdate(f))
	cmd.AddCommand(get.NewCmdGet(f))
	cmd.AddCommand(listlegacy.NewCmdListLegacy(f))

	return cmd
}
