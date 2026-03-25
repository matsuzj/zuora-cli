// Package status implements the "zr meter status" command.
package status

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdStatus creates the meter status command.
func NewCmdStatus(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <meterId> <version>",
		Short: "Get meter run status",
		Long: `Get the run status of a usage meter by meter ID and version.

Examples:
  zr meter status 402880e44c... 1
  zr meter status 402880e44c... 1 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runStatus(cmd *cobra.Command, f *factory.Factory, meterID, version string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/meters/%s/%s/runStatus", url.PathEscape(meterID), url.PathEscape(version))
	resp, err := client.Get(path, api.WithCheckSuccess())
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
		{Key: "Version", Value: getString(raw, "version")},
		{Key: "Status", Value: getString(raw, "status")},
		{Key: "Run Type", Value: getString(raw, "runType")},
		{Key: "Start Time", Value: getString(raw, "startTime")},
		{Key: "End Time", Value: getString(raw, "endTime")},
	}

	return output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
