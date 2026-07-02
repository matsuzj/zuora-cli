package config

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

func newCmdList(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configuration values",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, f)
		},
	}
}

func runList(cmd *cobra.Command, f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	active := cfg.ActiveEnvironment()
	envs := cfg.Environments()
	names := make([]string, 0, len(envs))
	for name := range envs {
		names = append(names, name)
	}
	sort.Strings(names)

	// Machine-format flags (--json/--jq/--template/--csv) must be honored, not
	// silently ignored (#453). The default keeps the human layout; the
	// format-flag paths route through the shared renderer with a synthesized
	// body (version.go pattern). environments is nested in JSON and flattened
	// into one detail row per environment for --template/--csv.
	fmtOpts := output.FromCmd(cmd)
	if fmtOpts.JSON || fmtOpts.JQ != "" || fmtOpts.Template != "" || fmtOpts.CSV {
		envData := make(map[string]interface{}, len(names))
		fields := []output.DetailField{
			{Key: "active_environment", Value: active},
			{Key: "zuora_version", Value: cfg.ZuoraVersion()},
			{Key: "default_output", Value: cfg.DefaultOutput()},
		}
		for _, name := range names {
			envData[name] = map[string]interface{}{
				"baseUrl": envs[name].BaseURL,
				"active":  name == active,
			}
			value := envs[name].BaseURL
			if name == active {
				value += " (active)"
			}
			fields = append(fields, output.DetailField{Key: "environment." + name, Value: value})
		}
		data := map[string]interface{}{
			"active_environment": active,
			"zuora_version":      cfg.ZuoraVersion(),
			"default_output":     cfg.DefaultOutput(),
			"environments":       envData,
		}
		rawJSON, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshaling config: %w", err)
		}
		return output.RenderDetail(f.IOStreams, rawJSON, fmtOpts, fields)
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "active_environment: %s\n", active)
	fmt.Fprintf(out, "zuora_version: %s\n", cfg.ZuoraVersion())
	fmt.Fprintf(out, "default_output: %s\n", cfg.DefaultOutput())
	fmt.Fprintln(out)

	fmt.Fprintln(out, "environments:")
	for _, name := range names {
		marker := " "
		if name == active {
			marker = "*"
		}
		fmt.Fprintf(out, "  %s %s (%s)\n", marker, name, envs[name].BaseURL)
	}

	return nil
}
