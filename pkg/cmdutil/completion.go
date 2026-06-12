// Completion helpers (P5-3b): dynamic shell completion for enum-valued flags
// and environment names. Suggestions are offered, never enforced — validation
// stays with the server or the command's own checks.
package cmdutil

import (
	"sort"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

// EnumCompletion returns a flag-completion function offering the given fixed
// values (file completion disabled).
func EnumCompletion(values ...string) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}

// EnvNamesCompletion completes --env with the configured environment names,
// sorted. Configuration errors degrade to "no suggestions" — completion must
// never break on a broken config.
func EnvNamesCompletion(f *factory.Factory) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		cfg, err := f.Config()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		envs := cfg.Environments()
		names := make([]string, 0, len(envs))
		for name := range envs {
			names = append(names, name)
		}
		sort.Strings(names)
		return names, cobra.ShellCompDirectiveNoFileComp
	}
}
