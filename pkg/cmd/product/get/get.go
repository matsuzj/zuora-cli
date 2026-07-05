// Package get implements the "zr product get" command.
package get

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the product get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <product-key>",
		Short: "Get a commerce product",
		Long:  `Get detailed information about a Zuora commerce product.`,
		Example: `  zr product get PROD-001
  zr product get PROD-001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, productKey string) error {
	// The official operation is POST /commerce/products/{key} ("Retrieve a
	// product by key", with an optional expand body we don't send) — it is a
	// read despite the verb, like the other Commerce query/list endpoints on
	// the read-only allowlist. Doc-verified 2026-07-05 (#435); Commerce is not
	// provisioned on this sandbox, so GET-vs-POST tolerance is doc-only.
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   fmt.Sprintf("/commerce/products/%s", url.PathEscape(productKey)),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// Real keys are camelCase (startDate/endDate — the old snake_case
			// start_date/end_date never existed, so the date rows were always
			// blank), and the product object carries no top-level description.
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "id")},
				{Key: "Name", Value: cmdutil.GetString(raw, "name")},
				{Key: "SKU", Value: cmdutil.GetString(raw, "sku")},
				{Key: "State", Value: cmdutil.GetString(raw, "state")},
				{Key: "Start Date", Value: cmdutil.GetString(raw, "startDate")},
				{Key: "End Date", Value: cmdutil.GetString(raw, "endDate")},
			}
		},
	})
}
