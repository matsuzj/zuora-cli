package get

import (
	"encoding/json"
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestOrderGet_Success(t *testing.T) {
	// Real-shaped GET /v1/orders/{id} response (nested under "order") loaded from
	// a golden fixture. The asserts below bite on nested, drift-prone keys — esp.
	// "existingAccountNumber" (NOT the flatter "accountNumber"): a swap to the
	// wrong key would render empty yet keep the test green without that assert.
	handler := cmdtest.OK(t, "GET", "/v1/orders/O-00000001",
		json.RawMessage(cmdtest.LoadFixture(t, "order_get")))

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "get", "O-00000001")
	require.NoError(t, err)
	assert.Contains(t, stdout, "O-00000001")   // orderNumber (nested)
	assert.Contains(t, stdout, "Completed")    // status (nested)
	assert.Contains(t, stdout, "ACCT-9000001") // existingAccountNumber (nested, drift-prone)
}

func TestOrderGet_RequiresArg(t *testing.T) {
	_, _, err := cmdtest.Run(t, "order", newCmd, nil, "order", "get")
	assert.Error(t, err)
}

func TestOrderGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 58730020, "Order not found")

	_, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "get", "O-99999999")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Order not found")
}

func TestOrderGet_FlatResponseFallsBackToRaw(t *testing.T) {
	// When a response is NOT nested under "order", the fallback (order = raw) must
	// render the detail fields from the top level instead of empty.
	handler := cmdtest.OK(t, "GET", "/v1/orders/O-00000001", map[string]interface{}{
		"orderNumber": "O-FLAT-1",
		"status":      "Completed",
		"success":     true,
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "get", "O-00000001")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Order Number:\s+O-FLAT-1$`, stdout)
	assert.Regexp(t, `(?m)^Status:\s+Completed$`, stdout)
}

func TestOrderGet_NonMapOrderKeyFallsBackToRaw(t *testing.T) {
	// raw["order"] is a non-map value: the type assertion fails (order stays nil),
	// so the fallback to raw must still render from the top level — and must not
	// panic on the unexpected shape.
	handler := cmdtest.OK(t, "GET", "/v1/orders/O-00000001", map[string]interface{}{
		"order":       []interface{}{"unexpected"}, // not a map
		"orderNumber": "O-RAW-1",
		"success":     true,
	})

	stdout, _, err := cmdtest.Run(t, "order", newCmd, handler, "order", "get", "O-00000001")
	require.NoError(t, err)
	assert.Regexp(t, `(?m)^Order Number:\s+O-RAW-1$`, stdout)
}
