// Package dqutil holds the shared helpers for the "zr data-query" command
// group (Zuora's asynchronous Data Query API). It lives in its own package so
// the parent command and every subcommand can import it without an import
// cycle (the parent imports the subcommand packages, so they cannot import the
// parent).
package dqutil

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmdutil"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// SubmitPath is the Data Query job-submission endpoint. It is NOT under /v1
// (Zuora exposes Data Query at /query/jobs); the read-only guard's allow path
// (isDataQueryWrite in internal/api) matches the normalized form "query/jobs".
const SubmitPath = "/query/jobs"

// JobPath returns the per-job path for a (percent-escaped) job id.
func JobPath(jobID string) string { return "/query/jobs/" + url.PathEscape(jobID) }

// Accepted enum values, offered as shell completions (suggestions only, like
// listcmd's Enum — not validated client-side).
var (
	OutputFormats = []string{"JSON", "CSV", "TSV", "DSV"}
	Compressions  = []string{"NONE", "GZIP", "ZIP"}
	Sources       = []string{"LIVE", "WAREHOUSE"}
	JobStatuses   = []string{"accepted", "in_progress", "completed", "failed", "cancelled"}
)

// SubmitFlags holds the Data Query job-submission options shared by `submit`
// and `run`.
type SubmitFlags struct {
	File            string
	OutputFormat    string
	Compression     string
	ColumnSeparator string
	Source          string
	ReadDeleted     bool
	UseIndexJoin    bool
	IdempotencyKey  string
}

// AddSubmitFlags registers the job-submission flags shared by `submit` and
// `run`.
func AddSubmitFlags(fs *pflag.FlagSet, sf *SubmitFlags) {
	fs.StringVar(&sf.File, "file", "", "Read the SQL from a file instead of the argument")
	fs.StringVar(&sf.OutputFormat, "output-format", "JSON", "Result file format: JSON|CSV|TSV|DSV")
	fs.StringVar(&sf.Compression, "compression", "NONE", "Result compression: NONE|GZIP|ZIP")
	fs.StringVar(&sf.ColumnSeparator, "column-separator", "", "Column separator for DSV output")
	fs.StringVar(&sf.Source, "source", "", "Query source: LIVE|WAREHOUSE (default LIVE)")
	fs.BoolVar(&sf.ReadDeleted, "read-deleted", false, "Include soft-deleted records (30-day retention)")
	fs.BoolVar(&sf.UseIndexJoin, "use-index-join", false, "Use index join (see Data Query best practices)")
	fs.StringVar(&sf.IdempotencyKey, "idempotency-key", "", "Idempotency-Key for the submit POST (prevents duplicate jobs on retry)")
}

// RegisterSubmitCompletions wires shell completion for the submit enum flags.
func RegisterSubmitCompletions(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("output-format", cmdutil.EnumCompletion(OutputFormats...))
	_ = cmd.RegisterFlagCompletionFunc("compression", cmdutil.EnumCompletion(Compressions...))
	_ = cmd.RegisterFlagCompletionFunc("source", cmdutil.EnumCompletion(Sources...))
}

// ResolveSQL returns the query text from exactly one of the positional argument
// or --file; supplying both, or neither, is an error.
func ResolveSQL(args []string, file string) (string, error) {
	hasPos := len(args) == 1 && strings.TrimSpace(args[0]) != ""
	switch {
	case hasPos && file != "":
		return "", fmt.Errorf("provide the SQL as an argument OR via --file, not both")
	case hasPos:
		return args[0], nil
	case file != "":
		b, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("reading --file %q: %w", file, err)
		}
		if strings.TrimSpace(string(b)) == "" {
			return "", fmt.Errorf("--file %q is empty", file)
		}
		return string(b), nil
	default:
		return "", fmt.Errorf("provide the SQL as an argument or via --file")
	}
}

// BuildSubmitBody assembles the POST /query/jobs request body. output.target is
// required by Zuora.
//
// The core body (query / outputFormat / compression / output.target) plus
// sourceData and useIndexJoin are live-verified against an apac-sandbox tenant
// (2026-06-29; the response echoes sourceData/useIndexJoin). columnSeparator
// (DSV) and readDeleted follow the API reference but were not exercised live.
func BuildSubmitBody(sql string, sf *SubmitFlags) ([]byte, error) {
	body := map[string]interface{}{
		"query":        sql,
		"outputFormat": sf.OutputFormat,
		"compression":  sf.Compression,
		"output":       map[string]interface{}{"target": "S3"},
	}
	if sf.ColumnSeparator != "" {
		body["columnSeparator"] = sf.ColumnSeparator
	}
	if sf.Source != "" {
		body["sourceData"] = sf.Source
	}
	if sf.ReadDeleted {
		body["readDeleted"] = true
	}
	if sf.UseIndexJoin {
		body["useIndexJoin"] = true
	}
	return json.Marshal(body)
}

// UnwrapData returns the "data" object of a Data Query response, or the raw map
// itself when there is no "data" envelope (defensive against shape drift).
func UnwrapData(raw map[string]interface{}) map[string]interface{} {
	if d, ok := raw["data"].(map[string]interface{}); ok {
		return d
	}
	return raw
}

// DecodeData parses a Data Query response body and returns its data object.
func DecodeData(body []byte) (map[string]interface{}, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return UnwrapData(raw), nil
}

// DetailFields renders the standard Data Query job fields for a detail view.
// The field names (queryStatus / outputRows / processingTime / dataFile) are
// live-verified against an apac-sandbox tenant (2026-06-29). outputRows and
// processingTime come back as JSON numbers, so GetDecimal renders them as plain
// decimals rather than scientific notation.
func DetailFields(d map[string]interface{}) []output.DetailField {
	return []output.DetailField{
		{Key: "ID", Value: cmdutil.GetString(d, "id")},
		{Key: "Status", Value: cmdutil.GetString(d, "queryStatus")},
		// Counts/times are numeric: GetDecimal renders a JSON-number float64 as a
		// plain decimal (1000000, not "1e+06") and passes a string value through.
		{Key: "Output Rows", Value: cmdutil.GetDecimal(d, "outputRows")},
		{Key: "Processing Time", Value: cmdutil.GetDecimal(d, "processingTime")},
		{Key: "Data File", Value: cmdutil.GetString(d, "dataFile")},
	}
}

// IsTerminalStatus reports whether a Data Query job has reached a terminal
// state. Zuora documents lowercase statuses; the US "canceled" spelling is also
// treated as terminal defensively.
func IsTerminalStatus(s string) bool {
	switch strings.ToLower(s) {
	case "completed", "failed", "cancelled", "canceled":
		return true
	}
	return false
}

// FirstNonEmpty returns the first non-empty string.
func FirstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// DownloadClientForTest overrides the hardened download client when non-nil.
// Test seam only — production code passes nil and DownloadStream consults this
// var itself, so production callers never reference *ForTest. When unset (nil),
// DownloadStream builds the hardened client (no auth, redirects refused,
// compression disabled).
var DownloadClientForTest *http.Client

// HardenedDownloadClient builds the HTTP client used to fetch the S3 dataFile.
// It carries NO Authorization (the URL is pre-signed and off the Zuora host, so
// the bearer token must never be attached), refuses redirects (a redirect could
// forward the signed query off-host), and disables compression so the exact
// bytes (including a GZIP/ZIP result) are preserved. The transport is cloned
// from the default so dial/TLS timeouts and proxy settings are kept.
func HardenedDownloadClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DisableCompression = true
	return &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return fmt.Errorf("refusing to follow redirect to %q (would leak the signed dataFile URL)", req.URL.Host)
		},
	}
}

// DownloadStream fetches rawURL (an S3 pre-signed https URL) and copies the
// exact bytes to w. client is injectable for tests; nil uses the hardened
// production client.
func DownloadStream(ctx context.Context, rawURL string, w io.Writer, client *http.Client) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid dataFile URL: %w", err)
	}
	if !strings.EqualFold(u.Scheme, "https") || u.Host == "" || u.User != nil {
		return fmt.Errorf("refusing to download dataFile from %q: require an https URL with no embedded credentials", rawURL)
	}
	if client == nil {
		// Production callers pass nil; honor the test seam here (not in command
		// code) so production paths never reference *ForTest. (#440)
		client = DownloadClientForTest
	}
	if client == nil {
		client = HardenedDownloadClient()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("downloading dataFile: HTTP %d", resp.StatusCode)
	}
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("writing dataFile: %w", err)
	}
	return nil
}

// DownloadToFile downloads rawURL to dstPath. dstPath "-" streams to stdout
// instead. A file download is atomic: it streams to a temp file in the target
// directory and renames on success, so a mid-stream failure leaves any existing
// file untouched. client is injectable for tests (nil = hardened production).
func DownloadToFile(ctx context.Context, rawURL, dstPath string, stdout io.Writer, client *http.Client) error {
	if dstPath == "-" {
		return DownloadStream(ctx, rawURL, stdout, client)
	}
	tmp, err := os.CreateTemp(filepath.Dir(dstPath), ".zr-dq-*")
	if err != nil {
		return fmt.Errorf("creating download temp file: %w", err)
	}
	tmpName := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()
	if err := DownloadStream(ctx, rawURL, tmp, client); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing download temp file: %w", err)
	}
	if err := os.Rename(tmpName, dstPath); err != nil {
		return fmt.Errorf("finalizing download to %q: %w", dstPath, err)
	}
	committed = true
	return nil
}
