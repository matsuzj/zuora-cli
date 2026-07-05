// Package list implements the "zr plan list" command.
package list

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type listOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdList creates the plan list command.
//
// Hand-written runE (documented listcmd exception): this list is a
// POST-with-body query, not a GET with query flags, so listcmd.Spec does
// not model it — same class as contact list's ZOQL POST.
func NewCmdList(f *factory.Factory) *cobra.Command {
	opts := &listOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List commerce plans",
		Long: `List Zuora commerce product rate plans.

The request body follows POST /commerce/plans/list: a required "filters"
array ({field, operator, value}) plus optional expand options.`,
		Example: `  zr plan list --body '{"filters":[{"field":"state","operator":"EQ","value":"active"}]}'
  zr plan list --body @query.json --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	f := opts.Factory
	fmtOpts := output.FromCmd(cmd)

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Post("/commerce/plans/list", bodyReader)
	if err != nil {
		return err
	}

	// Doc-verified envelope (#453): the 200 response is {"values":[...]} with
	// no success flag (Commerce is unprovisioned on the dev sandbox, so the
	// shape comes from the published API reference, like the #435 batch).
	var envelope struct {
		Values []map[string]interface{} `json:"values"`
	}
	if err := json.Unmarshal(resp.Body, &envelope); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	cols := []output.Column{
		{Header: "NAME"},
		{Header: "NUMBER"},
		{Header: "PRODUCT ID"},
		{Header: "STATE"},
		{Header: "START"},
		{Header: "END"},
		{Header: "ID"},
	}
	rows := make([][]string, len(envelope.Values))
	for i, p := range envelope.Values {
		rows[i] = []string{
			cmdutil.GetString(p, "name"),
			cmdutil.GetString(p, "productRatePlanNumber"),
			cmdutil.GetString(p, "productId"),
			cmdutil.GetString(p, "state"),
			cmdutil.GetString(p, "startDate"),
			cmdutil.GetString(p, "endDate"),
			cmdutil.GetString(p, "id"),
		}
	}
	return output.Render(f.IOStreams, resp.Body, fmtOpts, rows, cols)
}
