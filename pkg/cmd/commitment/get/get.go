// Package get implements the "zr commitment get" command.
package get

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdGet creates the commitment get command.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <commitment-key>",
		Short: "Get commitment details",
		Long: `Get detailed information about a Zuora commitment.

Examples:
  zr commitment get CMT-00000001
  zr commitment get CMT-00000001 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd, f, args[0])
		},
	}
	return cmd
}

func runGet(cmd *cobra.Command, f *factory.Factory, commitmentKey string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/v1/commitments/%s", url.PathEscape(commitmentKey)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// GET /v1/commitments/{key} returns the commitment at the top level keyed by
	// "id" and "commitmentNumber". "commitmentKey" is only the path-parameter name
	// (an id or number), not a response field, so it was always empty.
	fields := []output.DetailField{
		{Key: "ID", Value: cmdutil.GetString(raw, "id")},
		{Key: "Commitment Number", Value: cmdutil.GetString(raw, "commitmentNumber")},
		{Key: "Name", Value: cmdutil.GetString(raw, "name")},
		{Key: "Type", Value: cmdutil.GetString(raw, "type")},
		{Key: "Status", Value: cmdutil.GetString(raw, "status")},
		{Key: "Account Number", Value: cmdutil.GetString(raw, "accountNumber")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}
