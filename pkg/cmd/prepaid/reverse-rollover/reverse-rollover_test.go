package reverserollover

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
	prepaid := &cobra.Command{Use: "prepaid"}
	prepaid.AddCommand(NewCmdReverseRollover(f))
	root.AddCommand(prepaid)
	return root
}

func TestPrepaidReverseRollover_Success(t *testing.T) {
	var gotBody map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/ppdd/reverse-rollover", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"prepaid", "reverse-rollover", "--body", `{"subscriptionNumber":"A-S001"}`, "--confirm"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Equal(t, "A-S001", gotBody["subscriptionNumber"])
	assert.Contains(t, errOut.String(), "Prepaid reverse rollover completed.")
}

func TestPrepaidReverseRollover_RequiresConfirm(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(200)
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"prepaid", "reverse-rollover", "--body", `{"subscriptionNumber":"A-S001"}`})
	err := root.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
	assert.False(t, called, "no HTTP call should be made without --confirm")
}

func TestPrepaidReverseRollover_RequiresBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, "http://localhost", "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"prepaid", "reverse-rollover"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}

func TestPrepaidReverseRollover_SuccessFalse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"reasons": []map[string]interface{}{
				{"code": 50000000, "message": "Reverse rollover failed"},
			},
		})
	}))
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"prepaid", "reverse-rollover", "--body", `{"bad":"data"}`, "--confirm"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Reverse rollover failed")
}
