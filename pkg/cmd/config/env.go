package config

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdEnv(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "env <name>",
		Short: "Switch the active environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnv(cmd, f, args[0])
		},
	}
}

func runEnv(cmd *cobra.Command, f *factory.Factory, name string) error {
	fmtOpts := output.FromCmd(cmd)
	if err := output.RejectBareCSV(fmtOpts); err != nil {
		return err
	}

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

	// Machine flags get {"success": true}; the human message goes to stderr,
	// keeping stdout clean (#453/#519).
	return output.RenderSuccess(f.IOStreams, fmtOpts,
		fmt.Sprintf("Switched to environment %s (%s)\n", name, env.BaseURL))
}
