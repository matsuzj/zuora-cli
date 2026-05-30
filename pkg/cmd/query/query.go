// Package query implements the "zr query" command for ZOQL queries.
package query

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/matsuzj/zuora-cli/internal/api"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

type queryOptions struct {
	Factory *factory.Factory
	CSV     bool
	Export  string
	Limit   int
}

// NewCmdQuery creates the query command.
func NewCmdQuery(f *factory.Factory) *cobra.Command {
	opts := &queryOptions{Factory: f}

	cmd := &cobra.Command{
		Use:   `query "<ZOQL>"`,
		Short: "Execute a ZOQL query",
		Long: `Execute a ZOQL query against the Zuora API.

Automatically paginates through all results using queryMore.

Examples:
  zr query "SELECT Id, Name FROM Account"
  zr query "SELECT Id, Name FROM Account" --limit 10
  zr query "SELECT Id, Name FROM Account" --csv
  zr query "SELECT Id, Name FROM Account" --export results.csv --csv
  zr query "SELECT Id, Name FROM Account" --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runQuery(cmd, opts, args[0])
		},
	}

	cmd.Flags().BoolVar(&opts.CSV, "csv", false, "Output as CSV")
	cmd.Flags().StringVar(&opts.Export, "export", "", "Export results to file")
	cmd.Flags().IntVar(&opts.Limit, "limit", 0, "Maximum number of rows (0 = all)")

	return cmd
}

func runQuery(cmd *cobra.Command, opts *queryOptions, zoql string) (retErr error) {
	f := opts.Factory
	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// First query
	body, err := json.Marshal(map[string]string{"queryString": zoql})
	if err != nil {
		return fmt.Errorf("encoding query: %w", err)
	}

	resp, err := client.Post("/v1/action/query", bytes.NewReader(body), api.WithCheckSuccess())
	if err != nil {
		return err
	}

	var result queryResult
	if err := decodeQueryResult(resp.Body, &result); err != nil {
		return fmt.Errorf("parsing query response: %w", err)
	}

	allRecords := result.Records

	// Pagination: queryMore
	for !result.Done && result.QueryLocator != "" {
		if opts.Limit > 0 && len(allRecords) >= opts.Limit {
			break
		}
		moreBody, err := json.Marshal(map[string]string{"queryLocator": result.QueryLocator})
		if err != nil {
			return fmt.Errorf("encoding queryMore: %w", err)
		}
		resp, err = client.Post("/v1/action/queryMore", bytes.NewReader(moreBody), api.WithCheckSuccess())
		if err != nil {
			return err
		}
		result = queryResult{}
		if err := decodeQueryResult(resp.Body, &result); err != nil {
			return fmt.Errorf("parsing queryMore response: %w", err)
		}
		allRecords = append(allRecords, result.Records...)
	}

	// Apply limit — track whether we actually trimmed rows
	limitTrimmed := false
	if opts.Limit > 0 && len(allRecords) > opts.Limit {
		allRecords = allRecords[:opts.Limit]
		limitTrimmed = true
	}

	// Determine output destination. For --export, write to a temp file in the
	// target's directory and rename on success, so a mid-stream error leaves any
	// pre-existing target file intact (atomic, all-or-nothing export).
	var outWriter io.Writer = f.IOStreams.Out
	if opts.Export != "" {
		tmp, err := os.CreateTemp(filepath.Dir(opts.Export), ".zr-export-*")
		if err != nil {
			return fmt.Errorf("creating export file: %w", err)
		}
		tmpName := tmp.Name()
		outWriter = tmp
		defer func() {
			tmp.Close()
			if retErr == nil {
				retErr = os.Rename(tmpName, opts.Export)
			}
			if retErr != nil {
				os.Remove(tmpName)
			}
		}()
	}

	// Build an IOStreams pointing to the export destination (or original stdout)
	ios := f.IOStreams
	if opts.Export != "" {
		ios = &iostreams.IOStreams{
			In:     f.IOStreams.In,
			Out:    outWriter,
			ErrOut: f.IOStreams.ErrOut,
		}
	}

	// Format output based on flags
	fmtOpts := output.FromCmd(cmd)

	// Build combined JSON for --json/--jq/--template
	// done reflects whether the full result set is present (API complete AND no CLI truncation)
	isDone := result.Done && !limitTrimmed
	combinedMap := map[string]interface{}{
		"records": allRecords,
		"size":    len(allRecords),
		"done":    isDone,
	}
	if result.QueryLocator != "" && !limitTrimmed {
		combinedMap["queryLocator"] = result.QueryLocator
	}
	combined, err := json.Marshal(combinedMap)
	if err != nil {
		return fmt.Errorf("encoding combined results: %w", err)
	}

	if fmtOpts.JQ != "" {
		return output.PrintJSON(ios, combined, fmtOpts.JQ)
	}
	if fmtOpts.JSON {
		return output.PrintJSON(ios, combined, "")
	}
	if fmtOpts.Template != "" {
		return output.PrintTemplate(ios, combined, fmtOpts.Template)
	}

	// Extract column headers from records
	columns := extractColumns(allRecords)
	rows := buildRows(allRecords, columns)

	cols := make([]output.Column, len(columns))
	for i, c := range columns {
		cols[i] = output.Column{Header: c, Field: c}
	}

	if opts.CSV {
		return output.PrintCSV(outWriter, rows, cols)
	}

	// Default: table output
	return output.Render(ios, combined, fmtOpts, rows, cols)
}

type queryResult struct {
	Records      []map[string]interface{} `json:"records"`
	Size         int                      `json:"size"`
	Done         bool                     `json:"done"`
	QueryLocator string                   `json:"queryLocator"`
}

// decodeQueryResult decodes a ZOQL response using json.Number for all numbers,
// so large IDs and high-precision amounts in records are not rounded by float64.
func decodeQueryResult(body []byte, v *queryResult) error {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	return dec.Decode(v)
}

// extractColumns returns sorted column names from all records (union of all keys).
func extractColumns(records []map[string]interface{}) []string {
	if len(records) == 0 {
		return nil
	}
	seen := make(map[string]bool)
	for _, rec := range records {
		for k := range rec {
			seen[k] = true
		}
	}
	cols := make([]string, 0, len(seen))
	for k := range seen {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols
}

// buildRows converts records to a 2D string slice based on column order.
func buildRows(records []map[string]interface{}, columns []string) [][]string {
	rows := make([][]string, len(records))
	for i, rec := range records {
		row := make([]string, len(columns))
		for j, col := range columns {
			if v, ok := rec[col]; ok && v != nil {
				row[j] = formatCell(v)
			}
		}
		rows[i] = row
	}
	return rows
}

// formatCell renders a record value for a table/CSV cell. Scalars use their
// natural string form; nested objects/arrays are JSON-encoded so they read as
// JSON rather than Go's map[...] / [...] syntax.
func formatCell(v interface{}) string {
	switch v.(type) {
	case map[string]interface{}, []interface{}:
		if b, err := json.Marshal(v); err == nil {
			return string(b)
		}
	}
	return fmt.Sprintf("%v", v)
}
