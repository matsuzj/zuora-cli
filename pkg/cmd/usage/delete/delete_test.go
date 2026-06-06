package delete

import (
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
	usage := &cobra.Command{Use: "usage"}
	usage.AddCommand(NewCmdDelete(f))
	root.AddCommand(usage)
	return root
}

func TestUsageDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Equal(t, "/v1/object/usage/2c92a0f96bd", r.URL.Path)
		w.WriteHeader(204)
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "delete", "2c92a0f96bd", "--confirm"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, errOut.String(), "Usage record 2c92a0f96bd deleted.")
}

func TestUsageDelete_RequiresConfirm(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "delete", "2c92a0f96bd"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

func TestUsageDelete_RequiresArg(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "delete", "--confirm"})
	err := root.Execute()

	assert.Error(t, err)
}

func TestUsageDelete_BodyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"Id":"u-1"}`))
	}))
	defer server.Close()

	ios, _, out, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "delete", "u-1", "--confirm"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "u-1")
	assert.Contains(t, errOut.String(), "Usage record u-1 deleted.")
}

func TestUsageDelete_UnparseableBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"usage", "delete", "u-1", "--confirm"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse")
	assert.Contains(t, errOut.String(), "Unexpected response")
}
