// Package create implements the "zr usage create" command.
package create

import (
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type createOptions struct {
	Factory *factory.Factory
	Body    string
}

// NewCmdCreate creates the usage create command.
func NewCmdCreate(f *factory.Factory) *cobra.Command {
	opts := &createOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a usage record",
		Long:  `Create a new usage record via the CRUD API.`,
		Example: `  zr usage create --body @usage.json
  zr usage create --body '{"AccountId":"abc","Quantity":10,"StartDateTime":"2026-01-01","UOM":"Each"}'`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.Body == "" {
				return fmt.Errorf("--body is required")
			}
			return runCreate(cmd, opts)
		},
	}

	cmdutil.AddBodyFlag(cmd, &opts.Body, true)

	return cmd
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	f := opts.Factory
	bodyReader, err := cmdutil.ResolveBody(opts.Body, f.IOStreams.In)
	if err != nil {
		return err
	}

	return cmdutil.RunDetail(cmd, f, cmdutil.Action{
		Method: "POST",
		Path:   "/v1/object/usage",
		Body:   bodyReader,
		Fields: func(raw map[string]interface{}) []output.DetailField {
			return []output.DetailField{
				{Key: "ID", Value: cmdutil.GetString(raw, "Id")},
				{Key: "Success", Value: cmdutil.GetString(raw, "Success")},
			}
		},
		SuccessMsg: func(raw map[string]interface{}) string {
			if id := cmdutil.GetString(raw, "Id"); id != "" {
				return fmt.Sprintf("Usage record %s created.\n", id)
			}
			return ""
		},
	})
}
