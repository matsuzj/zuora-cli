// Package delete implements the "zr subscription delete" command.
package delete

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
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

Note: Zuora uses PUT (not HTTP DELETE) for this operation.`,
		Example: `  zr subscription delete A-S001 --confirm`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(confirm); err != nil {
				return err
			}
			return runDelete(cmd, f, args[0])
		},
	}

	cmdutil.AddConfirmFlag(cmd, &confirm, "deletion")
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
	)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	// Use the shared delete renderer, which guards against an empty/204
	// response body before unmarshalling (matching the other 8 delete
	// commands) — a raw json.Unmarshal would crash on an empty 200/204. (#425)
	return cmdutil.RenderDeleteResult(f.IOStreams, resp, fmtOpts,
		fmt.Sprintf("Subscription %s deleted.\n", key),
		func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		})
}
