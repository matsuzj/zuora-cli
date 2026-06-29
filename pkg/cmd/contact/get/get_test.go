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
		"address1": "1 Main St", "address2": "Suite 5",
		// Zuora returns the postal code under "zipCode" (not "postalCode").
		// The distinctive value guards the postalCode->zipCode fix: reverting
		// the key would leave the "Postal Code" row blank and fail the assertion.
		"zipCode": "1000000",
	})

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "get", "c-123")
	require.NoError(t, err)
	// Label-bound: each value must render under its OWN label, not merely appear
	// somewhere in the output (a value under the wrong label would pass a bare
	// substring check) — F-08.
	assert.Regexp(t, `(?m)^First Name:\s+John$`, stdout)
	assert.Regexp(t, `(?m)^Last Name:\s+Doe$`, stdout)
	assert.Regexp(t, `(?m)^Email:\s+j@example\.com$`, stdout)
	assert.Regexp(t, `(?m)^Country:\s+US$`, stdout)
	assert.Regexp(t, `(?m)^Address 1:\s+1 Main St$`, stdout)
	// Bites if the Address 2 row (address2) is dropped. (#427)
	assert.Regexp(t, `(?m)^Address 2:\s+Suite 5$`, stdout)
	assert.Regexp(t, `(?m)^Postal Code:\s+1000000$`, stdout)
}

func TestContactGet_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "get")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestContactGet_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Contact not found")

	_, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "get", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Contact not found")
}
