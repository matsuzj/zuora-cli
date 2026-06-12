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
		Use:   "summary <meterId>",
		Short: "Get meter run summary",
		Long: `Get a summary of meter runs for a usage meter.

The --run-type flag is required and specifies the type of run to summarize.
An optional --body flag can provide additional filter criteria.`,
		Example: `  zr meter summary 402880e44c... --run-type FULL
  zr meter summary 402880e44c... --run-type FULL --body '{"startDate":"2026-01-01"}'`,
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
			// The response could be a single object or contain nested data.
			// Render as detail output with common summary fields.
			return []output.DetailField{
				{Key: "Meter ID", Value: cmdutil.GetString(raw, "meterId")},
				{Key: "Run Type", Value: cmdutil.GetString(raw, "runType")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
	})
}
