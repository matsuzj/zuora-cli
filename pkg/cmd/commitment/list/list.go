// Package list implements the "zr commitment list" command.
package list

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil/listcmd"
	"github.com/spf13/cobra"
)

// NewCmdList creates the commitment list command.
//
// Table rendering is doc-verified (#453): GET /v1/commitments returns
// {total, page, page_size, commitments:[...]} with flat items — the same
// flat item shape the official schema documents for commitment get. This
// endpoint is not provisioned on the dev sandbox (404), so the shape comes
// from the published API reference, like the #486/#435 batch.
func NewCmdList(f *factory.Factory) *cobra.Command {
	cmd := listcmd.New(f, listcmd.Spec{
		Use:   "list",
		Short: "List commitments for an account",
		Long:  `List commitments associated with a Zuora account.`,
		Example: `  zr commitment list --account-number A00000001
  zr commitment list --account-number A00000001 --type MinCommitment
  zr commitment list --account-number A00000001 --json`,
		Flags: []listcmd.Flag{
			{Name: "account-number", Query: "accountNumber", Usage: "Account number (required)"},
			{Name: "type", Query: "type", Usage: "Filter by commitment type", Enum: []string{"MinCommitment", "MaxCommitment"}},
		},
		Path: func(args []string, flags map[string]string) string {
			return "/v1/commitments"
		},
		ItemsKey: "commitments",
		Columns: []listcmd.ColumnSpec{
			{Header: "NUMBER", Key: "commitmentNumber"},
			{Header: "NAME", Key: "name"},
			{Header: "TYPE", Key: "type"},
			{Header: "STATUS", Key: "status"},
			{Header: "START", Key: "startDate"},
			{Header: "END", Key: "endDate"},
			{Header: "AMOUNT", Key: "totalAmount", Money: true},
			{Header: "CURRENCY", Key: "currency"},
		},
	})
	// The endpoint requires accountNumber. With the deprecated --account
	// alias long gone (v0.7.0), cobra's own required-flag machinery applies
	// (#512 companion). An EXPLICIT empty value (--account-number "", e.g.
	// an unset shell variable) passes cobra's Changed-bit check, so a
	// value-level guard enforces non-emptiness with the same wording — the
	// P5-2 pattern (Codex catch on #512).
	_ = cmd.MarkFlagRequired("account-number")
	inner := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if v, _ := c.Flags().GetString("account-number"); v == "" {
			return fmt.Errorf(`required flag(s) "account-number" not set`)
		}
		return inner(c, args)
	}
	return cmd
}
