// Package summary implements the "zr meter summary" command.
package summary

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/internal/api"
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
An optional --body flag can provide additional filter criteria.

Examples:
  zr meter summary 402880e44c... --run-type FULL
  zr meter summary 402880e44c... --run-type FULL --body '{"startDate":"2026-01-01"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.RunType == "" {
				return fmt.Errorf("--run-type is required")
			}
			return runSummary(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVar(&opts.RunType, "run-type", "", "Run type (required)")
	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")

	return cmd
}

func runSummary(cmd *cobra.Command, opts *summaryOptions, meterID string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

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
	} else {
		bodyMap = make(map[string]interface{})
	}
	bodyMap["runType"] = opts.RunType

	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return fmt.Errorf("encoding body: %w", err)
	}

	path := fmt.Sprintf("/meters/%s/summary", url.PathEscape(meterID))
	resp, err := client.Post(path, strings.NewReader(string(bodyBytes)), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	// The response could be a single object or contain nested data.
	// Render as detail output with common summary fields.
	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Meter ID", Value: getString(raw, "meterId")},
		{Key: "Run Type", Value: getString(raw, "runType")},
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
