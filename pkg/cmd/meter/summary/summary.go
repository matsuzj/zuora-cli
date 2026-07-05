// Package summary implements the "zr meter summary" command.
package summary

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type summaryOptions struct {
	Factory *factory.Factory
	RunType string
	Body    string
}

// NewCmdSummary creates the meter summary command.
func NewCmdSummary(f *factory.Factory) *cobra.Command {
	opts := &summaryOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "summary <meter-id>",
		Short: "Get meter run summary",
		Long: `Get a summary of meter runs for a usage meter.

The --run-type flag is required and specifies the type of run to summarize.
An optional --body flag can provide additional filter criteria.`,
		Example: `  zr meter summary 402880e44c... --run-type NORMAL
  zr meter summary 402880e44c... --run-type NORMAL --body '{"startDate":"2026-01-01"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSummary(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.RunType, "run-type", "", "Run type (required)")
	_ = cmd.MarkFlagRequired("run-type")
	cmdutil.AddBodyFlag(cmd, &opts.Body, false)

	return cmd
}

func runSummary(cmd *cobra.Command, opts *summaryOptions, meterID string) error {
	f := opts.Factory
	// Build the JSON body: start with --body if provided, otherwise empty object,
	// then merge in --run-type.
	var bodyMap map[string]interface{}
	if opts.Body != "" {
		bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
		if err != nil {
			return err
		}
		dec := json.NewDecoder(bodyReader)
		if err := dec.Decode(&bodyMap); err != nil {
			return fmt.Errorf("parsing body: %w", err)
		}
		if bodyMap == nil {
			bodyMap = make(map[string]interface{})
		}
	} else {
		bodyMap = make(map[string]interface{})
	}
	bodyMap["runType"] = opts.RunType

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("encoding body: %w", err)
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   fmt.Sprintf("/meters/%s/summary", url.PathEscape(meterID)),
		Body:   bytes.NewReader(bodyBytes),
		Fields: func(raw map[string]interface{}) []output.DetailField {
			// Real shape per the official API reference (doc-verified 2026-07-05,
			// #486; this sandbox cannot probe mediation endpoints): the envelope is
			// {success, data:{requestId, requestTime, query:{runType,…}, output:[…]}}.
			// The previous flat meterId/runType keys do not exist; runType lives
			// nested under data.query. Full output groups are available via --json.
			data, _ := raw["data"].(map[string]interface{})
			query, _ := data["query"].(map[string]interface{})
			groups, _ := data["output"].([]interface{})
			return []output.DetailField{
				{Key: "Request ID", Value: cmdutil.GetString(data, "requestId")},
				{Key: "Request Time", Value: cmdutil.GetString(data, "requestTime")},
				{Key: "Run Type", Value: cmdutil.GetString(query, "runType")},
				{Key: "Output Groups", Value: fmt.Sprintf("%d", len(groups))},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
	})
}
