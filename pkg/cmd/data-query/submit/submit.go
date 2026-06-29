// Package submit implements the "zr data-query submit" command.
package submit

import (
	"bytes"
	"fmt"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/data-query/dqutil"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// NewCmdSubmit creates the data-query submit command.
func NewCmdSubmit(f *factory.Factory) *cobra.Command {
	sf := &dqutil.SubmitFlags{}
	cmd := &cobra.Command{
		Use:   `submit ["<SQL>"]`,
		Short: "Submit a Data Query job (async)",
		Long: `Submit a SQL query as an asynchronous Data Query job and print its id and status.

Provide the SQL as an argument or via --file (exactly one). Track it with
"zr data-query get <id>", or use "zr data-query run" to submit, poll, and
download in one step.`,
		Example: `  zr data-query submit "SELECT accountnumber, balance FROM account WHERE balance > 100"
  zr data-query submit --file query.sql --output-format CSV`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSubmit(cmd, f, sf, args)
		},
	}
	dqutil.AddSubmitFlags(cmd.Flags(), sf)
	dqutil.RegisterSubmitCompletions(cmd)
	return cmd
}

func runSubmit(cmd *cobra.Command, f *factory.Factory, sf *dqutil.SubmitFlags, args []string) error {
	sql, err := dqutil.ResolveSQL(args, sf.File)
	if err != nil {
		return err
	}
	body, err := dqutil.BuildSubmitBody(sql, sf)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	var reqOpts []api.RequestOption
	if sf.IdempotencyKey != "" {
		reqOpts = append(reqOpts, api.WithHeader("Idempotency-Key", sf.IdempotencyKey))
	}
	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method:  "POST",
		Path:    dqutil.SubmitPath,
		Body:    bytes.NewReader(body),
		ReqOpts: reqOpts,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return dqutil.DetailFields(dqutil.UnwrapData(raw))
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if id := cmdutil.GetString(dqutil.UnwrapData(raw), "id"); id != "" {
				return fmt.Sprintf("Data Query job %s submitted.\n", id)
			}
			return ""
		},
	})
}
