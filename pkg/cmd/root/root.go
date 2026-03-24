// Package root implements the root "zr" command.
package root

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/completion"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmd/version"
	"github.com/spf13/cobra"
)

// NewCmdRoot creates the root command for the CLI.
func NewCmdRoot(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "zr <command> <subcommand> [flags]",
		Short:         "Zuora CLI",
		Long:          "Work with Zuora from the command line.",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			jsonFlag, _ := cmd.Flags().GetBool("json")
			tmpl, _ := cmd.Flags().GetString("template")
			if jsonFlag && tmpl != "" {
				return fmt.Errorf("cannot use --json and --template together")
			}
			return nil
		},
	}

	// NOTE: Do NOT call cmd.SetOut()/cmd.SetErr() here.
	// Cobra has a known bug (https://github.com/spf13/cobra/issues/1708)
	// where SetOut causes some error messages to go to stdout instead of stderr.
	// Commands should write to f.IOStreams.Out/ErrOut directly instead.

	// Global flags
	cmd.PersistentFlags().StringP("env", "e", "", "Environment name")
	cmd.PersistentFlags().Bool("json", false, "Output as JSON")
	cmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression")
	cmd.PersistentFlags().String("template", "", "Format output with a Go template")
	cmd.PersistentFlags().String("zuora-version", "", "Override Zuora API version header")
	cmd.PersistentFlags().Bool("verbose", false, "Enable verbose/debug output")

	// Subcommands
	cmd.AddCommand(version.NewCmdVersion(f))
	cmd.AddCommand(completion.NewCmdCompletion(f))

	return cmd
}
