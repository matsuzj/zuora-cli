// Package root implements the root "zr" command.
package root

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/config"
	accountcmd "github.com/matsuzj/zuora-cli/pkg/cmd/account"
	apicmd "github.com/matsuzj/zuora-cli/pkg/cmd/api"
	authcmd "github.com/matsuzj/zuora-cli/pkg/cmd/auth"
	"github.com/matsuzj/zuora-cli/pkg/cmd/completion"
	configcmd "github.com/matsuzj/zuora-cli/pkg/cmd/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	subcmd "github.com/matsuzj/zuora-cli/pkg/cmd/subscription"
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
			// --json + --template mutual exclusion
			jsonFlag, _ := cmd.Flags().GetBool("json")
			tmpl, _ := cmd.Flags().GetString("template")
			if jsonFlag && tmpl != "" {
				return fmt.Errorf("cannot use --json and --template together")
			}

			// --env override (transient, does not persist to config.yml)
			if envName, _ := cmd.Flags().GetString("env"); envName != "" {
				origConfig := f.Config
				f.Config = func() (config.Config, error) {
					cfg, err := origConfig()
					if err != nil {
						return nil, err
					}
					// Validate environment exists
					if _, err := cfg.Environment(envName); err != nil {
						return nil, fmt.Errorf("invalid environment %q: %w", envName, err)
					}
					return &envOverrideConfig{Config: cfg, env: envName}, nil
				}
			}

			// --zuora-version override
			if zv, _ := cmd.Flags().GetString("zuora-version"); zv != "" {
				origHttpClient := f.HttpClient
				f.HttpClient = func() (*api.Client, error) {
					client, err := origHttpClient()
					if err != nil {
						return nil, err
					}
					client.SetZuoraVersion(zv)
					return client, nil
				}
			}

			// --verbose override
			if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
				origHttpClient := f.HttpClient
				f.HttpClient = func() (*api.Client, error) {
					client, err := origHttpClient()
					if err != nil {
						return nil, err
					}
					client.SetVerbose(f.IOStreams.ErrOut)
					return client, nil
				}
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
	cmd.AddCommand(authcmd.NewCmdAuth(f))
	cmd.AddCommand(configcmd.NewCmdConfig(f))
	cmd.AddCommand(apicmd.NewCmdAPI(f))
	cmd.AddCommand(accountcmd.NewCmdAccount(f))
	cmd.AddCommand(subcmd.NewCmdSubscription(f))

	return cmd
}

// envOverrideConfig wraps a Config to override ActiveEnvironment() without mutating the original.
type envOverrideConfig struct {
	config.Config
	env string
}

func (c *envOverrideConfig) ActiveEnvironment() string { return c.env }

// SetActiveEnvironment delegates to the underlying config so that
// explicit "config set active_environment" still persists.
func (c *envOverrideConfig) SetActiveEnvironment(name string) error {
	return c.Config.SetActiveEnvironment(name)
}
