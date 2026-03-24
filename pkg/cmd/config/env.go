package config

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdEnv(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "env <name>",
		Short: "Switch the active environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnv(f, args[0])
		},
	}
}

func runEnv(f *factory.Factory, name string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	env, err := cfg.Environment(name)
	if err != nil {
		return err
	}

	if err := cfg.SetActiveEnvironment(name); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "Switched to environment %s (%s)\n", name, env.BaseURL)
	return nil
}
