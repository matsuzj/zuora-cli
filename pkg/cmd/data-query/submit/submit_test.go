package submit

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSubmit(f) }

func TestSubmit_Success(t *testing.T) {
	var gotBody map[string]interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/query/jobs", r.URL.Path)
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"id": "job-9", "queryStatus": "accepted"}})
	})

	stdout, stderr, err := cmdtest.Run(t, "data-query", newCmd, handler, "data-query", "submit", "SELECT 1", "--json")
	require.NoError(t, err)
	assert.Equal(t, "SELECT 1", gotBody["query"])
	out, _ := gotBody["output"].(map[string]interface{})
	require.NotNil(t, out)
	assert.Equal(t, "S3", out["target"])
	assert.Contains(t, stdout, "job-9")
	assert.Contains(t, stderr, "submitted")
}

func TestSubmit_FileAndArgConflict(t *testing.T) {
	fp := filepath.Join(t.TempDir(), "q.sql")
	require.NoError(t, os.WriteFile(fp, []byte("SELECT 1"), 0o600))
	_, _, err := cmdtest.Run(t, "data-query", newCmd, nil, "data-query", "submit", "SELECT 1", "--file", fp)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not both")
}

func TestSubmit_RequiresSQL(t *testing.T) {
	_, _, err := cmdtest.Run(t, "data-query", newCmd, nil, "data-query", "submit")
	require.Error(t, err)
}
