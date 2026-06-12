// Package cancel implements the "zr subscription cancel" command.
package cancel

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

type cancelOptions struct {
	Body          string
	Policy        string
	EffectiveDate string
	Confirm       bool
}

// NewCmdCancel creates the subscription cancel command.
func NewCmdCancel(f *factory.Factory) *cobra.Command {
	opts := &cancelOptions{}

	cmd := &cobra.Command{
		Use:   "cancel <subscription-key>",
		Short: "Cancel a subscription",
		Long: `Cancel a Zuora subscription.

Use --policy and --effective-date flags, or --body for full control.`,
		Example: `  zr subscription cancel A-S001 --policy EndOfCurrentTerm
  zr subscription cancel A-S001 --policy SpecificDate --effective-date 2026-04-01
  zr sub cancel A-S001 --body @cancel.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Cobra's group check passes on EXPLICITLY-EMPTY values
			// (--policy ""); enforce the disjunction on the values too,
			// with cobra's wording (Codex, P5-2).
			if opts.Body == "" && opts.Policy == "" {
				return fmt.Errorf("at least one of the flags in the group [body policy] is required")
			}
			if opts.Body == "" && opts.Policy == "SpecificDate" && opts.EffectiveDate == "" {
				return fmt.Errorf("--effective-date is required when --policy is SpecificDate")
			}
			if err := cmdutil.RequireConfirm(opts.Confirm); err != nil {
				return err
			}
			return runCancel(cmd, f, opts, args[0])
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, false)
	cmd.Flags().StringVar(&opts.Policy, "policy", "", "Cancellation policy (EndOfCurrentTerm, EndOfLastInvoicePeriod, SpecificDate)")
	// body OR policy: cobra enforces the disjunction; the policy-conditional
	// date/period requirements stay handwritten in RunE.
	cmd.MarkFlagsOneRequired("body", "policy")
	cmd.Flags().StringVar(&opts.EffectiveDate, "effective-date", "", "Cancellation date (required for SpecificDate, YYYY-MM-DD)")
	cmdutil.AddConfirmFlag(cmd, &opts.Confirm, "cancellation")

	return cmd
}

func runCancel(cmd *cobra.Command, f *factory.Factory, opts *cancelOptions, key string) error {
	var bodyReader io.Reader
	var err error
	if opts.Body != "" {
		bodyReader, err = cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
		if err != nil {
			return err
		}
	} else {
		payload := map[string]interface{}{
			"cancellationPolicy": opts.Policy,
		}
		if opts.EffectiveDate != "" {
			payload["cancellationEffectiveDate"] = opts.EffectiveDate
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/subscriptions/%s/cancel", url.PathEscape(key)),
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "Subscription ID", Value: cmdutil.GetString(raw, "subscriptionId")},
				{Key: "Success", Value: cmdutil.GetString(raw, "success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			return fmt.Sprintf("Subscription %s cancelled.\n", key)
		},
	})
}
