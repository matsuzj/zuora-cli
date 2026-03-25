// Package activate implements the "zr order activate" command.
package activate

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdActivate creates the order activate command.
func NewCmdActivate(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <order-number>",
		Short: "Activate an order",
		Long: `Activate a Zuora order.

Examples:
  zr order activate O-00000001`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runActivate(cmd, f, args[0])
		},
	}
	return cmd
}

func runActivate(cmd *cobra.Command, f *factory.Factory, orderNumber string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Put(fmt.Sprintf("/v1/orders/%s/activate", url.PathEscape(orderNumber)), nil, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Order Number", Value: getString(raw, "orderNumber")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Order %s activated.\n", orderNumber)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
