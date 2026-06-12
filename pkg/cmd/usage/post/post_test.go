package post

import (
	"encoding/json"
	"io"
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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPost(f) }

func TestUsagePost_Success(t *testing.T) {
	// Create a temp CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "usage.csv")
	err := os.WriteFile(csvFile, []byte("ACCOUNT_ID,UOM,QTY,STARTDATE\nA001,Each,10,01/01/2026\n"), 0644)
	require.NoError(t, err)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/usage", r.URL.Path)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")

		// Verify the file is in the multipart body
		err := r.ParseMultipartForm(10 << 20)
		require.NoError(t, err)
		file, header, err := r.FormFile("file")
		require.NoError(t, err)
		defer file.Close()
		assert.Equal(t, "usage.csv", header.Filename)
		data, readErr := io.ReadAll(file)
		require.NoError(t, readErr)
		assert.Contains(t, string(data), "ACCOUNT_ID")

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":           true,
			"checkImportStatus": "https://rest.zuora.com/v1/usage/123/status",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "usage", newCmd, handler, "usage", "post", "--file", csvFile)

	require.NoError(t, err)
	assert.Contains(t, stdout, "Check Import Status")
	assert.Contains(t, stderr, "Usage file uploaded.")
}

func TestUsagePost_RequiresFile(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "post")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "file" not set`)
}

func TestUsagePost_FileNotFound(t *testing.T) {
	_, _, err := cmdtest.Run(t, "usage", newCmd, nil, "usage", "post", "--file", "/nonexistent/file.csv")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading file")
}
