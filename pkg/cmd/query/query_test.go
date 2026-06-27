package query

import (
	"encoding/json"
	"net/http"
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
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		callCount++

		if callCount == 1 {
			assert.Equal(t, "/v1/action/query", r.URL.Path)
			w.WriteHeader(200)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"records":      []map[string]interface{}{{"Id": "001"}},
				"size":         1,
				"done":         false,
				"queryLocator": "loc-123",
			})
			return
		}

		// Second call: queryMore
		assert.Equal(t, "/v1/action/queryMore", r.URL.Path)
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "loc-123", body["queryLocator"])

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": []map[string]interface{}{{"Id": "002"}},
			"size":    1,
			"done":    true,
		})
	})

	stdout, _, err := cmdtest.Run(t, "", newCmd, handler, "query", "SELECT Id FROM Account", "--json")

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
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

func TestQuery_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "query")

	assert.Error(t, err)
}
