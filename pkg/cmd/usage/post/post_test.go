package post

import (
	"encoding/json"
	httptest "github.com/matsuzj/zuora-cli/internal/testutil/httpmock"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	usage := &cobra.Command{Use: "usage"}
	usage.AddCommand(NewCmdPost(f))
	root.AddCommand(usage)
	return root
}

func TestUsagePost_Success(t *testing.T) {
	// Create a temp CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "usage.csv")
	err := os.WriteFile(csvFile, []byte("ACCOUNT_ID,UOM,QTY,STARTDATE\nA001,Each,10,01/01/2026\n"), 0644)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer server.Close()

	ios, _, out, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "post", "--file", csvFile})
	err = root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "Check Import Status")
	assert.Contains(t, errOut.String(), "Usage file uploaded.")
}

func TestUsagePost_RequiresFile(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "post"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--file is required")
}

func TestUsagePost_FileNotFound(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "post", "--file", "/nonexistent/file.csv"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading file")
}
