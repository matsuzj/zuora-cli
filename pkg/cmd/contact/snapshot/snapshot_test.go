package snapshot

import (
	"testing"

	"github.com/matsuzj/zuora-cli/pkg/cmd/factory"
	"github.com/matsuzj/zuora-cli/pkg/cmdtest"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCmd(f *factory.Factory) *cobra.Command { return NewCmdSnapshot(f) }

func TestContactSnapshot_Success(t *testing.T) {
	// The snapshot schema uses "postalCode" — unlike contact get's "zipCode"
	// (both doc-verified, #486). The distinctive value guards the key: reverting
	// prod to zipCode leaves the Postal Code row blank and fails the assertion.
	handler := cmdtest.OK(t, "GET", "/v1/contact-snapshots/snap-123", map[string]interface{}{
		"id": "snap-123", "firstName": "John", "lastName": "Doe",
		"workEmail": "j@example.com", "country": "US", "contactId": "c-456",
		"postalCode": "1000000",
	})

	stdout, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "snapshot", "snap-123")
	require.NoError(t, err)
	// Label-bound (F-08): each value must render under its OWN label, not merely
	// appear somewhere — a value under the wrong label would pass a bare
	// assert.Contains. Mirrors peer contact/get_test.
	assert.Regexp(t, `(?m)^ID:\s+snap-123$`, stdout)
	assert.Regexp(t, `(?m)^First Name:\s+John$`, stdout)
	assert.Regexp(t, `(?m)^Last Name:\s+Doe$`, stdout)
	assert.Regexp(t, `(?m)^Email:\s+j@example\.com$`, stdout)
	assert.Regexp(t, `(?m)^Country:\s+US$`, stdout)
	assert.Regexp(t, `(?m)^Postal Code:\s+1000000$`, stdout)
	assert.Regexp(t, `(?m)^Contact ID:\s+c-456$`, stdout)
}

func TestContactSnapshot_RequiresArgs(t *testing.T) {
	_, _, err := cmdtest.Run(t, "contact", newCmd, nil, "contact", "snapshot")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg(s), received 0")
}

func TestContactSnapshot_SuccessFalse(t *testing.T) {
	handler := cmdtest.Reasons(t, 50000040, "Contact snapshot not found")

	_, _, err := cmdtest.Run(t, "contact", newCmd, handler, "contact", "snapshot", "bad-id")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Contact snapshot not found")
}
