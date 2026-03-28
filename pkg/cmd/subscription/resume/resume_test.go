package resume

import (
	"encoding/json"
	"github.com/matsuzj/zuora-cli/internal/testutil"
	"net/http"
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
	sub.AddCommand(NewCmdResume(f))
	root.AddCommand(sub)
	return root
}

func TestResume_WithPolicy(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "SpecificDate", body["resumePolicy"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))

	ios, _, _, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "resume", "A-S001", "--policy", "SpecificDate", "--resume-date", "2026-05-01"})
	require.NoError(t, root.Execute())
	assert.Contains(t, errOut.String(), "resumed")
}

func TestResume_RequiresPolicyOrBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "resume", "A-S001"})
	assert.Error(t, root.Execute())
}

func TestResume_SpecificDateRequiresResumeDate(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "resume", "A-S001", "--policy", "SpecificDate"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--resume-date is required")
}

func TestResume_FixedPeriodsRequiresPeriodsAndType(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "resume", "A-S001", "--policy", "FixedPeriodsFromSuspendDate", "--periods", "1"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--periods and --periods-type are required")
}
