// Package delete implements the "zr subscription delete" command.
package delete

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdDelete creates the subscription delete command.
func NewCmdDelete(f *factory.Factory) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <subscription-key>",
		Short: "Delete a subscription",
		Long: `Delete a Zuora subscription. This action is irreversible.

Note: Zuora uses PUT (not HTTP DELETE) for this operation.

Examples:
  zr subscription delete A-S001 --confirm`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return fmt.Errorf("this action is irreversible. Use --confirm to proceed")
			}
			return runDelete(cmd, f, args[0])
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm the deletion")
	return cmd
}

func runDelete(cmd *cobra.Command, f *factory.Factory, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// Zuora uses PUT (not HTTP DELETE) for subscription deletion
	resp, err := client.Put(
		fmt.Sprintf("/v1/subscriptions/%s/delete", url.PathEscape(key)),
		strings.NewReader("{}"),
		api.WithCheckSuccess(),
	)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Success", Value: getString(raw, "success")},
	}
	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Subscription %s deleted.\n", key)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
