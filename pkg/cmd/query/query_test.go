package query

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdQuery(f) }

func TestQuery_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/action/query", r.URL.Path)

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "SELECT Id, Name FROM Account", body["queryString"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": []map[string]interface{}{
				{"Id": "001", "Name": "Acme"},
				{"Id": "002", "Name": "Beta"},
			},
			"size": 2,
			"done": true,
		})
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "query", "SELECT Id, Name FROM Account", "--json")

	require.NoError(t, err)
	assert.Contains(t, stdout, "001")
	assert.Contains(t, stdout, "Acme")
}

func TestQuery_Pagination(t *testing.T) {
	// ZOQLPages asserts the POST query -> queryMore path/locator contract and
	// that exactly two pages are fetched.
	handler := cmdtest.ZOQLPages(t,
		[]map[string]interface{}{{"Id": "001"}},
		[]map[string]interface{}{{"Id": "002"}},
	)

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "query", "SELECT Id FROM Account", "--json")

	require.NoError(t, err)
	assert.Contains(t, stdout, "001")
	assert.Contains(t, stdout, "002")
}

func TestQuery_Limit(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"records": []map[string]interface{}{
			{"Id": "001"},
			{"Id": "002"},
			{"Id": "003"},
		},
		"size": 3,
		"done": true,
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "query", "SELECT Id FROM Account", "--limit", "2", "--json")

	require.NoError(t, err)
	// Should have only 2 records
	var result struct {
		Records []map[string]interface{} `json:"records"`
		Size    int                      `json:"size"`
	}
	json.Unmarshal([]byte(stdout), &result)
	assert.Equal(t, 2, result.Size)
	assert.Len(t, result.Records, 2)
}

func TestQuery_CSV(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"records": []map[string]interface{}{
			{"Id": "001", "Name": "Acme"},
		},
		"size": 1,
		"done": true,
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "query", "SELECT Id, Name FROM Account", "--csv")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Id")
	assert.Contains(t, stdout, "Name")
	assert.Contains(t, stdout, "001")
	assert.Contains(t, stdout, "Acme")
}

// query precedence (query.go): the JSON family (--json/--jq/--template) is
// rendered BEFORE the --csv branch, so when combined with the inherited global
// --csv, the JSON family wins and --csv is ignored (F-15).
func TestQuery_JSONBeatsCSV(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"records": []map[string]interface{}{{"Id": "001", "Name": "Acme"}},
		"size":    1, "done": true,
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler,
		"query", "SELECT Id, Name FROM Account", "--csv", "--json")
	require.NoError(t, err)
	assert.True(t, json.Valid([]byte(stdout)), "output must be JSON — the JSON family wins over --csv")
	assert.Contains(t, stdout, `"records"`)
	assert.NotContains(t, stdout, "Id,Name", "must NOT be CSV when --json is also set")
}

func TestQuery_JQBeatsCSV(t *testing.T) {
	handler := cmdtest.OK(t, "", "", map[string]interface{}{
		"records": []map[string]interface{}{{"Id": "001", "Name": "Acme"}},
		"size":    1, "done": true,
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler,
		"query", "SELECT Id FROM Account", "--csv", "--jq", ".size")
	require.NoError(t, err)
	assert.Equal(t, "1\n", stdout, "jq output wins over --csv (the JSON family is rendered first)")
}

// TestQuery_JSONTemplateMutuallyExclusive pins that the --json/--template
// exclusion fires on a REAL command that shadows a global flag (query shadows
// --csv) — not just on the network-free version command (F-20). The nil handler
// proves it errors in PersistentPreRunE before any request.
func TestQuery_JSONTemplateMutuallyExclusive(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil,
		"query", "SELECT Id FROM Account", "--json", "--template", "{{.}}")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot use --json and --template together")
}

func TestQuery_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "query")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestQuery_ExportWritesFileAtomically(t *testing.T) {
	// --export writes the result to a temp file in the target's directory and
	// renames it on success. Verify the file lands with the data and that no
	// .zr-export-* temp file is left behind (the deferred rename/cleanup ran). (#436)
	handler := cmdtest.OK(t, "POST", "/v1/action/query", map[string]interface{}{
		"records": []map[string]interface{}{
			{"Id": "001", "Name": "Acme"},
			{"Id": "002", "Name": "Beta"},
		},
		"size": 2, "done": true,
	})

	dir := t.TempDir()
	out := filepath.Join(dir, "res.csv")
	_, _, err := cmdtest.Run(t, "", newCmd, handler, "query", "SELECT Id, Name FROM Account", "--csv", "--output", out)
	require.NoError(t, err)

	b, rerr := os.ReadFile(out)
	require.NoError(t, rerr)
	assert.Contains(t, string(b), "001")
	assert.Contains(t, string(b), "Acme")

	entries, rerr := os.ReadDir(dir)
	require.NoError(t, rerr)
	for _, e := range entries {
		assert.False(t, strings.HasPrefix(e.Name(), ".zr-export-"), "atomic export leaked a temp file: %s", e.Name())
	}
}

func TestQuery_ExportFormatsNestedObjectCells(t *testing.T) {
	// formatCell JSON-encodes non-scalar (map/slice) ZOQL record values. Drive
	// that path through --export and assert the cells are the JSON encodings. (#436)
	handler := cmdtest.OK(t, "POST", "/v1/action/query", map[string]interface{}{
		"records": []map[string]interface{}{
			{"Id": "001", "BillTo": map[string]interface{}{"city": "NYC"}, "Tags": []interface{}{"a", "b"}},
		},
		"size": 1, "done": true,
	})

	out := filepath.Join(t.TempDir(), "nested.csv")
	_, _, err := cmdtest.Run(t, "", newCmd, handler, "query", "SELECT Id, BillTo, Tags FROM Account", "--csv", "--output", out)
	require.NoError(t, err)

	b, rerr := os.ReadFile(out)
	require.NoError(t, rerr)
	rows, perr := csv.NewReader(bytes.NewReader(b)).ReadAll()
	require.NoError(t, perr)
	require.GreaterOrEqual(t, len(rows), 2, "header + one data row")

	// Flatten the data row and confirm the nested object/array rendered as JSON.
	cells := strings.Join(rows[1], "\x00")
	assert.Contains(t, cells, `{"city":"NYC"}`, "nested object cell must be JSON-encoded")
	assert.Contains(t, cells, `["a","b"]`, "array cell must be JSON-encoded")
}

// TestQuery_ExportAliasRemoved pins the #512 removal: only --output is
// accepted.
func TestQuery_ExportAliasRemoved(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "query", "SELECT Id FROM Account", "--export", "out.csv")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag: --export")
}
