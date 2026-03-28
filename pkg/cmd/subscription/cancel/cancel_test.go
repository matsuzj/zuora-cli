package cancel

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
	sub.AddCommand(NewCmdCancel(f))
	root.AddCommand(sub)
	return root
}

func TestCancel_WithPolicy(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/cancel")
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "EndOfCurrentTerm", body["cancellationPolicy"])
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "subscriptionId": "sub-1"})
	}))

	ios, _, out, errOut := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "cancel", "A-S001", "--policy", "EndOfCurrentTerm"})
	require.NoError(t, root.Execute())
	assert.Contains(t, out.String(), "true")
	assert.Contains(t, errOut.String(), "cancelled")
}

func TestCancel_WithBody(t *testing.T) {
	server := testutil.NewServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))

	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), server.URL, "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "cancel", "A-S001", "--body", `{"cancellationPolicy":"EndOfCurrentTerm"}`})
	require.NoError(t, root.Execute())
}

func TestCancel_RequiresPolicyOrBody(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "cancel", "A-S001"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--policy or --body")
}

func TestCancel_SpecificDateRequiresEffectiveDate(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	f := factory.NewTestFactory(ios, config.NewMockConfig(), "http://localhost", "tok")
	root := newTestRoot(f)
	root.SetArgs([]string{"subscription", "cancel", "A-S001", "--policy", "SpecificDate"})
	err := root.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--effective-date is required")
}
