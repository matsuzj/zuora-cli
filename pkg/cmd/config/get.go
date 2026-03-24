package config

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdGet(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
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
