package config

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdSet(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value.

Valid keys:
  active_environment  The active Zuora environment
  zuora_version       Default Zuora API version header
  default_output      Default output format (table, json)`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSet(f, args[0], args[1])
		},
	}
}

func runSet(f *factory.Factory, key, value string) error {
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

	fmt.Fprintf(f.IOStreams.Out, "Set %s to %s\n", key, value)
	return nil
}
