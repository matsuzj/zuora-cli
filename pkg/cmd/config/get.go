package config

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdGet(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use: "get <key>",
		// Complete the <key> argument with the known config keys (P5-3b).
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{"active_environment", "zuora_version", "default_output"}, cobra.ShellCompDirectiveNoFileComp
		},
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(f, args[0])
		},
	}
}

func runGet(f *factory.Factory, key string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	var value string
	switch key {
	case "active_environment":
		value = cfg.ActiveEnvironment()
	case "zuora_version":
		value = cfg.ZuoraVersion()
	case "default_output":
		value = cfg.DefaultOutput()
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}

	fmt.Fprintln(f.IOStreams.Out, value)
	return nil
}
