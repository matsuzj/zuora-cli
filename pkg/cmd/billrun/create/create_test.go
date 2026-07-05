package create

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdCreate(f) }

func TestBillRunCreate_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/bill-runs", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "AllBatches")
		// The full --body payload must reach the server intact. (#484)
		assert.JSONEq(t, `{"batches":["AllBatches"],"targetDate":"2026-06-30"}`, string(body))
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "br-001",
			"billRunNumber": "BR-00000001",
			"status":        "Pending",
			"success":       true,
		})
	})

	stdout, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "create", "--body", `{"batches":["AllBatches"],"targetDate":"2026-06-30"}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "BR-00000001")
	// Label-bound (#483): pin each rendered field under its own label.
	assert.Regexp(t, `(?m)^ID:\s+br-001$`, stdout)
	assert.Regexp(t, `(?m)^Bill Run Number:\s+BR-00000001$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Pending$`, stdout)
}

func TestBillRunCreate_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "billrun", newCmd, nil, "billrun", "create")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}

func TestBillRunCreate_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730020, "Invalid target date")

	_, _, err := cmdtest.Run(t, "billrun", newCmd, handler, "billrun", "create", "--body", `{}`)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid target date")
}
