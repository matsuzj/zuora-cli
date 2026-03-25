// Package run implements the "zr meter run" command.
package run

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdRun creates the meter run command.
func NewCmdRun(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run <meterId> <version>",
		Short: "Run a usage meter",
		Long: `Run a usage meter by meter ID and version.

Examples:
  zr meter run 402880e44c...  1
  zr meter run 402880e44c...  1 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMeter(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runMeter(cmd *cobra.Command, f *factory.Factory, meterID, version string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/meters/run/%s/%s", url.PathEscape(meterID), url.PathEscape(version))
	resp, err := client.Post(path, nil, api.WithCheckSuccess())
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
		{Key: "Message", Value: getString(raw, "message")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Meter run started.\n")
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
