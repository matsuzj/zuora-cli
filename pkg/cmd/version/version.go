// Package version implements the "zr version" command.
package version

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/build"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdVersion creates the version command.
func NewCmdVersion(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of zr",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion(f)
		},
	}
	return cmd
}

func runVersion(f *factory.Factory) error {
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
