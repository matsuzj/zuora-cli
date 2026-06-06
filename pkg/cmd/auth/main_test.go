package auth

import (
	"os"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestMain swaps the OS keyring for an in-memory mock for the whole test
// binary, so no test touches the real keychain. Without this, the
// `auth login` path (login.go) shells out to /usr/bin/security on macOS
// runners; the mock keeps the suite deterministic and side-effect-free.
func TestMain(m *testing.M) {
	keyring.MockInit()
	os.Exit(m.Run())
}
