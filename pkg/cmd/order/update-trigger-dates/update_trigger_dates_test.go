package updatetriggerdates

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdUpdateTriggerDates(f) }

func TestOrderUpdateTriggerDates_Success(t *testing.T) {
	var gotBody map[string]interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/orders/O-00000001/triggerDates", r.URL.Path)

		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
		})
	})

	_, stderr, err := cmdtest.Run(t, "order", newCmd, handler, "order", "update-trigger-dates", "O-00000001", "--body", `{"orderActions":[{"sequence":1}]}`)

	require.NoError(t, err)
	assert.NotNil(t, gotBody["orderActions"])
	assert.Contains(t, stderr, "Trigger dates updated for order O-00000001.")
}

func TestOrderUpdateTriggerDates_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 53100020, "Invalid trigger date")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "update-trigger-dates", "O-00000001", "--body", `{"bad":"data"}`)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid trigger date")
}

func TestOrderUpdateTriggerDates_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "update-trigger-dates", "O-00000001")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--body is required")
}
