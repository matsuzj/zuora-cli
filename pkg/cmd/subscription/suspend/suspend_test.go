package suspend

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
	sub := &cobra.Command{Use: "subscription"}
	sub.AddCommand(NewCmdSuspend(f))
	root.AddCommand(sub)
	return root
}

func TestSuspend_WithPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "FixedPeriodsFromToday", body["suspendPolicy"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer server.Close()

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "suspend", "A-S001", "--policy", "FixedPeriodsFromToday", "--periods", "3", "--periods-type", "Month"})
	require.NoError(t, root.Execute())
	assert.Contains(t, errOut.String(), "suspended")
}

func TestSuspend_RequiresPolicyOrBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "suspend", "A-S001"})
	assert.Error(t, root.Execute())
}

func TestSuspend_SpecificDateRequiresSuspendDate(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "suspend", "A-S001", "--policy", "SpecificDate"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--suspend-date is required")
}

func TestSuspend_FixedPeriodsRequiresPeriodsAndType(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "suspend", "A-S001", "--policy", "FixedPeriodsFromToday", "--periods", "3"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--periods and --periods-type are required")
}
