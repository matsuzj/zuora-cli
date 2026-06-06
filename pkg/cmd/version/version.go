// Package version implements the "zr version" command.
package version

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/build"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdVersion creates the version command.
func NewCmdVersion(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of zr",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(cmd, f)
		},
	}
	return cmd
}

func runVersion(cmd *cobra.Command, f *factory.Factory) error {
	fmtOpts := output.FromCmd(cmd)

	if fmtOpts.JSON || fmtOpts.JQ != "" || fmtOpts.Template != "" {
		data := map[string]string{
			"version": build.Version,
			"commit":  build.Commit,
			"date":    build.Date,
		}
		rawJSON, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshaling version: %w", err)
		}
		fields := []output.DetailField{
			{Key: "Version", Value: build.Version},
			{Key: "Commit", Value: build.Commit},
			{Key: "Date", Value: build.Date},
		}
		return output.RenderDetail(f.IOStreams, rawJSON, fmtOpts, fields)
	}

	out := build.Version
	if build.Commit != "" {
		out += fmt.Sprintf(" (commit: %s)", build.Commit)
	}
	if build.Date != "" {
		out += fmt.Sprintf(" (built: %s)", build.Date)
	}
	fmt.Fprintf(f.IOStreams.Out, "zr version %s\n", out)
	return nil
}
