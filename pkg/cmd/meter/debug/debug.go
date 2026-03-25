// Package debug implements the "zr meter debug" command.
package debug

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdDebug creates the meter debug command.
func NewCmdDebug(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug <meterId> <version>",
		Short: "Debug a usage meter",
		Long: `Debug a usage meter by meter ID and version.

Examples:
  zr meter debug 402880e44c... 1
  zr meter debug 402880e44c... 1 --json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDebug(cmd, f, args[0], args[1])
		},
	}
	return cmd
}

func runDebug(cmd *cobra.Command, f *factory.Factory, meterID, version string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/meters/debug/%s/%s", url.PathEscape(meterID), url.PathEscape(version))
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

	fmt.Fprintf(f.IOStreams.ErrOut, "Meter debug started.\n")
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
