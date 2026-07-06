// Package list implements the "zr commitment list" command.
package list

import (
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
	// cleanly (#512 companion — the hand-written value guard existed only
	// because the alias era's Changed-bit check couldn't see alias values).
	_ = cmd.MarkFlagRequired("account-number")
	return cmd
}
