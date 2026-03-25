// Package cancel implements the "zr subscription cancel" command.
package cancel

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type cancelOptions struct {
	Body          string
	Policy        string
	EffectiveDate string
}

// NewCmdCancel creates the subscription cancel command.
func NewCmdCancel(f *factory.Factory) *cobra.Command {
	opts := &cancelOptions{}

	cmd := &cobra.Command{
		Use:   "cancel <subscription-key>",
		Short: "Cancel a subscription",
		Long: `Cancel a Zuora subscription.

Use --policy and --effective-date flags, or --body for full control.

Examples:
  zr subscription cancel A-S001 --policy EndOfCurrentTerm
  zr subscription cancel A-S001 --policy SpecificDate --effective-date 2026-04-01
  zr sub cancel A-S001 --body @cancel.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" && opts.Policy == "" {
				return fmt.Errorf("--policy or --body is required")
			}
			if opts.Body == "" && opts.Policy == "SpecificDate" && opts.EffectiveDate == "" {
				return fmt.Errorf("--effective-date is required when --policy is SpecificDate")
			}
			return runCancel(cmd, f, opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Request body (JSON string, @file, or - for stdin)")
	cmd.Flags().StringVar(&opts.Policy, "policy", "", "Cancellation policy (EndOfCurrentTerm, EndOfLastInvoicePeriod, SpecificDate)")
	cmd.Flags().StringVar(&opts.EffectiveDate, "effective-date", "", "Cancellation date (required for SpecificDate, YYYY-MM-DD)")

	return cmd
}

func runCancel(cmd *cobra.Command, f *factory.Factory, opts *cancelOptions, key string) error {
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	var bodyReader io.Reader
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
		bodyReader = strings.NewReader(string(data))
	}

	resp, err := client.Put(fmt.Sprintf("/v1/subscriptions/%s/cancel", url.PathEscape(key)), bodyReader, api.WithCheckSuccess())
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	fields := []output.DetailField{
		{Key: "Subscription ID", Value: getString(raw, "subscriptionId")},
		{Key: "Success", Value: getString(raw, "success")},
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Subscription %s cancelled.\n", key)
	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}
