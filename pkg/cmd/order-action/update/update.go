// Package update implements the "zr order-action update" command.
package update

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdUpdate creates the order-action update command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	var body string

	cmd := &cobra.Command{
		Use:   "update <order-action-id>",
		Short: "Update an order action",
		Long: `Update a Zuora order action.

Examples:
  zr order-action update 2c92c0f9876... --body @action.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			return runUpdate(cmd, f, args[0], body)
		},
	}

	cmdutil.AddBodyFlag(cmd, &body, true)
	return cmd
}

func runUpdate(cmd *cobra.Command, f *factory.Factory, id, body string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	bodyReader, err := cmdutil.ResolveBody(body, f.IOStreams.In)
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/orderActions/%s", url.PathEscape(id)), bodyReader)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Order action %s updated.\n", id)
	return nil
}
