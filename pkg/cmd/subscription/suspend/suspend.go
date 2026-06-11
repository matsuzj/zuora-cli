// Package suspend implements the "zr subscription suspend" command.
package suspend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type suspendOptions struct {
	Body        string
	Policy      string
	SuspendDate string
	Periods     int
	PeriodsType string
}

// NewCmdSuspend creates the subscription suspend command.
func NewCmdSuspend(f *factory.Factory) *cobra.Command {
	opts := &suspendOptions{}

	cmd := &cobra.Command{
		Use:   "suspend <subscription-key>",
		Short: "Suspend a subscription",
		Long: `Suspend a Zuora subscription.

Examples:
  zr subscription suspend A-S001 --policy FixedPeriodsFromToday --periods 3 --periods-type Month
  zr subscription suspend A-S001 --policy SpecificDate --suspend-date 2026-04-01
  zr sub suspend A-S001 --body @suspend.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" && opts.Policy == "" {
				return fmt.Errorf("--policy or --body is required")
			}
			if opts.Body == "" {
				switch opts.Policy {
				case "SpecificDate":
					if opts.SuspendDate == "" {
						return fmt.Errorf("--suspend-date is required when --policy is SpecificDate")
					}
				case "FixedPeriodsFromToday":
					if opts.Periods == 0 || opts.PeriodsType == "" {
						return fmt.Errorf("--periods and --periods-type are required when --policy is FixedPeriodsFromToday")
					}
				}
			}
			return runSuspend(cmd, f, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)
	cmd.Flags().StringVar(&opts.Policy, "policy", "", "Suspend policy (Today, EndOfLastInvoicePeriod, SpecificDate, FixedPeriodsFromToday)")
	cmd.Flags().StringVar(&opts.SuspendDate, "suspend-date", "", "Suspend date (for SpecificDate, YYYY-MM-DD)")
	cmd.Flags().IntVar(&opts.Periods, "periods", 0, "Number of periods (for FixedPeriodsFromToday)")
	cmd.Flags().StringVar(&opts.PeriodsType, "periods-type", "", "Period type (Day, Week, Month, Year)")

	return cmd
}

func runSuspend(cmd *cobra.Command, f *factory.Factory, opts *suspendOptions, key string) error {
	var bodyReader io.Reader
	var err error
	if opts.Body != "" {
		bodyReader, err = cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
		if err != nil {
			return err
		}
	} else {
		payload := map[string]interface{}{
			"suspendPolicy": opts.Policy,
		}
		if opts.SuspendDate != "" {
			payload["suspendSpecificDate"] = opts.SuspendDate
		}
		if opts.Periods > 0 {
			payload["suspendPeriods"] = opts.Periods
			payload["suspendPeriodsType"] = opts.PeriodsType
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/subscriptions/%s/suspend", url.PathEscape(key)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Subscription %s suspended.\n", key)
		},
	})
}
