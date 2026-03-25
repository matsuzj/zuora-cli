// Package audit implements the "zr meter audit" command.
package audit

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type auditOptions struct {
	Factory    *factory.Factory
	ExportType string
	RunType    string
	From       string
	To         string
}

// NewCmdAudit creates the meter audit command.
func NewCmdAudit(f *factory.Factory) *cobra.Command {
	opts := &auditOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "audit <meterId>",
		Short: "Get meter audit trail entries",
		Long: `Get audit trail entries for a usage meter.

All flags (--export-type, --run-type, --from, --to) are required.

Examples:
  zr meter audit 402880e44c... --export-type CSV --run-type FULL --from 2026-01-01 --to 2026-01-31
  zr meter audit 402880e44c... --export-type CSV --run-type FULL --from 2026-01-01 --to 2026-01-31 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.ExportType == "" {
				return fmt.Errorf("--export-type is required")
			}
			if opts.RunType == "" {
				return fmt.Errorf("--run-type is required")
			}
			if opts.From == "" {
				return fmt.Errorf("--from is required")
			}
			if opts.To == "" {
				return fmt.Errorf("--to is required")
			}
			return runAudit(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.ExportType, "export-type", "", "Export type (required)")
	cmd.Flags().StringVar(&opts.RunType, "run-type", "", "Run type (required)")
	cmd.Flags().StringVar(&opts.From, "from", "", "Start date (required)")
	cmd.Flags().StringVar(&opts.To, "to", "", "End date (required)")

	return cmd
}

func runAudit(cmd *cobra.Command, opts *auditOptions, meterID string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/meters/%s/auditTrail/entries", url.PathEscape(meterID))
	resp, err := client.Get(path,
		api.WithCheckSuccess(),
		api.WithQuery("exportType", opts.ExportType),
		api.WithQuery("runType", opts.RunType),
		api.WithQuery("from", opts.From),
		api.WithQuery("to", opts.To),
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
		{Key: "Meter ID", Value: getString(raw, "meterId")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
