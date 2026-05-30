package periods

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
	commitment := &cobra.Command{Use: "commitment"}
	commitment.AddCommand(NewCmdPeriods(f))
	root.AddCommand(commitment)
	return root
}

func TestCommitmentPeriods_ByCommitment_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments/periods", r.URL.Path)
		assert.Equal(t, "CMT-00000001", r.URL.Query().Get("commitmentKey"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"periods": []map[string]interface{}{
				{
					"periodId":      "PRD-00000001",
					"commitmentKey": "CMT-00000001",
					"startDate":     "2026-01-01",
					"endDate":       "2026-12-31",
				},
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "periods", "--commitment", "CMT-00000001"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "PRD-00000001")
	assert.Contains(t, out.String(), "CMT-00000001")
}

func TestCommitmentPeriods_ByAccount_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/commitments/periods", r.URL.Path)
		assert.Equal(t, "A00000001", r.URL.Query().Get("accountNumber"))
		assert.Equal(t, "2026-01-01", r.URL.Query().Get("startDate"))
		assert.Equal(t, "2026-12-31", r.URL.Query().Get("endDate"))
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"periods": []map[string]interface{}{
				{
					"periodId":      "PRD-00000002",
					"accountNumber": "A00000001",
				},
			},
		})
	}))
	defer server.Close()

	ios, _, out, _ := iostreams.Test()
	cfg := config.NewMockConfig()
	f := factory.NewTestFactory(ios, cfg, server.URL, "test-token")

	root := newTestRoot(f)
	root.SetArgs([]string{"commitment", "periods", "--account", "A00000001", "--start-date", "2026-01-01", "--end-date", "2026-12-31"})
	err := root.Execute()

	require.NoError(t, err)
	assert.Contains(t, out.String(), "PRD-00000002")
	assert.Contains(t, out.String(), "A00000001")
}

func TestCommitmentPeriods_RequiresCommitmentOrAccount(t *testing.T) {
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
	root.SetArgs([]string{"commitment", "periods"})
	err := root.Execute()

	assert.Error(t, err)
	assert.False(t, called, "no HTTP call should be made when required flags are missing")
}

func TestCommitmentPeriods_MutuallyExclusive(t *testing.T) {
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
	root.SetArgs([]string{"commitment", "periods", "--commitment", "CMT-00000001", "--account", "A00000001"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
	assert.False(t, called, "no HTTP call should be made for mutually exclusive flags")
}

func TestCommitmentPeriods_AccountRequiresDates(t *testing.T) {
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
	root.SetArgs([]string{"commitment", "periods", "--account", "A00000001"})
	err := root.Execute()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start-date")
	assert.False(t, called, "no HTTP call should be made when dates are missing")
}
