package config

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdSet(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use: "set <key> <value>",
		// Complete the <key> argument with the known config keys (P5-3b).
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{"active_environment", "zuora_version", "default_output"}, cobra.ShellCompDirectiveNoFileComp
		},
		Short: "Set a configuration value",
		Long: `Set a configuration value.

Valid keys:
  active_environment  The active Zuora environment
  zuora_version       Default Zuora API version header
  default_output      Default output format (table, json)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(cmd, f, args[0], args[1])
		},
	}
}

func runSet(cmd *cobra.Command, f *factory.Factory, key, value string) error {
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	switch key {
	case "active_environment":
		if err := cfg.SetActiveEnvironment(value); err != nil {
			return err
		}
	case "zuora_version":
		if err := cfg.SetZuoraVersion(value); err != nil {
			return err
		}
	case "default_output":
		if err := cfg.SetDefaultOutput(value); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	// Machine flags get {"success": true}; the human message goes to stderr,
	// keeping stdout clean (#453/#519).
	return output.RenderSuccess(f.IOStreams, fmtOpts, fmt.Sprintf("Set %s to %s\n", key, value))
}
