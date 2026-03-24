package config

import (
	"fmt"
	"sort"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/spf13/cobra"
)

func newCmdList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(f)
		},
	}
}

func runList(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "active_environment: %s\n", cfg.ActiveEnvironment())
	fmt.Fprintf(out, "zuora_version: %s\n", cfg.ZuoraVersion())
	fmt.Fprintf(out, "default_output: %s\n", cfg.DefaultOutput())
	fmt.Fprintln(out)

	fmt.Fprintln(out, "environments:")
	envs := cfg.Environments()
	names := make([]string, 0, len(envs))
	for name := range envs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		marker := " "
		if name == cfg.ActiveEnvironment() {
			marker = "*"
		}
		fmt.Fprintf(out, "  %s %s (%s)\n", marker, name, envs[name].BaseURL)
	}

	return nil
}
