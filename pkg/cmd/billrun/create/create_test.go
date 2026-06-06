package create

import (
	"encoding/json"
	"io"
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
	root.PersistentFlags().Bool("csv", false, "")
	br := &cobra.Command{Use: "billrun"}
	br.AddCommand(NewCmdCreate(f))
	root.AddCommand(br)
	return root
}

func TestBillRunCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/bill-runs", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "AllBatches")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "br-001",
			"billRunNumber": "BR-00000001",
			"status":        "Pending",
			"success":       true,
		})
	}))
	defer server.Close()

	ios, in, out, _ := iostreams.Test()
	in.WriteString("")
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"billrun", "create", "--body", `{"batches":["AllBatches"],"targetDate":"2026-06-30"}`})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "BR-00000001")
}

func TestBillRunCreate_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"billrun", "create"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestBillRunCreate_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{{"code": 58730020, "message": "Invalid target date"}},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")
	root := newTestRoot(f)
	root.SetArgs([]string{"billrun", "create", "--body", `{}`})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid target date")
}
