// Package root implements the root "zr" command.
package root

import (
	"fmt"
	"os"
	"strings"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/internal/config"
	accountcmd "github.com/matsuzj/zuora-cli/pkg/cmd/account"
	aliascmd "github.com/matsuzj/zuora-cli/pkg/cmd/alias"
	apicmd "github.com/matsuzj/zuora-cli/pkg/cmd/api"
	authcmd "github.com/matsuzj/zuora-cli/pkg/cmd/auth"
	billruncmd "github.com/matsuzj/zuora-cli/pkg/cmd/billrun"
	chargecmd "github.com/matsuzj/zuora-cli/pkg/cmd/charge"
	commitmentcmd "github.com/matsuzj/zuora-cli/pkg/cmd/commitment"
	"github.com/matsuzj/zuora-cli/pkg/cmd/completion"
	configcmd "github.com/matsuzj/zuora-cli/pkg/cmd/config"
	contactcmd "github.com/matsuzj/zuora-cli/pkg/cmd/contact"
	creditmemocmd "github.com/matsuzj/zuora-cli/pkg/cmd/creditmemo"
	debitmemocmd "github.com/matsuzj/zuora-cli/pkg/cmd/debitmemo"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	fulfillmentcmd "github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment"
	fulfillmentitemcmd "github.com/matsuzj/zuora-cli/pkg/cmd/fulfillment-item"
	invoicecmd "github.com/matsuzj/zuora-cli/pkg/cmd/invoice"
	metercmd "github.com/matsuzj/zuora-cli/pkg/cmd/meter"
	omnichannelcmd "github.com/matsuzj/zuora-cli/pkg/cmd/omnichannel"
	ordercmd "github.com/matsuzj/zuora-cli/pkg/cmd/order"
	orderactioncmd "github.com/matsuzj/zuora-cli/pkg/cmd/order-action"
	orderlineitemcmd "github.com/matsuzj/zuora-cli/pkg/cmd/order-line-item"
	paymentcmd "github.com/matsuzj/zuora-cli/pkg/cmd/payment"
	plancmd "github.com/matsuzj/zuora-cli/pkg/cmd/plan"
	prepaidcmd "github.com/matsuzj/zuora-cli/pkg/cmd/prepaid"
	productcmd "github.com/matsuzj/zuora-cli/pkg/cmd/product"
	querycmd "github.com/matsuzj/zuora-cli/pkg/cmd/query"
	rampcmd "github.com/matsuzj/zuora-cli/pkg/cmd/ramp"
	rateplancmd "github.com/matsuzj/zuora-cli/pkg/cmd/rateplan"
	"github.com/matsuzj/zuora-cli/pkg/cmd/signup"
	subcmd "github.com/matsuzj/zuora-cli/pkg/cmd/subscription"
	usagecmd "github.com/matsuzj/zuora-cli/pkg/cmd/usage"
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
			// Output-format flags are mutually exclusive. The renderer picks one by
			// a fixed precedence (jq > json > template > csv), so passing two would
			// silently ignore the rest; reject the combination instead.
			jsonFlag, _ := cmd.Flags().GetBool("json")
			jq, _ := cmd.Flags().GetString("jq")
			tmpl, _ := cmd.Flags().GetString("template")
			csvFlag, _ := cmd.Flags().GetBool("csv")
			var set []string
			if jsonFlag {
				set = append(set, "--json")
			}
			if jq != "" {
				set = append(set, "--jq")
			}
			if tmpl != "" {
				set = append(set, "--template")
			}
			if csvFlag {
				set = append(set, "--csv")
			}
			if len(set) > 1 {
				return fmt.Errorf("output format flags are mutually exclusive: %s given together", strings.Join(set, " and "))
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

			zv, _ := cmd.Flags().GetString("zuora-version")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// --read-only flag takes precedence over the ZR_READ_ONLY env var.
			readOnly, _ := cmd.Flags().GetBool("read-only")
			if !cmd.Flags().Changed("read-only") {
				readOnly = envReadOnly()
			}

			// Apply all client overrides (context, version, verbose, read-only)
			// in a single wrapper captured from the original once, so the
			// overrides are not stacked cumulatively across invocations.
			ctx := cmd.Context()
			origHttpClient := f.HttpClient
			f.HttpClient = func() (*api.Client, error) {
				client, err := origHttpClient()
				if err != nil {
					return nil, err
				}
				if ctx != nil {
					client.SetContext(ctx)
				}
				if zv != "" {
					client.SetZuoraVersion(zv)
				}
				if verbose {
					client.SetVerbose(f.IOStreams.ErrOut)
				}
				if readOnly {
					client.SetReadOnly(true)
				}
				return client, nil
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
	cmd.PersistentFlags().Bool("csv", false, "Output as CSV")
	cmd.PersistentFlags().String("zuora-version", "", "Override Zuora API version header")
	cmd.PersistentFlags().Bool("verbose", false, "Enable verbose/debug output")
	cmd.PersistentFlags().Bool("read-only", false, "Block write operations (POST/PUT/DELETE/PATCH)")

	// Subcommands
	cmd.AddCommand(version.NewCmdVersion(f))
	cmd.AddCommand(completion.NewCmdCompletion(f))
	cmd.AddCommand(authcmd.NewCmdAuth(f))
	cmd.AddCommand(configcmd.NewCmdConfig(f))
	cmd.AddCommand(apicmd.NewCmdAPI(f))
	cmd.AddCommand(accountcmd.NewCmdAccount(f))
	cmd.AddCommand(subcmd.NewCmdSubscription(f))
	cmd.AddCommand(contactcmd.NewCmdContact(f))
	cmd.AddCommand(ordercmd.NewCmdOrder(f))
	cmd.AddCommand(orderactioncmd.NewCmdOrderAction(f))
	cmd.AddCommand(orderlineitemcmd.NewCmdOrderLineItem(f))
	cmd.AddCommand(signup.NewCmdSignup(f))
	cmd.AddCommand(productcmd.NewCmdProduct(f))
	cmd.AddCommand(plancmd.NewCmdPlan(f))
	cmd.AddCommand(chargecmd.NewCmdCharge(f))
	cmd.AddCommand(rateplancmd.NewCmdRatePlan(f))
	cmd.AddCommand(invoicecmd.NewCmdInvoice(f))
	cmd.AddCommand(creditmemocmd.NewCmdCreditMemo(f))
	cmd.AddCommand(debitmemocmd.NewCmdDebitMemo(f))
	cmd.AddCommand(billruncmd.NewCmdBillRun(f))
	cmd.AddCommand(paymentcmd.NewCmdPayment(f))
	cmd.AddCommand(usagecmd.NewCmdUsage(f))
	cmd.AddCommand(metercmd.NewCmdMeter(f))
	cmd.AddCommand(rampcmd.NewCmdRamp(f))
	cmd.AddCommand(commitmentcmd.NewCmdCommitment(f))
	cmd.AddCommand(fulfillmentcmd.NewCmdFulfillment(f))
	cmd.AddCommand(fulfillmentitemcmd.NewCmdFulfillmentItem(f))
	cmd.AddCommand(prepaidcmd.NewCmdPrepaid(f))
	cmd.AddCommand(querycmd.NewCmdQuery(f))
	cmd.AddCommand(omnichannelcmd.NewCmdOmnichannel(f))
	cmd.AddCommand(aliascmd.NewCmdAlias(f))

	return cmd
}

// envReadOnly reports whether ZR_READ_ONLY requests read-only mode. It accepts
// the conventional truthy/falsy spellings and, critically, fails safe: a
// non-empty but unrecognized value enables read-only rather than silently
// allowing writes.
func envReadOnly() bool {
	v := strings.TrimSpace(os.Getenv("ZR_READ_ONLY"))
	if v == "" {
		return false
	}
	switch strings.ToLower(v) {
	case "0", "f", "false", "no", "n", "off":
		return false
	default:
		// "1", "t", "true", "yes", "y", "on", and anything unrecognized.
		return true
	}
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
