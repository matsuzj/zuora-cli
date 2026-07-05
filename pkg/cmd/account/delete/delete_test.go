package delete

import (
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdDelete(f) }

func TestAccountDelete_Success(t *testing.T) {
	handler := cmdtest.Status(t, "DELETE", "/v1/accounts/A001", 204, nil)

	_, stderr, err := cmdtest.Run(t, "account", newCmd, handler, "account", "delete", "A001", "--confirm")

	require.NoError(t, err)
	assert.Contains(t, stderr, "Account A001 deleted.")
}

func TestAccountDelete_JSON(t *testing.T) {
	handler := cmdtest.Status(t, "", "", 204, nil)

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "delete", "A001", "--confirm", "--json")

	require.NoError(t, err)
	assert.Contains(t, stdout, `"success": true`)
}

func TestAccountDelete_RequiresConfirm(t *testing.T) {
	cmdtest.RequiresConfirm(t, "account", newCmd, "account", "delete", "A001")
}

func TestAccountDelete_BodyResponse(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"jobId":"job-1","jobStatus":"Pending"}`))
	})

	stdout, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "delete", "A001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "job-1")
	assert.Contains(t, stdout, "Pending")
}

func TestAccountDelete_AsyncRejection(t *testing.T) {
	// Async account delete returns HTTP 200 {"success":false} when the account
	// cannot be deleted (e.g. active subscriptions). This must be a non-zero exit,
	// not a silent success.
	handler := cmdtest.Reasons(t, "", "account has active subscriptions")

	_, _, err := cmdtest.Run(t, "account", newCmd, handler, "account", "delete", "A001", "--confirm")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "account has active subscriptions")
}

func TestAccountDelete_UnparseableBody(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not json"))
	})

	_, stderr, err := cmdtest.Run(t, "account", newCmd, handler, "account", "delete", "A001", "--confirm")
	require.NoError(t, err, "non-JSON 200 is a completed delete under the unified policy")
	assert.Contains(t, stderr, "deleted")
}
