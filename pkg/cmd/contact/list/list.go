// Package list implements the "zr contact list" command.
package list

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/output"
	"github.com/spf13/cobra"
)

// accountIDPattern bounds --account-id to the characters real Zuora object IDs
// use (hex-ish alphanumerics, with - and _ allowed defensively). Anything else —
// quotes, spaces, ZOQL syntax — is rejected before it can reach the ZOQL string
// literal. This is the PRIMARY injection defense: it does not depend on ZOQL's
// (unclear) quote-escaping semantics the way string-escaping alone would.
var accountIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// escapeZOQLString escapes a value for embedding in a single-quoted ZOQL
// string literal. Backslashes are escaped BEFORE quotes: escaping only the
// quote would let an input ending in `\` neutralize the escape (or the
// closing quote of the literal), e.g. `x\` -> `'x\'`.
func escapeZOQLString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	return strings.ReplaceAll(s, "'", `\'`)
}

// NewCmdList creates the contact list command.
func NewCmdList(f *factory.Factory) *cobra.Command {
	var accountID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contacts for an account",
		Long: `List contacts for a Zuora account using ZOQL query.

Requires --account-id (the Zuora account ID, not account number).
Use "zr account get <number> --jq .basicInfo.id" to find the account ID.`,
		Example: `  zr contact list --account-id 8aca822f12345
  zr contact list --account-id 8aca822f12345 --json`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, f, accountID)
		},
	}

	cmd.Flags().StringVar(&accountID, "account-id", "", "Zuora account ID (required)")
	_ = cmd.MarkFlagRequired("account-id")
	return cmd
}

type contactRecord struct {
	ID        string `json:"Id"`
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	WorkEmail string `json:"WorkEmail"`
}

type queryResponse struct {
	Records      []contactRecord `json:"records"`
	Size         int             `json:"size"`
	Done         bool            `json:"done"`
	QueryLocator string          `json:"queryLocator"`
}

func runList(cmd *cobra.Command, f *factory.Factory, accountID string) error {
	// Reject anything that isn't a plain Zuora object ID before it can reach the
	// ZOQL query — this is the real injection guard (see accountIDPattern).
	if !accountIDPattern.MatchString(accountID) {
		return fmt.Errorf("invalid --account-id %q: must be a Zuora object ID (letters, digits, - or _)", accountID)
	}

	client, err := f.HttpClient()
	if err != nil {
		return err
	}

	// Belt-and-suspenders: escape the value even though the pattern above
	// already forbids quotes and backslashes, so a future loosening of the
	// pattern can't silently re-open ZOQL injection.
	sanitized := escapeZOQLString(accountID)
	zoql := fmt.Sprintf(
		`SELECT Id, FirstName, LastName, WorkEmail FROM Contact WHERE AccountId = '%s'`,
		sanitized,
	)
	body := fmt.Sprintf(`{"queryString":%q}`, zoql)

	resp, err := client.Post("/v1/action/query", strings.NewReader(body))
	if err != nil {
		return err
	}

	var result queryResponse
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return fmt.Errorf("parsing response: %w", err)
	}

	allRecords := result.Records

	// Follow ZOQL pagination via queryMore
	for !result.Done && result.QueryLocator != "" {
		moreBody := fmt.Sprintf(`{"queryLocator":%q}`, result.QueryLocator)
		resp, err = client.Post("/v1/action/queryMore", strings.NewReader(moreBody))
		if err != nil {
			return err
		}
		result = queryResponse{}
		if err := json.Unmarshal(resp.Body, &result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
		allRecords = append(allRecords, result.Records...)
	}

	fmtOpts := output.FromCmd(cmd)

	cols := []output.Column{
		{Header: "ID"},
		{Header: "FIRST NAME"},
		{Header: "LAST NAME"},
		{Header: "EMAIL"},
	}

	rows := make([][]string, len(allRecords))
	for i, c := range allRecords {
		rows[i] = []string{c.ID, c.FirstName, c.LastName, c.WorkEmail}
	}

	// Build merged JSON with all paginated records for --json/--jq/--template
	merged := map[string]interface{}{
		"records": allRecords,
		"size":    len(allRecords),
		"done":    true,
	}
	rawJSON, err := json.Marshal(merged)
	if err != nil {
		return fmt.Errorf("building response: %w", err)
	}

	return output.Render(f.IOStreams, rawJSON, fmtOpts, rows, cols)
}
