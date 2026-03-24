// Package completion implements the "zr completion" command.
package completion

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// NewCmdCompletion creates the completion command.
func NewCmdCompletion(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion <shell>",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for bash, zsh, fish, or powershell.

To load completions:

Bash:
  $ source <(zr completion bash)

Zsh:
  $ source <(zr completion zsh)

Fish:
  $ zr completion fish | source`,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletion(f, cmd, args[0])
		},
	}
	return cmd
}

func runCompletion(f *factory.Factory, cmd *cobra.Command, shell string) error {
	rootCmd := cmd.Root()
	out := f.IOStreams.Out
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletionV2(out, true)
	case "zsh":
		return rootCmd.GenZshCompletion(out)
	case "fish":
		return rootCmd.GenFishCompletion(out, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}
