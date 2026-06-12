// Package resume implements the "zr subscription resume" command.
package resume

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

type resumeOptions struct {
	Body        string
	Policy      string
	ResumeDate  string
	Periods     int
	PeriodsType string
}

// NewCmdResume creates the subscription resume command.
func NewCmdResume(f *factory.Factory) *cobra.Command {
	opts := &resumeOptions{}

	cmd := &cobra.Command{
		Use:   "resume <subscription-key>",
		Short: "Resume a suspended subscription",
		Long:  `Resume a suspended Zuora subscription.`,
		Example: `  zr subscription resume A-S001 --policy FixedPeriodsFromSuspendDate --periods 1 --periods-type Month
  zr subscription resume A-S001 --policy SpecificDate --resume-date 2026-05-01
  zr sub resume A-S001 --body @resume.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				switch opts.Policy {
				case "SpecificDate":
					if opts.ResumeDate == "" {
						return fmt.Errorf("--resume-date is required when --policy is SpecificDate")
					}
				case "FixedPeriodsFromSuspendDate", "FixedPeriodsFromToday":
					if opts.Periods == 0 || opts.PeriodsType == "" {
						return fmt.Errorf("--periods and --periods-type are required when --policy is %s", opts.Policy)
					}
				}
			}
			return runResume(cmd, f, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, false)
	cmd.Flags().StringVar(&opts.Policy, "policy", "", "Resume policy (Today, SpecificDate, FixedPeriodsFromSuspendDate, FixedPeriodsFromToday)")
	// body OR policy: cobra enforces the disjunction; the policy-conditional
	// date/period requirements stay handwritten in RunE.
	cmd.MarkFlagsOneRequired("body", "policy")
	cmd.Flags().StringVar(&opts.ResumeDate, "resume-date", "", "Resume date (for SpecificDate, YYYY-MM-DD)")
	cmd.Flags().IntVar(&opts.Periods, "periods", 0, "Number of periods (for FixedPeriodsFromSuspendDate or FixedPeriodsFromToday)")
	cmd.Flags().StringVar(&opts.PeriodsType, "periods-type", "", "Period type (Day, Week, Month, Year)")

	return cmd
}

func runResume(cmd *cobra.Command, f *factory.Factory, opts *resumeOptions, key string) error {
	var bodyReader io.Reader
	var err error
	if opts.Body != "" {
		bodyReader, err = cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
		if err != nil {
			return err
		}
	} else {
		payload := map[string]interface{}{
			"resumePolicy": opts.Policy,
		}
		if opts.ResumeDate != "" {
			payload["resumeSpecificDate"] = opts.ResumeDate
		}
		if opts.Periods > 0 {
			payload["resumePeriods"] = opts.Periods
			payload["resumePeriodsType"] = opts.PeriodsType
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/subscriptions/%s/resume", url.PathEscape(key)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Subscription %s resumed.\n", key)
		},
	})
}
