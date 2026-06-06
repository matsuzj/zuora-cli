// Package writeoff implements the "zr invoice writeoff" command.
package writeoff

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type writeoffOptions struct {
	Factory *factory.Factory
	Body    string
	Confirm bool
}

// NewCmdWriteoff creates the invoice writeoff command.
func NewCmdWriteoff(f *factory.Factory) *cobra.Command {
	opts := &writeoffOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "writeoff <invoice-id>",
		Short: "Write off a posted invoice balance",
		Long: `Write off the balance of a posted Zuora invoice, creating a credit memo.

This action is irreversible. Use --confirm to proceed. An optional --body may
carry write-off details such as memoDate, comment, and reasonCode.

Examples:
  zr invoice writeoff 2c92c0f8... --confirm
  zr invoice writeoff 2c92c0f8... --confirm --body '{"comment":"bad debt","reasonCode":"Write-off"}'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.RequireConfirm(opts.Confirm); err != nil {
				return err
			}
			return runWriteoff(cmd, opts, args[0])
		},
	}

	cmd.Flags().StringVarP(&opts.Body, "body", "b", "", "Optional write-off details (JSON string, @file, or - for stdin)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Confirm the write-off")
	return cmd
}

func runWriteoff(cmd *cobra.Command, opts *writeoffOptions, invoiceID string) error {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// The write-off body is optional; only resolve a reader when one was given.
	var reqOpts []api.RequestOption
	reqOpts = append(reqOpts, api.WithCheckSuccess())
	if opts.Body != "" {
		bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
		if err != nil {
			return err
		}
		reqOpts = append(reqOpts, api.WithBody(bodyReader))
	}

	resp, err := client.Do("PUT", fmt.Sprintf("/v1/invoices/%s/write-off", url.PathEscape(invoiceID)), reqOpts...)
	if err != nil {
		return err
	}

	fmtOpts := output.FromCmd(cmd)

	var raw map[string]interface{}
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	// A successful write-off returns the generated credit memo, nested under
	// "creditMemo". Surface its id when present; the full object is in --json.
	fields := []output.DetailField{
		{Key: "Success", Value: cmdutil.GetString(raw, "success")},
	}
	if cm, ok := raw["creditMemo"].(map[string]interface{}); ok {
		fields = append(fields,
			output.DetailField{Key: "Credit Memo ID", Value: cmdutil.GetString(cm, "id")},
			output.DetailField{Key: "Credit Memo Number", Value: cmdutil.GetString(cm, "number")},
		)
	}

	if err := output.RenderDetail(f.IOStreams, resp.Body, fmtOpts, fields); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.ErrOut, "Invoice %s written off.\n", invoiceID)
	return nil
}
