package cmdtest

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/require"
)

//go:embed fixtures/*.json
var fixturesFS embed.FS

// LoadFixture returns the bytes of fixtures/<name>.json. Fixtures hold REAL
// Zuora response shapes (see AGENTS.md, "Build fixtures from REAL response
// shapes") so a command test renders against the real envelope — nesting and
// all — instead of a hand-written guess that can mask a wrong-key bug.
//
// Pass the bytes to a handler as json.RawMessage so they are sent verbatim:
//
//	OK(t, "GET", "/v1/orders/O-1", json.RawMessage(LoadFixture(t, "order_get")))
func LoadFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := fixturesFS.ReadFile("fixtures/" + name + ".json")
	require.NoError(t, err, "loading fixture %q", name)
	return b
}
