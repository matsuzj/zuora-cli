package post

import (
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdPost(f) }

func TestBillRunPost_Success(t *testing.T) {
	handler := cmdtest.OK(t, "PUT", "/v1/bill-runs/br-001/post", map[string]interface{}{
		"id":            "br-001",
		"billRunNumber": "BR-00000001",
		"status":        "Posted",
		"success":       true,
	})

	stdout, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "post", "br-001", "--confirm")
	require.NoError(t, err)
	assert.Contains(t, stdout, "Posted")
}

func TestBillRunPost_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "post")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestBillRunPost_RequiresConfirm(t *testing.T) {
	// Posting finalizes every invoice/credit memo the bill run generated and is
	// irreversible — it must require --confirm. (#424)
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "post", "br-001")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--confirm")
}

// TestBillRunPost_SendsEmptyJSONBody pins the 415 fix: Zuora's endpoint binds a Map body parameter
// and rejects requests without a Content-Type, which the client sets only
// when a body is present — the command must send an explicit "{}".
func TestBillRunPost_SendsEmptyJSONBody(t *testing.T) {
	inner := cmdtest.OK(t, "PUT", "/v1/bill-runs/br-001/post", map[string]interface{}{
		"id": "br-001", "status": "Posted", "success": true,
	})
	handler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		assert.Equal(t, "{}", string(b))
		inner(w, r)
	}

	_, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "post", "br-001", "--confirm")
	require.NoError(t, err)
}
