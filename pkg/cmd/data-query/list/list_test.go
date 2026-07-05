package list

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdList(f) }

func TestList_SendsStatusAndPageSize(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/query/jobs", r.URL.Path)
		assert.Equal(t, "completed", r.URL.Query().Get("queryStatus"))
		assert.Equal(t, "20", r.URL.Query().Get("pageSize"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []map[string]interface{}{
			{"id": "j1", "queryStatus": "completed"},
		}})
	})
	stdout, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "list", "--status", "completed", "--page-size", "20")
	require.NoError(t, err)
	assert.Contains(t, stdout, "j1")
	// Pin the second declared column's cell too (#483): Status (queryStatus)
	// was fixtured but unasserted.
	assert.Contains(t, stdout, "completed")
}

func TestList_NoFilters(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/query/jobs", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("queryStatus"))
		// page-size is OmitZero, so an unset (0) value must not be sent.
		assert.Empty(t, r.URL.Query().Get("pageSize"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []map[string]interface{}{}})
	})
	_, _, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "list")
	require.NoError(t, err)
}
