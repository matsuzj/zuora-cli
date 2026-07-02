// Package list implements the "zr order list" command.
package list

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the order list command. Besides the default all-orders
// listing it narrows by scope via mutually-exclusive flags — these fold in the
// former list-by-subscription / list-by-subscription-owner / list-by-invoice-owner
// commands (#454). list-pending stays a separate command for now (it needs a
// boolean the declarative list runner does not yet model).
func NewCmdList(f *factory.Factory) *cobra.Command {
	cmd := listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List orders",
		Long: `List Zuora orders.

By default lists every order. Narrow the scope with exactly one of
--subscription, --subscription-owner, or --invoice-owner.`,
		Example: `  zr order list
  zr order list --status Completed
  zr order list --subscription A-S00000001
  zr order list --invoice-owner A00000001 --json`,
		Flags: []listcmd.Flag{
			{Name: "status", Query: "status", Usage: "Filter by order status"},
			{Name: "subscription", Usage: "List orders for this subscription number/key"},
			{Name: "subscription-owner", Usage: "List orders whose subscription owner is this account number"},
			{Name: "invoice-owner", Usage: "List orders whose invoice owner is this account number"},
			{Name: "page", Query: "page", Usage: "Page number"},
			{Name: "page-size", Query: "pageSize", Usage: "Number of results per page", Int: true, OmitZero: true},
		},
		Path: func(args []string, flags map[string]string) string {
			switch {
			case flags["subscription"] != "":
				return fmt.Sprintf("/v1/orders/subscription/%s", url.PathEscape(flags["subscription"]))
			case flags["subscription-owner"] != "":
				return fmt.Sprintf("/v1/orders/subscriptionOwner/%s", url.PathEscape(flags["subscription-owner"]))
			case flags["invoice-owner"] != "":
				return fmt.Sprintf("/v1/orders/invoiceOwner/%s", url.PathEscape(flags["invoice-owner"]))
			default:
				return "/v1/orders"
			}
		},
		ItemsKey: "orders",
		Columns: []listcmd.ColumnSpec{
			{Header: "ORDER_NUMBER", Key: "orderNumber"},
			{Header: "STATUS", Key: "status"},
			{Header: "ORDER_DATE", Key: "orderDate"},
			{Header: "ACCOUNT", Key: "existingAccountNumber"},
			{Header: "DESCRIPTION", Key: "description"},
			{Header: "CREATED", Key: "createdDate"},
		},
		NextPage: listcmd.NextPage{Flag: "page", FromURL: "page"},
	})

	// The declarative list runner has no error-returning validation hook, so wrap
	// its RunE to reject more than one scope flag before the request. (RunE runs
	// after the root's PersistentPreRunE, so the client is already wired.)
	inner := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		scopes := 0
		for _, n := range []string{"subscription", "subscription-owner", "invoice-owner"} {
			if v, _ := c.Flags().GetString(n); v != "" {
				scopes++
			}
		}
		if scopes > 1 {
			return fmt.Errorf("specify at most one of --subscription, --subscription-owner, --invoice-owner")
		}
		return inner(c, args)
	}
	return cmd
}
