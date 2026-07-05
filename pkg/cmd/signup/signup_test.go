package signup

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

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSignup(f) }

func TestSignup_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/sign-up", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "accountInfo")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":            true,
			"accountId":          "a-new",
			"accountNumber":      "A00099",
			"subscriptionId":     "s-new",
			"subscriptionNumber": "A-S001",
			// Follow-on references the Sign-Up API returns (documented shape;
			// this tenant returns HTTP 500 on sign-up so it is not live-verified).
			"orderNumber":      "O-00001",
			"invoiceId":        "inv-id",
			"invoiceNumber":    "INV00001",
			"paymentId":        "pay-id",
			"paymentNumber":    "P-00001",
			"creditMemoId":     "cm-id",
			"creditMemoNumber": "CM00001",
			"paidAmount":       1080,
			"status":           "Active",
		})
	})

	stdout, stderr, err := cmdtest.Run(t, "", newCmd, handler, "signup", "--body", `{"accountInfo":{},"subscriptionInfo":{}}`)
	require.NoError(t, err)
	assert.Contains(t, stdout, "A00099")
	assert.Contains(t, stdout, "A-S001")
	// Follow-on references must render so an onboarding flow has the next-step
	// numbers without a second call (the #433 coverage gap).
	assert.Contains(t, stdout, "O-00001", "order number")
	assert.Contains(t, stdout, "INV00001", "invoice number")
	assert.Contains(t, stdout, "P-00001", "payment number")
	assert.Contains(t, stdout, "CM00001", "credit memo number")
	assert.Contains(t, stdout, "1080.00", "paid amount rendered as money (two decimals)")
	assert.Contains(t, stdout, "Active", "status")
	assert.Contains(t, stderr, "Sign-up complete. Account A00099 created.")
}

func TestSignup_RequiresBody(t *testing.T) {
	_, _, err := cmdtest.Run(t, "", newCmd, nil, "signup")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), `required flag(s) "body" not set`)
}
