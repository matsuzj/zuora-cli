package alias

import (
	"encoding/json"
	"fmt"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all aliases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, f)
		},
	}
	return cmd
}

func runList(cmd *cobra.Command, f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	store := NewStore(cfg.ConfigDir())
	if err := store.Load(); err != nil {
		return err
	}

	entries := store.All()

	// Machine-format flags (--json/--jq/--template/--csv) must be honored, not
	// silently ignored (#453). Render as a table through the shared pipeline;
	// an empty alias set becomes an empty JSON array / bare header, the correct
	// machine answer. The human default keeps the tab-separated form and the
	// "No aliases configured." notice.
	fmtOpts := output.FromCmd(cmd)
	if fmtOpts.JSON || fmtOpts.JQ != "" || fmtOpts.Template != "" || fmtOpts.CSV {
		type aliasJSON struct {
			Name    string `json:"name"`
			Command string `json:"command"`
		}
		list := make([]aliasJSON, 0, len(entries))
		rows := make([][]string, 0, len(entries))
		for _, e := range entries {
			list = append(list, aliasJSON(e))
			rows = append(rows, []string{e.Name, e.Command})
		}
		rawJSON, err := json.Marshal(list)
		if err != nil {
			return fmt.Errorf("marshaling aliases: %w", err)
		}
		cols := []output.Column{{Header: "NAME"}, {Header: "COMMAND"}}
		return output.Render(f.IOStreams, rawJSON, fmtOpts, rows, cols)
	}

	if len(entries) == 0 {
		fmt.Fprintln(f.IOStreams.ErrOut, "No aliases configured.")
		return nil
	}

	out := f.IOStreams.Out
	for _, e := range entries {
		fmt.Fprintf(out, "%s\t%s\n", e.Name, e.Command)
	}
	return nil
}
