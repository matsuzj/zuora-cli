// Package audit implements the "zr meter audit" command.
package audit

import (
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
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
		Use:   "audit <meter-id>",
		Short: "Get meter audit trail entries",
		Long: `Get audit trail entries for a usage meter.

All flags (--export-type, --run-type, --from, --to) are required.
Per the API reference, --export-type is SAMPLE or ERROR, --run-type is
DEBUG or NORMAL, and the time bounds are ISO 8601 timestamps. The entries
themselves are returned as an array; use --json to see them in full.`,
		Example: `  zr meter audit 402880e44c... --export-type SAMPLE --run-type NORMAL --from 2026-01-01T00:00:00Z --to 2026-01-31T00:00:00Z
  zr meter audit 402880e44c... --export-type ERROR --run-type NORMAL --from 2026-01-01T00:00:00Z --to 2026-01-31T00:00:00Z --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAudit(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.ExportType, "export-type", "", "Export type: SAMPLE or ERROR (required)")
	cmd.Flags().StringVar(&opts.RunType, "run-type", "", "Run type: DEBUG or NORMAL (required)")
	cmd.Flags().StringVar(&opts.From, "from", "", "Query start time, ISO 8601 (required; sent as queryFromTime)")
	cmd.Flags().StringVar(&opts.To, "to", "", "Query end time, ISO 8601 (required; sent as queryToTime)")
	_ = cmd.MarkFlagRequired("export-type")
	_ = cmd.MarkFlagRequired("run-type")
	_ = cmd.MarkFlagRequired("from")
	_ = cmd.MarkFlagRequired("to")

	return cmd
}

func runAudit(cmd *cobra.Command, opts *auditOptions, meterID string) error {
	f := opts.Factory
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "GET",
		Path:   fmt.Sprintf("/meters/%s/auditTrail/entries", url.PathEscape(meterID)),
		ReqOpts: []api.RequestOption{
			api.WithQuery("exportType", opts.ExportType),
			api.WithQuery("runType", opts.RunType),
			// The API's REQUIRED time-bound parameters are queryFromTime /
			// queryToTime (doc-verified 2026-07-05, #486). The previous "from" /
			// "to" names do not exist on this endpoint, so every call omitted the
			// required parameters.
			api.WithQuery("queryFromTime", opts.From),
			api.WithQuery("queryToTime", opts.To),
		},
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// Real shape per the official API reference (doc-verified 2026-07-05,
			// #486): {success, data:[{errorTime, timestamp, errorCode, errorMessage,
			// payload, …}, …]} — data is an ARRAY of entries; the previous flat
			// meterId key does not exist. The detail view summarizes; --json
			// carries the full entries.
			entries, _ := raw["data"].([]interface{})
			return []output.DetailField{
				{Key: "Entries", Value: fmt.Sprintf("%d", len(entries))},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
	})
}
