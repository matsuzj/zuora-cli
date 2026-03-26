package query

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matsuzj/zuora-cli/internal/config"
	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRoot(f *factory.Factory) *cobra.Command {
	root := &cobra.Command{Use: "zr"}
	root.PersistentFlags().Bool("json", false, "")
	root.PersistentFlags().String("jq", "", "")
	root.PersistentFlags().String("template", "", "")
	root.AddCommand(NewCmdQuery(f))
	return root
}

func TestQuery_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"query", "SELECT Id, Name FROM Account", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "001")
	assert.Contains(t, out.String(), "Acme")
}

func TestQuery_Pagination(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"query", "SELECT Id FROM Account", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, 2, callCount)
	assert.Contains(t, out.String(), "001")
	assert.Contains(t, out.String(), "002")
}

func TestQuery_Limit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": []map[string]interface{}{
				{"Id": "001"},
				{"Id": "002"},
				{"Id": "003"},
			},
			"size": 3,
			"done": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"query", "SELECT Id FROM Account", "--limit", "2", "--json"})
	err := root.Execute()

	require.NoError(t, err)
	// Should have only 2 records
	var result struct {
		Records []map[string]interface{} `json:"records"`
		Size    int                      `json:"size"`
	}
	json.Unmarshal(out.Bytes(), &result)
	assert.Equal(t, 2, result.Size)
	assert.Len(t, result.Records, 2)
}

func TestQuery_CSV(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"records": []map[string]interface{}{
				{"Id": "001", "Name": "Acme"},
			},
			"size": 1,
			"done": true,
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"query", "SELECT Id, Name FROM Account", "--csv"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Id")
	assert.Contains(t, out.String(), "Name")
	assert.Contains(t, out.String(), "001")
	assert.Contains(t, out.String(), "Acme")
}

func TestQuery_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"query"})
	err := root.Execute()

	assert.Error(t, err)
}
