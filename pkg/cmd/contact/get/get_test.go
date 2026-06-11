package get

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdGet(f) }

func TestContactGet_Success(t *testing.T) {
	handler := cmdtest.OK(t, "GET", "/v1/contacts/c-123", map[string]interface{}{
		"id": "c-123", "firstName": "John", "lastName": "Doe",
		"workEmail": "j@example.com", "country": "US",
		// Zuora returns the postal code under "zipCode" (not "postalCode").
		// The distinctive value guards the postalCode->zipCode fix: reverting
		// the key would leave the "Postal Code" row blank and fail the assertion.
		"zipCode": "1000000",
	})

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "get", "c-123")
	require.NoError(t, err)
	assert.Contains(t, stdout, "John")
	assert.Contains(t, stdout, "Doe")
	assert.Contains(t, stdout, "j@example.com")
	assert.Contains(t, stdout, "US")
	assert.Contains(t, stdout, "Postal Code")
	assert.Contains(t, stdout, "1000000")
}

func TestContactGet_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "get")
	assert.Error(t, err)
}

func TestContactGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Contact not found")

	_, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "get", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Contact not found")
}
